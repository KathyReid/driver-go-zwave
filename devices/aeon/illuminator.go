package aeon

import (
	"fmt"
	"time"

	"github.com/bitly/go-simplejson"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/CC"
)

const (
	maxDeviceBrightness = 100             // by experiment, a level of 100 does not work for this device
	maxDelay            = time.Second * 5 // maximum delay for apply calls
)

type illuminator struct {
	api        openzwave.API
	node       openzwave.Node
	bus        *ninja.DriverBus
	light      *devices.LightDevice
	brightness uint8 // brightness is a cache of the current brightness when the device is switched off.
	// It is updated from the device on a confirmed attempt to adjust the level to a non-zero value
	refresh chan struct{} // used to wait for confirmation of updates after a level change
}

func IlluminatorFactory(api openzwave.API, node openzwave.Node, bus *ninja.DriverBus) openzwave.Device {
	device := &illuminator{api, node, bus, nil, 0, make(chan struct{}, 0)}
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

		level := uint8(0)
		if state {
			level = device.brightness
		}
		return device.setDeviceLevel(level)
	}

	device.light.ApplyBrightness = func(state float64) error {

		var err error = nil
		if state < 0 {
			state = 0
		} else if state > 1.0 {
			state = 1.0
		}
		level, ok := device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).GetUint8()
		if ok {
			newLevel := uint8(state * maxDeviceBrightness)
			if level > 0 {
				err = device.setDeviceLevel(newLevel)
			} else {
				device.brightness = newLevel // to be applied when device is switched on
			}
		} else {
			err = fmt.Errorf("Unable to apply brightness - get failed.")
		}
		return err
	}

	device.light.ApplyLightState = func(state *devices.LightDeviceState) error {

		// TODO: synchronize with notification thread

		err1 := device.light.ApplyBrightness(*state.Brightness)
		err2 := device.light.ApplyOnOff(*state.OnOff)

		if err2 != nil {
			return err2
		} else {
			return err1
		}
	}
}

//
// Issue a set against the OpenZWave API, then wait until the refreshed
// value matches the requested level or until a timeout, issuing refreshes
// as required.
//
func (device *illuminator) setDeviceLevel(level uint8) error {

	val := device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0)

	if level >= maxDeviceBrightness {
		// aeon will reject attempts to set the level to exactly 100
		level = maxDeviceBrightness - 1
	}

	if !val.SetUint8(level) {
		return fmt.Errorf("Failed to set level to %d - set failed", level)
	}
	timer := time.NewTimer(maxDelay)

	// loop until timeout or until refresh yields expected level

	for {
		if !val.Refresh() {
			return fmt.Errorf("Failed to set required level to %d - refresh failed", level)
		}
		select {
		case timeout := <-timer.C:
			_ = timeout
			if level != 0 {
				device.brightness = level
			}
			return fmt.Errorf("Failed to set required level to %d - timeout", level)
		case refreshed := <-device.refresh:
			_ = refreshed
			current, ok := val.GetUint8()
			if ok && current == level {
				if level != 0 {
					device.brightness = level
				}
				return nil
			}
		}
	}
}

//
// This call is used to reflect notifications about the current
// state of the light back to towards the ninja network
//
func (device *illuminator) sendLightState() {
	state := &devices.LightDeviceState{}
	level, ok := device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).GetUint8()
	if ok {
		if device.brightness == maxDeviceBrightness-1 {
			device.brightness = 100
		}

		onOff := level != 0
		brightness := float64(device.brightness) / maxDeviceBrightness

		state.OnOff = &onOff
		state.Brightness = &brightness

		device.light.SetLightState(state)
	}
}

func (device *illuminator) NodeChanged() {
	select {
	case device.refresh <- struct{}{}:
	default:
		device.sendLightState()
	}
}

func (device *illuminator) NodeRemoved() {
}

func (device *illuminator) ValueChanged(v openzwave.Value) {
	select {
	case device.refresh <- struct{}{}:
	default:
		device.sendLightState()
	}
}
