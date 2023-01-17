package session

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/lnd/lnwire"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/payment"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
	"github.com/indra-labs/indra/pkg/wire/noop"
)

const (
	MagicString = "ss"
	Len         = magicbytes.Len + prv.KeyLen*2
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin session delivers a pair of private keys to a relay that represent
// the preimage referred to in a Lightning payment for a session.
//
// The preimage is hashed by the buyer to use in the payment, and when the relay
// receives it, it can then hash the two private keys to match it with the
// payment preimage hash, which proves the buyer paid, and simultaneously
// provides the session keys that both forward and reverse messages will use,
// each in different ways.
//
// Exit nodes are provided the ciphers to encrypt for the three hops back to the
// client, but they do not have either public or private part of the Header or
// Payload keys of the return hops, which are used to conceal their respective
// message sections.
//
// Thus, they cannot decrypt the header, but they can encrypt the payload with
// the three layers of encryption that the reverse path hops have the private
// keys to decrypt. By this, the path in the header is concealed and the payload
// is concealed to the hops except for the encryption layer they decrypt using
// their Payload key, delivered in this message.
type OnionSkin struct {
	Hop             byte
	Header, Payload *prv.Key
	types.Onion
}

func New() (x *OnionSkin) {
	var e error
	var hdrPrv, pldPrv *prv.Key
	if hdrPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	if pldPrv, e = prv.GenerateKey(); check(e) {
		return
	}

	return &OnionSkin{
		Header:  hdrPrv,
		Payload: pldPrv,
		Onion:   &noop.OnionSkin{},
	}
}

func (x *OnionSkin) PreimageHash() sha256.Hash {
	h, p := x.Header.ToBytes(), x.Payload.ToBytes()
	return sha256.Single(append(h[:], p[:]...))
}

func (x *OnionSkin) ToPayment(amount lnwire.
	MilliSatoshi) (p *payment.Payment) {

	p = &payment.Payment{
		ID:       nonce.NewID(),
		Preimage: x.PreimageHash(),
		Amount:   amount,
	}
	return
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int             { return Len + x.Onion.Len() }
func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	hdr := x.Header.ToBytes()
	pld := x.Payload.ToBytes()
	copy(b[*c:c.Inc(prv.KeyLen)], hdr[:])
	copy(b[*c:c.Inc(prv.KeyLen)], pld[:])
	x.Onion.Encode(b, c)
}
func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, string(Magic))
	}
	x.Header = prv.PrivkeyFromBytes(b[*c:c.Inc(prv.KeyLen)])
	x.Payload = prv.PrivkeyFromBytes(b[*c:c.Inc(prv.KeyLen)])
	return
}
