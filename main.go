package main

import (
	"os"

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

	zwaveEvents := func(api openzwave.API, event openzwave.Event) {
		switch event.(type) {
		case *openzwave.NodeAvailable:
			log.Infof("device available %v", event.GetNode())
			newDevice, err := buildDevice(bus, event.GetNode())
			if err != nil {
				log.Infof("error while creating device for node %v: %v", event.GetNode(), err)
				//TODO: generate notification
			} else {
				log.Infof("device created for node %v", event.GetNode())
			}
			event.GetNode().SetDevice(newDevice)
			break
		case *openzwave.NodeChanged:
			existingDevice := event.GetNode().GetDevice()
			_ = existingDevice
			break
		case *openzwave.NodeUnavailable:
			// TODO
		}
	}

	_ = bus
	_ = ipAddr

	os.Exit(openzwave.
		BuildAPI("/usr/local/etc/openzwave", "", "").
		SetLogger(log).
		AddBoolOption("ValidateValueChanges", true).
		AddBoolOption("SaveConfiguration", false).
		AddBoolOption("logging", false).
		AddStringOption("LogFileName", "/dev/null", false).
		AddBoolOption("ConsoleOutput", false).
		AddBoolOption("NotifyTransactions", true).
		AddIntOption("SaveLogLevel", LOG_LEVEL.NONE).
		AddIntOption("QueueLogLevel", LOG_LEVEL.NONE).
		AddIntOption("DumpTrigger", LOG_LEVEL.NONE).
		AddIntOption("PollInterval", 30).
		AddBoolOption("IntervalBetweenPolls", false).
		AddBoolOption("ValidateValueChanges", true).
		SetEventsCallback(zwaveEvents).
		Run())

}
