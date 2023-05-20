package onions

import (
	"net/netip"
	"time"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/node"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

type Skins []Onion

var nop = &End{}

func Encode(on Onion) (s *splice.Splice) {
	s = splice.New(on.Len())
	fails(on.Encode(s))
	return
}

// Assemble inserts the slice of Layer s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o Skins) Assemble() (on Onion) {
	// First item is the outer crypt.
	on = o[0]
	// Iterate through the remaining layers.
	for _, oc := range o[1:] {
		on.Wrap(oc)
		// Next step we are inserting inside the one we just inserted.
		on = oc
	}
	// At the end, the first element contains references to every element
	// inside it.
	return o[0]
}

func (o Skins) ForwardCrypt(s *sessions.Data, k *crypto.Prv, n nonce.IV) Skins {
	return o.Forward(s.Node.AddrPort).Crypt(s.Header.Pub, s.Payload.Pub, k,
		n, 0)
}

func (o Skins) ReverseCrypt(s *sessions.Data, k *crypto.Prv, n nonce.IV,
	seq int) (oo Skins) {

	if s == nil || k == nil {
		oo = append(o, &Reverse{})
		oo = append(oo, &Crypt{})
		return
	}
	return o.Reverse(s.Node.AddrPort).Crypt(s.Header.Pub, s.Payload.Pub, k, n,
		seq)
}

type Routing struct {
	Sessions [3]*sessions.Data
	Keys     crypto.Privs
	crypto.Nonces
}

type Headers struct {
	Forward, Return *Routing
	ReturnPubs      crypto.Pubs
}

func GetHeaders(alice, bob *sessions.Data, c sessions.Circuit,
	ks *crypto.KeySet) (h *Headers) {

	fwKeys := ks.Next3()
	rtKeys := ks.Next3()
	n := crypto.GenNonces(6)
	var rtNonces, fwNonces [3]nonce.IV
	copy(fwNonces[:], n[:3])
	copy(rtNonces[:], n[3:])
	var fwSessions, rtSessions [3]*sessions.Data
	copy(fwSessions[:], c[:2])
	fwSessions[2] = bob
	copy(rtSessions[:], c[3:])
	rtSessions[2] = alice
	var returnPubs crypto.Pubs
	returnPubs[0] = c[3].Payload.Pub
	returnPubs[1] = c[4].Payload.Pub
	returnPubs[2] = alice.Payload.Pub
	h = &Headers{
		Forward: &Routing{
			Sessions: fwSessions,
			Keys:     fwKeys,
			Nonces:   fwNonces,
		},
		Return: &Routing{
			Sessions: rtSessions,
			Keys:     rtKeys,
			Nonces:   rtNonces,
		},
		ReturnPubs: returnPubs,
	}
	return
}

type ExitPoint struct {
	*Routing
	ReturnPubs crypto.Pubs
}

func (h *Headers) ExitPoint() *ExitPoint {
	return &ExitPoint{
		Routing:    h.Return,
		ReturnPubs: h.ReturnPubs,
	}
}

func (o Skins) RoutingHeader(r *Routing) Skins {
	return o.
		ReverseCrypt(r.Sessions[0], r.Keys[0], r.Nonces[0], 3).
		ReverseCrypt(r.Sessions[1], r.Keys[1], r.Nonces[1], 2).
		ReverseCrypt(r.Sessions[2], r.Keys[2], r.Nonces[2], 1)
}

func (o Skins) ForwardSession(s *node.Node,
	k *crypto.Prv, n nonce.IV, sess *Session) Skins {

	return o.Forward(s.AddrPort).
		Crypt(s.Identity.Pub, nil, k, n, 0).
		Session(sess)
}

func (o Skins) Balance(id nonce.ID, amt lnwire.MilliSatoshi) Skins {

	return append(o, &Balance{ID: id, MilliSatoshi: amt})
}

func (o Skins) Confirmation(id nonce.ID, load byte) Skins {
	return append(o, &Confirmation{ID: id, Load: load})
}

func (o Skins) Crypt(toHdr, toPld *crypto.Pub, from *crypto.Prv, iv nonce.IV,
	depth int) Skins {

	return append(o, &Crypt{
		Depth:        depth,
		ToHeaderPub:  toHdr,
		ToPayloadPub: toPld,
		From:         from,
		IV:           iv,
		Onion:        nop,
	})
}

func (o Skins) Delay(d time.Duration) Skins {
	return append(o, &Delay{Duration: d, Onion: nop})
}

func (o Skins) Exit(id nonce.ID, port uint16, payload slice.Bytes,
	ep *ExitPoint) Skins {

	return append(o, &Exit{
		ID:      id,
		Ciphers: crypto.GenCiphers(ep.Keys, ep.ReturnPubs),
		Nonces:  ep.Nonces,
		Port:    port,
		Bytes:   payload,
		Onion:   nop,
	})
}

func (o Skins) Forward(addr *netip.AddrPort) Skins {
	return append(o, &Forward{AddrPort: addr, Onion: &End{}})
}

func (o Skins) GetBalance(id nonce.ID, ep *ExitPoint) Skins {
	return append(o, &GetBalance{
		ID:      id,
		Ciphers: crypto.GenCiphers(ep.Keys, ep.ReturnPubs),
		Nonces:  ep.Nonces,
		Onion:   nop,
	})
}

func (o Skins) HiddenService(in *Intro, point *ExitPoint) Skins {
	return append(o, &HiddenService{
		Intro:   *in,
		Ciphers: crypto.GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:  point.Nonces,
		Onion:   NewEnd(),
	})
}

func (o Skins) Intro(id nonce.ID, key *crypto.Prv, ap *netip.AddrPort,
	expires time.Time) (sk Skins) {
	return append(o, NewIntro(id, key, ap, 0, 0, expires))
}

func (o Skins) IntroQuery(id nonce.ID, hsk *crypto.Pub, exit *ExitPoint) Skins {
	return append(o, &IntroQuery{
		ID:      id,
		Ciphers: crypto.GenCiphers(exit.Keys, exit.ReturnPubs),
		Nonces:  exit.Nonces,
		Key:     hsk,
		Onion:   nop,
	})
}

func (o Skins) Message(msg *Message, ks *crypto.KeySet) Skins {
	return append(o.
		ForwardCrypt(msg.Forwards[0], ks.Next(), nonce.New()).
		ForwardCrypt(msg.Forwards[1], ks.Next(), nonce.New()),
		msg)
}

func (o Skins) Peer(id nonce.ID, key *crypto.Prv,
	expires time.Time) (sk Skins) {
	return append(o, NewPeer(id, key, expires))
}

func (o Skins) Addr(id nonce.ID, key *crypto.Prv,
	expires time.Time) (sk Skins) {
	return append(o, NewAddr(id, key, expires))
}

func (o Skins) Service(id nonce.ID, key *crypto.Prv,
	expires time.Time) (sk Skins) {
	return append(o, NewService(id, key, expires))
}

func (o Skins) Ready(id nonce.ID, addr *crypto.Pub, fwHeader,
	rvHeader RoutingHeaderBytes,
	fc, rc crypto.Ciphers, fn, rn crypto.Nonces) Skins {
	return append(o, &Ready{id, addr,
		&ReplyHeader{fwHeader, fc, fn},
		&ReplyHeader{rvHeader, rc, rn},
	})
}

func (o Skins) Response(id nonce.ID, res slice.Bytes, port uint16) Skins {
	return append(o, &Response{ID: id, Port: port, Bytes: res})
}

func (o Skins) Reverse(ip *netip.AddrPort) Skins {
	return append(o, &Reverse{AddrPort: ip, Onion: nop})
}

func (o Skins) Route(id nonce.ID, k *crypto.Pub, ks *crypto.KeySet,
	ep *ExitPoint) Skins {

	oo := &Route{
		HiddenService: k,
		Sender:        ks.Next(),
		IV:            nonce.New(),
		ID:            id,
		Ciphers:       crypto.GenCiphers(ep.Keys, ep.ReturnPubs),
		Nonces:        ep.Nonces,
		Onion:         &End{},
	}
	oo.SenderPub = crypto.DerivePub(oo.Sender)
	oo.HiddenCloaked = crypto.GetCloak(k)
	return append(o, oo)
}

func (o Skins) Session(sess *Session) Skins {
	// MakeSession can apply to from 1 to 5 nodes, if either key is nil then
	// this crypt just doesn't get added in the serialization process.
	if sess.Header == nil || sess.Payload == nil {
		return o
	}
	return append(o, &Session{
		Header:  sess.Header,
		Payload: sess.Payload,
		Onion:   &End{},
	})
}

func (o Skins) End() Skins {
	return append(o, &End{})
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
	ks *crypto.KeySet) Skins {

	n := crypto.GenPingNonces()
	return Skins{}.
		Crypt(s[0].Header.Pub, nil, ks.Next(), n[0], 0).
		ForwardCrypt(s[1], ks.Next(), n[1]).
		ForwardCrypt(s[2], ks.Next(), n[2]).
		ForwardCrypt(s[3], ks.Next(), n[3]).
		ForwardCrypt(s[4], ks.Next(), n[4]).
		ForwardCrypt(client, ks.Next(), n[5]).
		Confirmation(id, 0)
}

func MakeRoute(id nonce.ID, k *crypto.Pub, ks *crypto.KeySet,
	alice, bob *sessions.Data, c sessions.Circuit) Skins {

	headers := GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		Route(id, k, ks, headers.ExitPoint()).
		RoutingHeader(headers.Return)
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
func MakeExit(p ExitParams) Skins {
	headers := GetHeaders(p.Alice, p.Bob, p.S, p.KS)
	return Skins{}.
		RoutingHeader(headers.Forward).
		Exit(p.ID, p.Port, p.Payload, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

// MakeGetBalance sends out a request in a similar way to Exit except the node
// being queried can be any of the 5.
func MakeGetBalance(p GetBalanceParams) Skins {
	headers := GetHeaders(p.Alice, p.Bob, p.S, p.KS)
	return Skins{}.
		RoutingHeader(headers.Forward).
		GetBalance(p.ID, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func MakeHiddenService(in *Intro, alice, bob *sessions.Data,
	c sessions.Circuit, ks *crypto.KeySet) Skins {

	headers := GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		HiddenService(in, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func MakeIntroQuery(id nonce.ID, hsk *crypto.Pub, alice, bob *sessions.Data,
	c sessions.Circuit, ks *crypto.KeySet) Skins {

	headers := GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		IntroQuery(id, hsk, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func MakeSession(id nonce.ID, s [5]*Session,
	client *sessions.Data, hop []*node.Node, ks *crypto.KeySet) Skins {

	n := crypto.GenNonces(6)
	sk := Skins{}
	for i := range s {
		if i == 0 {
			sk = sk.Crypt(hop[i].Identity.Pub, nil, ks.Next(),
				n[i], 0).Session(s[i])
		} else {
			sk = sk.ForwardSession(hop[i], ks.Next(), n[i], s[i])
		}
	}
	return sk.
		ForwardCrypt(client, ks.Next(), n[5]).
		Confirmation(id, 0)
}

func WriteRoutingHeader(s *splice.Splice, b RoutingHeaderBytes) *splice.Splice {
	copy(s.GetAll()[s.GetCursor():s.Advance(RoutingHeaderLen,
		"routing header")], b[:])
	s.Segments = append(s.Segments,
		splice.NameOffset{Offset: s.GetCursor(), Name: "routingheader"})
	return s
}

func ReadRoutingHeader(s *splice.Splice, b *RoutingHeaderBytes) *splice.Splice {
	*b = GetRoutingHeaderFromCursor(s)
	s.Segments = append(s.Segments,
		splice.NameOffset{Offset: s.GetCursor(), Name: "routingheader"})
	return s
}

func GetRoutingHeaderFromCursor(s *splice.Splice) (r RoutingHeaderBytes) {
	rh := s.GetRange(s.GetCursor(), s.Advance(RoutingHeaderLen,
		"routing header"))
	copy(r[:], rh)
	return
}

type RoutingHeaderBytes [RoutingHeaderLen]byte

type ReplyHeader struct {
	RoutingHeaderBytes
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces
}

type RoutingLayer struct {
	*Reverse
	*Crypt
}

type RoutingHeader struct {
	Layers [3]RoutingLayer
}

func FormatReply(header RoutingHeaderBytes, ciphers crypto.Ciphers,
	nonces crypto.Nonces, res slice.Bytes) (rb *splice.Splice) {

	rl := RoutingHeaderLen
	rb = splice.New(rl + len(res))
	copy(rb.GetUntil(rl), header[:rl])
	copy(rb.GetFrom(rl), res)
	for i := range ciphers {
		blk := ciph.BlockFromHash(ciphers[i])
		ciph.Encipher(blk, nonces[i], rb.GetFrom(rl))
	}
	return
}
