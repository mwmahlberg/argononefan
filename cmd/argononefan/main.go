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
	l hclog.Logger
)

var cli struct {
	Debug      bool   `short:"d" long:"debug" help:"Enable debug mode" default:"false"`
	DeviceFile string `short:"f" long:"file" help:"File path in sysfs containing current CPU temperature" default:"/sys/class/thermal/thermal_zone0/temp"`
	Bus        int    `short:"b" long:"bus" help:"I2C bus the fan resides on" default:"0"`

	Daemon      daemonCmd      `kong:"cmd,help='Run the fan control daemon'"`
	Temperature temperatureCmd `kong:"cmd,help='Read the current CPU temperature'"`
	SetSpeed    setSpeedCmd    `kong:"cmd,help='Set the fan speed manually'"`
}

func main() {

	ctx := kong.Parse(&cli,
		kong.Name("argononefan"),
		kong.Description("Tools for fan control of the ArgonOne case"),
		kong.DefaultEnvars("ARGONONEFAN"),
		kong.Vars{
			"help_hysteresis": hystereisHelp,
			"help_thresholds": thresholdsHelp,
		},
	)
	ctx.Stderr = os.Stdout

	var level hclog.Level = hclog.Info
	if cli.Debug {
		level = hclog.Debug
	}

	l = hclog.New(&hclog.LoggerOptions{
		Name:  "argononefan",
		Level: level,
	})

	l.Debug("Executing", "command", ctx.Command())

	rerr := ctx.Run(&context{
		logger:            l,
		fanOptions:        []argononefan.FanOption{argononefan.OnBus(cli.Bus)},
		thermalDeviceFile: cli.DeviceFile,
	})
	ctx.FatalIfErrorf(rerr, "setting fan speed")

}
