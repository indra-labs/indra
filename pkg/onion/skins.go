package onion

import (
	"net/netip"
	"time"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/node"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/delay"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/directbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/noop"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/response"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/reverse"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/session"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Skins []types.Onion

var os = &noop.Layer{}

func (o Skins) ForwardCrypt(s *traffic.Session, k *prv.Key,
	n nonce.IV) Skins {

	return o.Forward(s.Peer.AddrPort).Crypt(s.HeaderPub, s.PayloadPub, k, n, 0)
}

func (o Skins) ReverseCrypt(s *traffic.Session, k *prv.Key, n nonce.IV, seq int) Skins {

	return o.Reverse(s.Peer.AddrPort).Crypt(s.HeaderPub, s.PayloadPub, k, n, seq)
}

func (o Skins) ForwardSession(s *node.Node,
	k *prv.Key, n nonce.IV, sess *session.Layer) Skins {

	return o.Forward(s.Peer.AddrPort).
		Crypt(s.Peer.IdentityPub, nil, k, n, 0).
		Session(sess)
}

func (o Skins) Balance(id, confID nonce.ID,
	amt lnwire.MilliSatoshi) Skins {

	return append(o, &balance.Layer{
		ID:           id,
		ConfID:       confID,
		MilliSatoshi: amt,
	})
}

func (o Skins) Confirmation(id nonce.ID) Skins {
	return append(o, &confirm.Layer{ID: id})
}

func (o Skins) Delay(d time.Duration) Skins {
	return append(o, &delay.Layer{Duration: d, Onion: os})
}

func (o Skins) DirectBalance(id, confID nonce.ID) Skins {
	return append(o, &directbalance.Layer{ID: id, ConfID: confID, Onion: os})
}

func (o Skins) Exit(port uint16, prvs [3]*prv.Key, pubs [3]*pub.Key,
	nonces [3]nonce.IV, id nonce.ID, payload slice.Bytes) Skins {

	return append(o, &exit.Layer{
		Port:    port,
		Ciphers: GenCiphers(prvs, pubs),
		Nonces:  nonces,
		ID:      id,
		Bytes:   payload,
		Onion:   os,
	})
}

func (o Skins) Forward(addr *netip.AddrPort) Skins {
	return append(o,
		&forward.Layer{
			AddrPort: addr,
			Onion:    &noop.Layer{},
		})
}

func (o Skins) GetBalance(id, confID nonce.ID, prvs [3]*prv.Key,
	pubs [3]*pub.Key, nonces [3]nonce.IV) Skins {

	return append(o, &getbalance.Layer{
		ID:      id,
		ConfID:  confID,
		Ciphers: GenCiphers(prvs, pubs),
		Nonces:  nonces,
		Onion:   os,
	})
}

func (o Skins) Crypt(toHdr, toPld *pub.Key, from *prv.Key, n nonce.IV, seq int) Skins {

	return append(o, &crypt.Layer{
		Seq:          seq,
		ToHeaderPub:  toHdr,
		ToPayloadPub: toPld,
		From:         from,
		Nonce:        n,
		Onion:        os,
	})
}

func (o Skins) Reverse(ip *netip.AddrPort) Skins {
	return append(o, &reverse.Layer{AddrPort: ip, Onion: os})
}

func (o Skins) Response(id nonce.ID, res slice.Bytes, port uint16) Skins {
	rs := response.Layer{ID: id, Port: port, Bytes: res}
	return append(o, &rs)
}

func (o Skins) Session(sess *session.Layer) Skins {
	// SendKeys can apply to from 1 to 5 nodes, if either key is nil then
	// this crypt just doesn't get added in the serialization process.
	if sess.Header == nil || sess.Payload == nil {
		return o
	}
	return append(o, &session.Layer{
		Header:  sess.Header,
		Payload: sess.Payload,
		Onion:   &noop.Layer{},
	})
}

// Assemble inserts the slice of Layer s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o Skins) Assemble() (on types.Onion) {
	// First item is the outer crypt.
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
