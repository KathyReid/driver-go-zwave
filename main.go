package main

import (
	"os"

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

	// TODO: what's this for???
	// statusJob.Start()

	zwaveEvents := func(api openzwave.API, event openzwave.Event) {
		switch event.(type) {
		case *openzwave.NodeAvailable:
			log.Infof("device available %v", event.GetNode())
			node := event.GetNode()
			newDevice, err := GetLibrary().GetDeviceFactory(*node.GetProductId())(bus, node)
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

	configurator := openzwave.
		BuildAPI("/usr/local/etc/openzwave", ".", "").
		SetLogger(log).
		SetEventsCallback(zwaveEvents)

	if configurator.GetBoolOption("logging", false) {
		callback := func(api openzwave.API, notification openzwave.Notification) {
			api.Logger().Infof("%v\n", notification)
		}

		configurator.SetNotificationCallback(callback)
	}

	os.Exit(configurator.Run())

	_ = statusJob

}
