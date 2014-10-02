package aeon

import (
	"time"
)

type filteredEmitter struct {
	last     *uint8
	lastTime time.Time
	filter   func(uint8)
}

//
// Creates a new, filtered emitter, such that the wrapped emitter is called at
// most once per minPeriod if the emitted value does not change.
//
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
