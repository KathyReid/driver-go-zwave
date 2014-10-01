package aeon

import (
	"fmt"
	"time"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/CC"

	"github.com/ninjasphere/driver-go-zwave/spi"
)

const (
	maxDeviceBrightness = 100             // by experiment, a level of 100 does not work for this device
	maxDelay            = time.Second * 5 // maximum delay for apply calls
)

type illuminator struct {
	driver    spi.Driver
	info      *model.Device
	sendEvent func(event string, payload interface{}) error

	node openzwave.Node

	onOffChannel      *channels.OnOffChannel
	brightnessChannel *channels.BrightnessChannel

	// brightness is a cache of the current brightness when the device is switched off.
	// It is updated from the device on a confirmed attempt to adjust the level to a non-zero value
	brightness uint8

	refresh chan struct{} // used to wait for confirmation of updates after a level change

	emitter *filteredEmitter
}

func IlluminatorFactory(driver spi.Driver, node openzwave.Node) openzwave.Device {
	device := &illuminator{
		info:       &model.Device{},
		driver:     driver,
		node:       node,
		brightness: 0,
		refresh:    make(chan struct{}, 0),
	}
	device.emitter = newFilteredEmitter(
		func(level uint8) {
			device.unconditionalSendLightState(level)
		},
		30*time.Second)

	return device
}

func (device *illuminator) GetDriver() ninja.Driver {
	return device.driver.Ninja()
}

func (device *illuminator) GetDeviceInfo() *model.Device {
	return device.info
}

func (device *illuminator) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	device.sendEvent = sendEvent
}

func (device *illuminator) NodeAdded() {
	var ok bool

	api := device.driver.ZWave()
	node := device.node
	productId := node.GetProductId()
	productDescription := node.GetProductDescription()

	sigs := make(map[string]string)

	sigs["zwave:manufacturerId"] = productId.ManufacturerId
	sigs["zwave:productId"] = productId.ProductId
	sigs["zwave:manufacturerName"] = productDescription.ManufacturerName
	sigs["zwave:productName"] = productDescription.ProductName
	sigs["zwave:productType"] = productDescription.ProductType
	sigs["ninja:thingType"] = "light"

	device.info.Signatures = &sigs

	//
	// This naming scheme won't survive reconfigurations of the network
	// where the network has two devices of the same type.
	//
	// So, we will need to investigate generating a unique token that
	// gets stored in the device. When this scheme is implemented we
	// should update the naming scheme used here from v0 to v1.
	//
	device.info.NaturalIDType = "ninja.zwave.v0"
	device.info.NaturalID = fmt.Sprintf(
		"%08x:%03d:%s:%s",
		node.GetHomeId(),
		node.GetId(),
		productId.ManufacturerId,
		productId.ProductId)

	device.info.Name = &productDescription.ProductName

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

	conn := device.driver.Connection()
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

	device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).SetPollingState(true)
}

func (device *illuminator) SetOnOff(state bool) error {
	level := uint8(0)
	if state {
		level = device.brightness
	}
	return device.setDeviceLevel(level)
}

func (device *illuminator) ToggleOnOff() error {
	level, ok := device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).GetUint8()
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
	level, ok := device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).GetUint8()
	if ok {
		newLevel := uint8(state * maxDeviceBrightness)
		if level > 0 {
			err = device.setDeviceLevel(newLevel)
		} else {
			device.brightness = newLevel // to be applied when device is switched on
		}
		device.emitter.reset()
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
	level, ok := device.node.GetValue(CC.SWITCH_MULTILEVEL, 1, 0).GetUint8()
	if ok {
		//
		// Emit the current state, but filter out levels that don't change
		// within a specified period.
		//
		device.emitter.emit(level)
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

type filteredEmitter struct {
	last     *uint8
	lastTime time.Time
	filter   func(uint8)
}

func newFilteredEmitter(emitter func(next uint8), minPeriod time.Duration) *filteredEmitter {
	var self *filteredEmitter

	self = &filteredEmitter{
		last:     nil,
		lastTime: time.Now(),
		filter: func(next uint8) {
			now := time.Now()
			if self.last != nil &&
				*(self.last) == next &&
				now.Sub(self.lastTime) < minPeriod {
				return
			} else {
				self.last = &next
				self.lastTime = now
				emitter(next)
			}
		},
	}

	return self

}

func (self *filteredEmitter) emit(next uint8) {
	self.filter(next)
}

func (self *filteredEmitter) reset() {
	self.last = nil
}
