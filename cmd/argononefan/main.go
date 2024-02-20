/*
 *  Copyright 2024 Markus W Mahlberg
 *
 *  main.go is part of argononefan
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/go-hclog"
	"github.com/mwmahlberg/argononefan"
)

const (
	readingTemperatureMsg = "readingTemperature"
)

var (
	l       hclog.Logger
	version = "dev"
)

var cli struct {
	Debug      bool   `short:"d" long:"debug" help:"Enable debug mode" default:"false"`
	DeviceFile string `short:"f" long:"file" help:"File path in sysfs containing current CPU temperature" default:"/sys/class/thermal/thermal_zone0/temp"`
	Bus        int    `short:"b" long:"bus" help:"I2C bus the fan resides on" default:"0"`

	Daemon      daemonCmd        `kong:"cmd,help='Run the fan control daemon'"`
	Temperature temperatureCmd   `kong:"cmd,help='Read the current CPU temperature'"`
	SetSpeed    setSpeedCmd      `kong:"cmd,help='Set the fan speed manually'"`
	Version     kong.VersionFlag `env:"-"`
}

func main() {

	ctx := kong.Parse(&cli,
		kong.Name("argononefan"),
		kong.Description("Tools for fan control of the ArgonOne case"),
		kong.DefaultEnvars("ARGONONEFAN"),
		kong.Vars{
			"version":         version,
			"help_hysteresis": hystereisHelp,
			"help_thresholds": thresholdsHelp,
		},
	)
	ctx.Stderr = os.Stdout

	var level hclog.Level = hclog.Info
	colored := hclog.ColorOff

	if cli.Debug {
		level = hclog.Debug
		colored = hclog.AutoColor
	}

	l = hclog.New(&hclog.LoggerOptions{
		DisableTime:     !cli.Debug,
		Color:           colored,
		IncludeLocation: cli.Debug,
		Level:           level,
	})

	l.Debug("Executing", "command", ctx.Command())

	// We need to bind the logger to that specific interface type
	// because kong's Bind function does not support binding interfaces
	// but only concrete types, of which it will determine the
	// reflection type and then bind to that.
	ctx.BindTo(l, (*hclog.Logger)(nil))
	ctx.Bind([]argononefan.ThermalReaderOption{argononefan.WithThermalDeviceFile(cli.DeviceFile)})
	ctx.Bind([]argononefan.FanOption{argononefan.OnBus(cli.Bus)})

	ctx.FatalIfErrorf(ctx.Run())

}
