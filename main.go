package main

import (
	"flag"
	"os"
)

func main() {

	var debug bool

	flagset := flag.NewFlagSet("driver-go-zwave", flag.ContinueOnError)
	flagset.BoolVar(&debug, "debug", false, "Enable debugging")
	flagset.Parse(os.Args)

	zwaveDriver, err := newZWaveDriver(debug)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(zwaveDriver.wait())
}
