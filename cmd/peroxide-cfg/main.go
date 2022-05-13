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
	"flag"
	"fmt"
	"os"

	"github.com/ljanyst/peroxide/pkg/bridge"
	"github.com/ljanyst/peroxide/pkg/logging"
)

var config = flag.String("config", "/etc/peroxide.conf", "configuration file")
var action = flag.String("action", "", "one of: gen-x509, list-accounts, delete-account, login-account, add-key, remove-key")
var x509Org = flag.String("x509-org", "", "organization name to be used in X509 certificate")
var x509Cn = flag.String("x509-cn", "", "common name to be used in X509 certificate")
var x509KeyFile = flag.String("x509-key", "key.pem", "output file for the RSA key")
var x509CertFile = flag.String("x509-cert", "cert.pem", "output file for the X509 certificate")
var accountName = flag.String("account-name", "", "account name")
var keyName = flag.String("key-name", "", "key name")
var logLevel = flag.String("log-level", "Warning", "account name")

func main() {
	flag.Parse()
	done := true

	logging.SetLevel(*logLevel)

	b := &bridge.Bridge{}

	err := b.Configure(*config)
	if err != nil {
		fmt.Printf("Failed to configure the bridge: %s\n", err)
		os.Exit(1)
	}

	switch *action {
	case "gen-x509":
		err = generateX509(*x509Org, *x509Cn, *x509CertFile, *x509KeyFile)
	case "list-accounts":
		listAccounts(b)
	case "delete-account":
		err = deleteAccount(b, *accountName)
	case "login-account":
		err = loginAccount(b, *accountName)
	case "add-key":
		err = addKey(b, *accountName, *keyName)
	case "remove-key":
		err = removeKey(b, *accountName, *keyName)
	default:
		done = false
	}

	if err != nil {
		fmt.Printf("Failed to execute command: %s\n", err)
		os.Exit(1)
	}

	if !done {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}
