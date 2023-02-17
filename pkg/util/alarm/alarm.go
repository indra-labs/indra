package alarm

import (
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func WakeAtTime(t time.Time, fn func(), quit qu.C) (cancel qu.C) {
	log.D.Ln("setting alarm", t)
	now := time.Now()
	if now.After(t) {
		log.D.Ln("already passed")
		return
	}
	until := t.Sub(now)
	cancel = qu.T()
	go func() {
		select {
		case <-cancel.Wait():
			log.D.Ln("canceled")
		case <-quit.Wait():
			log.D.Ln("quitted")
		case <-time.After(until):
			log.D.Ln("timed out")
			fn()
		}
	}()
	return
}
