package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/LOG_LEVEL"
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

	var notifications chan openzwave.Notification = make(chan openzwave.Notification)

	openzwave.
		NewAPI().
		StartOptions("/usr/local/etc/openzwave", "").
		AddIntOption("SaveLogLevel", LOG_LEVEL.DETAIL).
		AddIntOption("QueueLogLevel", LOG_LEVEL.DEBUG).
		AddIntOption("DumpTrigger", LOG_LEVEL.ERROR).
		AddIntOption("PollInterval", 500).
		AddBoolOption("IntervalBetweenPolls", true).
		AddBoolOption("ValidateValueChanges", true).
		EndOptions().
		CreateManager().
		SetNotificationChannel(notifications).
		StartDriver("")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		select {
		case notification := <-notifications:
			_ = notification
		case signal := <-signals:
			fmt.Println("Received signal: ", signal)
			os.Exit(1)
			break
		}
	}
}
