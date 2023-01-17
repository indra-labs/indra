package onion

import (
	"net/netip"
	"time"

	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/lnd/lnwire"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/balance"
	"github.com/indra-labs/indra/pkg/wire/confirm"
	"github.com/indra-labs/indra/pkg/wire/delay"
	"github.com/indra-labs/indra/pkg/wire/exit"
	"github.com/indra-labs/indra/pkg/wire/forward"
	"github.com/indra-labs/indra/pkg/wire/getbalance"
	"github.com/indra-labs/indra/pkg/wire/layer"
	"github.com/indra-labs/indra/pkg/wire/noop"
	"github.com/indra-labs/indra/pkg/wire/response"
	"github.com/indra-labs/indra/pkg/wire/reverse"
	"github.com/indra-labs/indra/pkg/wire/session"
	"github.com/indra-labs/indra/pkg/wire/token"
)

type Skins []types.Onion

var os = &noop.OnionSkin{}

func (o Skins) ForwardLayer(s *traffic.Session, k *prv.Key,
	n nonce.IV) Skins {

	return o.Forward(s.Peer.AddrPort).Layer(s.HeaderPub, k, n)
}

func (o Skins) ReverseLayer(s *traffic.Session, k *prv.Key,
	n nonce.IV) Skins {

	return o.Reverse(s.Peer.AddrPort).Layer(s.HeaderPub, k, n)
}

func (o Skins) ForwardSession(s *traffic.Session,
	k *prv.Key, n nonce.IV, hdr, pld *prv.Key) Skins {

	if hdr == nil || pld == nil {
		return o.Forward(s.Peer.AddrPort).Layer(s.HeaderPub, k, n)
	} else {
		return o.Forward(s.Peer.AddrPort).
			Layer(s.Peer.IdentityPub, k, n).
			Session(hdr, pld)
	}
}

func (o Skins) Balance(id nonce.ID,
	amt lnwire.MilliSatoshi) Skins {

	return append(o, &balance.OnionSkin{
		ID:           id,
		MilliSatoshi: amt,
	})
}

func (o Skins) Confirmation(id nonce.ID) Skins {
	return append(o, &confirm.OnionSkin{ID: id})
}

func (o Skins) Delay(d time.Duration) Skins {
	return append(o, &delay.OnionSkin{Duration: d, Onion: os})
}

func (o Skins) Exit(port uint16, prvs [3]*prv.Key, pubs [3]*pub.Key,
	nonces [3]nonce.IV, payload slice.Bytes) Skins {

	return append(o, &exit.OnionSkin{
		Port:    port,
		Ciphers: GenCiphers(prvs, pubs),
		Bytes:   payload,
		Nonces:  nonces,
		Onion:   os,
	})
}

func (o Skins) Forward(addr *netip.AddrPort) Skins {
	return append(o,
		&forward.OnionSkin{
			AddrPort: addr,
			Onion:    &noop.OnionSkin{},
		})
}

func (o Skins) GetBalance(id nonce.ID, prvs [3]*prv.Key,
	pubs [3]*pub.Key, nonces [3]nonce.IV) Skins {

	return append(o, &getbalance.OnionSkin{
		ID:      id,
		Ciphers: GenCiphers(prvs, pubs),
		Nonces:  nonces,
		Onion:   os,
	})
}

func (o Skins) Layer(to *pub.Key, from *prv.Key,
	n nonce.IV) Skins {

	return append(o, &layer.OnionSkin{
		To:    to,
		From:  from,
		Nonce: n,
		Onion: os,
	})
}

func (o Skins) Reverse(ip *netip.AddrPort) Skins {
	return append(o, &reverse.OnionSkin{AddrPort: ip, Onion: os})
}

func (o Skins) Response(hash sha256.Hash, res slice.Bytes) Skins {
	rs := response.OnionSkin{Hash: hash, Bytes: res}
	return append(o, &rs)
}

func (o Skins) Session(hdr, pld *prv.Key) Skins {
	// SendKeys can apply to from 1 to 5 nodes, if either key is nil then
	// this layer just doesn't get added in the serialization process.
	if hdr == nil || pld == nil {
		return o
	}
	return append(o, &session.OnionSkin{
		Header:  hdr,
		Payload: pld,
		Onion:   &noop.OnionSkin{},
	})
}

func (o Skins) Token(tok sha256.Hash) Skins {
	return append(o, (*token.OnionSkin)(&tok))
}

// Assemble inserts the slice of OnionSkin s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o Skins) Assemble() (on types.Onion) {
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
