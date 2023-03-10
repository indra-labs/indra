package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ExitMagic = "ex"
	ExitLen   = magic.Len + slice.Uint16Len + 3*sha256.Len +
		slice.Uint32Len + nonce.IVLen*3 + nonce.IDLen
)

type Exit struct {
	// Port identifies the type of service as well as being the port used by
	// the service to be relayed to. Notice there is no IP address, this is
	// because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way. This could be done by tunneling from
	// a local Socks5 proxy into Indranet and the exit node also having
	// this.
	Port uint16
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
	nonce.ID
	// Bytes are the message to be passed to the exit service.
	slice.Bytes
	Onion
}

func exitPrototype() Onion { return &Exit{} }

func init() { Register(ExitMagic, exitPrototype) }

func (o Skins) Exit(port uint16, prvs [3]*prv.Key, pubs [3]*pub.Key,
	nonces [3]nonce.IV, id nonce.ID, payload slice.Bytes) Skins {
	
	return append(o, &Exit{
		Port:    port,
		Ciphers: GenCiphers(prvs, pubs),
		Nonces:  nonces,
		ID:      id,
		Bytes:   payload,
		Onion:   nop,
	})
}

type ExitParams struct {
	Port    uint16
	Payload slice.Bytes
	ID      nonce.ID
	Client  *SessionData
	S       Circuit
	KS      *signer.KeySet
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
// The response is encrypted with the given layers, the ciphers are already
// given in reverse order, so they are decoded in given order to create the
// correct payload encryption to match the PayloadPub combined with the header's
// given public From key.
//
// The header remains a constant size and each node in the Reverse trims off
// their section at the top, moves the next crypt header to the top and pads the
// remainder with noise, so it always looks like the first hop.
func MakeExit(p ExitParams) Skins {
	
	forwardKeys := p.KS.Next3()
	returnKeys := p.KS.Next3()
	n := GenNonces(6)
	var returnNonces, forwardNonces [3]nonce.IV
	copy(returnNonces[:], n[3:])
	copy(forwardNonces[:], n[:3])
	var forwardSessions, returnSessions [3]*SessionData
	copy(forwardSessions[:], p.S[:3])
	copy(returnSessions[:], p.S[3:5])
	returnSessions[2] = p.Client
	var returnPubs [3]*pub.Key
	returnPubs[0] = p.S[3].PayloadPub
	returnPubs[1] = p.S[4].PayloadPub
	returnPubs[2] = p.Client.PayloadPub
	return Skins{}.
		RoutingHeader(forwardSessions, forwardKeys, forwardNonces).
		Exit(p.Port, returnKeys, returnPubs, returnNonces, p.ID, p.Payload).
		RoutingHeader(returnSessions, returnKeys, returnNonces)
}

func (x *Exit) Magic() string { return ExitMagic }

func (x *Exit) Encode(s *octet.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(ExitMagic).
		Uint16(x.Port).
		HashTriple(x.Ciphers).
		IVTriple(x.Nonces).
		ID(x.ID).
		Bytes(x.Bytes),
	)
}

func (x *Exit) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ExitLen-magic.Len, ExitMagic); check(e) {
		return
	}
	s.
		ReadUint16(&x.Port).
		ReadHashTriple(&x.Ciphers).
		ReadIVTriple(&x.Nonces).
		ReadID(&x.ID).
		ReadBytes(&x.Bytes)
	return
}

func (x *Exit) Len() int { return ExitLen + x.Bytes.Len() + x.Onion.Len() }

func (x *Exit) Wrap(inner Onion) { x.Onion = inner }

func (x *Exit) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	// payload is forwarded to a local port and the result is forwarded
	// back with a reverse header.
	var result slice.Bytes
	h := sha256.Single(x.Bytes)
	log.T.S(h)
	log.T.F("%s received exit id %x", ng.GetLocalNodeAddress(), x.ID)
	if e = ng.SendFromLocalNode(x.Port, x.Bytes); check(e) {
		return
	}
	timer := time.NewTicker(time.Second * 5)
	select {
	case result = <-ng.ReceiveToLocalNode(x.Port):
	case <-timer.C:
	}
	// We need to wrap the result in a message crypt.
	res := Encode(&Response{
		ID:    x.ID,
		Port:  x.Port,
		Load:  byte(ng.Load.Load()),
		Bytes: result,
	})
	rb := FormatReply(s.GetRange(s.GetCursor(), ReverseHeaderLen),
		res.GetRange(-1, -1), x.Ciphers, x.Nonces)
	switch on := p.(type) {
	case *Crypt:
		sess := ng.FindSessionByHeader(on.ToPriv)
		if sess == nil {
			break
		}
		for i := range sess.Services {
			if x.Port != sess.Services[i].Port {
				continue
			}
			in := sess.Services[i].RelayRate * s.Len() / 2
			out := sess.Services[i].RelayRate * rb.Len() / 2
			ng.DecSession(sess.ID, in+out, false, "exit")
			break
		}
	}
	ng.HandleMessage(rb, x)
	return
}

func (ng *Engine) SendExit(port uint16, msg slice.Bytes, id nonce.ID,
	target *SessionData, hook Callback) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := MakeExit(ExitParams{port, msg, id, se[len(se)-1], c, ng.KeySet})
	log.D.Ln("sending out exit onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}

func (ng *Engine) MakeExit(port uint16, msg slice.Bytes, id nonce.ID,
	exit *SessionData) (c Circuit, o Skins) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = exit
	se := ng.SelectHops(hops, s)
	copy(c[:], se)
	o = MakeExit(ExitParams{port, msg, id, se[len(se)-1], c, ng.KeySet})
	return
}

func (ng *Engine) SendExitNew(c Circuit, o Skins, hook Callback) {
	log.D.Ln("sending out exit onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}
