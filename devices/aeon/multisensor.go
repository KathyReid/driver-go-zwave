package aeon

import (
	"github.com/ninjasphere/go-openzwave"

	"github.com/ninjasphere/driver-go-zwave/spi"
)

type multisensor struct {
	driver spi.Driver
	node   openzwave.Node
}

func MultiSensorFactory(driver spi.Driver, node openzwave.Node) openzwave.Device {
	return &multisensor{driver, node}
}

func (device *multisensor) NodeAdded() {
}

func (device *multisensor) NodeChanged() {
}

func (device *multisensor) NodeRemoved() {
}

func (device *multisensor) ValueChanged(openzwave.Value) {
}
