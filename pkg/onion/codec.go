package onion

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/key/ecdh"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/onion/layers/balance"
	"github.com/indra-labs/indra/pkg/onion/layers/confirm"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/delay"
	"github.com/indra-labs/indra/pkg/onion/layers/directbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/exit"
	"github.com/indra-labs/indra/pkg/onion/layers/forward"
	"github.com/indra-labs/indra/pkg/onion/layers/getbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/magicbytes"
	"github.com/indra-labs/indra/pkg/onion/layers/response"
	"github.com/indra-labs/indra/pkg/onion/layers/reverse"
	"github.com/indra-labs/indra/pkg/onion/layers/session"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func Encode(on types.Onion) (b slice.Bytes) {
	b = make(slice.Bytes, on.Len())
	var sc slice.Cursor
	c := &sc
	on.Encode(b, c)
	return
}

func Peel(b slice.Bytes, c *slice.Cursor) (on types.Onion, e error) {
	switch b[*c:c.Inc(magicbytes.Len)].String() {
	case balance.MagicString:
		on = &balance.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case confirm.MagicString:
		on = &confirm.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case crypt.MagicString:
		var o crypt.Layer
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case delay.MagicString:
		on = &delay.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case directbalance.MagicString:
		on = &directbalance.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case exit.MagicString:
		on = &exit.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case forward.MagicString:
		on = &forward.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case getbalance.MagicString:
		on = &getbalance.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case reverse.MagicString:
		on = &reverse.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case response.MagicString:
		on = response.New()
		if e = on.Decode(b, c); check(e) {
			return
		}
	case session.MagicString:
		on = &session.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	default:
		e = fmt.Errorf("message magic not found")
		log.T.C(func() string {
			return fmt.Sprintln(e) + spew.Sdump(b.ToBytes())
		})
		return
	}
	return
}

func GenCiphers(prvs [3]*prv.Key, pubs [3]*pub.Key) (ciphers [3]sha256.Hash) {
	for i := range prvs {
		ciphers[2-i] = ecdh.Compute(prvs[i], pubs[i])
	}
	return
}

func Gen3Nonces() (n [3]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GenNonces(count int) (n []nonce.IV) {
	n = make([]nonce.IV, count)
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GenPingNonces() (n [6]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}
