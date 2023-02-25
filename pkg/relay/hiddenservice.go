package relay

func (eng *Engine) hiddenserviceBroadcaster() {
	log.D.Ln("propagating hidden service introduction")
	for {
		select {
		case <-eng.C.Wait():
			return
		}
	}
}
