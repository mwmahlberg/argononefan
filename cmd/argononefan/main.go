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
	"slices"
	"syscall"
	"time"

	"golang.org/x/exp/maps"

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
		Thresholds    map[float32]int `short:"t" long:"threshold" help:"Threshold map of °C to fan speed in %" type:"float32:int" default:"60=100;55=50;50=10"`
		CheckInterval time.Duration   `short:"i" long:"interval" help:"Check interval" default:"5s"`
	} `kong:"cmd,help='Run the fan control daemon'"`

	Temperature struct {
		Imperial bool `short:"i" long:"imperial" help:"Display temperature in imperial system" default:"false" env:"-"`
	} `kong:"cmd,help='Read the current CPU temperature'"`

	SetSpeed struct {
		Speed int `arg:"" help:"Fan speed" required:"" min:"0" max:"100"`
	} `kong:"cmd,help='Set the fan speed manually'"`
}

func main() {

	ctx := kong.Parse(&cli,
		kong.Name("argononefan"),
		kong.Description("Daemon to adjust the fan speed of the Argon One case"),
		kong.DefaultEnvars("ARGONONEFAN"),
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
	tr, err := argononefan.NewThermalReader(argononefan.WithThermalDeviceFile(cli.DeviceFile))
	ctx.FatalIfErrorf(err, "creating thermal reader")
	l.Debug("Executing", "command", ctx.Command())
	switch ctx.Command() {
	case "temperature":

		var t float32
		var frmt string
		var readErr error
		if cli.Temperature.Imperial {
			t, readErr = tr.Fahrenheit()
			frmt = "Temperature: %2.1f°F"
		} else {
			t, readErr = tr.Celsius()
			frmt = "Temperature: %2.1f°C"
		}
		ctx.FatalIfErrorf(readErr, readingTemperatureMsg)
		ctx.Printf(frmt, t)
		os.Exit(0)
	case "set-speed <speed>":
		if cli.SetSpeed.Speed < 0 || cli.SetSpeed.Speed > 100 {
			ctx.Fatalf("desired fan speed is out of range [0-100]: %d", cli.SetSpeed.Speed)
		}
		fan, err := argononefan.Connect(argononefan.OnBus(cli.Bus))
		ctx.FatalIfErrorf(err, "connecting to fan")
		ctx.FatalIfErrorf(fan.SetSpeed(cli.SetSpeed.Speed), "setting fan speed")
		os.Exit(0)
	}

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

	l.Debug("Starting goroutine reading temperature")
	tempC, done := readTemp(cli.Daemon.CheckInterval, tr)

	l.Debug("Starting adjust goroutine")
	go control(fan, cli.Daemon.Thresholds, tempC)

	l.Debug("Waiting for stop signal")
	<-stopsig
	defer fan.SetSpeed(100)
	l.Debug("Stop signal received")

	l.Debug("Closing temperature reading goroutine")
	done <- true
	l.Debug("Waiting for adjust goroutine to finish")

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
				l.Debug("Closing temperature channel")
				close(c)
				l.Debug("Exiting...")
				return

			case <-tick.C:

				t, err := tr.Celsius()
				if err != nil {
					ml.Error(readingTemperatureMsg, "error", err)
					continue
				}
				ml.Debug("Read temperature", "temperature", fmt.Sprintf("%2.1f", t))
				ml.Debug("Sending temperature to adjust goroutine")
				c <- t
			}
		}
	}()
	return c, done
}

func control(fan *argononefan.Fan, config map[float32]int, tempC <-chan float32) {

	ml := l.Named("control")
	// Ensure we are looking at the thresholds in descending order
	thresholds := maps.Keys(config)
	slices.Sort(thresholds)
	slices.Reverse(thresholds)
	ml.Debug("Thresholds", "thresholds", thresholds)

	var currentIdx int

	for currentTemperature := range tempC {
		ml.Debug("Received temperature from reading goroutine", "temperature", fmt.Sprintf("%2.1f", currentTemperature))

		// Find the index of the threshold matching the current temperature
		idx := slices.IndexFunc(thresholds, func(t float32) bool {
			// This requires the thresholds to be sorted from higher to lower
			return currentTemperature >= t
		})

		switch idx {
		case currentIdx:
			ml.Debug("Temperature is still within the same threshold, no need to adjust fan speed")
		case -1:
			ml.Debug("Temperature is lower than the lowest threshold, set fan to 0% speed")
			currentIdx = -1
			fan.SetSpeed(0)
		default:
			ml.Debug("Found threshold", "index", idx, "threshold", thresholds[idx], "fanSpeed", config[thresholds[idx]])
			currentIdx = idx
			fan.SetSpeed(config[thresholds[idx]])
		}
	}

}
