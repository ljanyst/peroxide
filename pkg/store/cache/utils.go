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

package cache

import (
	"path/filepath"

	"github.com/ljanyst/peroxide/pkg/config/settings"
)

// Memory cache was estimated by empirical usage in past and it was set to 100MB.
// NOTE: This value must not be less than maximal size of one email (~30MB).
const inMemoryCacheLimnit = 100 * (1 << 20)

// LoadMessageCache loads local cache in case it is enabled in settings and available.
// In any other case it is returning in-memory cache. Could also return an error in case
// local cache is enabled but unavailable (in-memory cache will be returned nevertheless).
func LoadMessageCache(s *settings.Settings) (Cache, error) {
	if !s.GetBool(settings.CacheEnabledKey) {
		return NewInMemoryCache(inMemoryCacheLimnit), nil
	}

	var compressor Compressor

	// NOTE(GODT-1158): Changing compression is not an option currently
	// available for user but, if user changes compression setting we have
	// to nuke the cache.
	if s.GetBool(settings.CacheCompressionKey) {
		compressor = &GZipCompressor{}
	} else {
		compressor = &NoopCompressor{}
	}

	path := filepath.Join(s.Get(settings.CacheDir), "messages")

	// To prevent memory peaks we set maximal write concurency for store
	// build jobs.
	//	store.SetBuildAndCacheJobLimit(s.GetInt(settings.CacheConcurrencyWrite))

	messageCache, err := NewOnDiskCache(path, compressor, Options{
		MinFreeAbs:      uint64(s.GetInt(settings.CacheMinFreeAbsKey)),
		MinFreeRat:      s.GetFloat64(settings.CacheMinFreeRatKey),
		ConcurrentRead:  s.GetInt(settings.CacheConcurrencyRead),
		ConcurrentWrite: s.GetInt(settings.CacheConcurrencyWrite),
	})

	if err != nil {
		return NewInMemoryCache(inMemoryCacheLimnit), err
	}

	return messageCache, nil
}
