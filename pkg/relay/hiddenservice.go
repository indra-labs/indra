package relay

import "git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"

type Referrers map[pub.Bytes][]pub.Bytes

func (eng *Engine) hiddenserviceBroadcaster(hsk pub.Bytes) {
	log.D.Ln("propagating hidden service introduction for %x", hsk)
	for {
		select {
		case <-eng.C.Wait():
			return
		}
		
	}
}
