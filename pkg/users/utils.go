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
	"strings"
)

// Extract the login and key slot from the login information
func DecodeLogin(login string) (string, string) {
	if login == "" {
		return "", "main"
	}

	splitLogin := strings.Split(login, "@")
	if len(splitLogin) > 2 || len(splitLogin) == 0 {
		return login, "main"
	}

	splitUser := strings.Split(splitLogin[0], "..")
	if len(splitUser) > 2 || len(splitUser) == 0 {
		return login, "main"
	}

	userName := splitUser[0]
	slot := "main"
	if len(splitUser) == 2 {
		slot = splitUser[1]
	}

	if len(splitLogin) == 2 {
		userName = userName + "@" + splitLogin[1]
	}

	return userName, slot
}
