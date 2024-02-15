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
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	Daemon struct {
		Thresholds    *thresholds   `short:"t" long:"threshold" help:"${help_thresholds}" default:"70=100;60=50;55=10"`
		Hysteresis    float32       `long:"hysteresis" help:"${help_hysteresis}" default:"1.0"`
		CheckInterval time.Duration `short:"i" long:"interval" help:"Check interval" default:"5s"`
	} `kong:"cmd,help='Run the fan control daemon'"`

	Temperature temperatureCmd `kong:"cmd,help='Read the current CPU temperature'"`

	SetSpeed struct {
		Speed int `arg:"" help:"Fan speed" required:"" min:"0" max:"100"`
	} `kong:"cmd,help='Set the fan speed manually'"`
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
		Name:  "argononefand",
		Level: level,
	})

	l.Debug("Executing", "command", ctx.Command())
	switch ctx.Command() {
	case "temperature":
		ctx.Run(&context{ThermalDeviceFile: cli.DeviceFile, Imperial: cli.Temperature.Imperial, logger: l.Named("temperature")})
	case "set-speed <speed>":
		if cli.SetSpeed.Speed < 0 || cli.SetSpeed.Speed > 100 {
			ctx.Fatalf("desired fan speed is out of range [0-100]: %d", cli.SetSpeed.Speed)
		}
		fan, err := argononefan.Connect(argononefan.OnBus(cli.Bus))
		ctx.FatalIfErrorf(err, "connecting to fan")
		ctx.FatalIfErrorf(fan.SetSpeed(cli.SetSpeed.Speed), "setting fan speed")
		os.Exit(0)
	}

	tr, err := argononefan.NewThermalReader(argononefan.WithThermalDeviceFile(cli.DeviceFile))
	ctx.FatalIfErrorf(err, "creating thermal reader")
	if l.IsDebug() {
		cli.Daemon.Thresholds.RLock()
		l.Debug("Index", "index", cli.Daemon.Thresholds.idx)
		cli.Daemon.Thresholds.RUnlock()
	}

	l.Debug("Connecting to fan")
	fan, err := argononefan.Connect(argononefan.OnBus(cli.Bus))

	if err != nil {
		l.Error("connecting to fan", "error", err)
		os.Exit(1)
	}

	fan.SetSpeed(100)

	l.Debug("Running with configuration", "config", cli)

	l.Debug("Setting up signal handling")
	var stopsig = make(chan os.Signal, 1)
	signal.Notify(stopsig, syscall.SIGTERM, syscall.SIGINT)

	l.Debug("Starting goroutine", "name", "read")
	tempC, done := readTemp(cli.Daemon.CheckInterval, tr)

	l.Debug("Starting goroutine", "name", "control")
	go control(fan, cli.Daemon.Thresholds, cli.Daemon.Hysteresis, tempC)

	l.Debug("Waiting for stop signal")
	<-stopsig
	defer fan.SetSpeed(100)
	l.Debug("Stop signal received")

	l.Debug("Shutting down goroutine", "name", "read")
	done <- true
	l.Debug("Waiting for goroutine to finish", "name", "control")

	lastTemp, err := tr.Celsius()
	ctx.FatalIfErrorf(err, readingTemperatureMsg)

	l.Warn("Fan control is shutting down, setting fan to 100% speed as a safety measure", "temperature", fmt.Sprintf("%2.1f°C", lastTemp))

}

func readTemp(interval time.Duration, tr *argononefan.ThermalReader) (<-chan float32, chan<- bool) {
	ml := l.Named("read")

	c := make(chan float32)
	done := make(chan bool)
	go func() {

		tick := time.NewTicker(interval)

		ml.Debug("Start reading temperature", "interval", interval)

		for {
			select {

			case <-done:
				ml.Debug("Received stop signal")
				ml.Debug("Closing temperature channel")
				close(c)
				ml.Debug("Exiting...")
				return

			case <-tick.C:

				t, err := tr.Celsius()
				if err != nil {
					ml.Error(readingTemperatureMsg, "error", err)
					continue
				}
				ml.Debug("Read temperature", "temperature", fmt.Sprintf("%2.1f", t))
				ml.Debug("Sending temperature to control")
				c <- t
			}
		}
	}()
	return c, done
}

func control(fan *argononefan.Fan, config *thresholds, hysteresis float32, tempC <-chan float32) {

	ml := l.Named("control")

	var currentSpeed int = -1

	for currentTemperature := range tempC {
		ml.Debug("Received temperature from read", "temperature", fmt.Sprintf("%2.1f", currentTemperature))

		speed := config.GetSpeed(currentTemperature)
		if speed < currentSpeed {
			speed = config.GetSpeedWithHysteresis(currentTemperature, hysteresis)
		}
		switch speed {
		case currentSpeed:
			ml.Debug("Temperature is still within the same threshold, no need to adjust fan speed")
		default:
			ml.Debug("Found threshold", "threshold", config.GetThreshold(currentTemperature), "computed fanSpeed with hystersis", config.GetSpeedWithHysteresis(currentTemperature, hysteresis))
			currentSpeed = speed
			fan.SetSpeed(speed)
		}
	}

}
