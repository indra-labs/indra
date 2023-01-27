package alarm

import (
	"time"

	"github.com/cybriq/qu"
)

func WakeAtTime(t time.Time, fn func()) (cancel qu.C) {
	now := time.Now()
	if now.After(t) {
		return
	}
	until := t.Sub(now)
	cancel = qu.T()
	go func() {
		select {
		case <-cancel.Wait():
		case <-time.After(until):
			fn()
		}
	}()
	return
}
