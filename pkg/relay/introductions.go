package relay

import (
	"sync"
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/messages/intro"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Intros map[pub.Bytes]slice.Bytes

type NotifiedIntroducers map[pub.Bytes][]nonce.ID

// Introductions is a map of existing known hidden service keys and the
// routing header for requesting a new one on behalf of the client.
//
// After a header is retrieved, the relay sends back a request to the hidden
// service using the headers in this store with the provided public key from the
// client which is then used to encrypt the provided header and prevent the
// introducing relay from also using the provided header.
type Introductions struct {
	sync.Mutex
	Intros
	NotifiedIntroducers
}

func NewIntroductions() *Introductions {
	return &Introductions{Intros: make(Intros),
		NotifiedIntroducers: make(NotifiedIntroducers)}
}

func (in *Introductions) Find(key pub.Bytes) (header slice.Bytes) {
	in.Lock()
	var ok bool
	if header, ok = in.Intros[key]; ok {
		// If found, the header is not to be used again.
		delete(in.Intros, key)
	}
	in.Unlock()
	return
}

func (in *Introductions) AddIntro(pk *pub.Key, header slice.Bytes) {
	in.Lock()
	var ok bool
	key := pk.ToBytes()
	if _, ok = in.Intros[key]; ok {
		log.D.Ln("entry already exists for key %x", key)
	} else {
		in.Intros[key] = header
		in.NotifiedIntroducers[key] = []nonce.ID{}
	}
	in.Unlock()
}

func (in *Introductions) AddNotified(nodeID nonce.ID, ident pub.Bytes) {
	in.Lock()
	var ok bool
	if _, ok = in.NotifiedIntroducers[ident]; ok {
		in.NotifiedIntroducers[ident] = append(in.NotifiedIntroducers[ident],
			nodeID)
	} else {
		in.NotifiedIntroducers[ident] = []nonce.ID{nodeID}
	}
	in.Unlock()
}

func (eng *Engine) SendIntro(id nonce.ID, target *Session, intr *intro.Layer,
	hook func(id nonce.ID, b slice.Bytes)) {
	
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := HiddenService(id, intr, se[len(se)-1], c, eng.KeySet)
	log.D.Ln("sending out intro onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}

func (eng *Engine) intro(intr *intro.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	if intr.Validate() {
		log.D.F("sending out intro to %s at %s to all known peers",
			intr.Key.ToBase32(), intr.AddrPort.String())
	}
}

func (eng *Engine) introductionBroadcaster(intr *intro.Layer) {
	log.D.F("propagating hidden service introduction for %x", intr.Key.ToBytes())
	done := qu.T()
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
