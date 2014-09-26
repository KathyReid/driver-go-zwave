package main

import (
	"fmt"

	"github.com/ninjasphere/driver-go-zwave/devices/aeon"
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/MF"
)

type DeviceFactory func(bus *ninja.DriverBus, node openzwave.Node) (openzwave.Device, error)
type libraryT map[openzwave.ProductId]DeviceFactory

var (
	library libraryT = make(map[openzwave.ProductId]DeviceFactory)

	AEON_MULTISENSOR = openzwave.ProductId{MF.AEON_LABS, "0005"}
	AEON_ILLUMINATOR = openzwave.ProductId{MF.AEON_LABS, "0008"}
)

type Library interface {
	GetDeviceFactory(id openzwave.ProductId) DeviceFactory
}

func GetLibrary() Library {
	if len(library) == 0 {
		library[AEON_MULTISENSOR] = aeon.MultiSensorFactory
		library[AEON_ILLUMINATOR] = aeon.IlluminatorFactory
	}
	return &library
}

func (lib *libraryT) GetDeviceFactory(id openzwave.ProductId) DeviceFactory {
	factory, ok := (*lib)[id]
	if ok {
		return factory
	} else {
		return func(bus *ninja.DriverBus, node openzwave.Node) (openzwave.Device, error) {
			return nil, error(fmt.Errorf("no support for product id %v", node))
		}
	}
}
