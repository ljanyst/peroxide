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
	"crypto/x509"
	"time"

	cfgCache "github.com/ljanyst/peroxide/pkg/config/cache"
	"github.com/ljanyst/peroxide/pkg/config/settings"
	"github.com/ljanyst/peroxide/pkg/store"
	"github.com/ljanyst/peroxide/pkg/store/cache"
	"github.com/pkg/errors"
)

const (
	// Memory cache was estimated by empirical usage in past and it was set to 100MB.
	// NOTE: This value must not be less than maximal size of one email (~30MB).
	inMemoryCacheLimnit = 100 * (1 << 20)
)

// GetConfig tries to load TLS config or generate new one which is then returned.
func loadTlsConfig(certPath, keyPath string) (*tls.Config, error) {
	c, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load cert and key")
	}

	c.Leaf, err = x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse the certificate")
	}

	if time.Now().Add(31 * 24 * time.Hour).After(c.Leaf.NotAfter) {
		return nil, errors.Wrap(err, "The X509 certificate is about to expire")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(c.Leaf)

	return &tls.Config{
		Certificates: []tls.Certificate{c},
		ServerName:   c.Leaf.Subject.CommonName,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
	}, nil
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

	path := cfg.GetDefaultMessageCacheDir()

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
