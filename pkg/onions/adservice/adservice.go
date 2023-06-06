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
	"github.com/indra-labs/indra/pkg/onions/adproto"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"time"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	Magic = "svad"
	Len   = adproto.Len +
		slice.Uint16Len +
		slice.Uint32Len
)

// Ad stores a specification for the fee rate and the service port, which
// must be a well known port to match with a type of service, eg 80 for web, 53
// for DNS, etc. These are also attached to the PeerAd entry via concatenating
// "/service/N" where N is the index of the entry. A zero value at an index
// signals to stop scanning for more subsequent values.
type Ad struct {
	adproto.Ad
	Port      uint16
	RelayRate uint32
}

var _ coding.Codec = &Ad{}

// NewServiceAd ...
func NewServiceAd(id nonce.ID, key *crypto.Prv, relayRate uint32, port uint16,
	expiry time.Time, ) (sv *Ad) {

	s := splice.New(adintro.Len)
	k := crypto.DerivePub(key)
	ServiceSplice(s, id, k, relayRate, port, expiry)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	sv = &Ad{
		Ad: adproto.Ad{
			ID:     id,
			Key:    k,
			Expiry: time.Now().Add(adproto.TTL),
			Sig:    sign,
		},
		RelayRate: relayRate,
		Port:      port,
	}
	return
}

func (x *Ad) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint16(&x.Port).
		ReadUint32(&x.RelayRate).
		ReadTime(&x.Expiry).
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
	ServiceSplice(s, x.ID, x.Key, x.RelayRate, x.Port, x.Expiry)
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

func ServiceSplice(s *splice.Splice, id nonce.ID, key *crypto.Pub, relayRate uint32, port uint16, expiry time.Time, ) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint16(port).
		Uint32(relayRate).
		Time(expiry)
}

func init() { reg.Register(Magic, serviceAdGen) }

func serviceAdGen() coding.Codec { return &Ad{} }
