package onions

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/services"
	"sync"
)

// Hidden is a collection of data related to hidden services.
// Introductions both own and other, hidden services.
type Hidden struct {
	sync.Mutex
	MyIntros
	KnownIntros
	Services
}

func (hr *Hidden) AddHiddenService(svc *services.Service, key *crypto.Prv,
	in *Intro, addr string) {
	pk := crypto.DerivePub(key).ToBytes()
	hr.Lock()
	log.I.F("%s added hidden service with key %s", addr, pk)
	hr.Services[pk] = &LocalHiddenService{
		Prv:     key,
		Service: svc,
	}
	hr.Services[pk].CurrentIntros = append(hr.Services[pk].
		CurrentIntros, in)
	hr.Unlock()
}

func (hr *Hidden) AddIntro(pk *crypto.Pub, intro *Introduction) {
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

func (hr *Hidden) AddIntroToHiddenService(key crypto.PubBytes, in *Intro) {
	hr.Lock()
	hr.Services[key].CurrentIntros = append(hr.Services[key].
		CurrentIntros, in)
	hr.Unlock()
}

func (hr *Hidden) Delete(key crypto.PubBytes) (header *Introduction) {
	hr.Lock()
	var ok bool
	if header, ok = hr.MyIntros[key]; ok {
		delete(hr.MyIntros, key)
	}
	hr.Unlock()
	return
}

func (hr *Hidden) DeleteIntroByID(id nonce.ID) {
	hr.Lock()
out:
	for i := range hr.Services {
		for j := range hr.Services[i].CurrentIntros {
			if hr.Services[i].CurrentIntros[j].ID == id {
				tmp := hr.Services[i].CurrentIntros
				tmp = append(tmp[:j], tmp[j+1:]...)
				hr.Services[i].CurrentIntros = tmp
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

func (hr *Hidden) DeleteKnownIntro(key crypto.PubBytes) (
	header *Introduction) {
	hr.Lock()
	var ok bool
	if _, ok = hr.KnownIntros[key]; ok {
		delete(hr.KnownIntros, key)
	}
	hr.Unlock()
	return
}

func (hr *Hidden) FindCloakedHiddenService(key crypto.PubKey) (
	pubKey *crypto.PubBytes) {
	for i := range hr.MyIntros {
		pubKey1 := hr.MyIntros[i].Intro.Key.ToBytes()
		if crypto.Match(key, pubKey1) {
			return &pubKey1
		}
	}
	for i := range hr.Services {
		if crypto.Match(key, i) {
			return &i
		}
	}
	for i := range hr.KnownIntros {
		if crypto.Match(key, i) {
			return &i
		}
	}
	return
}

func (hr *Hidden) FindHiddenService(key crypto.PubBytes) (
	hs *LocalHiddenService) {
	hr.Lock()
	var ok bool
	if hs, ok = hr.Services[key]; ok {
	}
	hr.Unlock()
	return
}

func (hr *Hidden) FindIntroduction(key crypto.PubBytes) (intro *Introduction) {
	hr.Lock()
	var ok bool
	if intro, ok = hr.MyIntros[key]; ok {
	}
	hr.Unlock()
	return
}

func (hr *Hidden) FindIntroductionUnsafe(
	key crypto.PubBytes) (intro *Introduction) {
	var ok bool
	if intro, ok = hr.MyIntros[key]; ok {
	}
	return
}

func (hr *Hidden) FindKnownIntro(key crypto.PubBytes) (intro *Intro) {
	hr.Lock()
	var ok bool
	if intro, ok = hr.KnownIntros[key]; ok {
	}
	hr.Unlock()
	return
}

func (hr *Hidden) FindKnownIntroUnsafe(key crypto.PubBytes) (intro *Intro) {
	var ok bool
	if intro, ok = hr.KnownIntros[key]; ok {
	}
	return
}

type KnownIntros map[crypto.PubBytes]*Intro
type LocalHiddenService struct {
	Prv           *crypto.Prv
	CurrentIntros []*Intro
	*services.Service
}
type MyIntros map[crypto.PubBytes]*Introduction
type Services map[crypto.PubBytes]*LocalHiddenService

func NewHiddenrouting() *Hidden {
	return &Hidden{
		MyIntros:    make(MyIntros),
		KnownIntros: make(KnownIntros),
		Services:    make(Services),
	}
}
