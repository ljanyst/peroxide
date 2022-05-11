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
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
)

type Credentials struct {
	UserID          string
	Name            string
	Emails          []string
	APIToken        string
	MailboxPassword []byte
	BridgePassword  string
}

func (s *Credentials) Logout() {
	s.APIToken = ""

	for i := range s.MailboxPassword {
		s.MailboxPassword[i] = 0
	}

	s.MailboxPassword = []byte{}
}

func (s *Credentials) IsConnected() bool {
	return s.APIToken != "" && len(s.MailboxPassword) != 0
}

func (s *Credentials) SplitAPIToken() (string, string, error) {
	split := strings.Split(s.APIToken, ":")

	if len(split) != 2 {
		return "", "", errors.New("malformed API token")
	}

	return split[0], split[1], nil
}

func (s *Credentials) CheckPassword(password string) error {
	if subtle.ConstantTimeCompare([]byte(s.BridgePassword), []byte(password)) != 1 {
		return fmt.Errorf("backend/credentials: incorrect password")
	}
	return nil

}
