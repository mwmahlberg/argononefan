package main

import (
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"golang.org/x/exp/maps"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/go-hclog"
	"github.com/samonzeweb/argononefan"
)

const (
	readingTemperatureMsg = "readingTemperature"
)

var (
	l hclog.Logger
)

var cli struct {
	Debug      bool   `short:"d" long:"debug" help:"Enable debug mode" default:"false"`
	DeviceFile string `short:"f" long:"file" help:"File path in sysfs containing current CPU temperature" default:"/sys/class/thermal/thermal_zone0/temp"`

	Daemon struct {
		Bus           int             `short:"b" long:"bus" help:"I2C bus the fan resides on" default:"0"`
		Thresholds    map[float32]int `short:"t" long:"threshold" help:"Thresholds" type:"float32:int" default:"60=100;55=50;50=10"`
		CheckInterval time.Duration   `short:"i" long:"interval" help:"Check interval" default:"5s"`
	} `kong:"cmd,help='Run the fan control daemon'"`

	Temperature struct {
		Imperial bool `short:"i" long:"imperial" help:"Display temperature in imperial system" default:"false"`
	} `kong:"cmd,help='Read the current CPU temperature'"`
}

func main() {

	ctx := kong.Parse(&cli,
		kong.Name("argononefand"),
		kong.Description("Daemon to adjust the fan speed of the Argon One case"),
		kong.DefaultEnvars("ARGONONEFAND"),
	)
	ctx.Stderr = os.Stdout

	var level hclog.Level = hclog.Info
	if cli.Debug {
		level = hclog.Debug
	}

	l = hclog.New(&hclog.LoggerOptions{
		Name:  "argononefand",
		Level: level,
	})

	switch ctx.Command() {
	case "temperature":

		t, err := argononefan.ReadCPUTemperature(cli.DeviceFile)
		ctx.FatalIfErrorf(err, readingTemperatureMsg)

		if cli.Temperature.Imperial {
			ctx.Printf("Temperature: %2.1f°F", toFahrenheit(t))
		} else {
			ctx.Printf("Temperature: %2.1f°C", t)
		}

		os.Exit(0)
	}

	l.Debug("Running with configuration", "config", cli)

	l.Debug("Setting up signal handling")
	var stopsig = make(chan os.Signal, 1)
	signal.Notify(stopsig, syscall.SIGTERM, syscall.SIGINT)

	l.Debug("Starting goroutine reading temperature")
	tempC, done := readTemp(cli.Daemon.CheckInterval)

	l.Debug("Starting adjust goroutine")
	go adjust(cli.Daemon.Bus, cli.Daemon.Thresholds, tempC)

	l.Debug("Waiting for stop signal")
	<-stopsig
	l.Debug("Stop signal received")

	l.Debug("Closing temperature reading goroutine")
	done <- true
	l.Debug("Waiting for adjust goroutine to finish")

	lastTemp, err := argononefan.ReadCPUTemperature(cli.DeviceFile)
	ctx.FatalIfErrorf(err, readingTemperatureMsg)

	l.Warn("Fan control is shutting down, setting fan to 100% speed as a safety measure", "temperature", fmt.Sprintf("%2.1f°C", lastTemp))
	argononefan.SetFanSpeed(cli.Daemon.Bus, 100)
}

func readTemp(interval time.Duration) (<-chan float32, chan<- bool) {
	ml := l.Named("readTemp")

	c := make(chan float32)
	done := make(chan bool)
	go func() {

		tick := time.NewTicker(interval)

		ml.Debug("Start reading temperature", "interval", interval)
		for {
			select {

			case <-done:
				ml.Debug("Received stop signal")
				l.Debug("Closing temperature channel")
				close(c)
				l.Debug("Exiting...")
				return

			case <-tick.C:
				t, err := argononefan.ReadCPUTemperature(cli.DeviceFile)

				if err != nil {
					ml.Error(readingTemperatureMsg, "error", err)
					continue
				}
				ml.Debug("Read temperature", "temperature", t)
				ml.Debug("Sending temperature to adjust goroutine")
				c <- t
			}
		}
	}()
	return c, done
}

func adjust(bus int, config map[float32]int, tempC <-chan float32) {

	ml := l.Named("adjust")
	// Ensure we are looking at the thresholds in descending order
	thresholds := maps.Keys(config)
	slices.Sort(thresholds)
	slices.Reverse(thresholds)
	ml.Debug("Thresholds", "thresholds", thresholds)

	var currentIdx int

	for currentTemperature := range tempC {
		ml.Debug("Received temperature from reading goroutine", "temperature", currentTemperature)

		// Find the index of the threshold matching the current temperature
		idx := slices.IndexFunc(thresholds, func(t float32) bool {
			// This requires the thresholds to be sorted from higher to lower
			return currentTemperature >= t
		})

		switch idx {
		case currentIdx:
			ml.Debug("Temperature is still within the same threshold, no need to adjust fan speed")
		case -1:
			ml.Debug("Temperature is lower than the lowest threshold, set fan to 0% speed")
			currentIdx = -1
			argononefan.SetFanSpeed(bus, 0)
		default:
			ml.Debug("Found threshold", "index", idx, "threshold", thresholds[idx], "fanSpeed", config[thresholds[idx]])
			currentIdx = idx
			argononefan.SetFanSpeed(bus, config[thresholds[idx]])
		}
	}

}

func toFahrenheit(c float32) float32 {
	return c*9/5 + 32
}
