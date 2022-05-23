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
	"os/signal"
	"syscall"

	"github.com/ljanyst/peroxide/pkg/bridge"
	"github.com/ljanyst/peroxide/pkg/logging"
	"github.com/sirupsen/logrus"
)

var config = flag.String("config", "/etc/peroxide.conf", "configuration file")
var logLevel = flag.String("log-level", "Warning", "account name")
var logFile = flag.String("log-file", "", "output file for diagnostics")

func setLogFile(filePath string) *os.File {
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Cannot open log file: " + err.Error())
	}
	logrus.SetOutput(logFile)
	return logFile
}

func rotateLogFile(filePath string) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGHUP)

	f := setLogFile(filePath)

	for {
		select {
		case <-signalCh:
			f.Close()
			f = setLogFile(filePath)
			logrus.Debug("Logfile rotated")
		}
	}
}

func main() {
	flag.Parse()

	logging.SetLevel(*logLevel)

	if *logFile != "" {
		go rotateLogFile(*logFile)
	}

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
