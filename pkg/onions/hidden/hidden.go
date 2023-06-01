package hidden

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/onions/consts"
	"github.com/indra-labs/indra/pkg/onions/intro"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"sync"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
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
	in *intro.Ad, addr string) {
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

func (hr *Hidden) AddIntroToHiddenService(key crypto.PubBytes, in *intro.Ad) {
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

type RoutingHeaderBytes [consts.RoutingHeaderLen]byte
type Services map[crypto.PubBytes]*LocalHiddenService

func FormatReply(header RoutingHeaderBytes, ciphers crypto.Ciphers,
	nonces crypto.Nonces, res slice.Bytes) (rb *splice.Splice) {

	rl := consts.RoutingHeaderLen
	rb = splice.New(rl + len(res))
	copy(rb.GetUntil(rl), header[:rl])
	copy(rb.GetFrom(rl), res)
	for i := range ciphers {
		blk := ciph.BlockFromHash(ciphers[i])
		ciph.Encipher(blk, nonces[i], rb.GetFrom(rl))
	}
	return
}

func GetRoutingHeaderFromCursor(s *splice.Splice) (r RoutingHeaderBytes) {
	rh := s.GetRange(s.GetCursor(), s.Advance(consts.RoutingHeaderLen,
		"routing header"))
	copy(r[:], rh)
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

func (hr *Hidden) FindKnownIntro(key crypto.PubBytes) (intro *intro.Ad) {
	hr.Lock()
	var ok bool
	if intro, ok = hr.KnownIntros[key]; ok {
	}
	hr.Unlock()
	return
}

func (hr *Hidden) FindKnownIntroUnsafe(key crypto.PubBytes) (intro *intro.Ad) {
	var ok bool
	if intro, ok = hr.KnownIntros[key]; ok {
	}
	return
}

type Introduction struct {
	Intro *intro.Ad
	ReplyHeader
}
type KnownIntros map[crypto.PubBytes]*intro.Ad
type LocalHiddenService struct {
	Prv           *crypto.Prv
	CurrentIntros []*intro.Ad
	*services.Service
}
type MyIntros map[crypto.PubBytes]*Introduction
type ReplyHeader struct {
	RoutingHeaderBytes
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces
}

func NewHiddenrouting() *Hidden {
	return &Hidden{
		MyIntros:    make(MyIntros),
		KnownIntros: make(KnownIntros),
		Services:    make(Services),
	}
}

func ReadRoutingHeader(s *splice.Splice, b *RoutingHeaderBytes) *splice.Splice {
	*b = GetRoutingHeaderFromCursor(s)
	s.Segments = append(s.Segments,
		splice.NameOffset{Offset: s.GetCursor(), Name: "routingheader"})
	return s
}

func WriteRoutingHeader(s *splice.Splice, b RoutingHeaderBytes) *splice.Splice {
	copy(s.GetAll()[s.GetCursor():s.Advance(consts.RoutingHeaderLen,
		"routing header")], b[:])
	s.Segments = append(s.Segments,
		splice.NameOffset{Offset: s.GetCursor(), Name: "routingheader"})
	return s
}
