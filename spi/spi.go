// Provides abstractions used by the various ZWave/Ninja device adapters foudn in devices
package spi

import (
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-openzwave"
)

type ZWaveDriver interface {
	GetOpenZWaveAPI() openzwave.API
	GetNinjaDriver() ninja.Driver
	GetNinjaConnection() *ninja.Connection
}
