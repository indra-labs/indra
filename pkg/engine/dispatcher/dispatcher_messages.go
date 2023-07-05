package dispatcher

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

// Magic bytes message prefixes for Dispatcher messages.
const (
	NewKeyMagic      = "newK"
	AcknowledgeMagic = "ackn"
	OnionMagic       = "onio"
)

type (
	// Acknowledge wraps up an RxRecord to tell the other side how a message
	// transmission went.
	Acknowledge struct {
		*RxRecord
	}

	// NewKey delivers a new public key for the other side to use to encrypt
	// messages.
	NewKey struct {
		NewPubkey *crypto.Pub
	}

	// Onion is an onion, intended to be processed by the recipient, its layer
	// decoded and the enclosed message received and processed appropriately.
	Onion struct {
		slice.Bytes // contains an encoded Onion.
	}
)

// Decode a splice with the cursor at the first byte after the magic.
func (a *Acknowledge) Decode(s *splice.Splice) (e error) {
	if s.Len() < a.Len() {
		return fmt.Errorf("message too short, got %d, require %d", a.Len(),
			s.Len())
	}
	s.ReadID(&a.ID).
		ReadHash(&a.Hash).
		ReadTime(&a.First).
		ReadTime(&a.Last).
		ReadUint64(&a.Size).
		ReadUint64(&a.Received).
		ReadDuration(&a.Ping)
	return
}

// Encode an Acknowledge message to a splice.
func (a *Acknowledge) Encode(s *splice.Splice) (e error) {
	s.Magic(AcknowledgeMagic).
		ID(a.ID).
		Hash(a.Hash).
		Time(a.First).
		Time(a.Last).
		Uint64(a.Size).
		Uint64(a.Received).
		Duration(a.Ping)
	return
}

// AcknowledgeGen is a factory function that will be added to the registry for
// recognition and generation.
func AcknowledgeGen() codec.Codec { return &Acknowledge{&RxRecord{}} }

// GetOnion returns nil because there is no onion inside an Acknowledge.
func (a *Acknowledge) GetOnion() interface{} { return nil }

// Len returns the length of an Acknowledge message in bytes.
func (a *Acknowledge) Len() int {
	return 4 + nonce.IDLen + sha256.Len + 5*slice.Uint64Len
}

// Magic is the identifying 4 byte prefix of an Acknowledge in binary form.
func (a *Acknowledge) Magic() string { return AcknowledgeMagic }

// InitRekeyGen is a factory function to generate a NewKey.
func InitRekeyGen() codec.Codec { return &NewKey{} }

// Decode a NewKey out of a splice with cursor pointing to the first byte after
// the magic.
func (k *NewKey) Decode(s *splice.Splice) (e error) {
	if s.Len() < k.Len() {
		return fmt.Errorf("message too short, got %d, require %d", k.Len(),
			s.Len())
	}
	s.ReadPubkey(&k.NewPubkey)
	if k.NewPubkey == nil {
		return fmt.Errorf("invalid public key")
	}
	return
}

// Encode a NewKey into the provided splice.
func (k *NewKey) Encode(s *splice.Splice) (e error) {
	s.Magic(NewKeyMagic).Pubkey(k.NewPubkey)
	return
}

// GetOnion returns nil because there is no onion inside an NewKey.
func (k *NewKey) GetOnion() interface{} { return nil }

// Len returns the length of an NewKey message in bytes.
func (k *NewKey) Len() int { return 4 + crypto.PubKeyLen }

// Magic is the identifying 4 byte prefix of an NewKey in binary form.
func (k *NewKey) Magic() string { return NewKeyMagic }

// Decode an Onion out of a splice with cursor pointing to the first byte after
// the magic.
func (o *Onion) Decode(s *splice.Splice) (e error) {
	if s.Len() < o.Len() {
		return fmt.Errorf("message too short, got %d, require %d", o.Len(),
			s.Len())
	}
	s.ReadBytes(&o.Bytes)
	return
}

// Encode an Onion into the provided splice.
func (o *Onion) Encode(s *splice.Splice) (e error) {
	s.Magic(OnionMagic).Bytes(o.Bytes)
	return
}

// OnionGen is a factory function for creating a new Onion.
func OnionGen() codec.Codec { return &Onion{} }

// GetOnion invockes Unpack, which returns the onion.
func (o *Onion) GetOnion() interface{} { return o.Unpack() }

func (o *Onion) Len() int {
	return 4 + len(o.Bytes) + 4
}
func (o *Onion) Magic() string { return OnionMagic }

func (o Onion) Unpack() (mu ont.Onion) {
	s := splice.NewFrom(o.Bytes)
	mm := reg.Recognise(s)
	var ok bool
	if mu, ok = mm.(ont.Onion); !ok {
		log.D.Ln("type not recognised as a onion")
	}
	return
}

func init() {
	reg.Register(NewKeyMagic, InitRekeyGen)
	reg.Register(AcknowledgeMagic, AcknowledgeGen)
	reg.Register(OnionMagic, OnionGen)
}
