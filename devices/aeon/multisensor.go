package aeon

import (
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-openzwave"
)

func MultiSensorFactory(bus *ninja.DriverBus, node openzwave.Node) (openzwave.Device, error) {
	return &struct{}{}, nil
}
