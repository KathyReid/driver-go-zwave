package utils

import (
	"time"
)

type Emitter interface {
	Emit(next Equatable)
	Reset()
}

type filteredEmitter struct {
	last     Equatable
	lastTime time.Time
	filter   func(Equatable)
}

//
// Creates a new, filtered emitter, such that the wrapped emitter is called at
// most once per minPeriod if the emitted value does not change.
//
func Filter(emitter func(next Equatable), minPeriod time.Duration) Emitter {
	var self *filteredEmitter

	self = &filteredEmitter{
		last:     nil,
		lastTime: time.Now(),
		filter: func(next Equatable) {
			now := time.Now()
			if self.last != nil &&
				self.last.Equals(next) &&
				now.Sub(self.lastTime) < minPeriod {
				return
			} else {
				self.last = next
				self.lastTime = now
				emitter(next)
			}
		},
	}

	return self

}

func (self *filteredEmitter) Emit(next Equatable) {
	self.filter(next)
}

func (self *filteredEmitter) Reset() {
	self.last = nil
}
