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

package credentials

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound         = errors.New("Credentials not found")
	ErrLocked           = errors.New("Credentials are locked")
	ErrDecryptionFailed = errors.New("Decryption of credentials failed")
	ErrEncryptionFailed = errors.New("Encryption of credentials failed")
	log                 = logrus.WithField("pkg", "credentials")
	secretKey           [32]byte
)

// Store is an encrypted credentials store.
type Store struct {
	lock     sync.RWMutex
	creds    map[string]*Credentials
	filePath string
}

// NewStore creates a new encrypted credentials store.
func NewStore(filePath string) (*Store, error) {
	s := &Store{
		creds:    make(map[string]*Credentials),
		filePath: filePath,
	}

	key := os.Getenv("PEROXIDE_CREDENTIALS_KEY")
	if key == "" {
		return nil, fmt.Errorf("PEROXIDE_CREDENTIALS_KEY envvar not set")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode the credentials key: %s", err)
	}

	if len(keyBytes) != len(secretKey) {
		return nil, fmt.Errorf("Decoded credentials key is not %d bytes long", len(secretKey))
	}

	copy(secretKey[:], keyBytes)

	err = s.loadCredentials()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) Add(userID, userName, uid, ref string, mailboxPassword []byte, emails []string) (*Credentials, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.WithFields(logrus.Fields{
		"user":     userID,
		"username": userName,
		"emails":   emails,
	}).Trace("Adding new credentials")

	creds := &Credentials{
		UserID: userID,
		Name:   userName,
		Emails: emails,
		Secret: Secret{
			APIToken:        uid + ":" + ref,
			MailboxPassword: mailboxPassword,
		},
		key: secretKey,
	}

	currentCredentials, ok := s.creds[userID]
	if ok {
		log.Info("Updating credentials of existing user")
		creds.Secret.BridgePassword = currentCredentials.Secret.BridgePassword
	} else {
		log.Info("Generating credentials for new user")
		creds.Secret.BridgePassword = generatePassword()
	}

	if err := creds.encrypt(); err != nil {
		return nil, err
	}

	s.creds[userID] = creds

	if err := s.saveCredentials(); err != nil {
		return nil, err
	}

	return creds, nil
}

func (s *Store) UpdateEmails(userID string, emails []string) (*Credentials, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	credentials, ok := s.creds[userID]
	if !ok {
		return nil, ErrNotFound
	}

	credentials.Emails = emails

	return credentials, s.saveCredentials()
}

func (s *Store) UpdatePassword(userID string, password []byte) (*Credentials, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	credentials, ok := s.creds[userID]
	if !ok {
		return nil, ErrNotFound
	}

	if credentials.locked() {
		return nil, ErrLocked
	}

	credentials.Secret.MailboxPassword = password
	if err := credentials.encrypt(); err != nil {
		return nil, err
	}

	return credentials, s.saveCredentials()
}

func (s *Store) UpdateToken(userID, uid, ref string) (*Credentials, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	credentials, ok := s.creds[userID]
	if !ok {
		return nil, ErrNotFound
	}

	if credentials.locked() {
		return nil, ErrLocked
	}

	credentials.Secret.APIToken = uid + ":" + ref
	if err := credentials.encrypt(); err != nil {
		return nil, err
	}

	return credentials, s.saveCredentials()
}

func (s *Store) Logout(userID string) (*Credentials, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	credentials, ok := s.creds[userID]
	if !ok {
		return nil, ErrNotFound
	}

	if credentials.locked() {
		return nil, ErrLocked
	}

	if err := credentials.encrypt(); err != nil {
		return nil, err
	}

	credentials.logout()

	return credentials, s.saveCredentials()
}

// List returns a list of usernames that have credentials stored.
func (s *Store) List() ([]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	log.Trace("Listing credentials in credentials store")

	userIDs := []string{}
	for id := range s.creds {
		userIDs = append(userIDs, id)
	}
	sort.Strings(userIDs)

	return userIDs, nil
}

func (s *Store) Get(userID string) (creds *Credentials, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	creds, ok := s.creds[userID]
	if !ok {
		return nil, ErrNotFound
	}
	return creds, nil
}

// Delete removes credentials from the store.
func (s *Store) Delete(userID string) (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.creds[userID]
	if !ok {
		return ErrNotFound
	}

	delete(s.creds, userID)
	return nil
}

func (s *Store) saveCredentials() error {
	f, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(s.creds)
}

func (s *Store) loadCredentials() error {
	f, err := os.Open(s.filePath)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&s.creds); err != nil {
		return err
	}

	for _, v := range s.creds {
		v.key = secretKey
		v.decrypt()
	}

	return nil
}
