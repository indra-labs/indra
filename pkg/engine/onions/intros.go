package onions

import (
	"git-indra.lan/indra-labs/indra/pkg/util/qu"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"

	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
)

type Introduction struct {
	*Intro
	ReplyHeader
}

func GossipAd(onion Onion, sm *sess.Manager, c qu.C) {
	switch on := onion.(type) {
	case *Intro:
		log.D.F("propagating hidden service intro for %s",
			on.Key.ToBase32Abbreviated())
		done := qu.T()
		msg := splice.New(on.Len())
		if fails(on.Encode(msg)) {
			return
		}
		nPeers := sm.NodesLen()
		peerIndices := make([]int, nPeers)
		for i := 1; i < nPeers; i++ {
			peerIndices[i] = i
		}
		cryptorand.Shuffle(nPeers, func(i, j int) {
			peerIndices[i], peerIndices[j] = peerIndices[j], peerIndices[i]
		})
		var cursor int
		for {
			select {
			case <-c.Wait():
				return
			case <-done:
				return
			default:
			}
			n := sm.FindNodeByIndex(peerIndices[cursor])
			n.Transport.Send(msg.GetAll())
			cursor++
			if cursor > len(peerIndices)-1 {
				break
			}
		}
		log.T.Ln("finished broadcasting intro")

	}
}
