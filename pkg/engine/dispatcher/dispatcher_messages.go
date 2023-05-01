package dispatcher

import (
	"fmt"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/onions"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

const (
	NewKeyMagic      = "newK"
	AcknowledgeMagic = "ackn"
	OnionMagic       = "onio"
)

type NewKey struct {
	NewPubkey *crypto.Pub
}

func InitRekeyGen() coding.Codec        { return &NewKey{} }
func init()                             { onions.Register(NewKeyMagic, InitRekeyGen) }
func (k *NewKey) Magic() string         { return NewKeyMagic }
func (k *NewKey) GetOnion() interface{} { return nil }
func (k *NewKey) Len() int              { return 4 + crypto.PubKeyLen }

func (k *NewKey) Encode(s *splice.Splice) (e error) {
	s.Magic(NewKeyMagic).Pubkey(k.NewPubkey)
	return
}

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

type Acknowledge struct {
	*RxRecord
}

func AcknowledgeGen() coding.Codec           { return &Acknowledge{&RxRecord{}} }
func init()                                  { onions.Register(AcknowledgeMagic, AcknowledgeGen) }
func (a *Acknowledge) Magic() string         { return AcknowledgeMagic }
func (a *Acknowledge) GetOnion() interface{} { return nil }

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

func (a *Acknowledge) Len() int {
	return 4 + nonce.IDLen + sha256.Len + 5*slice.Uint64Len
}

type Onion struct {
	slice.Bytes // contains an encoded Onion.
}

func (m Onion) Unpack() (mu onions.Onion) {
	s := splice.NewFrom(m.Bytes)
	mm := onions.Recognise(s)
	var ok bool
	if mu, ok = mm.(onions.Onion); !ok {
		log.D.Ln("type not recognised as a onion")
	}
	return
}

func OnionGen() coding.Codec           { return &Onion{} }
func init()                            { onions.Register(OnionMagic, OnionGen) }
func (m *Onion) Magic() string         { return OnionMagic }
func (m *Onion) GetOnion() interface{} { return nil }

func (m *Onion) Encode(s *splice.Splice) (e error) {
	s.Magic(OnionMagic).Bytes(m.Bytes)
	return
}

func (m *Onion) Decode(s *splice.Splice) (e error) {
	if s.Len() < m.Len() {
		return fmt.Errorf("message too short, got %d, require %d", m.Len(),
			s.Len())
	}
	s.ReadBytes(&m.Bytes)
	return
}

func (m *Onion) Len() int {
	return 4 + len(m.Bytes) + 4
}
