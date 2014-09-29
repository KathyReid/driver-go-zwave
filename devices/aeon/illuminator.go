package aeon

import (
	"fmt"

	"github.com/bitly/go-simplejson"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/CC"
)

const (
	maxDeviceBrightness = 99 // by experiment, a level of 100 does not work for this device
)

type illuminator struct {
	api        openzwave.API
	node       openzwave.Node
	bus        *ninja.DriverBus
	light      *devices.LightDevice
	brightness uint8
}

func IlluminatorFactory(api openzwave.API, node openzwave.Node, bus *ninja.DriverBus) openzwave.Device {
	device := &illuminator{api, node, bus, nil, 0}
	return device
}

func (device *illuminator) NodeAdded() {
	var ok bool

	node := device.node
	api := device.api

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

	// initialize brightness from the current level

	device.brightness, ok = device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).GetUint8()
	if !ok || device.brightness == 0 {
		// we have to reset brightness to 100 since we apply brightness when
		// we switch it on

		//
		// one implication of this is that if the controller is removed
		// then replaced, while the light is off, the original brightness
		// will be lost
		//

		device.brightness = maxDeviceBrightness
	}

	deviceBus, err := device.bus.AnnounceDevice(address, "light", label, sigs)
	if err != nil {
		api.Logger().Infof("failed to announce device: %v", node)
		return
	}

	device.light, err = devices.CreateLightDevice(label, deviceBus)
	if err != nil {
		api.Logger().Infof("failed to create light device: %v", node)
		return
	}

	if err := device.light.EnableOnOffChannel(); err != nil {
		api.Logger().Infof("Failed to enable on/off channel: %v", node)
		return
	}

	if err := device.light.EnableBrightnessChannel(); err != nil {
		api.Logger().Infof("Failed to brightness channel: %v", node)
		return
	}

	device.light.ApplyOnOff = func(state bool) error {

		// TODO: synchronize with notification thread

		level := uint8(0)
		if state {
			level = device.brightness
		}
		if device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).SetUint8(level) {
			return nil
		} else {
			return fmt.Errorf("Failed to change on/off state")
		}
	}

	device.light.ApplyBrightness = func(state float64) error {

		// TODO: synchronize with notification thread

		var err error = nil
		if state < 0 {
			state = 0
		} else if state > 1.0 {
			state = 1.0
		}
		level, ok := device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).GetUint8()
		if ok {
			device.brightness = uint8(state * maxDeviceBrightness)
			if device.brightness < 1 {
				device.brightness = 1
			}
			if level > 0 {
				if !device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).SetUint8(device.brightness) {
					err = fmt.Errorf("Failed to change brightness")
				}
			}
		} else {
			err = fmt.Errorf("Failed to read existing level from device")
		}
		return err
	}

	device.light.ApplyLightState = func(state *devices.LightDeviceState) error {

		// TODO: synchronize with notification thread

		device.light.ApplyOnOff(*state.OnOff)
		device.light.ApplyBrightness(*state.Brightness)
		return nil
	}
}

func (device *illuminator) NodeChanged() {
	state := &devices.LightDeviceState{}
	level, ok := device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).GetUint8()
	if ok {
		if level > maxDeviceBrightness {
			level = maxDeviceBrightness
		}

		if level != 0 {
			device.brightness = level
		}

		onOff := level != 0
		brightness := float64(device.brightness) / maxDeviceBrightness

		state.OnOff = &onOff
		state.Brightness = &brightness

		device.light.SetLightState(state)
	}
}

func (device *illuminator) NodeRemoved() {
}

func (device *illuminator) ValueChanged(openzwave.Value) {
}
