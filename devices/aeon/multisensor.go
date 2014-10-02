package aeon

import (
	"github.com/ninjasphere/driver-go-zwave/spi"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/CC"
)

type multisensor struct {
	spi.Device
	motionChannel *channels.MotionChannel
}

func MultiSensorFactory(driver spi.Driver, node openzwave.Node) openzwave.Device {
	device := &multisensor{}

	device.Init(driver, node)

	(*device.Info.Signatures)["ninja:thingType"] = "sensor"
	return device
}

func (device *multisensor) NodeAdded() {
	node := device.Node
	api := device.Driver.ZWave()
	conn := device.Driver.Connection()

	err := conn.ExportDevice(device)
	if err != nil {
		api.Logger().Infof("failed to export node: %v as device: %s", node, err)
		return
	}

	device.motionChannel = channels.NewMotionChannel(device)
	err = conn.ExportChannel(device, device.motionChannel, "motion")
	if err != nil {
		api.Logger().Infof("failed to export motion channel for %v: %s", node, err)
		return
	}

}

func (device *multisensor) NodeChanged() {
}

func (device *multisensor) NodeRemoved() {
}

func (device *multisensor) ValueChanged(value openzwave.Value) {
	switch value.Id().CommandClassId {
	case CC.SENSOR_BINARY:
		flag, ok := value.GetBool()
		if ok && flag {
			device.motionChannel.SendMotion()
		}
	}
}
