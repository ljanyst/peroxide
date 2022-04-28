// Copyright (c) 2022 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

// Package constants contains variables that are set via ldflags during build.
package constants

import "fmt"

const VendorName = "protonmail"

// nolint[gochecknoglobals]
var (
	// Version of the build.
	Version = "1000.1000.1000+git"

	// Revision is current hash of the build.
	Revision = "bf2a7aaf0f"

	// BuildTime stamp of the build.
	BuildTime = "2022-01-01T17:20:35+0100"

	// BuildVersion is derived from LongVersion and BuildTime.
	BuildVersion = fmt.Sprintf("%v (%v) %v", Version, Revision, BuildTime)
)
