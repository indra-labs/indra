package engine

import (
	"sync"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Intros map[pub.Bytes]slice.Bytes

type KnownIntros map[pub.Bytes]*Intro

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
	KnownIntros
}

func NewIntroductions() *Introductions {
	return &Introductions{Intros: make(Intros),
		KnownIntros: make(KnownIntros)}
}

func (in *Introductions) Find(key pub.Bytes) (header slice.Bytes) {
	in.Lock()
	var ok bool
	if header, ok = in.Intros[key]; ok {
	}
	in.Unlock()
	return
}

func (in *Introductions) FindUnsafe(key pub.Bytes) (header slice.Bytes) {
	var ok bool
	log.D.S(in.Intros)
	if header, ok = in.Intros[key]; ok {
	}
	return
}

func (in *Introductions) FindKnownIntro(key pub.Bytes) (intro *Intro) {
	in.Lock()
	log.D.S(in.KnownIntros)
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

func (in *Introductions) Delete(key pub.Bytes) (header slice.Bytes) {
	in.Lock()
	var ok bool
	if header, ok = in.Intros[key]; ok {
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
	}
	in.Unlock()
}

func SendIntro(id nonce.ID, target *SessionData,
	intro *Intro, sm *SessionManager, ks *signer.KeySet,
	p *PendingResponses) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = target
	se := sm.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := MakeHiddenService(id, intro, se[len(se)-1], c, ks)
	log.D.Ln("sending out intro onion")
	res := sm.PostAcctOnion(o)
	sm.SendWithOneHook(c[0].AddrPort, res, func(id nonce.ID,
		b slice.Bytes) (e error) {
		log.I.Ln("received routing header request for %s", intro.Key.ToBase32())
		return
	}, p)
}

func GossipIntro(intr *Intro, sm *SessionManager, c qu.C) {
	log.D.F("propagating hidden service intro for %x", intr.Key.ToBytes())
	done := qu.T()
	msg := octet.New(IntroLen)
	if check(intr.Encode(msg)) {
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
