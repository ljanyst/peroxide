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

func main() {
	flag.Parse()
	done := false

	if *genKey {
		var key [32]byte
		if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
			fmt.Fprintf(os.Stderr, "Can't read random bytes: %s:\n", err)
			os.Exit(1)
		}
		password := base64.StdEncoding.EncodeToString(key[:])
		fmt.Println(password)
		done = true
	}

	if !done {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}
