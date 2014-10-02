package spi

import (
	"fmt"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-openzwave"
)

type Device struct {
	Driver    Driver
	Info      *model.Device
	SendEvent func(event string, payload interface{}) error
	Node      openzwave.Node
}

func (device *Device) GetDriver() ninja.Driver {
	return device.Driver.Ninja()
}

func (device *Device) GetDeviceInfo() *model.Device {
	return device.Info
}

func (device *Device) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	device.SendEvent = sendEvent
}

func (device *Device) Init(driver Driver, node openzwave.Node) {
	device.Driver = driver
	device.Node = node

	productId := node.GetProductId()
	productDescription := node.GetProductDescription()

	sigs := make(map[string]string)

	sigs["zwave:manufacturerId"] = productId.ManufacturerId
	sigs["zwave:productId"] = productId.ProductId
	sigs["zwave:manufacturerName"] = productDescription.ManufacturerName
	sigs["zwave:productName"] = productDescription.ProductName
	sigs["zwave:productType"] = productDescription.ProductType

	device.Info.Signatures = &sigs

	//
	// This naming scheme won't survive reconfigurations of the network
	// where the network has two devices of the same type.
	//
	// So, we will need to investigate generating a unique token that
	// gets stored in the device. When this scheme is implemented we
	// should update the naming scheme used here from v0 to v1.
	//
	device.Info.NaturalIDType = "ninja.zwave.v0"
	device.Info.NaturalID = fmt.Sprintf(
		"%08x:%03d:%s:%s",
		node.GetHomeId(),
		node.GetId(),
		productId.ManufacturerId,
		productId.ProductId)

	device.Info.Name = &productDescription.ProductName

	// initialize brightness from the current level

}
