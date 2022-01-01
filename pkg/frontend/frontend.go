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
	"github.com/ljanyst/peroxide/pkg/config/useragent"
	"github.com/ljanyst/peroxide/pkg/frontend/cli"
	"github.com/ljanyst/peroxide/pkg/frontend/qt"
	"github.com/ljanyst/peroxide/pkg/frontend/types"
	"github.com/ljanyst/peroxide/pkg/locations"
	"github.com/ljanyst/peroxide/pkg/updater"
	"github.com/ljanyst/peroxide/pkg/listener"
)

type Frontend interface {
	Loop() error
	NotifyManualUpdate(update updater.VersionInfo, canInstall bool)
	SetVersion(update updater.VersionInfo)
	NotifySilentUpdateInstalled()
	NotifySilentUpdateError(error)
	WaitUntilFrontendIsReady()
}

// New returns initialized frontend based on `frontendType`, which can be `cli` or `qt`.
func New(
	version,
	buildVersion,
	programName,
	frontendType string,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	locations *locations.Locations,
	settings *settings.Settings,
	eventListener listener.Listener,
	updater types.Updater,
	userAgent *useragent.UserAgent,
	bridge *bridge.Bridge,
	noEncConfirmator types.NoEncConfirmator,
	restarter types.Restarter,
) Frontend {
	bridgeWrap := types.NewBridgeWrap(bridge)
	switch frontendType {
	case "qt":
		return qt.New(
			version,
			buildVersion,
			programName,
			showWindowOnStart,
			panicHandler,
			locations,
			settings,
			eventListener,
			updater,
			userAgent,
			bridgeWrap,
			noEncConfirmator,
			restarter,
		)
	case "cli":
		return cli.New(
			panicHandler,
			locations,
			settings,
			eventListener,
			updater,
			bridgeWrap,
			restarter,
		)
	default:
		return nil
	}
}