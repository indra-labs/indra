package engine

import (
	"sync"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
)

type Introduction struct {
	*Intro
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces
	RoutingHeaderBytes
}

type MyIntros map[pub.Bytes]*Introduction

type KnownIntros map[pub.Bytes]*Intro

type LocalHiddenService struct {
	Prv           *prv.Key
	CurrentIntros []*Intro
	*Service
}

type HiddenServices map[pub.Bytes]*LocalHiddenService

// HiddenRouting is a collection of data related to hidden services.
// Introductions both own and other, hidden services.
type HiddenRouting struct {
	sync.Mutex
	MyIntros
	KnownIntros
	HiddenServices
}

func NewHiddenrouting() *HiddenRouting {
	return &HiddenRouting{
		MyIntros:       make(MyIntros),
		KnownIntros:    make(KnownIntros),
		HiddenServices: make(HiddenServices),
	}
}

func (hr *HiddenRouting) AddHiddenService(svc *Service, key *prv.Key,
	in *Intro, addr string) {
	
	pk := pub.Derive(key).ToBytes()
	hr.Lock()
	log.I.F("%s added hidden service with key %s", addr, pk)
	hr.HiddenServices[pk] = &LocalHiddenService{
		Prv:     key,
		Service: svc,
	}
	hr.HiddenServices[pk].CurrentIntros = append(hr.HiddenServices[pk].
		CurrentIntros, in)
	hr.Unlock()
}

func (hr *HiddenRouting) AddIntroToHiddenService(key pub.Bytes, in *Intro) {
	hr.Lock()
	hr.HiddenServices[key].CurrentIntros = append(hr.HiddenServices[key].
		CurrentIntros, in)
	hr.Unlock()
}

func (hr *HiddenRouting) DeleteIntroByID(id nonce.ID) {
	hr.Lock()
out:
	for i := range hr.HiddenServices {
		for j := range hr.HiddenServices[i].CurrentIntros {
			if hr.HiddenServices[i].CurrentIntros[j].ID == id {
				tmp := hr.HiddenServices[i].CurrentIntros
				tmp = append(tmp[:j], tmp[j+1:]...)
				hr.HiddenServices[i].CurrentIntros = tmp
				break out
			}
		}
	}
	for i := range hr.KnownIntros {
		if hr.KnownIntros[i].ID == id {
			delete(hr.KnownIntros, i)
			break
		}
	}
	hr.Unlock()
	
}

func (hr *HiddenRouting) FindCloakedHiddenService(key cloak.PubKey) (
	pubKey *pub.Bytes) {
	
	for i := range hr.MyIntros {
		pubKey1 := hr.MyIntros[i].Key.ToBytes()
		if cloak.Match(key, pubKey1) {
			return &pubKey1
		}
	}
	for i := range hr.HiddenServices {
		if cloak.Match(key, i) {
			return &i
		}
	}
	for i := range hr.KnownIntros {
		if cloak.Match(key, i) {
			return &i
		}
	}
	return
}

func (hr *HiddenRouting) FindHiddenService(key pub.Bytes) (
	hs *LocalHiddenService) {
	
	hr.Lock()
	var ok bool
	if hs, ok = hr.HiddenServices[key]; ok {
	}
	hr.Unlock()
	return
}

func (hr *HiddenRouting) FindIntroduction(key pub.Bytes) (intro *Introduction) {
	hr.Lock()
	var ok bool
	if intro, ok = hr.MyIntros[key]; ok {
	}
	hr.Unlock()
	return
}

func (hr *HiddenRouting) FindIntroductionUnsafe(
	key pub.Bytes) (intro *Introduction) {
	
	var ok bool
	if intro, ok = hr.MyIntros[key]; ok {
	}
	return
}

func (hr *HiddenRouting) FindKnownIntro(key pub.Bytes) (intro *Intro) {
	hr.Lock()
	var ok bool
	if intro, ok = hr.KnownIntros[key]; ok {
	}
	hr.Unlock()
	return
}

func (hr *HiddenRouting) FindKnownIntroUnsafe(key pub.Bytes) (intro *Intro) {
	var ok bool
	if intro, ok = hr.KnownIntros[key]; ok {
	}
	return
}

func (hr *HiddenRouting) Delete(key pub.Bytes) (header *Introduction) {
	hr.Lock()
	var ok bool
	if header, ok = hr.MyIntros[key]; ok {
		delete(hr.MyIntros, key)
	}
	hr.Unlock()
	return
}

func (hr *HiddenRouting) DeleteKnownIntro(key pub.Bytes) (
	header *Introduction) {
	hr.Lock()
	var ok bool
	if _, ok = hr.KnownIntros[key]; ok {
		delete(hr.KnownIntros, key)
	}
	hr.Unlock()
	return
}

func (hr *HiddenRouting) AddIntro(pk *pub.Key, intro *Introduction) {
	hr.Lock()
	var ok bool
	key := pk.ToBytes()
	if _, ok = hr.MyIntros[key]; ok {
		log.D.Ln("entry already exists for key %x", key)
	} else {
		hr.MyIntros[key] = intro
	}
	hr.Unlock()
}

func GossipIntro(intro *Intro, sm *SessionManager, c qu.C) {
	log.D.F("propagating hidden service intro for %s",
		intro.Key.ToBase32Abbreviated())
	done := qu.T()
	msg := NewSplice(IntroLen)
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
		n.Transport.Send(msg.GetAll())
		cursor++
		if cursor > len(peerIndices)-1 {
			break
		}
	}
	log.T.Ln("finished broadcasting intro")
}
