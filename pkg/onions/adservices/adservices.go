package adservices

import (
	"fmt"
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
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic      = "svad"
	ServiceLen = slice.Uint16Len + slice.Uint32Len
)

type Service struct {
	Port      uint16
	RelayRate uint32
}

// Ad stores a specification for the fee rate and the service port, which
// must be a well known port to match with a type of service, eg 80 for web, 53
// for DNS, etc. These are also attached to the PeerAd entry via concatenating
// "/service/N" where N is the index of the entry. A zero value at an index
// signals to stop scanning for more subsequent values.
type Ad struct {
	adproto.Ad
	Services []Service
}

var _ coding.Codec = &Ad{}

// NewServiceAd ...
func NewServiceAd(id nonce.ID, key *crypto.Prv, services []Service,
	expiry time.Time) (sv *Ad) {

	s := splice.New(adintro.Len)
	k := crypto.DerivePub(key)
	ServiceSplice(s, id, k, services, expiry)
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
		Services: services,
	}
	return
}

func (x *Ad) Decode(s *splice.Splice) (e error) {
	var i, count uint32
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint32(&count)
	x.Services = make([]Service, count)
	for ; i < count; i++ {
		s.ReadUint16(&x.Services[i].Port).
			ReadUint32(&x.Services[i].RelayRate)
	}
	s.
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

func (x *Ad) Len() int { return adproto.Len + len(x.Services)*ServiceLen + slice.Uint32Len }

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
	ServiceSplice(s, x.ID, x.Key, x.Services, x.Expiry)
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

func ServiceSplice(s *splice.Splice, id nonce.ID, key *crypto.Pub, services []Service, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint32(uint32(len(services)))
	for i := range services {
		s.
			Uint16(services[i].Port).
			Uint32(services[i].RelayRate)
	}
	s.Time(expiry)
}

func init() { reg.Register(Magic, serviceAdGen) }

func serviceAdGen() coding.Codec { return &Ad{} }
