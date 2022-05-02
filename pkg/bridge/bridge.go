// Copyright (c) 2022 Proton Technologies AG
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

// Package bridge provides core functionality of Bridge app.
package bridge

import (
	"errors"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	cacheCfg "github.com/ljanyst/peroxide/pkg/config/cache"
	"github.com/ljanyst/peroxide/pkg/config/settings"
	"github.com/ljanyst/peroxide/pkg/config/tls"
	"github.com/ljanyst/peroxide/pkg/cookies"
	"github.com/ljanyst/peroxide/pkg/events"
	"github.com/ljanyst/peroxide/pkg/imap"
	"github.com/ljanyst/peroxide/pkg/keychain"
	"github.com/ljanyst/peroxide/pkg/listener"
	"github.com/ljanyst/peroxide/pkg/logging"
	"github.com/ljanyst/peroxide/pkg/message"
	"github.com/ljanyst/peroxide/pkg/pmapi"
	"github.com/ljanyst/peroxide/pkg/smtp"
	"github.com/ljanyst/peroxide/pkg/store/cache"
	"github.com/ljanyst/peroxide/pkg/users"
	"github.com/ljanyst/peroxide/pkg/users/credentials"

	logrus "github.com/sirupsen/logrus"
)

var log = logrus.WithField("pkg", "bridge") //nolint[gochecknoglobals]

var ErrLocalCacheUnavailable = errors.New("local cache is unavailable")

type Bridge struct {
	Users *users.Users

	settings      SettingsProvider
	clientManager pmapi.Manager
	cacheProvider *cacheCfg.Cache
	cache         cache.Cache
	listener      listener.Listener
}

func (b *Bridge) Configure(configFile string) error {
	rand.Seed(time.Now().UnixNano())

	if err := logging.Init(); err != nil {
		return err
	}

	settingsObj := settings.New(configFile)

	cacheConf, err := cacheCfg.New(settingsObj.Get(settings.CacheDir), "c11")
	if err != nil {
		return err
	}
	if err := cacheConf.RemoveOldVersions(); err != nil {
		return err
	}

	listener := listener.New()
	events.SetupEvents(listener)

	kc, err := keychain.NewKeychain("bridge")
	if err != nil {
		return err
	}

	cfg := pmapi.NewConfig()
	cfg.UpgradeApplicationHandler = func() { listener.Emit(events.UpgradeApplicationEvent, "") }
	cfg.TLSIssueHandler = func() { listener.Emit(events.TLSCertIssue, "") }

	cm := pmapi.New(cfg)

	cm.AddConnectionObserver(pmapi.NewConnectionObserver(
		func() { listener.Emit(events.InternetOffEvent, "") },
		func() { listener.Emit(events.InternetOnEvent, "") },
	))

	jar, err := cookies.NewCookieJar(settingsObj.Get(settings.CookieJar))
	if err != nil {
		return err
	}

	cm.SetCookieJar(jar)

	cache, err := LoadMessageCache(settingsObj, cacheConf)
	if err != nil {
		return err
	}

	builder := message.NewBuilder(
		settingsObj.GetInt(settings.FetchWorkers),
		settingsObj.GetInt(settings.AttachmentWorkers),
	)

	if settingsObj.GetBool(settings.AllowProxyKey) {
		cm.AllowProxy()
	}

	u := users.New(
		listener,
		cm,
		credentials.NewStore(kc),
		NewStoreFactory(cacheConf, listener, cache, builder),
	)

	b.Users = u
	b.settings = settingsObj
	b.clientManager = cm
	b.cacheProvider = cacheConf
	b.cache = cache
	b.listener = listener
	return nil
}

func (b *Bridge) Run() error {
	tlsCfg := tls.New(b.settings.Get(settings.TLSDir))

	tlsConfig, err := loadTLSConfig(tlsCfg)
	if err != nil {
		return err
	}

	imapBackend := imap.NewIMAPBackend(b.listener, b.cacheProvider, b.settings, b.Users)
	smtpBackend := smtp.NewSMTPBackend(b.listener, b.Users)

	go func() {
		imapPort := b.settings.GetInt(settings.IMAPPortKey)
		imap.NewIMAPServer(
			false, // log client
			false, // log server
			imapPort, tlsConfig, imapBackend, b.listener).ListenAndServe()
	}()

	go func() {
		smtpPort := b.settings.GetInt(settings.SMTPPortKey)
		useSSL := b.settings.GetBool(settings.SMTPSSLKey)
		smtp.NewSMTPServer(
			false,
			smtpPort, useSSL, tlsConfig, smtpBackend, b.listener).ListenAndServe()
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	return nil
}

// FactoryReset will remove all local cache and settings.
// It will also downgrade to latest stable version if user is on early version.
func (b *Bridge) FactoryReset() {
	if err := b.Users.ClearData(); err != nil {
		log.WithError(err).Error("Failed to remove bridge data")
	}

	if err := b.Users.ClearUsers(); err != nil {
		log.WithError(err).Error("Failed to remove bridge users")
	}
}
