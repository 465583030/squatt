package server

import (
	"errors"
	"time"
)

var errKeepAliveTimeout = errors.New("keep-alive timeout")

type watchdog struct {
	expire time.Duration
	*time.Timer
}

func newWatchdog(expire time.Duration, callback func()) *watchdog {
	return &watchdog{expire: expire, Timer: time.AfterFunc(expire, callback)}
}

func (w *watchdog) Kick() {
	if w == nil {
		return
	}
	if w.Stop() {
		w.Reset(w.expire)
	}
}

func (w *watchdog) Stop() bool {
	if w == nil {
		return false
	}
	return w.Timer.Stop()
}
