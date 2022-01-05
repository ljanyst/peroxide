// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

// Package bridge implements the bridge CLI application.
package bridge

import (
	"crypto/tls"
	"time"

	"github.com/ljanyst/peroxide/pkg/api"
	"github.com/ljanyst/peroxide/pkg/app/base"
	pkgBridge "github.com/ljanyst/peroxide/pkg/bridge"
	"github.com/ljanyst/peroxide/pkg/config/settings"
	pkgTLS "github.com/ljanyst/peroxide/pkg/config/tls"
	"github.com/ljanyst/peroxide/pkg/constants"
	"github.com/ljanyst/peroxide/pkg/frontend"
	"github.com/ljanyst/peroxide/pkg/frontend/types"
	"github.com/ljanyst/peroxide/pkg/imap"
	"github.com/ljanyst/peroxide/pkg/message"
	"github.com/ljanyst/peroxide/pkg/smtp"
	"github.com/ljanyst/peroxide/pkg/store"
	"github.com/ljanyst/peroxide/pkg/store/cache"
	"github.com/ljanyst/peroxide/pkg/updater"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// Memory cache was estimated by empirical usage in past and it was set to 100MB.
	// NOTE: This value must not be less than maximal size of one email (~30MB).
	inMemoryCacheLimnit = 100 * (1 << 20)
)

func MailLoop(b *base.Base) error { // nolint[funlen]
	tlsConfig, err := loadTLSConfig(b)
	if err != nil {
		return err
	}

	cache, cacheErr := loadMessageCache(b)
	if cacheErr != nil {
		logrus.WithError(cacheErr).Error("Could not load local cache.")
	}

	builder := message.NewBuilder(
		b.Settings.GetInt(settings.FetchWorkers),
		b.Settings.GetInt(settings.AttachmentWorkers),
	)

	bridge := pkgBridge.New(
		b.Locations,
		b.Cache,
		b.Settings,
		b.Listener,
		cache,
		builder,
		b.CM,
		b.Creds,
		b.Updater,
		b.Versioner,
		b.Autostart,
	)
	imapBackend := imap.NewIMAPBackend(b.Listener, b.Cache, b.Settings, bridge)
	smtpBackend := smtp.NewSMTPBackend(b.Listener, b.Settings, bridge)

	if cacheErr != nil {
		bridge.AddError(pkgBridge.ErrLocalCacheUnavailable)
	}

	go func() {
		api.NewAPIServer(b.Settings, b.Listener).ListenAndServe()
	}()

	go func() {
		imapPort := b.Settings.GetInt(settings.IMAPPortKey)
		imap.NewIMAPServer(
			false, // log client
			false, // log server
			imapPort, tlsConfig, imapBackend, b.UserAgent, b.Listener).ListenAndServe()
	}()

	go func() {
		smtpPort := b.Settings.GetInt(settings.SMTPPortKey)
		useSSL := b.Settings.GetBool(settings.SMTPSSLKey)
		smtp.NewSMTPServer(
			false,
			smtpPort, useSSL, tlsConfig, smtpBackend, b.Listener).ListenAndServe()
	}()

	f := frontend.New(
		b.Locations,
		b.Settings,
		b.Listener,
		b.Updater,
		bridge,
		b,
	)

	// Watch for updates routine
	go func() {
		ticker := time.NewTicker(constants.UpdateCheckInterval)

		for {
			checkAndHandleUpdate(b.Updater, f, b.Settings.GetBool(settings.AutoUpdateKey))
			<-ticker.C
		}
	}()

	return f.Loop()
}

func loadTLSConfig(b *base.Base) (*tls.Config, error) {
	if !b.TLS.HasCerts() {
		if err := generateTLSCerts(b); err != nil {
			return nil, err
		}
	}

	tlsConfig, err := b.TLS.GetConfig()
	if err == nil {
		return tlsConfig, nil
	}

	logrus.WithError(err).Error("Failed to load TLS config, regenerating certificates")

	if err := generateTLSCerts(b); err != nil {
		return nil, err
	}

	return b.TLS.GetConfig()
}

func generateTLSCerts(b *base.Base) error {
	template, err := pkgTLS.NewTLSTemplate()
	if err != nil {
		return errors.Wrap(err, "failed to generate TLS template")
	}

	if err := b.TLS.GenerateCerts(template); err != nil {
		return errors.Wrap(err, "failed to generate TLS certs")
	}

	if err := b.TLS.InstallCerts(); err != nil {
		return errors.Wrap(err, "failed to install TLS certs")
	}

	return nil
}

func checkAndHandleUpdate(u types.Updater, f frontend.Frontend, autoUpdate bool) {
	log := logrus.WithField("pkg", "app/bridge")
	version, err := u.Check()
	if err != nil {
		log.WithError(err).Error("An error occurred while checking for updates")
		return
	}

	f.WaitUntilFrontendIsReady()

	// Update links in UI
	f.SetVersion(version)

	if !u.IsUpdateApplicable(version) {
		log.Info("No need to update")
		return
	}

	log.WithField("version", version.Version).Info("An update is available")

	if !autoUpdate {
		f.NotifyManualUpdate(version, u.CanInstall(version))
		return
	}

	if !u.CanInstall(version) {
		log.Info("A manual update is required")
		f.NotifySilentUpdateError(updater.ErrManualUpdateRequired)
		return
	}

	if err := u.InstallUpdate(version); err != nil {
		if errors.Cause(err) == updater.ErrDownloadVerify {
			log.WithError(err).Warning("Skipping update installation due to temporary error")
		} else {
			log.WithError(err).Error("The update couldn't be installed")
			f.NotifySilentUpdateError(err)
		}

		return
	}

	f.NotifySilentUpdateInstalled()
}

// loadMessageCache loads local cache in case it is enabled in settings and available.
// In any other case it is returning in-memory cache. Could also return an error in case
// local cache is enabled but unavailable (in-memory cache will be returned nevertheless).
func loadMessageCache(b *base.Base) (cache.Cache, error) {
	if !b.Settings.GetBool(settings.CacheEnabledKey) {
		return cache.NewInMemoryCache(inMemoryCacheLimnit), nil
	}

	var compressor cache.Compressor

	// NOTE(GODT-1158): Changing compression is not an option currently
	// available for user but, if user changes compression setting we have
	// to nuke the cache.
	if b.Settings.GetBool(settings.CacheCompressionKey) {
		compressor = &cache.GZipCompressor{}
	} else {
		compressor = &cache.NoopCompressor{}
	}

	var path string

	if customPath := b.Settings.Get(settings.CacheLocationKey); customPath != "" {
		path = customPath
	} else {
		path = b.Cache.GetDefaultMessageCacheDir()
		// Store path so it will allways persist if default location
		// will be changed in new version.
		b.Settings.Set(settings.CacheLocationKey, path)
	}

	// To prevent memory peaks we set maximal write concurency for store
	// build jobs.
	store.SetBuildAndCacheJobLimit(b.Settings.GetInt(settings.CacheConcurrencyWrite))

	messageCache, err := cache.NewOnDiskCache(path, compressor, cache.Options{
		MinFreeAbs:      uint64(b.Settings.GetInt(settings.CacheMinFreeAbsKey)),
		MinFreeRat:      b.Settings.GetFloat64(settings.CacheMinFreeRatKey),
		ConcurrentRead:  b.Settings.GetInt(settings.CacheConcurrencyRead),
		ConcurrentWrite: b.Settings.GetInt(settings.CacheConcurrencyWrite),
	})

	if err != nil {
		return cache.NewInMemoryCache(inMemoryCacheLimnit), err
	}

	return messageCache, nil
}
