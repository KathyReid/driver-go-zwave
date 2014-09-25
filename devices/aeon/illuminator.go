package aeon

import (
	"fmt"

	"github.com/bitly/go-simplejson"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-openzwave"
)

func IlluminatorFactory(bus *ninja.DriverBus, node openzwave.Node) (openzwave.Device, error) {

	productId := node.GetProductId()
	productDescription := node.GetProductDescription()

	sigs := simplejson.New()
	sigs.Set("ninja:manufacturer", productDescription.ManufacturerName)
	sigs.Set("ninja:productName", productDescription.ProductName)
	sigs.Set("ninja:productType", productDescription.ProductType)
	sigs.Set("ninja:thingType", "light")

	address := fmt.Sprintf(
		"%08x:%03d:%s:%s",
		node.GetHomeId(),
		node.GetId(),
		productId.ManufacturerId,
		productId.ProductId)

	label := node.GetNodeName()

	deviceBus, err := bus.AnnounceDevice(address, "light", label, sigs)
	if err != nil {
		return nil, err
	}

	light, err := devices.CreateLightDevice(label, deviceBus)
	if err != nil {
		return nil, err
	}

	if err := light.EnableOnOffChannel(); err != nil {
		return nil, err
	}

	if err := light.EnableBrightnessChannel(); err != nil {
		return nil, err
	}

	light.ApplyLightState = func(state *devices.LightDeviceState) error {
		return nil
	}

	light.ApplyOnOff = func(state bool) error {
		return nil
	}

	light.ApplyBrightness = func(state float64) error {
		return nil
	}

	return &struct{}{}, nil
}
