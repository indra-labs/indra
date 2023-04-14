package engine

import (
	"fmt"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	KeyChangeInitiateMagic = "kchi"
	KeyChangeReplyMagic    = "kchr"
	AcknowledgementMagic   = "ackn"
	MungMagic              = "mung"
)

type KeyChangeInitiate struct {
	NewPubkey *crypto.Pub
}

func (k *KeyChangeInitiate) Encode(s *Splice) (e error) {
	s.Magic4(KeyChangeInitiateMagic).Pubkey(k.NewPubkey)
	return
}

func (k *KeyChangeInitiate) Decode(s *Splice) (e error) {
	if s.Len() < k.Len() {
		return fmt.Errorf("message too short, got %d, require %d", k.Len(),
			s.Len())
	}
	var magic string
	s.ReadMagic4(&magic)
	if magic != KeyChangeInitiateMagic {
		return fmt.Errorf("incorrect magic, got '%s', require '%s'", magic,
			KeyChangeInitiateMagic)
	}
	s.ReadPubkey(&k.NewPubkey)
	if k.NewPubkey == nil {
		return fmt.Errorf("invalid public key")
	}
	return
}

func (k *KeyChangeInitiate) Len() int {
	return 4 + crypto.PubKeyLen
}

type KeyChangeReply struct {
	NewPubkey *crypto.Pub
}

func (r *KeyChangeReply) Encode(s *Splice) (e error) {
	s.Magic4(KeyChangeReplyMagic).Pubkey(r.NewPubkey)
	return
}

func (r *KeyChangeReply) Decode(s *Splice) (e error) {
	if s.Len() < r.Len() {
		return fmt.Errorf("message too short, got %d, require %d", r.Len(),
			s.Len())
	}
	var magic string
	s.ReadMagic4(&magic)
	if magic != KeyChangeReplyMagic {
		return fmt.Errorf("incorrect magic, got '%s', require '%s'", magic,
			KeyChangeReplyMagic)
	}
	s.ReadPubkey(&r.NewPubkey)
	if r.NewPubkey == nil {
		return fmt.Errorf("invalid public key")
	}
	return
}

func (r *KeyChangeReply) Len() int {
	return 4 + crypto.PubKeyLen
}

type Acknowledgement struct {
	RxRecord
}

func (a *Acknowledgement) Encode(s *Splice) (e error) {
	s.Magic4(AcknowledgementMagic).
		ID(a.ID).
		Hash(a.Hash).
		Time(a.First).
		Time(a.Last).
		Uint64(uint64(a.Received)).
		Duration(a.Ping)
	return
}

func (a *Acknowledgement) Decode(s *Splice) (e error) {
	if s.Len() < a.Len() {
		return fmt.Errorf("message too short, got %d, require %d", a.Len(),
			s.Len())
	}
	var magic string
	s.ReadMagic4(&magic)
	if magic != AcknowledgementMagic {
		return fmt.Errorf("incorrect magic, got '%s', require '%s'", magic,
			AcknowledgementMagic)
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

type Mung struct {
	slice.Bytes
}

func (m *Mung) Encode(s *Splice) (e error) {
	s.Magic4(MungMagic).Bytes(m.Bytes)
	return
}

func (m *Mung) Decode(s *Splice) (e error) {
	if s.Len() < m.Len() {
		return fmt.Errorf("message too short, got %d, require %d", m.Len(),
			s.Len())
	}
	var magic string
	s.ReadMagic4(&magic)
	if magic != MungMagic {
		return fmt.Errorf("incorrect magic, got '%s', require '%s'", magic,
			MungMagic)
	}
	s.ReadBytes(&m.Bytes)
	return
}

func (m *Mung) Len() int {
	return 4 + len(m.Bytes) + 4
}
