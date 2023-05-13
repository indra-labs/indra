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

// GossipIntro -
// todo: this should update a peer subkey on DHT for
//   introductions, as well as the link to the hidden service peer data
func GossipIntro(intro *Intro, sm *sess.Manager, c qu.C) {
	log.D.F("propagating hidden service intro for %s",
		intro.Key.ToBase32Abbreviated())
	done := qu.T()
	msg := splice.New(IntroLen)
	if fails(intro.Encode(msg)) {
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
	// We broadcast the received introduction to two other randomly selected
	// nodes, which guarantees the entire network will see the intro at least
	// once.
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
