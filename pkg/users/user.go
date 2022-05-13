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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ljanyst/peroxide/pkg/events"
	"github.com/ljanyst/peroxide/pkg/listener"
	"github.com/ljanyst/peroxide/pkg/pmapi"
	"github.com/ljanyst/peroxide/pkg/store"
	"github.com/ljanyst/peroxide/pkg/users/credentials"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrLoggedOutUser is sent to IMAP and SMTP if user exists, password is OK but user is logged out from the app.
var ErrLoggedOutUser = errors.New("account is logged out, use the app to login again")

// User is a struct on top of API client and credentials store.
type User struct {
	log           *logrus.Entry
	listener      listener.Listener
	client        pmapi.Client
	clientManager pmapi.Manager
	credStorer    CredentialsStorer

	storeFactory StoreMaker
	store        *store.Store

	userID string
	creds  *credentials.Credentials

	usedBytes, totalBytes int64

	lock sync.RWMutex
}

// newUser creates a new user.
// The user is initially disconnected and must be connected by calling connect().
func newUser(
	userID string,
	eventListener listener.Listener,
	credStorer CredentialsStorer,
	storeFactory StoreMaker,
	clientManager pmapi.Manager,
) (*User, error) {
	log := log.WithField("user", userID)

	log.Debug("Creating or loading user")

	creds, err := credStorer.Get(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load user credentials")
	}

	return &User{
		log:           log,
		listener:      eventListener,
		clientManager: clientManager,
		credStorer:    credStorer,
		storeFactory:  storeFactory,
		userID:        userID,
		creds:         creds,
	}, nil
}

// connect connects a user. This includes
// - providing it with an authorised API client
// - loading its credentials from the credentials store
// - loading and unlocking its PGP keys
// - loading its store.
func (u *User) connect(client pmapi.Client) error {
	u.log.Info("Connecting user")

	// Connected users have an API client.
	u.client = client

	u.client.AddAuthRefreshHandler(u.handleAuthRefresh)

	// Connected users have unlocked keys.
	if err := u.unlockIfNecessary(); err != nil {
		return err
	}

	// Connected users have a store.
	if err := u.loadStore(); err != nil { //nolint[revive] easier to read
		return err
	}

	// If the client is already unlocked, we can unlock the store cache as well.
	if client.IsUnlocked() {
		kr, err := client.GetUserKeyRing()
		if err != nil {
			return err
		}

		if err := u.store.UnlockCache(kr); err != nil {
			return err
		}

		u.store.StartWatcher()
	}

	u.UpdateSpace(nil)

	return nil
}

func (u *User) loadStore() error {
	// Logged-out user keeps store running to access offline data.
	// Therefore it is necessary to close it before re-init.
	if u.store != nil {
		if err := u.store.Close(); err != nil {
			log.WithError(err).Error("Not able to close store")
		}
		u.store = nil
	}

	store, err := u.storeFactory.New(u, u.creds.IsConnected())
	if err != nil {
		return errors.Wrap(err, "failed to create store")
	}

	u.store = store

	return nil
}

func (u *User) handleAuthRefresh(auth *pmapi.AuthRefresh) {
	u.log.Debug("User received auth refresh update")

	if auth == nil {
		if err := u.Logout(); err != nil {
			log.WithError(err).
				WithField("userID", u.userID).
				Error("User logout failed while watching API auths")
		}
		return
	}

	creds, err := u.credStorer.UpdateToken(u.userID, auth.UID, auth.RefreshToken)
	if err != nil {
		u.log.WithError(err).Error("Failed to update refresh token in credentials store")
		return
	}

	u.creds = creds
}

// clearStore removes the database.
func (u *User) clearStore() error {
	u.log.Trace("Clearing user store")

	if u.store != nil {
		if err := u.store.Remove(); err != nil {
			return errors.Wrap(err, "failed to remove store")
		}
	} else {
		u.log.Warn("Store is not initialized: cleaning up store files manually")
		if err := u.storeFactory.Remove(u.userID); err != nil {
			return errors.Wrap(err, "failed to remove store manually")
		}
	}
	return nil
}

// closeStore just closes the store without deleting it.
func (u *User) closeStore() error {
	u.log.Trace("Closing user store")

	if u.store != nil {
		if err := u.store.Close(); err != nil {
			return errors.Wrap(err, "failed to close store")
		}
	}

	return nil
}

// ID returns the user's userID.
func (u *User) ID() string {
	return u.userID
}

// UsedBytes returns number of bytes used on server.
func (u *User) UsedBytes() int64 {
	return u.usedBytes
}

// TotalBytes returns number of bytes available on server.
func (u *User) TotalBytes() int64 {
	return u.totalBytes
}

// UpdateSpace will update TotalBytes and UsedBytes values from API user. If
// pointer is nill it will get fresh user from API. API user can come from
// update event which means it doesn't contain all data. Therefore only
// positive values will be updated.
func (u *User) UpdateSpace(apiUser *pmapi.User) {
	// If missing get latest pmapi.User from API instead of using cached
	// values from client.CurrentUser()
	if apiUser == nil {
		var err error
		apiUser, err = u.GetClient().GetUser(pmapi.ContextWithoutRetry(context.Background()))
		if err != nil {
			u.log.WithError(err).Warning("Cannot update user space")
			return
		}
	}

	if apiUser.UsedSpace != nil {
		u.usedBytes = *apiUser.UsedSpace
	}
	if apiUser.MaxSpace != nil {
		u.totalBytes = *apiUser.MaxSpace
	}
}

// Username returns the user's username as found in the user's credentials.
func (u *User) Username() string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.Name
}

// IsConnected returns whether user is logged in.
func (u *User) IsConnected() bool {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.IsConnected()
}

func (u *User) GetClient() pmapi.Client {
	if err := u.unlockIfNecessary(); err != nil {
		u.log.WithError(err).Error("Failed to unlock user")
	}
	return u.client
}

// unlockIfNecessary will not trigger keyring unlocking if it was already successfully unlocked.
func (u *User) unlockIfNecessary() error {
	if !u.creds.IsConnected() {
		return nil
	}

	if u.client.IsUnlocked() {
		return nil
	}

	// unlockIfNecessary is called with every access to underlying pmapi
	// client. Unlock should only finish unlocking when connection is back up.
	// That means it should try it fast enough and not retry if connection
	// is still down.
	err := u.client.Unlock(pmapi.ContextWithoutRetry(context.Background()), u.creds.Secret.MailboxPassword)
	if err == nil {
		return nil
	}

	if pmapi.IsFailedAuth(err) || pmapi.IsFailedUnlock(err) {
		if logoutErr := u.logout(); logoutErr != nil {
			u.log.WithError(logoutErr).Warn("Could not logout user")
		}
		return errors.Wrap(err, "failed to unlock user")
	}

	switch errors.Cause(err) {
	case pmapi.ErrNoConnection, pmapi.ErrUpgradeApplication:
		u.log.WithError(err).Warn("Skipping unlock for known reason")
	default:
		u.log.WithError(err).Error("Unknown unlock issue")
	}

	return nil
}

// GetPrimaryAddress returns the user's original address (which is
// not necessarily the same as the primary address, because a primary address
// might be an alias and be in position one).
func (u *User) GetPrimaryAddress() string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.Emails[0]
}

// GetStoreAddresses returns all addresses used by the store (so in combined mode,
// that's just the original address, but in split mode, that's all active addresses).
func (u *User) GetStoreAddresses() []string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.Emails[:1]
}

// GetAddresses returns list of all addresses.
func (u *User) GetAddresses() []string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.Emails
}

// GetAddressID returns the API ID of the given address.
func (u *User) GetAddressID(address string) (id string, err error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	if u.store != nil {
		address = strings.ToLower(address)
		return u.store.GetAddressID(address)
	}

	if u.client == nil {
		return "", errors.New("bridge account is not fully connected to server")
	}

	addresses := u.client.Addresses()
	pmapiAddress := addresses.ByEmail(address)
	if pmapiAddress != nil {
		return pmapiAddress.ID, nil
	}
	return "", errors.New("address not found")
}

func (u *User) CheckCredentials(slot, password string) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	verified := false
	if u.creds.Locked() {
		if err := u.creds.Unlock(slot, password); err != nil {
			return err
		}
		verified = true
	}

	if !u.creds.IsConnected() {
		return ErrLoggedOutUser
	}

	if verified {
		return nil
	}

	return u.creds.Unlock(slot, password)
}

func (u *User) UnlockCredentials(slot, password string) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	return u.creds.Unlock(slot, password)
}

func (u *User) BringOnline(username, password string) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	if u.client != nil {
		return nil
	}

	if u.creds.Locked() {
		if err := u.creds.Unlock("main", password); err != nil {
			return err
		}
	}

	if !u.creds.IsConnected() {
		return u.connect(u.clientManager.NewClient("", "", "", time.Time{}))
	}

	uid, ref, err := u.creds.SplitAPIToken()
	if err != nil {
		return errors.Wrap(err, "could not get user's refresh token")
	}

	ctx := pmapi.ContextWithoutRetry(context.Background())
	client, auth, err := u.clientManager.NewClientWithRefresh(ctx, uid, ref)
	if err != nil {
		connectErr := u.connect(u.clientManager.NewClient(uid, "", ref, time.Time{}))

		switch errors.Cause(err) {
		case pmapi.ErrNoConnection, pmapi.ErrUpgradeApplication:
			return connectErr
		}

		if pmapi.IsFailedAuth(connectErr) {
			if logoutErr := u.logout(); logoutErr != nil {
				logrus.WithError(logoutErr).Warn("Could not logout user")
			}
		}
		return errors.Wrap(err, "could not refresh token")
	}

	// Update the user's credentials with the latest auth used to connect this user.
	if u.creds, err = u.credStorer.UpdateToken(u.creds.UserID, auth.UID, auth.RefreshToken); err != nil {
		return errors.Wrap(err, "could not create get user's refresh token")
	}

	return u.connect(client)
}

// UpdateUser updates user details from API and saves to the credentials.
func (u *User) UpdateUser(ctx context.Context) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	user, err := u.client.UpdateUser(ctx)
	if err != nil {
		return err
	}

	if err := u.client.ReloadKeys(ctx, u.creds.Secret.MailboxPassword); err != nil {
		return errors.Wrap(err, "failed to reload keys")
	}

	creds, err := u.credStorer.UpdateEmails(u.userID, u.client.Addresses().ActiveEmails())
	if err != nil {
		return err
	}

	u.creds = creds

	u.UpdateSpace(user)

	return nil
}

// Logout logs out the user from pmapi, the credentials store, the mail store, and tries to remove as much
// sensitive data as possible.
func (u *User) Logout() error {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.logout()
}

func (u *User) logout() error {
	u.log.Debug("Logging out user")

	if !u.creds.IsConnected() {
		return nil
	}

	if u.client == nil {
		u.log.Warn("Failed to delete auth: no client")
	} else if err := u.client.AuthDelete(context.Background()); err != nil {
		u.log.WithError(err).Warn("Failed to delete auth")
	}

	creds, err := u.credStorer.Logout(u.userID)
	if err != nil {
		u.log.WithError(err).Warn("Could not log user out from credentials store")

		if err := u.credStorer.Delete(u.userID); err != nil {
			u.log.WithError(err).Error("Could not delete user from credentials store")
		}
	} else {
		u.creds = creds
	}

	// Do not close whole store, just event loop. Some information might be needed offline (e.g. addressID)
	u.closeEventLoopAndCacher()

	u.CloseAllConnections()

	runtime.GC()

	return nil
}

func (u *User) ListKeySlots() ([]string, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.credStorer.ListKeySlots(u.userID)
}

func (u *User) RemoveKeySlot(slot string) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	return u.credStorer.RemoveKeySlot(u.userID, slot)
}

func (u *User) AddKeySlot(slot, mainKey string) (string, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	return u.credStorer.AddKeySlot(u.userID, slot, mainKey)
}

func (u *User) closeEventLoopAndCacher() {
	if u.store == nil {
		return
	}

	u.store.CloseEventLoopAndCacher()
}

// CloseAllConnections calls CloseConnection for all users addresses.
func (u *User) CloseAllConnections() {
	for _, address := range u.creds.Emails {
		u.CloseConnection(address)
	}

	if u.store != nil {
		u.store.SetChangeNotifier(nil)
	}
}

// CloseConnection emits closeConnection event on `address` which should close all active connection.
func (u *User) CloseConnection(address string) {
	u.listener.Emit(events.CloseConnectionEvent, address)
}

func (u *User) GetStore() *store.Store {
	return u.store
}
