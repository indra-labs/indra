package relay

import (
	"sync"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
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
