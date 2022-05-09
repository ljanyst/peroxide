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
	spacing := "%3d: %-20s %-20s %-15s %-15s "
	for idx, user := range b.Users.GetUsers() {
		connected := "disconnected"
		if user.IsConnected() {
			connected = "connected"
		}

		mode := "split"
		if user.IsCombinedAddressMode() {
			mode = "combined"
		}

		fmt.Printf(spacing, idx, user.Username(), user.GetPrimaryAddress(), connected, mode)

		for _, address := range user.GetAddresses() {
			fmt.Printf("%-20s", address)
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

func addAccount(b *bridge.Bridge, accountName string) error {
	if accountName == "" {
		return fmt.Errorf("Missing account name")
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

	user, err := b.Users.FinishLogin(client, auth, mailboxPassword)
	if err != nil {
		return fmt.Errorf("Login of account %s failed: %s", accountName, err)
	}

	fmt.Printf("Account %s has been added successfully.\n", user.Username())

	return nil
}
