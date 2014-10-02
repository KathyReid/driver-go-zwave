package aeon

import (
	"time"

	"github.com/ninjasphere/go-openzwave"
	"github.com/ninjasphere/go-openzwave/CC"

	"github.com/ninjasphere/go-ninja/channels"

	"github.com/ninjasphere/driver-go-zwave/spi"
	"github.com/ninjasphere/driver-go-zwave/utils"
)

var (
	motion_sensor      = openzwave.ValueID{CC.SENSOR_BINARY, 1, 0}
	temperature_sensor = openzwave.ValueID{CC.SENSOR_MULTILEVEL, 1, 1}
	illuminance_sensor = openzwave.ValueID{CC.SENSOR_MULTILEVEL, 1, 3}
	humidity_sensor    = openzwave.ValueID{CC.SENSOR_MULTILEVEL, 1, 5}
)

type multisensor struct {
	spi.Device
	motionChannel      *channels.MotionChannel
	motionSensor       utils.Emitter
	temperatureChannel *channels.TemperatureChannel
	illuminanceChannel *channels.IlluminanceChannel
	humidityChannel    *channels.HumidityChannel
}

func MultiSensorFactory(driver spi.Driver, node openzwave.Node) openzwave.Device {
	device := &multisensor{}

	device.Init(driver, node)

	(*device.Info.Signatures)["ninja:thingType"] = "sensor"

	device.motionSensor = utils.Filter(func(next utils.Equatable) {
		device.motionChannel.SendMotion()
	}, 1*time.Second)
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

	device.illuminanceChannel = channels.NewIlluminanceChannel(device)
	err = conn.ExportChannel(device, device.illuminanceChannel, "illuminance")
	if err != nil {
		api.Logger().Infof("failed to export illuminance channel for %v: %s", node, err)
		return
	} else {
		node.GetValueWithId(illuminance_sensor).SetPollingState(true)
	}

	device.temperatureChannel = channels.NewTemperatureChannel(device)
	err = conn.ExportChannel(device, device.temperatureChannel, "temperature")
	if err != nil {
		api.Logger().Infof("failed to export temperature channel for %v: %s", node, err)
		return
	} else {
		node.GetValueWithId(temperature_sensor).SetPollingState(true)
	}

	device.humidityChannel = channels.NewHumidityChannel(device)
	err = conn.ExportChannel(device, device.humidityChannel, "humidity")
	if err != nil {
		api.Logger().Infof("failed to export humidity channel for %v: %s", node, err)
		return
	} else {
		node.GetValueWithId(humidity_sensor).SetPollingState(true)
	}
}

func (device *multisensor) NodeChanged() {
}

func (device *multisensor) NodeRemoved() {

}

func (device *multisensor) ValueChanged(value openzwave.Value) {
	switch value.Id() {
	case motion_sensor: // motion
		flag, ok := value.GetBool()
		if ok && flag && device.motionChannel != nil {
			device.motionSensor.Emit(utils.WrapBool(true))
		}
	case temperature_sensor: // temperature
		valF, ok := value.GetFloat()
		if ok && device.temperatureChannel != nil {
			device.temperatureChannel.SendState(valF)
		}
	case illuminance_sensor: // luminance
		valF, ok := value.GetFloat()
		if ok && device.illuminanceChannel != nil {
			device.illuminanceChannel.SendState(valF)
		}
	case humidity_sensor: // relative humidity
		valF, ok := value.GetFloat()
		if ok {
			device.humidityChannel.SendState(valF)
		}
	}
}
