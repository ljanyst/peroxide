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
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"golang.org/x/crypto/nacl/secretbox"
)

type Secret struct {
	APIToken        string
	MailboxPassword []byte
}

type Credentials struct {
	UserID       string
	Name         string
	Emails       []string
	Secret       Secret `json:"-"`
	SealedSecret []byte
	SealedKeys   map[string][]byte
	Key          [32]byte `json:"-"`
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

func (s *Credentials) Unlock(slot, password string) error {
	sealedKey, ok := s.SealedKeys[slot]
	if !ok {
		return ErrUnauthorized
	}

	pb, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return ErrUnauthorized
	}

	if len(pb) != len(s.Key) {
		return ErrUnauthorized
	}

	var passBytes [32]byte
	copy(passBytes[:], pb)

	if len(sealedKey) < 24 {
		return ErrUnauthorized
	}

	// Read the nonce
	var nonce [24]byte
	copy(nonce[:], sealedKey[:24])

	// Decrypt
	keyBytes, ok := secretbox.Open(nil, sealedKey[24:], &nonce, &passBytes)
	if !ok {
		return ErrUnauthorized
	}

	if s.Locked() {
		copy(s.Key[:], keyBytes)
		return s.Decrypt()
	}

	return nil
}

func (s *Credentials) SealKey(slot string, key [32]byte) error {
	if s.Locked() {
		return ErrLocked
	}

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	s.SealedKeys[slot] = secretbox.Seal(nonce[:], s.Key[:], &nonce, &key)
	return nil
}

func (s *Credentials) Encrypt() error {
	if s.Locked() {
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
	s.SealedSecret = secretbox.Seal(nonce[:], []byte(secret), &nonce, &s.Key)

	return nil
}

func (s *Credentials) Decrypt() error {
	if len(s.SealedSecret) < 24 || s.Locked() {
		return ErrDecryptionFailed
	}

	// Read the nonce
	var nonce [24]byte
	copy(nonce[:], s.SealedSecret[:24])

	// Decrypt
	decrypted, ok := secretbox.Open(nil, s.SealedSecret[24:], &nonce, &s.Key)
	if !ok {
		return ErrDecryptionFailed
	}

	if err := json.Unmarshal(decrypted, &s.Secret); err != nil {
		return ErrDecryptionFailed
	}

	return nil
}

func (s *Credentials) Locked() bool {
	for _, v := range s.Key {
		if v != 0 {
			return false
		}
	}
	return true
}
