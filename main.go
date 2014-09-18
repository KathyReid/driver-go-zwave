package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-openzwave"
)

const driverName = "driver-zwave"

var log = logger.GetLogger(driverName)

func main() {

	log.Infof("Starting " + driverName)

	conn, err := ninja.Connect("com.ninjablocks.zwave")
	if err != nil {
		log.FatalError(err, "Could not connect to MQTT")
	}

	pwd, _ := os.Getwd()

	bus, err := conn.AnnounceDriver("com.ninjablocks.zwave", driverName, pwd)
	if err != nil {
		log.FatalError(err, "Could not get driver bus")
	}

	statusJob, err := ninja.CreateStatusJob(conn, driverName)

	if err != nil {
		log.FatalError(err, "Could not setup status job")
	}

	statusJob.Start()

	ipAddr, err := ninja.GetNetAddress()
	if err != nil {
		log.FatalError(err, "Could not get net address")
	}

	_ = bus
	_ = ipAddr

	var notifications chan *interface{} = make(chan *interface{})

	go func() {

		for {
			notification := <-notifications
			_ = notification
		}
	}()

	openzwave.
		NewAPI().
		CreateOptions("/usr/local/etc/openzwave", "").
		AddIntOption("SaveLogLevel", openzwave.LogLevel_Detail).
		AddIntOption("QueueLogLevel", openzwave.LogLevel_Debug).
		AddIntOption("DumpTrigger", openzwave.LogLevel_Error).
		AddIntOption("PollInterval", 500).
		AddBoolOption("IntervalBetweenPolls", true).
		AddBoolOption("ValidateValueChanges", true).
		LockOptions().
		AddWatcher(notifications).
		AddDriver("/dev/ttyUSB0")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

}
