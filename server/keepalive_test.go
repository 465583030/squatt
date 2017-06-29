package server

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testTimer = 50 * time.Millisecond
	testDelta = 10 * time.Millisecond
)

func TestKeepAlive(t *testing.T) {
	Convey(`Given a watchdog`, t, func() {
		var called bool
		w := newWatchdog(testTimer, func() { called = true })
		Reset(func() {
			w.Kick()
			called = false
		})
		Convey(`When waiting for less than the timer`, func() {
			time.Sleep(testTimer - testDelta)
			Convey(`Then the timer should not have expired`, func() { So(called, ShouldBeFalse) })
			Convey(`When kicking the timer`, func() {
				w.Kick()
				Convey(`When waiting for more than the timer`, func() {
					time.Sleep(testTimer + testDelta)
					Convey(`Then the timer should have expired`, func() { So(called, ShouldBeTrue) })
				})
				Convey(`When waiting for less than the timer`, func() {
					time.Sleep(testTimer - testDelta)
					Convey(`Then the timer should not have expired`, func() { So(called, ShouldBeFalse) })
				})
			})
			Convey(`When stopping the timer`, func() {
				w.Stop()
				Convey(`When waiting for more than the timer`, func() {
					time.Sleep(testTimer + testDelta)
					Convey(`Then the timer should not have expired`, func() { So(called, ShouldBeFalse) })
				})
			})
		})
		Convey(`When waiting for more than the timer`, func() {
			time.Sleep(testTimer + testDelta)
			Convey(`Then the timer should have expired`, func() { So(called, ShouldBeTrue) })
		})
	})
}
