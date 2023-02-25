package relay

import (
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/messages/intro"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Referrers map[pub.Bytes][]pub.Bytes

func (eng *Engine) hiddenserviceBroadcaster(hsk *pub.Key) {
	log.D.F("propagating hidden service introduction for %x", hsk.ToBytes())
	done := qu.T()
	me := eng.GetLocalNodeAddress()
	intr := &intro.Layer{
		Key: hsk, AddrPort: me,
	}
	msg := make(slice.Bytes, intro.Len)
	c := slice.NewCursor()
	intr.Encode(msg, c)
	nPeers := eng.NodesLen()
	peerIndices := make([]int, nPeers)
	for i := 0; i < nPeers; i++ {
		peerIndices[i] = i
	}
	cryptorand.Shuffle(nPeers, func(i, j int) {
		peerIndices[i], peerIndices[j] = peerIndices[j], peerIndices[i]
	})
	// Since relays will also gossip this information, we will start a ticker
	// that sends out the hidden service introduction once a second until it
	// runs out of known relays to gossip to.
	ticker := time.NewTicker(time.Second)
	var cursor int
	for {
		select {
		case <-eng.C.Wait():
			return
		case <-done:
			return
		case <-ticker.C:
			n := eng.FindNodeByIndex(peerIndices[cursor])
			n.Transport.Send(msg)
			cursor++
		}
	}
}
