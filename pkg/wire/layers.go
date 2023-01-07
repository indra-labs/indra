package wire

import (
	"net/netip"
	"time"

	"github.com/indra-labs/indra/pkg/key/address"
	"github.com/indra-labs/indra/pkg/key/ecdh"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/cipher"
	"github.com/indra-labs/indra/pkg/wire/confirm"
	"github.com/indra-labs/indra/pkg/wire/delay"
	"github.com/indra-labs/indra/pkg/wire/exit"
	"github.com/indra-labs/indra/pkg/wire/forward"
	"github.com/indra-labs/indra/pkg/wire/layer"
	"github.com/indra-labs/indra/pkg/wire/noop"
	"github.com/indra-labs/indra/pkg/wire/purchase"
	"github.com/indra-labs/indra/pkg/wire/response"
	"github.com/indra-labs/indra/pkg/wire/reverse"
	"github.com/indra-labs/indra/pkg/wire/session"
	"github.com/indra-labs/indra/pkg/wire/token"
)

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

func GenPingNonces() (n [4]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

type OnionSkins []types.Onion

var os = &noop.OnionSkin{}

func (o OnionSkins) Cipher(hdr, pld *prv.Key) OnionSkins {
	return append(o, &cipher.OnionSkin{
		Header:  hdr,
		Payload: pld,
		Onion:   &noop.OnionSkin{},
	})
}

func (o OnionSkins) Confirmation(id nonce.ID) OnionSkins {
	return append(o, &confirm.OnionSkin{ID: id})
}

func (o OnionSkins) Delay(d time.Duration) OnionSkins {
	return append(o, &delay.OnionSkin{Duration: d, Onion: os})
}

func (o OnionSkins) Exit(port uint16, prvs [3]*prv.Key, pubs [3]*pub.Key,
	nonces [3]nonce.IV, payload slice.Bytes) OnionSkins {

	return append(o, &exit.OnionSkin{
		Port:    port,
		Ciphers: GenCiphers(prvs, pubs),
		Bytes:   payload,
		Nonces:  nonces,
		Onion:   os,
	})
}

func (o OnionSkins) Forward(addr *netip.AddrPort) OnionSkins {
	return append(o, &forward.OnionSkin{AddrPort: addr, Onion: &noop.OnionSkin{}})
}

func (o OnionSkins) OnionSkin(to *address.Sender, from *prv.Key, n nonce.IV) OnionSkins {
	return append(o, &layer.OnionSkin{
		To:    to,
		From:  from,
		Nonce: n,
		Onion: os,
	})
}
func (o OnionSkins) Purchase(id nonce.ID, nBytes uint64, prvs [3]*prv.Key,
	pubs [3]*pub.Key, n [3]nonce.IV) OnionSkins {

	oo := append(o, &purchase.OnionSkin{
		ID:      id,
		NBytes:  nBytes,
		Ciphers: GenCiphers(prvs, pubs),
		Nonces:  n,
		Onion:   os,
	})

	return oo
}
func (o OnionSkins) Reverse(ip *netip.AddrPort) OnionSkins {
	return append(o, &reverse.OnionSkin{AddrPort: ip, Onion: os})
}
func (o OnionSkins) Response(hash sha256.Hash, res slice.Bytes) OnionSkins {
	rs := response.OnionSkin{Hash: hash, Bytes: res}
	return append(o, &rs)
}
func (o OnionSkins) Session(hdr, pld *pub.Key) OnionSkins {
	return append(o, &session.OnionSkin{
		HeaderKey:  hdr,
		PayloadKey: pld,
		Onion:      os,
	})
}
func (o OnionSkins) Token(tok sha256.Hash) OnionSkins {
	return append(o, (*token.OnionSkin)(&tok))
}

// Assemble inserts the slice of OnionSkin s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o OnionSkins) Assemble() (on types.Onion) {
	// First item is the outer layer.
	on = o[0]
	// Iterate through the remaining layers.
	for _, oc := range o[1:] {
		on.Insert(oc)
		// Next step we are inserting inside the one we just inserted.
		on = oc
	}
	// At the end, the first element contains references to every element
	// inside it.
	return o[0]
}
