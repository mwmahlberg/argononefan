package argononefan

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DefaultThermalDeviceFile is the path in sysfs containing current CPU temperature
const DefaultThermalDeviceFile = "/sys/class/thermal/thermal_zone0/temp"

// The temperature multiplier
const multiplier = float32(1000)

// ReadCPUTemperature reads the current CPU temperature
func ReadCPUTemperature(filepath string) (float32, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return 0, fmt.Errorf("unable to read temperature file %s : %w", filepath, err)
	}

	stringTemperature := strings.TrimSuffix(string(content), "\n")
	rawTemperature, err := strconv.Atoi(stringTemperature)
	if err != nil {
		return 0, fmt.Errorf("unable to parse temperature %s : %w", content, err)
	}

	return (float32(rawTemperature) / multiplier), nil
}
