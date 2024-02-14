package main

import (
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/go-hclog"
	"github.com/samonzeweb/argononefan"
)

var (
	l hclog.Logger
)

var cli struct {
	Bus           int             `short:"b" long:"bus" help:"I2C bus the fan resides on" default:"0"`
	Debug         bool            `short:"d" long:"debug" help:"Enable debug mode" default:"false"`
	ConfigFile    string          `short:"c" long:"config" help:"Configuration file" default:"./adjustfan.json" type:"existingfile"`
	Thresholds    map[float32]int `short:"t" long:"threshold" help:"Thresholds" type:"float32:int" default:"60=100;55=50;50=10"`
	CheckInterval time.Duration   `short:"i" long:"interval" help:"Check interval" default:"5s"`
}

func main() {
	ctx := kong.Parse(&cli)
	ctx.Stderr = os.Stdout

	var level hclog.Level = hclog.Info
	if cli.Debug {
		level = hclog.Debug
	}

	l = hclog.New(&hclog.LoggerOptions{
		Name:  "argononefan",
		Level: level,
	})
	l.Info("Starting adjustfan", "bus", cli.Bus, "debug", cli.Debug)

	l.Debug("Setting up signal handling")
	var stopsig = make(chan os.Signal, 1)
	signal.Notify(stopsig, syscall.SIGTERM|syscall.SIGINT)

	l.Debug("Starting goroutine reading temperature")
	tempC, done := readTemp(cli.CheckInterval)

	wg := sync.WaitGroup{}
	wg.Add(1)

	l.Debug("Starting adjust goroutine")
	go adjust(cli.Bus, cli.Thresholds, tempC, &wg)

	l.Debug("Waiting for stop signal")
	<-stopsig
	l.Debug("Stop signal received")
	l.Debug("Closing temperature reading goroutine")
	done <- true
	l.Debug("Waiting for adjust goroutine to finish")
	wg.Wait()
	// adjustFanLoop(bus, configuration, stopsig)
	// Ensure the fan is reset to 100% speed when the program ends
	argononefan.SetFanSpeed(cli.Bus, 100)
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
					l.Error("Error reading temperature", "error", err)
					continue
				}
				l.Debug("Reading temperature", "temperature", cpuTemparature)
				l.Debug("Sending temperature to adjust goroutine")
				c <- cpuTemparature
			}
		}
	}()
	return c, done
}

func adjust(bus int, config map[float32]int, tempC <-chan float32, wg *sync.WaitGroup) {
	thresholds := make([]float32, 0, len(config))
	for t := range config {
		thresholds = append(thresholds, t)
	}
	defer wg.Done()
	for currentTemperature := range tempC {
		l.Debug("Received temperature from reading goroutine", "temperature", currentTemperature)
		idx := slices.IndexFunc(thresholds, func(t float32) bool {
			// This requires the thresholds to be sorted from higher to lower
			return currentTemperature >= t
		})
		switch idx {
		case -1:
			l.Debug("Temperature is lower than the lowest threshold, set fan to 0% speed")
			argononefan.SetFanSpeed(bus, 0)
		default:
			l.Debug("Found threshold", "index", idx, "threshold", thresholds[idx], "fanSpeed", config[thresholds[idx]])
			argononefan.SetFanSpeed(bus, config[thresholds[idx]])
		}
	}
	l.Debug("Temperature channel is closed, set fan to 100% speed as a safety measure")
	// Channel is closed, set fan to 100% speed as a safety measure
	argononefan.SetFanSpeed(bus, 100)
}
