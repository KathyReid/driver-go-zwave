// Provides abstractions used by the various ZWave/Ninja device adapters foudn in devices
package spi

import (
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-openzwave"
)

type Driver interface {
	ZWave() openzwave.API
	Ninja() ninja.Driver
	Connection() *ninja.Connection
}
