/* 
 *  Copyright 2024 Markus W Mahlberg
 *  
 *  temperature.go is part of argononefan
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

package argononefan

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	// DefaultThermalDeviceFile is the path in sysfs containing current CPU temperature
	DefaultThermalDeviceFile = "/sys/class/thermal/thermal_zone0/temp"
	// The temperature multiplier
	multiplier = float32(1000)
)

// ThermalReaderOption is a function that configures a ThermalReader instance.
type ThermalReaderOption func(*ThermalReader) error

// ThermalReader is a type that represents a thermal reader to read the CPU temperature.
type ThermalReader struct {
	filepath string
}

// NewThermalReader creates a new ThermalReader instance.
// The default device file is /sys/class/thermal/thermal_zone0/temp.
// This can be overridden using the WithThermalDeviceFile option.
func NewThermalReader(opts ...ThermalReaderOption) (*ThermalReader, error) {
	tr := &ThermalReader{
		filepath: DefaultThermalDeviceFile,
	}

	for _, opt := range opts {
		if err := opt(tr); err != nil {
			return nil, fmt.Errorf("creating thermal reader: %w", err)
		}
	}

	return tr, nil
}

// WithThermalDeviceFile is an option that sets the file path of the devicefile in sysfs
// containing current CPU temperature.
// Returns an error if the file does not exist, is not accessible, or is a directory.
func WithThermalDeviceFile(filepath string) ThermalReaderOption {
	return func(tr *ThermalReader) error {
		tr.filepath = filepath
		info, err := os.Stat(tr.filepath)
		if os.IsNotExist(err) {
			return fmt.Errorf("file '%s' does not exist: %w", tr.filepath, err)
		} else if os.IsPermission(err) {
			return fmt.Errorf("file '%s' is not accessible: %w", tr.filepath, err)
		} else if info.IsDir() {
			return fmt.Errorf("file '%s' is a directory: %w", tr.filepath, err)
		}
		return nil
	}
}

// Celsius returns the current CPU temperature in Celsius.
func (tr *ThermalReader) Celsius() (float32, error) {
	in, err := os.OpenFile(tr.filepath, os.O_RDONLY, 0)
	if err != nil {
		return 0, fmt.Errorf("opening temperature file: %w", err)
	}
	defer in.Close()

	t, err := readCPUTemperature(in)
	if err != nil {
		return 0, fmt.Errorf("reading temperature: %w", err)
	}
	return float32(t) / multiplier, nil
}

// Fahrenheit returns the current CPU temperature in Fahrenheit.
func (tr *ThermalReader) Fahrenheit() (float32, error) {
	c, err := tr.Celsius()
	if err != nil {
		return 0, fmt.Errorf("obtaining temperature in Celsius: %w", err)
	}
	return (c * 9 / 5) + 32, nil
}

func readCPUTemperature(in io.Reader) (int, error) {
	b, err := io.ReadAll(in)
	if err != nil {
		return 0, fmt.Errorf("reading temperature: %w", err)
	}
	t, err := strconv.Atoi(string(b[:len(b)-1]))
	if err != nil {
		return 0, fmt.Errorf("parsing temperature: %w", err)
	}
	return t, nil
}
