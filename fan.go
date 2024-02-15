/* 
 *  Copyright 2024 Markus W Mahlberg
 *  
 *  fan.go is part of argononefan
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

	"gobot.io/x/gobot/platforms/raspi"
)

// DefaultFanAddress is the default address of the fan on the i2c bus
// of the ArgonOne case.
// This should never change, as it is a hardware constant.
// However it can be overridden using the WithAddress option.
const DefaultFanAddress = 0x1A

// FanOption is a function that configures a Fan instance.
type FanOption func(*Fan) error

// Fan is a type that represents the fan in the ArgonOne case.
type Fan struct {
	bus     int
	address int
}

// Connect opens a connection to the fan on bus 0 at address 0x1A.
// Those values can be overridden using the WithBus and WithAddress options.
func Connect(opts ...FanOption) (*Fan, error) {
	f := &Fan{
		bus:     0,
		address: DefaultFanAddress,
	}

	for _, opt := range opts {
		if err := opt(f); err != nil {
			return nil, fmt.Errorf("error creating fan: %w", err)
		}
	}

	return f, nil
}

// OnBus is an option that sets the bus the fan resides on.
func OnBus(bus int) FanOption {
	return func(f *Fan) error {
		f.bus = bus
		return nil
	}
}

// WithAddress is an option that sets the address of the fan on the i2c bus.
// This should never change, as it is a hardware constant.
// However it can be overridden using the this option.
// USE WITH CAUTION, as it can cause the fan to stop working if set to an incorrect value.
// You can use the i2cdetect command to find the correct address.
func WithAddress(address int) FanOption {
	return func(f *Fan) error {
		f.address = address
		return nil
	}
}

// SetSpeed sets the fan speed.
func (f *Fan) SetSpeed(speed int) error {

	if speed < 0 || speed > 100 {
		return fmt.Errorf("desired fan speed is out of range: %d", speed)
	}

	a := raspi.NewAdaptor()
	defer a.Finalize()

	conn, err := a.GetConnection(f.address, f.bus)
	if err != nil {
		return fmt.Errorf("can't connect to i2c bus: %w", err)
	}
	defer conn.Close()

	err = conn.WriteByte(byte(speed))
	if err != nil {
		return fmt.Errorf("can't write fan seed: %W", err)
	}

	return nil
}
