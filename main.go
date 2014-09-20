package main

import (
       "fmt"
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

	loop := func(notifications chan openzwave.Notification, quit chan bool) {
		for {
		    select {
		    	case notification := <- notifications:
				_ = notification
			case quitReceived := <- quit:
			     _ = quitReceived; // TODO: something useful
			     fmt.Printf("TODO: quit received\n");
		    }
		}
	}

	os.Exit(openzwave.
		API("/usr/local/etc/openzwave", "", "").
		AddIntOption("SaveLogLevel", LOG_LEVEL.DETAIL).
		AddIntOption("QueueLogLevel", LOG_LEVEL.DEBUG).
		AddIntOption("DumpTrigger", LOG_LEVEL.ERROR).
		AddIntOption("PollInterval", 500).
		AddBoolOption("IntervalBetweenPolls", true).
		AddBoolOption("ValidateValueChanges", true).
		Run(loop));

}
