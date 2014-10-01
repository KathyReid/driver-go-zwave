package main

import (
	"os"

	"flag"
)

func main() {

	var debug bool

	flag.BoolVar(&debug, "debug", false, "Enable debugging")
	flag.Parse()

	zwaveDriver, err := newZWaveDriver(debug)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(zwaveDriver.wait())
}
