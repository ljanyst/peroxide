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

package smtp

import (
	"strings"
	"time"

	goSMTPBackend "github.com/emersion/go-smtp"
	"github.com/ljanyst/peroxide/pkg/listener"
	"github.com/ljanyst/peroxide/pkg/users"
	"github.com/pkg/errors"
)

type settingsProvider interface {
	GetBool(string) bool
}

type smtpBackend struct {
	eventListener listener.Listener
	settings      settingsProvider
	users         *users.Users
	sendRecorder  *sendRecorder
}

// NewSMTPBackend returns struct implementing go-smtp/backend interface.
func NewSMTPBackend(
	eventListener listener.Listener,
	settings settingsProvider,
	users *users.Users,
) *smtpBackend { //nolint[golint]
	return newSMTPBackend(eventListener, settings, users)
}

func newSMTPBackend(
	eventListener listener.Listener,
	settings settingsProvider,
	users *users.Users,
) *smtpBackend {
	return &smtpBackend{
		eventListener: eventListener,
		settings:      settings,
		users:         users,
		sendRecorder:  newSendRecorder(),
	}
}

// Login authenticates a user.
func (sb *smtpBackend) Login(_ *goSMTPBackend.ConnectionState, username, password string) (goSMTPBackend.Session, error) {
	username = strings.ToLower(username)

	user, err := sb.users.GetUser(username)
	if err != nil {
		log.Warn("Cannot get user: ", err)
		return nil, err
	}
	if err := user.CheckBridgeLogin(password); err != nil {
		log.WithError(err).Error("Could not check bridge password")
		// Apple Mail sometimes generates a lot of requests very quickly. It's good practice
		// to have a timeout after bad logins so that we can slow those requests down a little bit.
		time.Sleep(10 * time.Second)
		return nil, err
	}
	// Client can log in only using address so we can properly close all SMTP connections.
	addressID, err := user.GetAddressID(username)
	if err != nil {
		log.Error("Cannot get addressID: ", err)
		return nil, err
	}
	// AddressID is only for split mode--it has to be empty for combined mode.
	if user.IsCombinedAddressMode() {
		addressID = ""
	}
	return newSMTPUser(sb.eventListener, sb, user, username, addressID)
}

func (sb *smtpBackend) AnonymousLogin(_ *goSMTPBackend.ConnectionState) (goSMTPBackend.Session, error) {
	return nil, errors.New("anonymous login not supported")
}
