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

	_ = bus
	_ = ipAddr

	loop := func(api openzwave.API) {
		for {
		    select {
		    	case notification := <- api.Notifications():
			     log.Infof("notification received <- %v\n", notification);
			     api.FreeNotification(notification);
			case quitReceived := <- api.QuitSignal():
			     _ = quitReceived
			     return;
		    }
		}
	}

	os.Exit(openzwave.
		BuildAPI("/usr/local/etc/openzwave", "", "").
		AddIntOption("SaveLogLevel", LOG_LEVEL.NONE).
		AddIntOption("QueueLogLevel", LOG_LEVEL.NONE).
		AddIntOption("DumpTrigger", LOG_LEVEL.NONE).
		AddIntOption("PollInterval", 60).
		AddBoolOption("IntervalBetweenPolls", true).
		AddBoolOption("ValidateValueChanges", true).
		Run(loop));

}
