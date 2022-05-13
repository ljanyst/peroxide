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

//go:build !imaptest
// +build !imaptest

package credentials

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

func GenerateKey(size uint) []byte {
	key := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	return key
}

func generatePassword() string {
	return base64.RawURLEncoding.EncodeToString(GenerateKey(16))
}

func Decrypt(msg []byte, key [32]byte) ([]byte, error) {
	if len(msg) < 24 {
		return nil, ErrDecryptionFailed
	}

	// Read the nonce
	var nonce [24]byte
	copy(nonce[:], msg[:24])

	// Decrypt
	decrypted, ok := secretbox.Open(nil, msg[24:], &nonce, &key)
	if !ok {
		return nil, ErrDecryptionFailed
	}

	return decrypted, nil
}

func Encrypt(msg []byte, key [32]byte) ([]byte, error) {
	// Nonce must be different for each message
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	// Encrypt the secret and append the result to the nonce
	return secretbox.Seal(nonce[:], msg, &nonce, &key), nil
}
