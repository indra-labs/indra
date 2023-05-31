package onions

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

const (
	PeerMagic = "peer"
	PeerLen   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen + 1 +
		slice.Uint32Len +
		crypto.SigLen
)

type PeerAd struct {
	ID        nonce.ID    // This ensures never a repeated signed message.
	Key       *crypto.Pub // Identity key.
	RelayRate uint32
	Sig       crypto.SigBytes
}

func NewPeer(id nonce.ID, key *crypto.Prv,
	relayRate uint32) (peerAd *PeerAd) {

	pk := crypto.DerivePub(key)
	s := splice.New(IntroLen - magic.Len)
	s.ID(id).
		Pubkey(pk).
		Uint32(relayRate)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	peerAd = &PeerAd{
		ID:        id,
		Key:       pk,
		RelayRate: relayRate,
		Sig:       sign,
	}
	return
}

func (x *PeerAd) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), PeerLen-magic.Len,
		PeerMagic); fails(e) {

		return
	}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint32(&x.RelayRate).
		ReadSignature(&x.Sig)
	return
}

func (x *PeerAd) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Sig,
	)
	x.Splice(s.Magic(PeerMagic))
	return
}

func (x *PeerAd) GetOnion() interface{} { return x }

func (x *PeerAd) Gossip(sm *sess.Manager, c qu.C) {
	log.D.F("propagating peer info for %s",
		x.Key.ToBased32Abbreviated())
	Gossip(x, sm, c)
	log.T.Ln("finished broadcasting peer info")
}

func (x *PeerAd) Len() int { return PeerLen }

func (x *PeerAd) Magic() string { return PeerMagic }

func (x *PeerAd) Splice(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Key).
		Uint32(x.RelayRate).
		Signature(x.Sig)
}

func (x *PeerAd) Validate() bool {
	s := splice.New(PeerLen - magic.Len)
	s.ID(x.ID).
		Pubkey(x.Key).
		Uint32(x.RelayRate)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	key, e := x.Sig.Recover(hash)
	if fails(e) {
		return false
	}
	if key.Equals(x.Key) {
		return true
	}
	return false
}

func init() { Register(PeerMagic, peerGen) }

func peerGen() coding.Codec { return &PeerAd{} }
