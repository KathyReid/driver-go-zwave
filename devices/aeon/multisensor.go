package aeon

import (
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-openzwave"
)

type multisensor struct {
	api  openzwave.API
	node openzwave.Node
	bus  *ninja.DriverBus
}

func MultiSensorFactory(api openzwave.API, node openzwave.Node, bus *ninja.DriverBus) openzwave.Device {
	return &multisensor{api, node, bus}
}

func (device *multisensor) NodeAdded() {
}

func (device *multisensor) NodeChanged() {
}

func (device *multisensor) NodeRemoved() {
}

func (device *multisensor) ValueChanged(openzwave.Value) {
}
