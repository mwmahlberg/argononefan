/*
 *  Copyright 2024 Markus W Mahlberg
 *
 *  daemon_metrics.go is part of argononefan
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	readings = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "temperature_readings_total",
		Help:      "The total number of temperature readings performed by argononefan in daemon mode",
		Subsystem: "argonone",
	})
	
	readingsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "temperature_readings_failed_total",
		Help:      "The total number of failed temperature readings performed by argononefan in daemon mode",
		Subsystem: "argonone",
	})

	temperatureK = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "temperature",
		Help:      "The current CPU temperature in degrees Kelvin",
		Subsystem: "argonone",
	})

	fanSpeed = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "fan_speed",
		Help:      "The current fan speed in percent",
		Subsystem: "argonone",
	})
	fanSpeedSet = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "speed_set_total",
		Help:      "The total number of fan speed changes performed by argononefan in daemon mode",
		Subsystem: "argonone",
	})
	fanSpeedSetFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "fan_speed_set_failed_total",
		Help:      "The total number of failed fan speed changes performed by argononefan in daemon mode",
		Subsystem: "argonone",
	})
)
