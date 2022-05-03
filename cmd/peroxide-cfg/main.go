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
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
)

var genKey = flag.Bool("gen-key", false, "generate a random key for the encryption of credentials")
var genX509 = flag.Bool("gen-x509", false, "generate a self-signed X509 certificate")
var x509Org = flag.String("x509-org", "", "organization name to be used in X509 certificate")
var x509Cn = flag.String("x509-cn", "", "common name to be used in X509 certificate")
var x509KeyFile = flag.String("x509-key", "key.pem", "output file for the RSA key")
var x509CertFile = flag.String("x509-cert", "cert.pem", "output file for the X509 certificate")

func main() {
	flag.Parse()
	done := false

	if *genKey {
		var key [32]byte
		if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
			fmt.Fprintf(os.Stderr, "Can't read random bytes: %s\n", err)
			os.Exit(1)
		}
		password := base64.StdEncoding.EncodeToString(key[:])
		fmt.Println(password)
		done = true
	} else if *genX509 {
		if err := generateX509(*x509Org, *x509Cn, *x509CertFile, *x509KeyFile); err != nil {
			fmt.Fprintf(os.Stderr, "Can't generate an X509 certificate: %s\n", err)
			os.Exit(1)
		}
		done = true
	}

	if !done {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}
