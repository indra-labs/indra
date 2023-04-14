package engine

import (
	"fmt"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	InitRekeyMagic   = "kchi"
	RekeyReplyMagic  = "kchr"
	AcknowledgeMagic = "ackn"
	MungedMagic      = "mung"
)

type InitRekey struct {
	NewPubkey *crypto.Pub
}

func InitRekeyPrototype() Codec    { return &InitRekey{} }
func init()                        { Register(InitRekeyMagic, InitRekeyPrototype) }
func (k *InitRekey) Magic() string { return InitRekeyMagic }
func (k *InitRekey) GetMung() Mung { return nil }

func (k *InitRekey) Encode(s *Splice) (e error) {
	s.Magic4(InitRekeyMagic).Pubkey(k.NewPubkey)
	return
}

func (k *InitRekey) Decode(s *Splice) (e error) {
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

func (k *InitRekey) Len() int {
	return 4 + crypto.PubKeyLen
}

type RekeyReply struct {
	NewPubkey *crypto.Pub
}

func RekeyReplyPrototype() Codec    { return &RekeyReply{} }
func init()                         { Register(RekeyReplyMagic, RekeyReplyPrototype) }
func (r *RekeyReply) Magic() string { return RekeyReplyMagic }
func (r *RekeyReply) GetMung() Mung { return nil }

func (r *RekeyReply) Encode(s *Splice) (e error) {
	s.Magic4(RekeyReplyMagic).Pubkey(r.NewPubkey)
	return
}

func (r *RekeyReply) Decode(s *Splice) (e error) {
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

type Acknowledgement struct {
	*RxRecord
}

func AcknowledgementPrototype() Codec    { return &Acknowledgement{} }
func init()                              { Register(AcknowledgeMagic, AcknowledgementPrototype) }
func (a *Acknowledgement) Magic() string { return AcknowledgeMagic }
func (a *Acknowledgement) GetMung() Mung { return nil }

func (a *Acknowledgement) Encode(s *Splice) (e error) {
	s.Magic4(AcknowledgeMagic).
		ID(a.ID).
		Hash(a.Hash).
		Time(a.First).
		Time(a.Last).
		Uint64(a.Received).
		Duration(a.Ping)
	return
}

func (a *Acknowledgement) Decode(s *Splice) (e error) {
	if s.Len() < a.Len() {
		return fmt.Errorf("message too short, got %d, require %d", a.Len(),
			s.Len())
	}
	s.ReadID(&a.ID).
		ReadHash(&a.Hash).
		ReadTime(&a.First).
		ReadTime(&a.Last).
		ReadUint64(&a.Received).
		Duration(a.Ping)
	return
}

func (a *Acknowledgement) Len() int {
	return 4 + nonce.IDLen + sha256.Len + 4*slice.Uint64Len
}

type Munged struct {
	slice.Bytes // contains an encoded Mung.
}

func (m Munged) Unpack() (mu Mung) {
	s := NewSpliceFrom(m.Bytes)
	mm := Recognise(s)
	var ok bool
	if mu, ok = mm.(Mung); !ok {
		log.D.Ln("type not recognised as a mung")
	}
	return
}

func MungedPrototype() Codec    { return &Munged{} }
func init()                     { Register(MungedMagic, MungedPrototype) }
func (m *Munged) Magic() string { return MungedMagic }
func (m *Munged) GetMung() Mung { return nil }

func (m *Munged) Encode(s *Splice) (e error) {
	s.Magic4(MungedMagic).Bytes(m.Bytes)
	return
}

func (m *Munged) Decode(s *Splice) (e error) {
	if s.Len() < m.Len() {
		return fmt.Errorf("message too short, got %d, require %d", m.Len(),
			s.Len())
	}
	s.ReadBytes(&m.Bytes)
	return
}

func (m *Munged) Len() int {
	return 4 + len(m.Bytes) + 4
}
