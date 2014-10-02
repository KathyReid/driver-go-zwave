package aeon

import (
	"fmt"
	"time"

	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/CC"

	"github.com/ninjasphere/go-ninja/channels"

	"github.com/ninjasphere/driver-go-zwave/spi"
	"github.com/ninjasphere/driver-go-zwave/utils"
)

const (
	maxDeviceBrightness = 100             // by experiment, a level of 100 does not work for this device
	maxDelay            = time.Second * 5 // maximum delay for apply calls
)

var (
	level_switch = openzwave.ValueID{CC.SWITCH_MULTILEVEL, 1, 0}
)

type illuminator struct {
	spi.Device

	onOffChannel      *channels.OnOffChannel
	brightnessChannel *channels.BrightnessChannel

	// brightness is a cache of the current brightness when the device is switched off.
	// It is updated from the device on a confirmed attempt to adjust the level to a non-zero value
	brightness uint8

	refresh chan struct{} // used to wait for confirmation of updates after a level change

	emitter utils.Emitter
}

func IlluminatorFactory(driver spi.Driver, node openzwave.Node) openzwave.Device {
	device := &illuminator{}

	device.Init(driver, node)

	var ok bool

	device.brightness, ok = device.Node.GetValueWithId(level_switch).GetUint8()
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

	(*device.Info.Signatures)["ninja:thingType"] = "light"

	device.refresh = make(chan struct{}, 0)

	device.emitter = utils.Filter(
		func(level utils.Equatable) {
			device.unconditionalSendLightState(level.(*utils.WrappedUint8).Unwrap())
		},
		30*time.Second)

	return device
}

// ZWave protocols

func (device *illuminator) NodeAdded() {

	node := device.Node
	api := device.Driver.ZWave()
	conn := device.Driver.Connection()

	err := conn.ExportDevice(device)
	if err != nil {
		api.Logger().Infof("failed to export node: %v as device: %s", node, err)
		return
	}

	device.onOffChannel = channels.NewOnOffChannel(device)
	err = conn.ExportChannel(device, device.onOffChannel, "on-off")
	if err != nil {
		api.Logger().Infof("failed to export on-off channel for %v: %s", node, err)
		return
	}

	device.brightnessChannel = channels.NewBrightnessChannel(device)
	err = conn.ExportChannel(device, device.brightnessChannel, "brightness")
	if err != nil {
		api.Logger().Infof("failed to export brightness channel for %v: %s", node, err)
	}

	device.Node.GetValueWithId(level_switch).SetPollingState(true)

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

// Ninja protocols

func (device *illuminator) SetOnOff(state bool) error {
	level := uint8(0)
	if state {
		level = device.brightness
	}
	return device.setDeviceLevel(level)
}

func (device *illuminator) ToggleOnOff() error {
	level, ok := device.Node.GetValueWithId(level_switch).GetUint8()
	if !ok {
		return fmt.Errorf("Unable to determine current state of switch")
	}
	if level == 0 {
		return device.setDeviceLevel(device.brightness)
	} else {
		return device.setDeviceLevel(0)
	}
}

func (device *illuminator) SetBrightness(state float64) error {

	var err error = nil
	if state < 0 {
		state = 0
	} else if state > 1.0 {
		state = 1.0
	}
	level, ok := device.Node.GetValueWithId(level_switch).GetUint8()
	if ok {
		newLevel := uint8(state * maxDeviceBrightness)
		if level > 0 {
			err = device.setDeviceLevel(newLevel)
		} else {
			device.brightness = newLevel // to be applied when device is switched on
			device.emitter.Reset()
		}
	} else {
		err = fmt.Errorf("Unable to apply brightness - get failed.")
	}
	return err
}

//
// Issue a set against the OpenZWave API, then wait until the refreshed
// value matches the requested level or until a timeout, issuing refreshes
// as required.
//
func (device *illuminator) setDeviceLevel(level uint8) error {

	val := device.Node.GetValueWithId(level_switch)

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
				device.emitter.Reset()
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
	level, ok := device.Node.GetValueWithId(level_switch).GetUint8()
	if ok {
		//
		// Emit the current state, but filter out levels that don't change
		// within a specified period.
		//
		device.emitter.Emit(utils.WrapUint8(level))
	}
}

func (device *illuminator) unconditionalSendLightState(level uint8) {
	if device.brightness == maxDeviceBrightness-1 {
		device.brightness = 100
	}

	onOff := level != 0
	brightness := float64(device.brightness) / maxDeviceBrightness

	device.onOffChannel.SendState(onOff)
	device.brightnessChannel.SendState(brightness)
}
