package intro

import (
	"net"
	"net/netip"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/sig"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/messages/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	MagicString = "in"
	AddrLen     = net.IPv6len + 3
	Len         = magicbytes.Len + pub.KeyLen + AddrLen + sig.Len
)

var (
	Magic = slice.Bytes(MagicString)
)

type Layer struct {
	Key      *pub.Key
	AddrPort *netip.AddrPort
	Bytes    sig.Bytes
}

func New(key *prv.Key, ap *netip.AddrPort) (in *Layer) {
	pk := pub.Derive(key)
	bap, _ := ap.MarshalBinary()
	pkb := pk.ToBytes()
	hash := sha256.Single(append(pkb[:], bap...))
	var e error
	var sign sig.Bytes
	if sign, e = sig.Sign(key, hash); check(e) {
		return nil
	}
	in = &Layer{
		Key:      pk,
		AddrPort: ap,
		Bytes:    sign,
	}
	return
}

func (im *Layer) Validate() bool {
	bap, _ := im.AddrPort.MarshalBinary()
	pkb := im.Key.ToBytes()
	hash := sha256.Single(append(pkb[:], bap...))
	key, e := im.Bytes.Recover(hash)
	if check(e) {
		return false
	}
	kb := key.ToBytes()
	if kb.Equals(pkb) {
		return true
	}
	return false
}

func (im *Layer) Insert(o types.Onion) {}
func (im *Layer) Len() int             { return Len }

func (im *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	splice.Splice(b, c).
		Magic(Magic).
		Pubkey(im.Key).
		AddrPort(im.AddrPort).
		Signature(im.Bytes)
	return
}

func (im *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	splice.Splice(b, c).
		ReadPubkey(&im.Key).
		ReadAddrPort(&im.AddrPort).
		ReadSignature(&im.Bytes)
	return
}
