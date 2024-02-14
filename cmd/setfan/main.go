package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/samonzeweb/argononefan"
)

var bus int = 0

func init() {
	flag.IntVar(&bus, "bus", 0, "I2C bus the fan resides on")
}

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		displayUsageAndExit()
	}

	fanspeed, err := strconv.Atoi(flag.Arg(0))
	if err != nil || fanspeed < 0 || fanspeed > 100 {
		fmt.Printf("Invalid fanspeed \"%#v\":%s\n---\n", flag.Arg(0), err)
		displayUsageAndExit()
	}

	err = argononefan.SetFanSpeed(bus, fanspeed)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func displayUsageAndExit() {
	fmt.Fprintf(os.Stderr, "usage : %s fanspeed\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "with fanspeed between 0 and 100")
	os.Exit(1)
}
