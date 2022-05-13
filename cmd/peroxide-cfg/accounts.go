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

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/ljanyst/peroxide/pkg/bridge"
	"github.com/ljanyst/peroxide/pkg/users"
)

func askPass(prompt string) ([]byte, error) {
	f := os.Stdin
	if !isatty.IsTerminal(f.Fd()) {
		// This can happen if stdin is used for piping data
		var err error
		if f, err = os.Open("/dev/tty"); err != nil {
			return nil, err
		}
		defer f.Close()
	}
	fmt.Fprintf(os.Stderr, "%v: ", prompt)
	b, err := terminal.ReadPassword(int(f.Fd()))
	if err == nil {
		fmt.Fprintf(os.Stderr, "\n")
	}
	return b, err
}

func listAccounts(b *bridge.Bridge) {
	for idx, user := range b.Users.GetUsers() {
		fmt.Printf("%3d: %s ", idx, user.Username())

		fmt.Printf("| addresses: ")
		for _, address := range user.GetAddresses() {
			fmt.Printf("%s ", address)
		}

		fmt.Printf("| keys: ")
		slots, _ := user.ListKeySlots()
		for _, slot := range slots {
			fmt.Printf("%s ", slot)
		}

		fmt.Println()
	}
}

func deleteAccount(b *bridge.Bridge, accountName string) error {
	if accountName == "" {
		return fmt.Errorf("Missing account name")
	}

	userArr := b.Users.GetUsers()
	if len(userArr) == 0 {
		return fmt.Errorf("No registered user accounts")
	}

	var user *users.User

	for _, u := range userArr {
		if u.Username() == accountName {
			user = u
			break
		}
	}

	if user == nil {
		return fmt.Errorf("Account %s not found", accountName)
	}

	if err := user.Logout(); err != nil {
		return fmt.Errorf("Logout of account %s failed: %s", accountName, err)
	}

	if err := b.Users.DeleteUser(user.ID(), true); err != nil {
		return fmt.Errorf("Deletion of account %s failed: %s", accountName, err)
	}

	return nil
}

func loginAccount(b *bridge.Bridge, accountName string) error {
	if accountName == "" {
		return fmt.Errorf("Missing account name")
	}

	user, _ := b.Users.GetUser(accountName)
	if user != nil {
		mainKey, err := askPass("Main key")
		if err != nil {
			return fmt.Errorf("The main key is required to modify an existing user: %s", err)
		}

		if len(mainKey) == 0 {
			return fmt.Errorf("The main key is required to modify an existing user")
		}

		if err := user.UnlockCredentials("main", string(mainKey)); err != nil {
			return fmt.Errorf("Unable to unlock credentials: %s", err)
		}

		if err := user.Logout(); err != nil {
			return fmt.Errorf("Unable to logout previous session: %s", err)
		}
	}

	password, err := askPass("Password")
	if err != nil {
		return fmt.Errorf("Unable to read password: %s", err)
	}

	if len(password) == 0 {
		return fmt.Errorf("Empty password")
	}

	fmt.Printf("Authenticating %s...\n", accountName)
	client, auth, err := b.Users.Login(accountName, password)
	if err != nil {
		return fmt.Errorf("Login of account %s failed: %s", accountName, err)
	}

	if auth.HasTwoFactor() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Printf("2FA TOTP code: ")
		scanner.Scan()
		code := scanner.Text()

		if code == "" {
			return fmt.Errorf("Empty 2FA TOTP code")
		}

		err = client.Auth2FA(context.Background(), code)
		if err != nil {
			return fmt.Errorf("2FA of account %s failed: %s", accountName, err)
		}
	}

	mailboxPassword := password
	if auth.HasMailboxPassword() {
		mailboxPassword, err = askPass("Mailbox password: ")
		if err != nil {
			return fmt.Errorf("Unable to read mailbox password: %s", err)
		}
	}

	if len(mailboxPassword) == 0 {
		return fmt.Errorf("Empty mailbox password")
	}

	user, key, err := b.Users.FinishLogin(client, auth, mailboxPassword, "")
	if err != nil {
		return fmt.Errorf("Login of account %s failed: %s", accountName, err)
	}

	fmt.Printf("Account %s has been added successfully.\n", user.Username())
	if len(key) != 0 {
		fmt.Printf("Main key: %s\n", key)
		fmt.Printf("PLEASE MAKE SURE TO NOTE THE KEY. IT'S NOT STORED ANYWHERE.\n")
	}

	return nil
}

func addKey(b *bridge.Bridge, accountName, keyName string) error {
	if accountName == "" || keyName == "" {
		return fmt.Errorf("Key name or account name empty")
	}

	user, err := b.Users.GetUser(accountName)
	if err != nil {
		return fmt.Errorf("Cannot get user data: %s", err)
	}

	mainKey, err := askPass("Main key")
	if err != nil {
		return fmt.Errorf("The main key is required to add a new key: %s", err)
	}

	if len(mainKey) == 0 {
		return fmt.Errorf("The main key is required to add a new key")
	}

	key, err := user.AddKeySlot(keyName, string(mainKey))
	if err != nil {
		return fmt.Errorf("Cannot add key slot: %s", err)
	}

	fmt.Printf("Added key %s: %s\n", keyName, key)
	fmt.Printf("PLEASE MAKE SURE TO NOTE THE KEY. IT'S NOT STORED ANYWHERE.\n")

	return nil
}

func removeKey(b *bridge.Bridge, accountName, keyName string) error {
	if accountName == "" || keyName == "" {
		return fmt.Errorf("Key name or account name empty")
	}

	user, err := b.Users.GetUser(accountName)
	if err != nil {
		return fmt.Errorf("Cannot get user data: %s", err)
	}

	return user.RemoveKeySlot(keyName)
}
