// Copyright (c) 2022 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
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

// Package bridge implements the bridge CLI application.
package bridge

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/pkg/errors"
)

// GetConfig tries to load TLS config or generate new one which is then returned.
func loadTlsConfig(certPath, keyPath string) (*tls.Config, error) {
	c, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load cert and key")
	}

	c.Leaf, err = x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse the certificate")
	}

	if time.Now().Add(31 * 24 * time.Hour).After(c.Leaf.NotAfter) {
		return nil, errors.Wrap(err, "The X509 certificate is about to expire")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(c.Leaf)

	return &tls.Config{
		Certificates: []tls.Certificate{c},
		ServerName:   c.Leaf.Subject.CommonName,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
	}, nil
}
