package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/samonzeweb/argononefan"
)

// Scan temperature (and adust fan speed) with the given internval
const adjustInterval = 5 * time.Second

// The fan speed is maintained for at least X intervals
// ie if interval is 5 seconds, and interval count is equal to 3, then
// the fan will not slow down for at least 15 secondes (5 * 3).
// This will not prevent the fan to speed up.
const maintainSpeedInIntervalCount = 12

// Configuration file (in current directory)
const configurationFile = "adjustfan.json"

var (
	bus   int
	debug bool
	l     hclog.Logger
)

func init() {
	flag.IntVar(&bus, "bus", 0, "I2C bus the fan resides on")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
}

func main() {
	flag.Parse()

	var level hclog.Level = hclog.Info
	if debug {
		level = hclog.Debug
	}
	l = hclog.New(&hclog.LoggerOptions{
		Name:  "argononefan",
		Level: level,
	})
	l.Info("Starting adjustfan", "bus", bus, "debug", debug)

	l.Debug("Reading configuration", "file", configurationFile)
	configuration, err := readConfiguration(configurationFile)
	if err != nil {
		dislayErrorAndExit(err)
	}

	l.Debug("Setting up signal handling")
	var stopsig = make(chan os.Signal, 1)
	signal.Notify(stopsig, syscall.SIGTERM|syscall.SIGINT)

	l.Debug("Starting goroutine reading temperature")
	tempC, done := readTemp(adjustInterval)
	wg := sync.WaitGroup{}
	wg.Add(1)
	l.Debug("Starting adjust goroutine")
	go adjust(bus, configuration, tempC, &wg)

	l.Debug("Waiting for stop signal")
	<-stopsig
	l.Debug("Stop signal received")
	l.Debug("Closing temperature reading goroutine")
	done <- true
	l.Debug("Waiting for adjust goroutine to finish")
	wg.Wait()
	// adjustFanLoop(bus, configuration, stopsig)
	// Ensure the fan is reset to 100% speed when the program ends
	argononefan.SetFanSpeed(bus, 100)
}

func readTemp(interval time.Duration) (<-chan float32, chan<- bool) {
	c := make(chan float32)
	done := make(chan bool)

	go func() {
		tick := time.NewTicker(interval)
		l.Debug("Start reading temperature", "interval", interval)
		for {
			select {
			case <-done:
				l.Debug("Stop reading temperature")
				l.Debug("Closing temperature channel")
				close(c)
				l.Debug("Exiting temperature reading goroutine")
				return

			case <-tick.C:
				cpuTemparature, err := argononefan.ReadCPUTemperature()
				if err != nil {
					dislayErrorAndExit(err)
				}
				l.Debug("Reading temperature", "temperature", cpuTemparature)
				l.Debug("Sending temperature to adjust goroutine")
				c <- cpuTemparature
			}
		}
	}()
	return c, done
}

func adjust(bus int, configuration *Configuration, tempC <-chan float32, wg *sync.WaitGroup) {
	defer wg.Done()
	for currentTemperature := range tempC {
		l.Debug("Received temperature from reading goroutine", "temperature", currentTemperature)
		idx := slices.IndexFunc(configuration.Thresholds, func(t Threshold) bool {
			// This requires the thresholds to be sorted from higher to lower
			return currentTemperature >= t.Temperature
		})
		switch idx {
		case -1:
			l.Info("Temperature is lower than the lowest threshold, set fan to 0% speed")
			argononefan.SetFanSpeed(bus, 0)
		default:
			l.Debug("Found threshold", "index", idx, "threshold", configuration.Thresholds[idx], "fanSpeed", configuration.Thresholds[idx].FanSpeed)
			argononefan.SetFanSpeed(bus, configuration.Thresholds[idx].FanSpeed)
		}
	}
	l.Debug("Temperature channel is closed, set fan to 100% speed as a safety measure")
	// Channel is closed, set fan to 100% speed as a safety measure
	argononefan.SetFanSpeed(bus, 100)
}

func dislayErrorAndExit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
