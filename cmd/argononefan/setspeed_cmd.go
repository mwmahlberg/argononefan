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

	"github.com/mwmahlberg/argononefan"
)

type setSpeedCmd struct {
	Speed int `arg:"" help:"Fan speed" required:"" min:"0" max:"100"`
}

func (c *setSpeedCmd) Run(ctx *context) error {
	ctx.logger.Debug("Connecting to fan", "options", ctx.fanOptions)
	fan, err := argononefan.Connect(ctx.fanOptions...)
	if err != nil {
		ctx.logger.Error("Error connecting to fan", "error", err)
		return fmt.Errorf("Error connecting to fan: %w", err)
	}
	return fan.SetSpeed(c.Speed)
}
