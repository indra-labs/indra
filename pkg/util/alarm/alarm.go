package alarm

import (
	"time"
	
	"github.com/cybriq/qu"
)

func WakeAtTime(t time.Time, fn func(), quit qu.C) (cancel qu.C) {
	now := time.Now()
	if now.After(t) {
		return
	}
	until := t.Sub(now)
	cancel = qu.T()
	go func() {
		select {
		case <-cancel.Wait():
		case <-quit.Wait():
		case <-time.After(until):
			fn()
		}
	}()
	return
}
