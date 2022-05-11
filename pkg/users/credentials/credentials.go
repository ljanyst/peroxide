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

// Package credentials implements our struct stored in keychain.
// Store struct is kind of like a database client.
// Credentials struct is kind of like one record from the database.
package credentials

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/nacl/secretbox"
)

type Secret struct {
	APIToken        string
	MailboxPassword []byte
	BridgePassword  string
}

type Credentials struct {
	UserID       string
	Name         string
	Emails       []string
	Secret       Secret `json:"-"`
	SealedSecret []byte
	key          [32]byte `json:"-"`
}

func (s *Credentials) logout() {
	s.Secret.APIToken = ""

	for i := range s.Secret.MailboxPassword {
		s.Secret.MailboxPassword[i] = 0
	}

	s.Secret.MailboxPassword = []byte{}
}

func (s *Credentials) IsConnected() bool {
	return s.Secret.APIToken != "" && len(s.Secret.MailboxPassword) != 0
}

func (s *Credentials) SplitAPIToken() (string, string, error) {
	split := strings.Split(s.Secret.APIToken, ":")

	if len(split) != 2 {
		return "", "", errors.New("malformed API token")
	}

	return split[0], split[1], nil
}

func (s *Credentials) CheckPassword(password string) error {
	if subtle.ConstantTimeCompare([]byte(s.Secret.BridgePassword), []byte(password)) != 1 {
		return fmt.Errorf("backend/credentials: incorrect password")
	}
	return nil
}

func (s *Credentials) encrypt() error {
	if s.locked() {
		return ErrEncryptionFailed
	}

	secret, err := json.Marshal(s.Secret)
	if err != nil {
		return ErrEncryptionFailed
	}

	// Nonce must be different for each message
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	// Encrypt the secret and append the result to the nonce
	s.SealedSecret = secretbox.Seal(nonce[:], []byte(secret), &nonce, &s.key)

	return nil
}

func (s *Credentials) decrypt() error {
	if len(s.SealedSecret) < 24 || s.locked() {
		return ErrDecryptionFailed
	}

	// Read the nonce
	var nonce [24]byte
	copy(nonce[:], s.SealedSecret[:24])

	// Decrypt
	decrypted, ok := secretbox.Open(nil, s.SealedSecret[24:], &nonce, &s.key)
	if !ok {
		return ErrDecryptionFailed
	}

	if err := json.Unmarshal(decrypted, &s.Secret); err != nil {
		return ErrDecryptionFailed
	}

	return nil
}

func (s *Credentials) locked() bool {
	for _, v := range s.key {
		if v != 0 {
			return false
		}
	}
	return true
}
