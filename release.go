// +build release

package main

import (
	"github.com/bugsnag/bugsnag-go"
	"github.com/juju/loggo"
	"github.com/ninjasphere/go-ninja/logger"
)

func init() {
	logger.GetLogger("").SetLogLevel(loggo.INFO)

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       "7b6271abd03c178b20942eaff4dfa2e2",
		ReleaseStage: "production",
	})
}
