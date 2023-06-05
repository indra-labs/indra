package adservice

import (
	"fmt"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/onions/adintro"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	Magic = "svad"
	Len   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen +
		slice.Uint16Len +
		slice.Uint32Len +
		crypto.SigLen
)

// Ad stores a specification for the fee rate and the service port, which
// must be a well known port to match with a type of service, eg 80 for web, 53
// for DNS, etc. These are also attached to the PeerAd entry via concatenating
// "/service/N" where N is the index of the entry. A zero value at an index
// signals to stop scanning for more subsequent values.
type Ad struct {
	ID        nonce.ID    // To ensure no repeating message.
	Key       *crypto.Pub // Server offering service.
	Port      uint16
	RelayRate uint32
	Sig       crypto.SigBytes
}

var _ coding.Codec = &Ad{}

func NewServiceAd(
	id nonce.ID,
	key *crypto.Prv,
	relayRate uint32,
	port uint16,
) (sv *Ad) {

	s := splice.New(adintro.Len)
	k := crypto.DerivePub(key)
	ServiceSplice(s, id, k, relayRate, port)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	sv = &Ad{
		ID:        id,
		Key:       k,
		RelayRate: relayRate,
		Port:      port,
		Sig:       sign,
	}
	return
}

func (x *Ad) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint16(&x.Port).
		ReadUint32(&x.RelayRate).
		ReadSignature(&x.Sig)
	return
}

func (x *Ad) Encode(s *splice.Splice) (e error) {
	x.Splice(s)
	return
}

func (x *Ad) GetOnion() interface{} { return nil }

func (x *Ad) Gossip(sm *sess.Manager, c qu.C) {}

func (x *Ad) Len() int { return Len }

func (x *Ad) Magic() string { return "" }

func (x *Ad) Sign(prv *crypto.Prv) (e error) {
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

func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

func (x *Ad) SpliceNoSig(s *splice.Splice) {
	ServiceSplice(s, x.ID, x.Key, x.RelayRate, x.Port)
}

func (x *Ad) Validate() (valid bool) {
	s := splice.New(adintro.Len - magic.Len)
	x.SpliceNoSig(s)
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

func ServiceSplice(
	s *splice.Splice,
	id nonce.ID,
	key *crypto.Pub,
	relayRate uint32,
	port uint16,
) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint16(port).
		Uint32(relayRate)
}

func init() { reg.Register(Magic, serviceAdGen) }

func serviceAdGen() coding.Codec { return &Ad{} }
