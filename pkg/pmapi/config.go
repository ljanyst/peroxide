// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package pmapi

type Config struct {
	// HostURL is the base URL of API.
	HostURL string

	// AppVersion sets version to headers of each request.
	AppVersion string

	// UpgradeApplicationHandler is used to notify when there is a force upgrade.
	UpgradeApplicationHandler func()

	// TLSIssueHandler is used to notify when there is a TLS issue.
	TLSIssueHandler func()
}

func NewConfig() Config {
	return Config{
		HostURL:    getRootURL(),
		AppVersion: "LinuxBridge_1000.1000.1000+git",
	}
}

func (c *Config) getUserAgent() string {
	return "UnknownClient/0.0.1"
}
