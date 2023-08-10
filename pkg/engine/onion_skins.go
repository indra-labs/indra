package engine

import (
	"git.indra-labs.org/dev/ind/pkg/codec/ad/intro"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/confirmation"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/end"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/response"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/crypt"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/exit"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/forward"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/getbalance"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/hidden/introquery"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/hidden/ready"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/hidden/route"
	hiddenservice "git.indra-labs.org/dev/ind/pkg/codec/onion/hidden/services"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/hidden/whisper"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/reverse"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/session"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/engine/node"
	"git.indra-labs.org/dev/ind/pkg/engine/protocols"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	headers2 "git.indra-labs.org/dev/ind/pkg/headers"
	"git.indra-labs.org/dev/ind/pkg/hidden"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"github.com/multiformats/go-multiaddr"
)

// func (o Skins) Balance(id nonce.ID, amt lnwire.MilliSatoshi) Skins {
//	return append(o, balance.NewBalance(id, amt))
// }

func (o Skins) Confirmation(id nonce.ID, load byte) Skins {
	return append(o, confirmation.New(id))
}

// func (o Skins) Delay(d time.Duration) Skins { return append(o, delay.NewDelay(d)) }

type (
	// Skins is a slice of onions that can be assembled into a message.
	//
	// Note that if an address is not found for the enabled protocols the segment
	// will be omitted from the chain, and report the error. It is thus expected that
	// sessions and nodes referenced in a Skins constructor were selected because
	// they have an available protocol.
	//
	// It would be very uncommon to not have at least IPv4 anyway, but we handle this
	// in an ugly way that should never see the light outside a dev lab.
	Skins []ont.Onion

	// RoutingLayer is the reverse and crypt message layers used in a RoutingHeader.
	RoutingLayer struct {
		*reverse.Reverse
		*crypt.Crypt
	}

	// RoutingHeader is a stack of 3 RoutingLayer used to construct replies in a
	// source routed architecture.
	RoutingHeader struct {
		Layers [3]RoutingLayer
	}
)

// End is a terminator, construction of a message will halt when it hits one of
// these.
func (o Skins) End() Skins {
	return append(o, &end.End{})
}

// Exit inserts an exit message with its request payload.
func (o Skins) Exit(id nonce.ID, port uint16, payload slice.Bytes,
	ep *exit.ExitPoint) Skins {

	return append(o, exit.New(id, port, payload, ep))
}

// Forward is a simple forwarding instruction, usually followed by a crypt for the recipient.
func (o Skins) Forward(addr multiaddr.Multiaddr) Skins {
	return append(o, forward.New(addr))
}

// ForwardCrypt is a forwarding and encryption layer for simple forwarding of a
// message. Used with hidden service messages and pings for legs of the route
// that do not need to carry a reply.
func (o Skins) ForwardCrypt(s *sessions.Data, k *crypto.Prv, n nonce.IV, p protocols.NetworkProtocols) Skins {
	ma := s.Node.PickAddress(p)
	return o.Forward(ma).Crypt(s.Header.Pub, s.Payload.Pub, k,
		n, 0)
}

// ForwardSession is an onion layer that delivers the two session keys matching a
// payment preimage to establish a new session.
func (o Skins) ForwardSession(s *node.Node,
	k *crypto.Prv, n nonce.IV, sess *session.Session,
	p protocols.NetworkProtocols) Skins {

	ma := s.PickAddress(p)
	return o.Forward(ma).
		Crypt(s.Identity.Pub, nil, k, n, 0).
		Session(sess)
}

// GetBalance is a message request to reply using a RoutingHeader with the
// balance of the session, indicated by the key of the Crypt in the previous
// layer.
func (o Skins) GetBalance(id nonce.ID, ep *exit.ExitPoint) Skins {
	return append(o, getbalance.NewGetBalance(id, ep))
}

// HiddenService is a message that delivers an intro and a referral RoutingHeader
// to enable a relay to introduce a client to a hidden service.
func (o Skins) HiddenService(in *intro.Ad, point *exit.ExitPoint) Skins {
	return append(o, hiddenservice.New(in, point))
}

// IntroQuery is a message that carries a query for a hidden service with a given
// public key, if the recipient has this intro, it returns it and the client can
// then form a Route message to establish a connection to it.
func (o Skins) IntroQuery(id nonce.ID, hsk *crypto.Pub, exit *exit.ExitPoint) Skins {
	return append(o, introquery.New(id, hsk, exit))
}

// Reverse is a special variant of the Forward message that is only used in
// groups of three to create a RoutingHeader;
func (o Skins) Reverse(ip multiaddr.Multiaddr) Skins {
	return append(o, reverse.New(ip))
}

// Crypt is a layer providing the recipient cloaked address, one-time sender
// private key and nonce needed to decrypt the remainder of the onion message.
func (o Skins) Crypt(toHdr, toPld *crypto.Pub, from *crypto.Prv, iv nonce.IV,
	depth int) Skins {

	return append(o, crypt.New(toHdr, toPld, from, iv, depth))
}

// ReverseCrypt is a single layer of a RoutingHeader designating the session and relay for one of the hops in a Route.
func (o Skins) ReverseCrypt(s *sessions.Data, k *crypto.Prv, n nonce.IV,
	seq int, p protocols.NetworkProtocols) (oo Skins) {

	if s == nil || k == nil {
		oo = append(o, &reverse.Reverse{})
		oo = append(oo, &crypt.Crypt{})
		return
	}
	ma := s.Node.PickAddress(p)
	return o.Reverse(ma).Crypt(s.Header.Pub, s.Payload.Pub, k, n,
		seq)
}

// RoutingHeader constructs a RoutingHeader using inputs from an exit.Routing.
// These are used in all hidden service messages too as the path back to both
// ends of the path provided by the recipient at it.
func (o Skins) RoutingHeader(r *exit.Routing,
	p protocols.NetworkProtocols) Skins {
	return o.
		ReverseCrypt(r.Sessions[0], r.Keys[0], r.Nonces[0], 3, p).
		ReverseCrypt(r.Sessions[1], r.Keys[1], r.Nonces[1], 2, p).
		ReverseCrypt(r.Sessions[2], r.Keys[2], r.Nonces[2], 1, p)
}

// MakeExit constructs a message containing an arbitrary payload to a node (3rd
// hop) with a set of 3 ciphers derived from the hidden PayloadPub of the return
// hops that are layered progressively after the Exit message.
//
// The Exit node forwards the packet it receives to the local port specified in
// the Exit message, and then uses the ciphers to encrypt the reply with the
// three ciphers provided, which don't enable it to decrypt the header, only to
// encrypt the payload.
//
// The header remains a constant size and each node in the Reverse trims off
// their section at the top, moves the next crypt header to the top and pads the
// remainder with noise, so it always looks like the first hop.
func MakeExit(p exit.Params, np protocols.NetworkProtocols) Skins {
	headers := headers2.GetHeaders(p.Alice, p.Bob, p.S, p.KS)
	return Skins{}.
		RoutingHeader(headers.Forward, np).
		Exit(p.ID, p.Port, p.Payload, headers.ExitPoint()).
		RoutingHeader(headers.Return, np)
}

// MakeGetBalance sends out a request in a similar way to Exit except the node
// being queried can be any of the 5.
func MakeGetBalance(p getbalance.GetBalanceParams,
	np protocols.NetworkProtocols) Skins {
	headers := headers2.GetHeaders(p.Alice, p.Bob, p.S, p.KS)
	return Skins{}.
		RoutingHeader(headers.Forward, np).
		GetBalance(p.ID, headers.ExitPoint()).
		RoutingHeader(headers.Return, np)
}

// MakeHiddenService constructs a services message for designating
// introducers to refer clients to hidden services.
func MakeHiddenService(in *intro.Ad, alice, bob *sessions.Data,
	c sessions.Circuit, ks *crypto.KeySet,
	p protocols.NetworkProtocols) Skins {

	headers := headers2.GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward, p).
		HiddenService(in, headers.ExitPoint()).
		RoutingHeader(headers.Return, p)
}

// MakeIntroQuery constructs a message to query a peer for an intro for a designated hidden service address.
func MakeIntroQuery(id nonce.ID, hsk *crypto.Pub, alice, bob *sessions.Data,
	c sessions.Circuit, ks *crypto.KeySet,
	p protocols.NetworkProtocols) Skins {

	headers := headers2.GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward, p).
		IntroQuery(id, hsk, headers.ExitPoint()).
		RoutingHeader(headers.Return, p)
}

// MakeRoute constructs a Route message, which a client sends a RoutingHeader to
// a hidden service for them to reply with a Ready message for establishing a
// session with a hidden service.
func MakeRoute(id nonce.ID, k *crypto.Pub, ks *crypto.KeySet,
	alice, bob *sessions.Data, c sessions.Circuit,
	p protocols.NetworkProtocols) Skins {

	headers := headers2.GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward, p).
		Route(id, k, ks, headers.ExitPoint()).
		RoutingHeader(headers.Return, p)
}

// MakeSession constructs a set of 5 ForwardSession messages to deliver the
// session keys to newly established sessions paid for via LN.
func MakeSession(id nonce.ID, s [5]*session.Session,
	client *sessions.Data, hop []*node.Node, ks *crypto.KeySet,
	p protocols.NetworkProtocols) Skins {

	n := crypto.GenNonces(6)
	sk := Skins{}
	for i := range s {
		if i == 0 {
			sk = sk.Crypt(hop[i].Identity.Pub, nil, ks.Next(),
				n[i], 0).Session(s[i])
		} else {
			sk = sk.ForwardSession(hop[i], ks.Next(), n[i], s[i], p)
		}
	}
	return sk.
		ForwardCrypt(client, ks.Next(), n[5], p).
		Confirmation(id, 0)
}

// Message constructs a hidden service message, used on both ends, with the
// RoutingHeader received in the previosu message from the client or hidden
// service.
func (o Skins) Message(msg *whisper.Message, ks *crypto.KeySet,
	p protocols.NetworkProtocols) Skins {

	return append(o.
		ForwardCrypt(msg.Forwards[0], ks.Next(), nonce.New(), p).
		ForwardCrypt(msg.Forwards[1], ks.Next(), nonce.New(), p),
		msg)
}

// Ping is a message which checks the liveness of relays by ensuring they are
// correctly relaying messages.
//
// The pending ping records keep the identifiers of the 5 nodes that were in
// a ping onion and when the Confirmation is correctly received these nodes get
// an increment of their liveness score. By using this scheme, when nodes are
// offline their scores will fall to zero after a time whereas live nodes will
// have steadily increasing scores from successful pings.
func Ping(id nonce.ID, client *sessions.Data, s sessions.Circuit,
	ks *crypto.KeySet, p protocols.NetworkProtocols) Skins {

	n := crypto.GenPingNonces()
	return Skins{}.
		Crypt(s[0].Header.Pub, nil, ks.Next(), n[0], 0).
		ForwardCrypt(s[1], ks.Next(), n[1], p).
		ForwardCrypt(s[2], ks.Next(), n[2], p).
		ForwardCrypt(s[3], ks.Next(), n[3], p).
		ForwardCrypt(s[4], ks.Next(), n[4], p).
		ForwardCrypt(client, ks.Next(), n[5], p).
		Confirmation(id, 0)
}

// Ready is a message carrying a reply RoutingHeader for a client to send their
// first request message to a hidden service.
func (o Skins) Ready(id nonce.ID, addr *crypto.Pub, fwHdr,
	rvHdr hidden.RoutingHeaderBytes, fc, rc crypto.Ciphers, fn, rn crypto.Nonces) Skins {

	return append(o, ready.New(id, addr, fwHdr, rvHdr, fc, rc, fn, rn))
}

// Response constructs a message sent back from an Exit service to a client in
// response to an Exit message, containing a request.
func (o Skins) Response(id nonce.ID, res slice.Bytes, port uint16, load byte) Skins {
	return append(o, response.New(id, port, res, load))
}

// Route constructs a new Route message to establish a hidden service connection.
func (o Skins) Route(id nonce.ID, k *crypto.Pub, ks *crypto.KeySet, ep *exit.ExitPoint) Skins {
	return append(o, route.New(id, k, ks, ep))
}

// Session generates a message layer carrying the two private keys of a session.
func (o Skins) Session(sess *session.Session) Skins {
	//	MakeSession can apply to from 1 to 5 nodes, if either key is nil then
	//	this crypt just doesn't get added in the serialization process.
	if sess.Header == nil || sess.Payload == nil {
		return o
	}
	return append(o, sess)
}
