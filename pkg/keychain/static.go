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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/nacl/secretbox"

	"github.com/ljanyst/peroxide/pkg/files"
)

const path = "~/.peroxide-creds"

var secretKey [32]byte

// Stores credentials in a file in user's home directory
type Static struct {
	credentials map[string]Credentials
}

func (s *Static) Add(c *Credentials) error {
	s.credentials[c.ServerURL] = *c
	return dumpStaticStore(path, s.credentials)
}

func (s *Static) Delete(serverURL string) error {
	delete(s.credentials, serverURL)
	return dumpStaticStore(path, s.credentials)
}

func (s *Static) Get(serverURL string) (string, string, error) {
	if cred, ok := s.credentials[serverURL]; ok {
		return cred.Username, cred.Secret, nil
	}
	return "", "", fmt.Errorf("no usernames for %s", serverURL)
}

func (s *Static) List() (map[string]string, error) {
	list := make(map[string]string)
	for _, v := range s.credentials {
		list[v.ServerURL] = v.Username
	}
	return list, nil
}

func decryptSecrets(credentials map[string]Credentials) (map[string]Credentials, error) {
	creds := make(map[string]Credentials)
	for k, v := range credentials {
		encrypted, err := base64.StdEncoding.DecodeString(v.Secret)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode the secret for server %s: %s", k, err)
		}

		// Read the nonce
		var nonce [24]byte
		copy(nonce[:], encrypted[:24])

		// Decrypt
		decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, &secretKey)
		if !ok {
			return nil, fmt.Errorf("Unable to decrypt the secret for server %s", k)
		}

		creds[k] = Credentials{
			ServerURL: v.ServerURL,
			Username:  v.Username,
			Secret:    string(decrypted),
		}
	}
	return creds, nil
}

func encryptSecrets(credentials map[string]Credentials) (map[string]Credentials, error) {
	creds := make(map[string]Credentials)
	for k, v := range credentials {
		// Nonce must be different for each message
		var nonce [24]byte
		if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
			return nil, err
		}

		// Encrypt the secret and append the result to the nonce
		encrypted := secretbox.Seal(nonce[:], []byte(v.Secret), &nonce, &secretKey)
		creds[k] = Credentials{
			ServerURL: v.ServerURL,
			Username:  v.Username,
			Secret:    base64.StdEncoding.EncodeToString(encrypted),
		}
	}
	return creds, nil
}

func loadStaticStore(path string) (map[string]Credentials, error) {
	fileName := files.ExpandTilde(path)
	data, err := ioutil.ReadFile(fileName)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("Unable to read the credential store file %s: %s", fileName, err)
	}

	creds := make(map[string]Credentials)

	if os.IsNotExist(err) {
		return creds, nil
	}

	err = json.Unmarshal(data, &creds)
	if err != nil {
		return nil, fmt.Errorf("Malformed credential store %s: %s", fileName, err)
	}

	return decryptSecrets(creds)
}

func dumpStaticStore(path string, credentials map[string]Credentials) error {
	creds, err := encryptSecrets(credentials)
	fileName := files.ExpandTilde(path)
	if err != nil {
		return fmt.Errorf("Unable to encrypt the credential store %s: %s", fileName, err)
	}

	data, err := json.MarshalIndent(creds, "", "    ")
	if err != nil {
		return fmt.Errorf("Unable to serialize the credential store %s: %s", fileName, err)
	}

	err = ioutil.WriteFile(fileName, data, 0600)
	if err != nil {
		return fmt.Errorf("Unable to write the credential store file %s: %s", fileName, err)
	}

	return nil
}

func newStaticKeychain() (Helper, error) {
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

	creds, err := loadStaticStore(path)
	if err != nil {
		return nil, err
	}
	return &Static{creds}, nil
}
