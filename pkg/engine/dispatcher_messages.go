package engine

import (
	"fmt"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/onions"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	InitRekeyMagic   = "kchi"
	RekeyReplyMagic  = "kchr"
	AcknowledgeMagic = "ackn"
	OnionMagic       = "onio"
)

type InitRekey struct {
	NewPubkey *crypto.Pub
}

func InitRekeyGen() coding.Codec           { return &InitRekey{} }
func init()                                { onions.Register(InitRekeyMagic, InitRekeyGen) }
func (k *InitRekey) Magic() string         { return InitRekeyMagic }
func (k *InitRekey) GetOnion() interface{} { return nil }
func (k *InitRekey) Len() int              { return 4 + crypto.PubKeyLen }

func (k *InitRekey) Encode(s *splice.Splice) (e error) {
	s.Magic(InitRekeyMagic).Pubkey(k.NewPubkey)
	return
}

func (k *InitRekey) Decode(s *splice.Splice) (e error) {
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

type RekeyReply struct {
	NewPubkey *crypto.Pub
}

func RekeyReplyGen() coding.Codec           { return &RekeyReply{} }
func init()                                 { onions.Register(RekeyReplyMagic, RekeyReplyGen) }
func (r *RekeyReply) Magic() string         { return RekeyReplyMagic }
func (r *RekeyReply) GetOnion() interface{} { return nil }

func (r *RekeyReply) Encode(s *splice.Splice) (e error) {
	s.Magic(RekeyReplyMagic).Pubkey(r.NewPubkey)
	return
}

func (r *RekeyReply) Decode(s *splice.Splice) (e error) {
	if s.Len() < r.Len() {
		return fmt.Errorf("message too short, got %d, require %d", r.Len(),
			s.Len())
	}
	s.ReadPubkey(&r.NewPubkey)
	if r.NewPubkey == nil {
		return fmt.Errorf("invalid public key")
	}
	return
}

func (r *RekeyReply) Len() int {
	return 4 + crypto.PubKeyLen
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
		Uint64(a.Received).
		Duration(a.Ping)
	return
}

func (a *Acknowledge) Decode(s *splice.Splice) (e error) {
	if s.Len() < a.Len() {
		return fmt.Errorf("message too short, got %d, require %d", a.Len(),
			s.Len())
	}
	log.D.S("ack", a)
	s.ReadID(&a.ID).
		ReadHash(&a.Hash).
		ReadTime(&a.First).
		ReadTime(&a.Last).
		ReadUint64(&a.Received).
		Duration(a.Ping)
	return
}

func (a *Acknowledge) Len() int {
	return 4 + nonce.IDLen + sha256.Len + 4*slice.Uint64Len
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
