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

package users

import (
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestLoginDecoder(t *testing.T) {
	test := func(login, parsedLogin, parsedKeySlot string) {
		l, ks := DecodeLogin(login)
		r.Equal(t, parsedLogin, l)
		r.Equal(t, parsedKeySlot, ks)
	}

	test("", "", "main")
	test("foo@bar@baz", "foo@bar@baz", "main")
	test("foo..test@bar@baz", "foo..test@bar@baz", "main")
	test("foo..test..test@bar@baz", "foo..test..test@bar@baz", "main")
	test("foo", "foo", "main")
	test("foo@bar", "foo@bar", "main")
	test("foo..test@bar", "foo@bar", "test")
}
