/*
 *  Copyright 2024 Markus W Mahlberg
 *
 *  help.go is part of argononefan
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

	"github.com/hashicorp/go-hclog"
	"github.com/mwmahlberg/argononefan"
)

type daemonCmd struct {
	Thresholds    *thresholds   `short:"t" long:"threshold" help:"${help_thresholds}" default:"70=100;60=50;55=10"`
	Hysteresis    float32       `long:"hysteresis" help:"${help_hysteresis}" default:"1.0"`
	CheckInterval time.Duration `short:"i" long:"interval" help:"Check interval" default:"5s"`
	logger        hclog.Logger  `kong:"-"`
}

func (d *daemonCmd) Run(ctx *context) error {
	d.logger = ctx.logger.Named("daemon")
	d.logger.Info("Starting daemon", "thresholds", d.Thresholds, "hysteresis", d.Hysteresis, "interval", d.CheckInterval)

	d.logger.Debug("Creating thermal reader", "device", ctx.thermalDeviceFile)
	tr, err := argononefan.NewThermalReader(argononefan.WithThermalDeviceFile(ctx.thermalDeviceFile))
	if err != nil {
		return fmt.Errorf("creating thermal reader: %w", err)
	}

	d.logger.Debug("Connecting to fan", "options", ctx.fanOptions)
	fan, err := argononefan.Connect(ctx.fanOptions...)
	if err != nil {
		return fmt.Errorf("connecting to fan: %w", err)
	}

	// Set the fan speed to a safe 100% to start
	d.logger.Info("Setting initial fan speed to 100% as a safety measure", "reason", "we don't know the current CPU temperature yet")
	if err := fan.SetSpeed(100); err != nil {
		return fmt.Errorf("setting fan speed: %w", err)
	}

	// Ensure the fan speed is reset to 100% when the daemon exits
	defer func() {
		lastTemp, err := tr.Celsius()
		if err != nil {
			d.logger.Error("Reading temperatur", "error", fmt.Errorf("reading temperature: %w", err))
		}
		d.logger.Warn("Fan control is shutting down, setting fan to 100% speed as a safety measure", "temperature", fmt.Sprintf("%2.1f°C", lastTemp))
		fan.SetSpeed(100)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	tempC, done := d.readTemp(d.CheckInterval, tr)
	go d.control(fan, d.Thresholds, d.Hysteresis, tempC)
	<-sigs

	// Notify the temperature reading goroutine to stop
	// Since it will close the tempC channel, the control goroutine will also stop
	done <- true

	return nil
}

func (d *daemonCmd) readTemp(interval time.Duration, tr *argononefan.ThermalReader) (<-chan float32, chan<- bool) {
	ml := d.logger.Named("read")

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

func (d *daemonCmd) control(fan *argononefan.Fan, config *thresholds, hysteresis float32, tempC <-chan float32) {

	ml := d.logger.Named("control")

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
