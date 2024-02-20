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
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/mwmahlberg/argononefan"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type daemonCmd struct {
	Thresholds     *thresholds   `short:"t" long:"threshold" help:"${help_thresholds}" default:"70=100;60=50;55=10"`
	Hysteresis     float32       `long:"hysteresis" help:"${help_hysteresis}" default:"1.0"`
	CheckInterval  time.Duration `short:"i" long:"interval" help:"Check interval" default:"5s"`
	logger         hclog.Logger  `kong:"-"`
	PrometheusBind string        `long:"promehteus-bind" help:"Address to bind the Prometheus metrics server to" default:"localhost:8080"`
}

func (d *daemonCmd) Run(
	logger hclog.Logger,
	readerOptions []argononefan.ThermalReaderOption,
	fanOptions []argononefan.FanOption,
) error {

	d.logger = logger

	d.logger.Info("Starting daemon", "thresholds", d.Thresholds.thresholds, "hysteresis", d.Hysteresis, "interval", d.CheckInterval)

	d.logger.Info("Starting Prometheus metrics server", "address", d.PrometheusBind)

	http.Handle("/metrics", promhttp.Handler())
	srv := http.Server{
		Addr: d.PrometheusBind,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			d.logger.Error("Starting Prometheus metrics server", "error", err)
		} else if err == http.ErrServerClosed {
			d.logger.Info("Prometheus metrics server stopped")
		}
	}()

	d.logger.Debug("Creating thermal reader")
	tr, err := argononefan.NewThermalReader(readerOptions...)
	if err != nil {
		return fmt.Errorf("creating thermal reader: %w", err)
	}

	d.logger.Debug("Connecting to fan")
	fan, err := argononefan.Connect(fanOptions...)
	if err != nil {
		return fmt.Errorf("connecting to fan: %w", err)
	}

	// Set the fan speed to a safe 100% to start
	d.logger.Info("Setting initial fan speed to 100% as a safety measure", "reason", "we don't know the current CPU temperature yet")
	if err := fan.SetSpeed(100); err != nil {
		fanSpeedSetFailed.Inc()
		return fmt.Errorf("setting fan speed: %w", err)
	}
	fanSpeedSet.Inc()

	// Ensure the fan speed is reset to 100% when the daemon exits
	defer func() {
		lastTemp, err := tr.Celsius()
		if err != nil {
			d.logger.Error("Reading temperatur", "error", fmt.Errorf("reading temperature: %w", err))
		}
		d.logger.Warn("Fan control is shutting down, setting fan to 100% speed as a safety measure", "temperature", fmt.Sprintf("%2.1f°C", lastTemp))
		fan.SetSpeed(100)
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go d.control(signalCtx, fan, tr, d.Thresholds, d.Hysteresis)
	<-signalCtx.Done()
	srv.Shutdown(nil)

	return nil
}

func (d *daemonCmd) control(ctx context.Context, fan *argononefan.Fan, tr *argononefan.ThermalReader, config *thresholds, hysteresis float32) {

	var (
		currentSpeed       int = -1
		currentTemperature float32
		once               sync.Once
		tick               = time.NewTicker(5 * time.Second)
		errC               = make(chan error)
		err                error
	)

	for {
		select {
		case <-tick.C:
			if currentTemperature, err = tr.Celsius(); err != nil {
				errC <- fmt.Errorf("reading temperature: %w", err)
			}
			targetSpeed := config.GetSpeed(currentTemperature)

			if targetSpeed < currentSpeed {
				targetSpeed = config.GetSpeedWithHysteresis(currentTemperature, hysteresis)
			}

			switch targetSpeed {

			case currentSpeed:
				d.logger.Debug("Temperature is still within the same threshold, no need to adjust fan speed")

			default:
				d.logger.Debug("Found threshold", "threshold", config.GetThreshold(currentTemperature), "computed fanSpeed with hystersis", config.GetSpeedWithHysteresis(currentTemperature, hysteresis))

				currentSpeed = targetSpeed
				if err = fan.SetSpeed(targetSpeed); err != nil {
					d.logger.Error("Setting fan speed", "error", err)
					errC <- fmt.Errorf("setting fan speed: %w", err)
					fanSpeedSetFailed.Inc()
					continue
				}

				fanSpeed.Set(float64(targetSpeed))
				fanSpeedSet.Inc()

				once.Do(func() {
					d.logger.Info("Set initial fan speed based on readings", "temperature", currentTemperature, "speed", currentSpeed)
				})
			}

		case <-ctx.Done():
			d.logger.Debug("Received stop signal")
			d.logger.Debug("Exiting goroutine...")
			return
		}
	}
}
