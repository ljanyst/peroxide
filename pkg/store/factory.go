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

package store

import (
	"fmt"
	"path/filepath"

	"github.com/ljanyst/peroxide/pkg/config/settings"
	"github.com/ljanyst/peroxide/pkg/listener"
	"github.com/ljanyst/peroxide/pkg/message"
	"github.com/ljanyst/peroxide/pkg/store/cache"
)

type StoreFactory struct {
	settings *settings.Settings
	listener listener.Listener
	events   *Events
	cache    cache.Cache
	builder  *message.Builder
}

func NewStoreFactory(
	settingsObj *settings.Settings,
	listener listener.Listener,
	cache cache.Cache,
	builder *message.Builder,
) *StoreFactory {
	eventsCachePath := filepath.Join(settingsObj.Get(settings.CacheDir), "store_events.json")
	return &StoreFactory{
		settings: settingsObj,
		listener: listener,
		events:   NewEvents(eventsCachePath),
		cache:    cache,
		builder:  builder,
	}
}

// New creates new store for given user.
func (f *StoreFactory) New(user BridgeUser) (*Store, error) {
	return New(
		user,
		f.listener,
		f.cache,
		f.builder,
		getUserStorePath(f.settings.Get(settings.CacheDir), user.ID()),
		f.events,
	)
}

// Remove removes all store files for given user.
func (f *StoreFactory) Remove(userID string) error {
	return RemoveStore(
		f.events,
		getUserStorePath(f.settings.Get(settings.CacheDir), userID),
		userID,
	)
}

// getUserStorePath returns the file path of the store database for the given userID.
func getUserStorePath(storeDir string, userID string) (path string) {
	return filepath.Join(storeDir, fmt.Sprintf("mailbox-%v.db", userID))
}
