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

const hystereisHelp = `
hysteresis is the value in °C the temperature must drop below a threshold
before the fan is slowed down to the according speed. This is to prevent the fan from constantly switching between two speeds.

Note that this only applies to the fan slowing down coming from a higher threshold, not when speeding up.
`
const thresholdsHelp = `thresholds is map of °C to fan speed in %`
