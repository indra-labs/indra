// Package hidden is a manager for hidden services.
package hidden

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/onions/ad/intro"
	"github.com/indra-labs/indra/pkg/onions/consts"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"sync"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// Hidden is a collection of data related to hidden services.
// Introductions both own and other, hidden services.
type Hidden struct {

	// Mutex because this needs to be concurrent safe.
	sync.Mutex

	// MyIntros are hidden services being hosted in this current engine.
	MyIntros

	// KnownIntros are hidden services we have heard of and can return to a query if we have them.
	//
	// todo: this probably belongs with the peerstore...
	KnownIntros

	// Services are the service specifications for the services we are providing from this engine.
	Services
}

// AddHiddenService adds a new hidden service with a given private key, intro and address.
//
// todo: looks like that addr parameter should be in a logging closure derived from the key.
func (hr *Hidden) AddHiddenService(svc *services.Service, key *crypto.Prv,
	in *intro.Ad, addr string) {
	pk := crypto.DerivePub(key).ToBytes()
	hr.Lock()
	log.D.F("%s added hidden service with key %s", addr, pk)
	hr.Services[pk] = &LocalHiddenService{
		Prv:     key,
		Service: svc,
	}
	hr.Services[pk].CurrentIntros = append(hr.Services[pk].
		CurrentIntros, in)
	hr.Unlock()
}

// AddIntro adds an intro for a newly created hidden service.
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

// AddIntroToHiddenService adds an intro to a given hidden service.
//
// todo: this looks like it isn't used.
func (hr *Hidden) AddIntroToHiddenService(key crypto.PubBytes, in *intro.Ad) {
	hr.Lock()
	hr.Services[key].CurrentIntros = append(hr.Services[key].CurrentIntros, in)
	hr.Unlock()
}

// Delete removes a hidden service identified by its public key bytes.
func (hr *Hidden) Delete(key crypto.PubBytes) (header *Introduction) {
	hr.Lock()
	var ok bool
	if header, ok = hr.MyIntros[key]; ok {
		delete(hr.MyIntros, key)
	}
	hr.Unlock()
	return
}

// DeleteIntroByID removes the Intro with matching nonce.ID.
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

// DeleteKnownIntro evicts a known intro.
//
// todo: this really should be integrated with the peerstore.
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

// RoutingHeaderBytes is a raw bytes form of a 3 layer RoutingHeader.
type RoutingHeaderBytes [consts.RoutingHeaderLen]byte

// Services is a map of local hidden services keyed to their public key.
type Services map[crypto.PubBytes]*LocalHiddenService

// FormatReply constructs a reply message with RoutingHeader, ciphers and nonces,
// and the response payload, encrypting it and writing it into a splice.
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

// GetRoutingHeaderFromCursor extracts what should be a routing header from a
// splice.
func GetRoutingHeaderFromCursor(s *splice.Splice) (r RoutingHeaderBytes) {
	rh := s.GetRange(s.GetCursor(), s.Advance(consts.RoutingHeaderLen,
		"routing header"))
	copy(r[:], rh)
	return
}

// FindCloakedHiddenService checks known local hidden service keys to match a
// "To" cloaked public key, for which we should have the private key to form the
// ECDH secret to decrypt the message.
func (hr *Hidden) FindCloakedHiddenService(key crypto.CloakedPubKey) (
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

// FindHiddenService searches for a hidden service from local hidden services.
func (hr *Hidden) FindHiddenService(key crypto.PubBytes) (
	hs *LocalHiddenService) {
	hr.Lock()
	var ok bool
	if hs, ok = hr.Services[key]; ok {
	}
	hr.Unlock()
	return
}

// FindIntroduction returns the intro.Ad matching a provided public key bytes.
func (hr *Hidden) FindIntroduction(key crypto.PubBytes) (intro *Introduction) {
	hr.Lock()
	intro = hr.FindIntroductionUnsafe(key)
	hr.Unlock()
	return
}

// FindIntroductionUnsafe does the same thing as FindIntroduction without
// locking. Used when the lock is already acquired.
func (hr *Hidden) FindIntroductionUnsafe(
	key crypto.PubBytes) (intro *Introduction) {
	var ok bool
	if intro, ok = hr.MyIntros[key]; ok {
	}
	return
}

// FindKnownIntro searches non-local intros for a matching public key bytes.
//
// todo: this definitely should be part of peerstore.
func (hr *Hidden) FindKnownIntro(key crypto.PubBytes) (intro *intro.Ad) {
	hr.Lock()
	intro = hr.FindKnownIntroUnsafe(key)
	hr.Unlock()
	return
}

// FindKnownIntroUnsafe searches for a KnownIntro without locking.
func (hr *Hidden) FindKnownIntroUnsafe(key crypto.PubBytes) (intro *intro.Ad) {
	var ok bool
	if intro, ok = hr.KnownIntros[key]; ok {
	}
	return
}

// Introduction is the combination of an intro.Ad and a ReplyHeader, used to
// forward the Route message from a client and establish the connection between
// client and hidden service.
type Introduction struct {
	Intro *intro.Ad
	ReplyHeader
}

// KnownIntros is a key/value store of hidden service intros we know of.
//
// todo: This definitely should be peerstore
type KnownIntros map[crypto.PubBytes]*intro.Ad

// LocalHiddenService is a hidden service being served from this node.
type LocalHiddenService struct {

	// Prv is the private key for the hidden service.
	Prv *crypto.Prv

	// CurrentIntros are intro.Ad that are current for this hidden service. Not sure
	// yet how many this should be. 6 or more, it really depends, perhaps have it
	// scale up if demand exceeds supply to some sort of reasonable ceiling.
	CurrentIntros []*intro.Ad

	// Service is the definition of the hidden service. There should be a server
	// listening on or forwarding from the service port on localhost that provides
	// the service.
	*services.Service
}

// MyIntros is a key value store of the hidden service introductions we have got
// currently available and sent out to introducing nodes.
type MyIntros map[crypto.PubBytes]*Introduction

// ReplyHeader is the bundle of routing header, payload encryption secrets and
// the nonces to be used, which match also what is inside the RoutingHeaderBytes.
type ReplyHeader struct {

	// RoutingHeaderBytes contains the 3 layer RoutingHeader that holds the path
	// instructions in three progressively encrypted layers in reverse order so to be
	// unwrapped progressively.
	RoutingHeaderBytes

	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers

	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces
}

// NewHiddenRouting instantiates a new Hidden for managing hidden services.
func NewHiddenRouting() *Hidden {
	return &Hidden{
		MyIntros:    make(MyIntros),
		KnownIntros: make(KnownIntros),
		Services:    make(Services),
	}
}

// ReadRoutingHeader extracts a RoutingHeaderBytes from a splice at the current
// cursor.
func ReadRoutingHeader(s *splice.Splice, b *RoutingHeaderBytes) *splice.Splice {
	*b = GetRoutingHeaderFromCursor(s)
	s.Segments = append(s.Segments,
		splice.NameOffset{Offset: s.GetCursor(), Name: "routingheader"})
	return s
}

// WriteRoutingHeader copies RoutingHeaderBytes into a splice.
func WriteRoutingHeader(s *splice.Splice, b RoutingHeaderBytes) *splice.Splice {
	copy(s.GetAll()[s.GetCursor():s.Advance(consts.RoutingHeaderLen,
		"routing header")], b[:])
	s.Segments = append(s.Segments,
		splice.NameOffset{Offset: s.GetCursor(), Name: "routingheader"})
	return s
}
