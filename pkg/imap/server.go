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

package imap

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	imapid "github.com/ProtonMail/go-imap-id"
	"github.com/emersion/go-imap"
	imapappendlimit "github.com/emersion/go-imap-appendlimit"
	imapmove "github.com/emersion/go-imap-move"
	imapquota "github.com/emersion/go-imap-quota"
	imapunselect "github.com/emersion/go-imap-unselect"
	"github.com/emersion/go-imap/backend"
	imapserver "github.com/emersion/go-imap/server"
	"github.com/emersion/go-sasl"
	"github.com/ljanyst/peroxide/pkg/bridge"
	"github.com/ljanyst/peroxide/pkg/config/useragent"
	"github.com/ljanyst/peroxide/pkg/imap/id"
	"github.com/ljanyst/peroxide/pkg/imap/idle"
	"github.com/ljanyst/peroxide/pkg/imap/uidplus"
	"github.com/ljanyst/peroxide/pkg/listener"
	"github.com/ljanyst/peroxide/pkg/serverutil"
)

// Server takes care of IMAP listening serving. It implements serverutil.Server.
type Server struct {
	userAgent   *useragent.UserAgent
	debugClient bool
	debugServer bool
	port        int

	server     *imapserver.Server
	controller serverutil.Controller
}

// NewIMAPServer constructs a new IMAP server configured with the given options.
func NewIMAPServer(
	debugClient, debugServer bool,
	port int,
	tls *tls.Config,
	imapBackend backend.Backend,
	userAgent *useragent.UserAgent,
	eventListener listener.Listener,
) *Server {
	server := &Server{
		userAgent:   userAgent,
		debugClient: debugClient,
		debugServer: debugServer,
		port:        port,
	}

	server.server = newGoIMAPServer(tls, imapBackend, server.Address(), userAgent)
	server.controller = serverutil.NewController(server, eventListener)
	return server
}

func newGoIMAPServer(tls *tls.Config, backend backend.Backend, address string, userAgent *useragent.UserAgent) *imapserver.Server {
	server := imapserver.New(backend)
	server.TLSConfig = tls
	server.AllowInsecureAuth = true
	server.ErrorLog = serverutil.NewServerErrorLogger(serverutil.IMAP)
	server.AutoLogout = 30 * time.Minute
	server.Addr = address

	serverID := imapid.ID{
		imapid.FieldName:       "ProtonMail Bridge",
		imapid.FieldVendor:     "Proton Technologies AG",
		imapid.FieldSupportURL: "https://protonmail.com/support",
	}

	server.EnableAuth(sasl.Login, func(conn imapserver.Conn) sasl.Server {
		return sasl.NewLoginServer(func(address, password string) error {
			user, err := conn.Server().Backend.Login(nil, address, password)
			if err != nil {
				return err
			}

			ctx := conn.Context()
			ctx.State = imap.AuthenticatedState
			ctx.User = user
			return nil
		})
	})

	server.Enable(
		idle.NewExtension(),
		imapmove.NewExtension(),
		id.NewExtension(serverID, userAgent),
		imapquota.NewExtension(),
		imapappendlimit.NewExtension(),
		imapunselect.NewExtension(),
		uidplus.NewExtension(),
	)

	return server
}

// ListenAndServe will run server and all monitors.
func (s *Server) ListenAndServe() { s.controller.ListenAndServe() }

// Close turns off server and monitors.
func (s *Server) Close() { s.controller.Close() }

// Implements serverutil.Server interface.

func (Server) Protocol() serverutil.Protocol { return serverutil.IMAP }
func (s *Server) UseSSL() bool               { return false }
func (s *Server) Address() string            { return fmt.Sprintf("%s:%d", bridge.Host, s.port) }
func (s *Server) TLSConfig() *tls.Config     { return s.server.TLSConfig }

func (s *Server) DebugServer() bool { return s.debugServer }
func (s *Server) DebugClient() bool { return s.debugClient }

func (s *Server) SetLoggers(localDebug, remoteDebug io.Writer) {
	s.server.Debug = imap.NewDebugWriter(localDebug, remoteDebug)

	if !s.userAgent.HasClient() {
		s.userAgent.SetClient("UnknownClient", "0.0.1")
	}
}

func (s *Server) DisconnectUser(address string) {
	log.Info("Disconnecting all open IMAP connections for ", address)
	s.server.ForEachConn(func(conn imapserver.Conn) {
		connUser := conn.Context().User
		if connUser != nil && strings.EqualFold(connUser.Username(), address) {
			if err := conn.Close(); err != nil {
				log.WithError(err).Error("Failed to close the connection")
			}
		}
	})
}

func (s *Server) Serve(listener net.Listener) error { return s.server.Serve(listener) }
func (s *Server) StopServe() error                  { return s.server.Close() }
