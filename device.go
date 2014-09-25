package main

import (
	"github.com/ninjasphere/driver-go-zwave/devices/aeon"
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/MF"
)

type zwaveDevice struct {
	node openzwave.Node
}

type DeviceFactory func(bus *ninja.DriverBus, node openzwave.Node) (openzwave.Device, error)

var (
	builders map[openzwave.ProductId]DeviceFactory

	AEON_MULTISENSOR = openzwave.ProductId{MF.AEON_LABS, "0005"}
	AEON_ILLUMINATOR = openzwave.ProductId{MF.AEON_LABS, "000d"}
)

func buildDevice(bus *ninja.DriverBus, node openzwave.Node) (openzwave.Device, error) {
	if len(builders) == 0 {
		builders = make(map[openzwave.ProductId]DeviceFactory)
		builders[AEON_MULTISENSOR] = aeon.MultiSensorFactory
		builders[AEON_ILLUMINATOR] = aeon.IlluminatorFactory
	}

	builder, ok := builders[*node.GetProductId()]
	if ok {
		return builder(bus, node)
	} else {
		return &struct{}{}, nil
	}

}
