package aeon

import (
	"fmt"

	"github.com/bitly/go-simplejson"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/CC"
)

type illuminator struct {
	node  openzwave.Node
	light *devices.LightDevice
}

func IlluminatorFactory(bus *ninja.DriverBus, node openzwave.Node) (openzwave.Device, error) {

	device := &illuminator{node, nil}

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

	device.light, err = devices.CreateLightDevice(label, deviceBus)
	if err != nil {
		return nil, err
	}

	if err := device.light.EnableOnOffChannel(); err != nil {
		return nil, err
	}

	if err := device.light.EnableBrightnessChannel(); err != nil {
		return nil, err
	}

	device.light.ApplyLightState = func(state *devices.LightDeviceState) error {
		return nil
	}

	device.light.ApplyOnOff = func(state bool) error {
		level := uint8(0)
		if state {
			level = 255
		}
		if device.node.SetUint8Value(CC.SWITCH_MULTILEVEL, 1, 0, level) {
			return nil
		} else {
			return fmt.Errorf("Failed to change state")
		}
	}

	device.light.ApplyBrightness = func(state float64) error {
		return nil
	}

	return device, nil
}
