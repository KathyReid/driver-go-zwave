package aeon

import (
	"github.com/ninjasphere/go-openzwave"
)

func IlluminatorFactory(node openzwave.Node) openzwave.Device {
	return &struct{}{}
}

