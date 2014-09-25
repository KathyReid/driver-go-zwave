package aeon

import (
	"github.com/ninjasphere/go-openzwave"
)

func MultiSensorFactory(node openzwave.Node) openzwave.Device {
	return &struct{}{}
}
