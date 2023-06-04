package adpeer

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/ad"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/onions/intro"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	Magic = "peer"
	Len   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen + 1 +
		slice.Uint32Len +
		crypto.SigLen
)

type Ad struct {
	ID        nonce.ID    // This ensures never a repeated signed message.
	Key       *crypto.Pub // Identity key.
	RelayRate uint32
	Sig       crypto.SigBytes
}

func New(id nonce.ID, key *crypto.Prv,
	relayRate uint32) (peerAd *Ad) {

	pk := crypto.DerivePub(key)
	s := splice.New(intro.Len - magic.Len)
	s.ID(id).
		Pubkey(pk).
		Uint32(relayRate)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	peerAd = &Ad{
		ID:        id,
		Key:       pk,
		RelayRate: relayRate,
		Sig:       sign,
	}
	return
}

func (x *Ad) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {

		return
	}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint32(&x.RelayRate).
		ReadSignature(&x.Sig)
	return
}

func (x *Ad) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Sig,
	)
	x.Splice(s.Magic(Magic))
	return
}

func (x *Ad) GetOnion() interface{} { return x }

func (x *Ad) Gossip(sm *sess.Manager, c qu.C) {
	log.D.F("propagating peer info for %s",
		x.Key.ToBased32Abbreviated())
	ad.Gossip(x, sm, c)
	log.T.Ln("finished broadcasting peer info")
}

func (x *Ad) Len() int { return Len }

func (x *Ad) Magic() string { return Magic }

func (x *Ad) Splice(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Key).
		Uint32(x.RelayRate).
		Signature(x.Sig)
}

func (x *Ad) Validate() bool {
	s := splice.New(Len - magic.Len)
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

func init() { reg.Register(Magic, peerGen) }

func peerGen() coding.Codec { return &Ad{} }
