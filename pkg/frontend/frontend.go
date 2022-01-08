// Copyright (c) 2021 Proton Technologies AG
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

// Package frontend provides all interfaces of the Bridge.
package frontend

import (
	"github.com/ljanyst/peroxide/pkg/bridge"
	"github.com/ljanyst/peroxide/pkg/config/settings"
	"github.com/ljanyst/peroxide/pkg/frontend/cli"
	"github.com/ljanyst/peroxide/pkg/frontend/types"
	"github.com/ljanyst/peroxide/pkg/listener"
)

type Frontend interface {
	Loop() error
}

// New returns initialized frontend based on `frontendType`, which can be `cli` or `qt`.
func New(
	settings *settings.Settings,
	eventListener listener.Listener,
	bridge *bridge.Bridge,
) Frontend {
	bridgeWrap := types.NewBridgeWrap(bridge)
	return cli.New(
		settings,
		eventListener,
		bridgeWrap,
	)
}
