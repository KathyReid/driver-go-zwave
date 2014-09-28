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

	zwaveDeviceFactory := func(api openzwave.API, node openzwave.Node) openzwave.Device {
		return GetLibrary().GetDeviceFactory(*node.GetProductId())(api, node, bus)

	}

	configurator := openzwave.
		BuildAPI("/usr/local/etc/openzwave", ".", "").
		SetLogger(log).
		SetDeviceFactory(zwaveDeviceFactory)

	if configurator.GetBoolOption("logging", false) {
		callback := func(api openzwave.API, notification openzwave.Notification) {
			api.Logger().Infof("%v\n", notification)
		}

		configurator.SetNotificationCallback(callback)
	}

	os.Exit(configurator.Run())

	_ = statusJob

}
