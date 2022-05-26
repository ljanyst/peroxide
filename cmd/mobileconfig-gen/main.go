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
	"io/ioutil"
	"os"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/google/uuid"
)

type Config struct {
	UUID       string
	Identifier string
	CardDAV    []DAV
	CalDAV     []DAV
	Email      []Email
}

type DAV struct {
	AccountDescription string
	HostName           string
	Username           string
	Password           string
	UseSSL             bool
	Port               int
	PrincipalURL       string
	UUID               string
	Identifier         string
}

type Email struct {
	AccountName        string
	AccountDescription string
	Address            string
	IncomingHostName   string
	IncomingPortNumber int
	IncomingUsername   string
	IncomingPassword   string
	IncomingUseSSL     bool
	OutgoingHostName   string
	OutgoingPortNumber int
	OutgoingUsername   string
	OutgoingPassword   string
	OutgoingUseSSL     bool
	UUID               string
	Identifier         string
}

var configTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Inc//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>PayloadVersion</key>
    <integer>1</integer>

    <key>PayloadUUID</key>
    <string>{{ .UUID }}</string>

    <key>PayloadType</key>
    <string>Configuration</string>

    <key>PayloadIdentifier</key>
    <string>{{ .Identifier }}</string>

    <key>Label</key>
    <string>Account Configuration</string>

    <key>PayloadContent</key>
    <array>
{{ range .CardDAV }}
    <dict>
        <key>CardDAVAccountDescription</key>
        <string>{{ .AccountDescription }}</string>

        <key>CardDAVHostName</key>
        <string>{{ .HostName }}</string>

        <key>CardDAVPrincipalURL</key>
        <string>{{ .PrincipalURL }}</string>

        <key>CardDAVUsername</key>
        <string>{{ .Username }}</string>

        <key>CardDAVPassword</key>
        <string>{{ .Password }}</string>

        <key>CardDAVUseSSL</key>
        <{{ .UseSSL }} />

        <key>CardDAVPort</key>
        <integer>{{ .Port }}</integer>

        <key>PayloadVersion</key>
        <integer>1</integer>

        <key>PayloadType</key>
        <string>com.apple.carddav.account</string>

        <key>PayloadIdentifier</key>
        <string>{{ .Identifier }}</string>

        <key>PayloadUUID</key>
        <string>{{ .UUID }}</string>

        <key>PayloadOrganization</key>
        <string>A nice company</string>

        <key>PayloadDescription</key>
        <string>Configures CardDAV account</string>

    </dict>
{{ end }}
{{ range .CalDAV }}
    <dict>
        <key>CalDAVAccountDescription</key>
        <string>{{ .AccountDescription }}</string>

        <key>CalDAVHostName</key>
        <string>{{ .HostName }}</string>

        <key>CalDAVPrincipalURL</key>
        <string>{{ .PrincipalURL }}</string>

        <key>CalDAVUsername</key>
        <string>{{ .Username }}</string>

        <key>CalDAVPassword</key>
        <string>{{ .Password }}</string>

        <key>CalDAVUseSSL</key>
        <{{ .UseSSL }} />

        <key>CalDAVPort</key>
        <integer>{{ .Port }}</integer>

        <key>PayloadVersion</key>
        <integer>1</integer>

        <key>PayloadType</key>
        <string>com.apple.caldav.account</string>

        <key>PayloadIdentifier</key>
        <string>{{ .Identifier }}</string>

        <key>PayloadUUID</key>
        <string>{{ .UUID }}</string>

        <key>PayloadOrganization</key>
        <string>A nice company</string>

        <key>PayloadDescription</key>
        <string>Configures CalDAV account</string>

    </dict>
{{ end }}
{{ range .Email }}
    <dict>
        <key>EmailAccountDescription</key>
        <string>{{ .AccountDescription }}</string>

        <key>EmailAccountName</key>
        <string>{{ .AccountName }}</string>

        <key>EmailAccountType</key>
        <string>EmailTypeIMAP</string>

        <key>EmailAddress</key>
        <string>{{ .Address }}</string>

        <key>IncomingMailServerAuthentication</key>
        <string>EmailAuthPassword</string>

        <key>IncomingMailServerHostName</key>
        <string>{{ .IncomingHostName }}</string>

        <key>IncomingMailServerPortNumber</key>
        <integer>{{ .IncomingPortNumber }}</integer>

        <key>IncomingMailServerUsername</key>
        <string>{{ .IncomingUsername }}</string>

        <key>IncomingPassword</key>
        <string>{{ .IncomingPassword }}</string>

        <key>IncomingMailServerUseSSL</key>
        <{{ .IncomingUseSSL }} />

        <key>IncomingMailServerIMAPPathPrefix</key>
        <string></string>

        <key>OutgoingMailServerAuthentication</key>
        <string>EmailAuthPassword</string>

        <key>OutgoingMailServerHostName</key>
        <string>{{ .OutgoingHostName }}</string>

        <key>OutgoingMailServerPortNumber</key>
        <integer>{{ .OutgoingPortNumber }}</integer>

        <key>OutgoingMailServerUsername</key>
        <string>{{ .OutgoingUsername }}</string>

        <key>OutgoingPassword</key>
        <string>{{ .OutgoingPassword }}</string>

        <key>OutgoingMailServerUseSSL</key>
        <{{ .OutgoingUseSSL }} />

        <key>PayloadVersion</key>
        <integer>1</integer>

        <key>PayloadType</key>
        <string>com.apple.mail.managed</string>

        <key>PayloadIdentifier</key>
        <string>{{ .Identifier }}</string>

        <key>PayloadUUID</key>
        <string>{{ .UUID }}</string>

        <key>PayloadOrganization</key>
        <string>A nice company</string>

        <key>PayloadDescription</key>
        <string>Configures an Email account</string>

    </dict>
{{ end }}
    </array>
</dict>
</plist>
`

var out = flag.String("out", "account.mobileconfig", "output config")
var in = flag.String("in", "account.json", "configuration data")

func main() {
	flag.Parse()
	templ := template.Must(template.New("mobileconfig").Parse(configTmpl))

	outFile, err := os.OpenFile(*out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Cannot open the output file for writing: %s\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	data, err := ioutil.ReadFile(*in)
	if err != nil {
		fmt.Printf("Cannot read the configuration data: %s\n", err)
		os.Exit(1)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("Cannot unmarshal the configuration data: %s\n", err)
		os.Exit(1)
	}

	cfg.UUID = uuid.New().String()
	for i := 0; i < len(cfg.CardDAV); i++ {
		cfg.CardDAV[i].UUID = uuid.New().String()
		cfg.CardDAV[i].Identifier = uuid.New().String()
	}

	for i := 0; i < len(cfg.CalDAV); i++ {
		cfg.CalDAV[i].UUID = uuid.New().String()
		cfg.CalDAV[i].Identifier = uuid.New().String()
	}

	for i := 0; i < len(cfg.Email); i++ {
		cfg.Email[i].UUID = uuid.New().String()
		cfg.Email[i].Identifier = uuid.New().String()
	}

	if cfg.Identifier == "" {
		cfg.Identifier = "com.example.account"
	}

	if err := templ.Execute(outFile, cfg); err != nil {
		fmt.Printf("Cannot execute the config template: %s\n", err)
		os.Exit(1)
	}
}
