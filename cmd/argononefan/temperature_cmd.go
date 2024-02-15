/*
 *  Copyright 2024 Markus W Mahlberg
 *
 *  temperature_cmd.go is part of argononefan
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

	"github.com/hashicorp/go-hclog"
	"github.com/mwmahlberg/argononefan"
)

type context struct {
	ThermalDeviceFile string
	Imperial          bool
	logger            hclog.Logger
}

type temperatureCmd struct {
	Imperial bool `short:"i" long:"imperial" help:"Display temperature in imperial system" default:"false" env:"-"`
}

func (tc *temperatureCmd) Run(ctx *context) error {
	ctx.logger.Debug("Creating thermal reader", "device", ctx.ThermalDeviceFile)
	tr, err := argononefan.NewThermalReader(argononefan.WithThermalDeviceFile(ctx.ThermalDeviceFile))
	if err != nil {
		return fmt.Errorf("creating thermal reader: %w", err)
	}

	t, readErr := tr.Celsius()
	frmt := "Temperature: %2.1f°C\n"
	if ctx.Imperial {
		t, readErr = tr.Fahrenheit()
		frmt = "Temperature: %3.1f°F\n"
	}

	if readErr != nil {
		return fmt.Errorf("reading temperature: %w", readErr)
	}
	_, werr := fmt.Printf(frmt, t)

	return werr
}
