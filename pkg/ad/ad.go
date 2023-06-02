package ad

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/sess"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/splice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

// Ad is an interface for the signed messages stored in the PeerStore of the
// libp2p host inside an indra engine.
type Ad interface {
	coding.Codec
	Splice(s *splice.Splice)
	Validate() bool
	Gossip(sm *sess.Manager, c qu.C)
}

// Gossip writes a new Ad out to the p2p network.
//
// todo: this will be changed to use the engine host peer store. An interface
//  will be required.
func Gossip(x Ad, sm *sess.Manager, c qu.C) {
	done := qu.T()
	msg := splice.New(x.Len())
	if fails(x.Encode(msg)) {
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
}
