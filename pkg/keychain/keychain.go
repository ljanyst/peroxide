// Copyright (c) 2021 Proton Technologies AG
// Copyright (c) 2022 Lukasz Janyst <lukasz@jany.st>
//
// This file is part of Peroxide.
//
// Peroxide is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Peroxide is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Peroxide.  If not, see <https://www.gnu.org/licenses/>.

package keychain

import (
	"fmt"
	"sync"

	"github.com/ljanyst/peroxide/pkg/config/settings"
)

// Version is the keychain data version.
const Version = "k11"

// NewKeychain creates a new native keychain.
func NewKeychain(s *settings.Settings, keychainName string) (*Keychain, error) {
	helper, err := newStaticKeychain()
	if err != nil {
		return nil, err
	}

	return newKeychain(helper, fmt.Sprintf("protonmail/%v/users", keychainName)), nil
}

func newKeychain(helper Helper, url string) *Keychain {
	return &Keychain{
		helper: helper,
		url:    url,
		locker: &sync.Mutex{},
	}
}

type Keychain struct {
	helper Helper
	url    string
	locker sync.Locker
}

func (kc *Keychain) List() ([]string, error) {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	userIDsByURL, err := kc.helper.List()
	if err != nil {
		return nil, err
	}

	var userIDs []string // nolint[prealloc]

	for url, userID := range userIDsByURL {
		if url != kc.secretURL(userID) {
			continue
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (kc *Keychain) Delete(userID string) error {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	userIDsByURL, err := kc.helper.List()
	if err != nil {
		return err
	}

	if _, ok := userIDsByURL[kc.secretURL(userID)]; !ok {
		return nil
	}

	return kc.helper.Delete(kc.secretURL(userID))
}

// Get returns the username and secret for the given userID.
func (kc *Keychain) Get(userID string) (string, string, error) {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	return kc.helper.Get(kc.secretURL(userID))
}

func (kc *Keychain) Put(userID, secret string) error {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	return kc.helper.Add(&Credentials{
		ServerURL: kc.secretURL(userID),
		Username:  userID,
		Secret:    secret,
	})
}

// secretURL returns the URL referring to a userID's secrets.
func (kc *Keychain) secretURL(userID string) string {
	return fmt.Sprintf("%v/%v", kc.url, userID)
}
