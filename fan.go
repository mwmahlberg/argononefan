package argononefan

import (
	"fmt"

	"gobot.io/x/gobot/platforms/raspi"
)

// Fan address on i2c bus
const fanAddress = 0x1A

// SetFanSpeed sets the fan speed
func SetFanSpeed(bus, speed int) error {

	if speed < 0 || speed > 100 {
		return fmt.Errorf("desired fan speed is out of range : %d", speed)
	}

	adapter := raspi.NewAdaptor()
	defer adapter.Finalize()

	conn, err := adapter.GetConnection(fanAddress, bus)
	if err != nil {
		return fmt.Errorf("can't connect to i2c bus : %w", err)
	}
	defer conn.Close()

	err = conn.WriteByte(byte(speed))
	if err != nil {
		return fmt.Errorf("can't write fan seed : %W", err)
	}

	return nil
}
