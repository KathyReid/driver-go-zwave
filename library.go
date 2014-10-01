package main

import (
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/MF"

	"github.com/ninjasphere/driver-go-zwave/devices/aeon"
	"github.com/ninjasphere/driver-go-zwave/spi"
)

type NinjaDeviceFactory func(spi spi.ZWaveDriver, node openzwave.Node) openzwave.Device
type libraryT map[openzwave.ProductId]NinjaDeviceFactory

var (
	library libraryT = make(map[openzwave.ProductId]NinjaDeviceFactory)

	AEON_MULTISENSOR = openzwave.ProductId{MF.AEON_LABS, "0005"}
	AEON_ILLUMINATOR = openzwave.ProductId{MF.AEON_LABS, "0008"}
)

type Library interface {
	GetDeviceFactory(id openzwave.ProductId) NinjaDeviceFactory
}

func GetLibrary() Library {
	if len(library) == 0 {
		library[AEON_MULTISENSOR] = aeon.MultiSensorFactory
		library[AEON_ILLUMINATOR] = aeon.IlluminatorFactory
	}
	return &library
}

type unsupportedDevice struct {
}

func (lib *libraryT) GetDeviceFactory(id openzwave.ProductId) NinjaDeviceFactory {
	factory, ok := (*lib)[id]
	if ok {
		return factory
	} else {
		return func(spi spi.ZWaveDriver, node openzwave.Node) openzwave.Device {
			return &unsupportedDevice{}
		}
	}
}

func (*unsupportedDevice) NodeAdded() {
}

func (*unsupportedDevice) NodeChanged() {
}

func (*unsupportedDevice) ValueChanged(openzwave.Value) {
}

func (*unsupportedDevice) NodeRemoved() {
}
