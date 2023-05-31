package onions

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

const (
	PeerAdMagic = "prad"
	PeerAdLen   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen +
		slice.Uint32Len +
		crypto.SigLen
)

var _ Ad = &PeerAd{}

// PeerAd is the root identity document for an Indra peer. It is indexed by the
// Identity field, its public key. The slices found below it are derived via
// concatenation of strings with the keys and hashing to generate a derived
// field index, used to search the DHT for matches.
//
// The data stored for Peer must be signed with the key claimed by the Identity.
// For hidden services the address fields are signed in the DHT by the hidden
// service from their introduction solicitation, and the index from the current
// set is given by the hidden service.
type PeerAd struct {
	nonce.ID              // To ensure no repeating message
	Identity  *crypto.Pub // Must match signature.
	RelayRate uint32      // Zero means not relaying.
	Sig       crypto.SigBytes
}

func NewPeerAd(
	id nonce.ID,
	key *crypto.Prv,
	relayRate uint32,
) (pa *PeerAd) {

	pa = &PeerAd{
		ID:        id,
		Identity:  crypto.DerivePub(key),
		RelayRate: relayRate,
	}
	var e error
	if e = pa.Sign(key); fails(e) {
		return
	}
	return
}

func (x *PeerAd) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	//TODO implement me
	panic("implement me")
}

func (x *PeerAd) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadPubkey(&x.Identity).
		ReadUint32(&x.RelayRate).
		ReadSignature(&x.Sig)
	return nil
}

func (x *PeerAd) Encode(s *splice.Splice) (e error) {
	s.Magic(PeerAdMagic)
	x.Splice(s)
	return nil
}

func (x *PeerAd) GetOnion() interface{}           { return nil }
func (x *PeerAd) Gossip(sm *sess.Manager, c qu.C) {}
func (x *PeerAd) Handle(s *splice.Splice, p Onion, ni Ngin) (e error) {
	return nil
}
func (x *PeerAd) Len() int      { return PeerAdLen }
func (x *PeerAd) Magic() string { return PeerAdMagic }

func (x *PeerAd) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(x.Len())
	if e = x.Encode(s); fails(e) {
		return
	}
	var b []byte
	if b, e = prv.Sign(s.GetUntil(s.GetCursor())); fails(e) {
		return
	}
	if len(b) != crypto.SigLen {
		return fmt.Errorf("signature incorrect length, got %d expected %d",
			len(b), crypto.SigLen)
	}
	copy(x.Sig[:], b)
	return nil
}

func (x *PeerAd) Splice(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Identity).
		Uint32(x.RelayRate).
		Signature(x.Sig)
}

func (x *PeerAd) SpliceNoSig(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Identity).
		Uint32(x.RelayRate).
		Signature(x.Sig)
}

func (x *PeerAd) Validate() (valid bool) {
	s := splice.New(x.Len())
	x.SpliceNoSig(s)
	h := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var pk *crypto.Pub
	if pk, e = x.Sig.Recover(h); fails(e) {
	}
	return pk != nil
}

func (x *PeerAd) Wrap(inner Onion) {}
func init()                        { Register(PeerAdMagic, peerAdGen) }
func peerAdGen() coding.Codec      { return &PeerAd{} }
