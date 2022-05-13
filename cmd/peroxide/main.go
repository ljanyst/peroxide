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
	"os"

	"github.com/ljanyst/peroxide/pkg/bridge"
	"github.com/ljanyst/peroxide/pkg/logging"
	"github.com/sirupsen/logrus"
)

var config = flag.String("config", "/etc/peroxide.conf", "configuration file")
var logLevel = flag.String("log-level", "Warning", "account name")

func main() {
	flag.Parse()

	logging.SetLevel(*logLevel)

	b := &bridge.Bridge{}

	if err := b.Configure(*config); err != nil {
		logrus.WithError(err).Fatal("Failed to configure the bridge")
		os.Exit(1)
	}

	if err := b.Run(); err != nil {
		logrus.WithError(err).Fatal("Bridge exited with error")
		os.Exit(1)
	}
}
