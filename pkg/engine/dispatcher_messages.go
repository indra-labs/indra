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

func InitRekeyGen() Codec            { return &InitRekey{} }
func init()                          { Register(InitRekeyMagic, InitRekeyGen) }
func (k *InitRekey) Magic() string   { return InitRekeyMagic }
func (k *InitRekey) GetOnion() Onion { return nil }

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

func RekeyReplyGen() Codec            { return &RekeyReply{} }
func init()                           { Register(RekeyReplyMagic, RekeyReplyGen) }
func (r *RekeyReply) Magic() string   { return RekeyReplyMagic }
func (r *RekeyReply) GetOnion() Onion { return nil }

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

type Acknowledge struct {
	*RxRecord
}

func AcknowledgeGen() Codec            { return &Acknowledge{} }
func init()                            { Register(AcknowledgeMagic, AcknowledgeGen) }
func (a *Acknowledge) Magic() string   { return AcknowledgeMagic }
func (a *Acknowledge) GetOnion() Onion { return nil }

func (a *Acknowledge) Encode(s *Splice) (e error) {
	s.Magic4(AcknowledgeMagic).
		ID(a.ID).
		Hash(a.Hash).
		Time(a.First).
		Time(a.Last).
		Uint64(a.Received).
		Duration(a.Ping)
	return
}

func (a *Acknowledge) Decode(s *Splice) (e error) {
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

func (a *Acknowledge) Len() int {
	return 4 + nonce.IDLen + sha256.Len + 4*slice.Uint64Len
}

type Munged struct {
	slice.Bytes // contains an encoded Onion.
}

func (m Munged) Unpack() (mu Onion) {
	s := NewSpliceFrom(m.Bytes)
	mm := Recognise(s)
	var ok bool
	if mu, ok = mm.(Onion); !ok {
		log.D.Ln("type not recognised as a mung")
	}
	return
}

func MungedGen() Codec            { return &Munged{} }
func init()                       { Register(MungedMagic, MungedGen) }
func (m *Munged) Magic() string   { return MungedMagic }
func (m *Munged) GetOnion() Onion { return nil }

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
