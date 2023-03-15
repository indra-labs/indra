package engine

import (
	"sync"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Introduction struct {
	*Intro
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
	slice.Bytes
}

type MyIntros map[pub.Bytes]*Introduction

type KnownIntros map[pub.Bytes]*Intro

// Introductions is a map of existing known hidden service keys and the
// routing header for requesting a new one on behalf of the client.
//
// After a header is retrieved, the relay sends back a request to the hidden
// service using the headers in this store with the provided public recv from the
// client which is then used to encrypt the provided header and prevent the
// introducing relay from also using the provided header.
type Introductions struct {
	sync.Mutex
	MyIntros
	KnownIntros
}

func NewIntroductions() *Introductions {
	return &Introductions{
		MyIntros:    make(MyIntros),
		KnownIntros: make(KnownIntros),
	}
}

func (in *Introductions) Find(key pub.Bytes) (header *Introduction) {
	in.Lock()
	var ok bool
	if header, ok = in.MyIntros[key]; ok {
	}
	in.Unlock()
	return
}

func (in *Introductions) FindUnsafe(key pub.Bytes) (header *Introduction) {
	var ok bool
	log.D.S(in.MyIntros)
	if header, ok = in.MyIntros[key]; ok {
	}
	return
}

func (in *Introductions) FindKnownIntro(key pub.Bytes) (intro *Intro) {
	in.Lock()
	var ok bool
	if intro, ok = in.KnownIntros[key]; ok {
	}
	in.Unlock()
	return
}

func (in *Introductions) FindKnownIntroUnsafe(key pub.Bytes) (intro *Intro) {
	var ok bool
	if intro, ok = in.KnownIntros[key]; ok {
	}
	return
}

func (in *Introductions) Delete(key pub.Bytes) (header *Introduction) {
	in.Lock()
	var ok bool
	if header, ok = in.MyIntros[key]; ok {
		delete(in.MyIntros, key)
	}
	in.Unlock()
	return
}

func (in *Introductions) DeleteKnownIntro(key pub.Bytes) (
	header *Introduction) {
	in.Lock()
	var ok bool
	if _, ok = in.KnownIntros[key]; ok {
		delete(in.KnownIntros, key)
	}
	in.Unlock()
	return
}

func (in *Introductions) AddIntro(pk *pub.Key, intro *Introduction) {
	in.Lock()
	var ok bool
	key := pk.ToBytes()
	if _, ok = in.MyIntros[key]; ok {
		log.D.Ln("entry already exists for recv %x", key)
	} else {
		in.MyIntros[key] = intro
	}
	in.Unlock()
}

func GossipIntro(intro *Intro, sm *SessionManager, c qu.C) {
	log.D.F("propagating hidden service intro for %x", intro.Key.ToBytes())
	done := qu.T()
	msg := octet.New(IntroLen)
	if check(intro.Encode(msg)) {
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
		n.Transport.Send(msg.GetRange(-1, -1))
		cursor++
		if cursor > len(peerIndices)-1 {
			break
		}
	}
	log.T.Ln("finished broadcasting intro")
}
