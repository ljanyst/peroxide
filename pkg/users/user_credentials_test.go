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

package users

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/ljanyst/peroxide/pkg/events"
	"github.com/ljanyst/peroxide/pkg/pmapi"
	"github.com/pkg/errors"
	r "github.com/stretchr/testify/require"
)

func TestUpdateUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().UpdateUser(gomock.Any()).Return(testPMAPIUser, nil),
		m.pmapiClient.EXPECT().ReloadKeys(gomock.Any(), testCredentials.Secret.MailboxPassword).Return(nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		m.credentialsStore.EXPECT().UpdateEmails("user", []string{testPMAPIAddress.Email}).Return(testCredentials, nil),
	)

	r.NoError(t, user.UpdateUser(context.Background()))
}

func TestLogoutUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().AuthDelete(gomock.Any()).Return(nil),
		m.credentialsStore.EXPECT().Logout("user").Return(testCredentialsDisconnected, nil),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me"),
	)

	err := user.Logout()
	r.NoError(t, err)
}

func TestLogoutUserFailsLogout(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().AuthDelete(gomock.Any()).Return(nil),
		m.credentialsStore.EXPECT().Logout("user").Return(nil, errors.New("logout failed")),
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me"),
	)

	err := user.Logout()
	r.NoError(t, err)
}

func TestCheckBridgeLogin(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	err := user.UnlockCredentials("main", testMainKeyString)
	r.NoError(t, err)
}

func TestCheckBridgeLoginLoggedOut(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	gomock.InOrder(
		// Mock init of user.
		m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil),
		m.pmapiClient.EXPECT().AddAuthRefreshHandler(gomock.Any()),
		m.pmapiClient.EXPECT().ListLabels(gomock.Any()).Return(nil, pmapi.ErrUnauthorized),
		m.pmapiClient.EXPECT().Addresses().Return(nil),
	)

	user, err := newUser("user", m.eventListener, m.credentialsStore, m.storeMaker, m.clientManager)
	r.NoError(t, err)

	err = user.connect(m.pmapiClient)
	r.Error(t, err)
	defer cleanUpUserData(user)

	err = user.CheckCredentials("main", "asdf")
	r.Equal(t, ErrLoggedOutUser, err)
}

func TestCheckBridgeLoginBadPassword(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	err := user.UnlockCredentials("main", "wrong!")
	r.EqualError(t, err, "Bridge credentials checking failed")
}
