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

// Package bridge implements the bridge CLI application.
package bridge

import (
	"crypto/tls"

	cfgCache "github.com/ljanyst/peroxide/pkg/config/cache"
	"github.com/ljanyst/peroxide/pkg/config/settings"
	pkgTLS "github.com/ljanyst/peroxide/pkg/config/tls"
	"github.com/ljanyst/peroxide/pkg/store"
	"github.com/ljanyst/peroxide/pkg/store/cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// Memory cache was estimated by empirical usage in past and it was set to 100MB.
	// NOTE: This value must not be less than maximal size of one email (~30MB).
	inMemoryCacheLimnit = 100 * (1 << 20)
)

func loadTLSConfig(cfg *pkgTLS.TLS) (*tls.Config, error) {
	if !cfg.HasCerts() {
		if err := generateTLSCerts(cfg); err != nil {
			return nil, err
		}
	}

	tlsConfig, err := cfg.GetConfig()
	if err == nil {
		return tlsConfig, nil
	}

	logrus.WithError(err).Error("Failed to load TLS config, regenerating certificates")

	if err := generateTLSCerts(cfg); err != nil {
		return nil, err
	}

	return cfg.GetConfig()
}

func generateTLSCerts(cfg *pkgTLS.TLS) error {
	template, err := pkgTLS.NewTLSTemplate()
	if err != nil {
		return errors.Wrap(err, "failed to generate TLS template")
	}

	if err := cfg.GenerateCerts(template); err != nil {
		return errors.Wrap(err, "failed to generate TLS certs")
	}

	if err := cfg.InstallCerts(); err != nil {
		return errors.Wrap(err, "failed to install TLS certs")
	}

	return nil
}

// loadMessageCache loads local cache in case it is enabled in settings and available.
// In any other case it is returning in-memory cache. Could also return an error in case
// local cache is enabled but unavailable (in-memory cache will be returned nevertheless).
func LoadMessageCache(s *settings.Settings, cfg *cfgCache.Cache) (cache.Cache, error) {
	if !s.GetBool(settings.CacheEnabledKey) {
		return cache.NewInMemoryCache(inMemoryCacheLimnit), nil
	}

	var compressor cache.Compressor

	// NOTE(GODT-1158): Changing compression is not an option currently
	// available for user but, if user changes compression setting we have
	// to nuke the cache.
	if s.GetBool(settings.CacheCompressionKey) {
		compressor = &cache.GZipCompressor{}
	} else {
		compressor = &cache.NoopCompressor{}
	}

	var path string

	if customPath := s.Get(settings.CacheLocationKey); customPath != "" {
		path = customPath
	} else {
		path = cfg.GetDefaultMessageCacheDir()
		// Store path so it will allways persist if default location
		// will be changed in new version.
		s.Set(settings.CacheLocationKey, path)
	}

	// To prevent memory peaks we set maximal write concurency for store
	// build jobs.
	store.SetBuildAndCacheJobLimit(s.GetInt(settings.CacheConcurrencyWrite))

	messageCache, err := cache.NewOnDiskCache(path, compressor, cache.Options{
		MinFreeAbs:      uint64(s.GetInt(settings.CacheMinFreeAbsKey)),
		MinFreeRat:      s.GetFloat64(settings.CacheMinFreeRatKey),
		ConcurrentRead:  s.GetInt(settings.CacheConcurrencyRead),
		ConcurrentWrite: s.GetInt(settings.CacheConcurrencyWrite),
	})

	if err != nil {
		return cache.NewInMemoryCache(inMemoryCacheLimnit), err
	}

	return messageCache, nil
}
