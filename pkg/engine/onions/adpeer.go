package onions

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

const PeerAdLen = magic.Len +
	nonce.IDLen +
	slice.Uint64Len +
	crypto.SigLen

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
	nonce.ID  // To ensure no repeating message
	Identity  crypto.PubBytes
	RelayRate int
	Sig       crypto.SigBytes
	// Addresses - first is address, nil for hidden services,
	// hidden services have more than one, 6 or more are kept active.
	Addresses    []*Address
	ServiceInfos []ServiceAd
}

func (p *PeerAd) Decode(s *splice.Splice) (e error) {
	var v uint64
	s.ReadID(&p.ID).ReadUint64(&v)
	s.ReadSignature(&p.Sig)
	p.RelayRate = int(v)
	return nil
}

func (p *PeerAd) Encode(s *splice.Splice) (e error) {
	s.ID(p.ID).Uint64(uint64(p.RelayRate))
	return nil
}

func (p *PeerAd) GetOnion() interface{} {
	return nil
}

func (p *PeerAd) Len() int {
	return PeerAdLen
}

func (p *PeerAd) Magic() string {
	return ""
}

func (p *PeerAd) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(p.Len())
	if e = p.Encode(s); fails(e) {
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
	copy(p.Sig[:], b)
	return nil
}

func (p *PeerAd) Validate(s *splice.Splice) (pk *crypto.Pub) {
	h := sha256.Single(s.GetRange(0, nonce.IDLen+slice.Uint64Len))
	var e error
	if pk, e = p.Sig.Recover(h); fails(e) {
	}
	return
}
