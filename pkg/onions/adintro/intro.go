package adintro

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"net/netip"
	"reflect"
	"time"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	Magic = "inad"
	Len   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen + 1 + 4 +
		splice.AddrLen +
		slice.Uint16Len +
		slice.Uint32Len +
		slice.Uint64Len +
		crypto.SigLen
)

type Ad struct {
	ID        nonce.ID        // Ensures never a repeated signature.
	Key       *crypto.Pub     // Hidden service address.
	AddrPort  *netip.AddrPort // Introducer address.
	Port      uint16          // Well known port of protocol available.
	RelayRate uint32          // mSat/Mb
	Expiry    time.Time
	Sig       crypto.SigBytes
}

var _ coding.Codec = &Ad{}

func (x *Ad) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {

		return
	}
	s.
		ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadAddrPort(&x.AddrPort).
		ReadUint32(&x.RelayRate).
		ReadUint16(&x.Port).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Ad) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.AddrPort.String(), x.Expiry, x.Sig,
	)
	x.Splice(s)
	return
}

func (x *Ad) GetOnion() interface{} { return x }

// Gossip means adding to the node's peer message list which will be gossiped by
// the libp2p network of Indra peers.
func (x *Ad) Gossip(sm *sess.Manager, c qu.C) {
	log.D.F("propagating hidden service intro for %s",
		x.Key.ToBased32Abbreviated())
}

func (x *Ad) Len() int { return Len }

func (x *Ad) Magic() string { return Magic }

func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

func (x *Ad) SpliceNoSig(s *splice.Splice) {
	IntroSplice(s, x.ID, x.Key, x.AddrPort, x.RelayRate, x.Port, x.Expiry)
}

func (x *Ad) Validate() bool {
	s := splice.New(Len - magic.Len)
	x.SpliceNoSig(s)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	key, e := x.Sig.Recover(hash)
	if fails(e) {
		return false
	}
	if key.Equals(x.Key) && x.Expiry.After(time.Now()) {
		return true
	}
	return false
}

func IntroSplice(
	s *splice.Splice,
	id nonce.ID,
	key *crypto.Pub,
	ap *netip.AddrPort,
	relayRate uint32,
	port uint16,
	expires time.Time,
) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		AddrPort(ap).
		Uint32(relayRate).
		Uint16(port).
		Time(expires)
}

func NewIntroAd(
	id nonce.ID,
	key *crypto.Prv,
	ap *netip.AddrPort,
	relayRate uint32,
	port uint16,
	expires time.Time,
) (in *Ad) {

	pk := crypto.DerivePub(key)
	s := splice.New(Len)
	IntroSplice(s, id, pk, ap, relayRate, port, expires)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	in = &Ad{
		ID:        id,
		Key:       pk,
		AddrPort:  ap,
		RelayRate: relayRate,
		Port:      port,
		Expiry:    expires,
		Sig:       sign,
	}
	return
}

func init() { reg.Register(Magic, introGen) }

func introGen() coding.Codec { return &Ad{} }
