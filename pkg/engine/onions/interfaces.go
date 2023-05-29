package onions

import (
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/qu"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"

	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/responses"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
)

type Ngin interface {
	HandleMessage(s *splice.Splice, pr Onion)
	GetLoad() byte
	SetLoad(byte)
	Mgr() *sess.Manager
	Pending() *responses.Pending
	GetHidden() *Hidden
	KillSwitch() qu.C
	Keyset() *crypto.KeySet
}

// Onion are messages that can be layered over each other and have
// a set of processing instructions for the data in them, and, if relevant,
// how to account for them in sessions.
type Onion interface {
	coding.Codec
	Wrap(inner Onion)
	Handle(s *splice.Splice, p Onion, ni Ngin) (e error)
	Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
		last bool) (skip bool, sd *sessions.Data)
}
type PeerInfo interface {
	Onion
	Splice(s *splice.Splice)
	Validate() bool
	Gossip(sm *sess.Manager, c qu.C)
}

func Gossip(x PeerInfo, sm *sess.Manager, c qu.C) {
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
