package aeon

import (
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-openzwave"
)

type multisensor struct {
	bus  *ninja.DriverBus
	node openzwave.Node
}

func MultiSensorFactory(bus *ninja.DriverBus, node openzwave.Node) openzwave.Device {
	return &multisensor{bus, node}
}

func (device *multisensor) Notify(api openzwave.API, event openzwave.Event) {
}
