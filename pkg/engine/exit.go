package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

const (
	ExitMagic = "ex"
	ExitLen   = magic.Len + slice.Uint16Len + 3*sha256.Len +
		slice.Uint32Len + nonce.IVLen*3 + nonce.IDLen
)

type Exit struct {
	zip.Reply
	// Port identifies the type of service as well as being the port used by
	// the service to be relayed to. Notice there is no IP address, this is
	// because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way. This could be done by tunneling from
	// a local Socks5 proxy into Indranet and the exit node also having
	// this.
	Port uint16
	// Bytes are the message to be passed to the exit service.
	slice.Bytes
	Onion
}

func exitPrototype() Onion { return &Exit{} }

func init() { Register(ExitMagic, exitPrototype) }

func (o Skins) Exit(id nonce.ID, port uint16, payload slice.Bytes,
	ep *ExitPoint) Skins {
	
	return append(o, &Exit{
		Reply: zip.Reply{
			ID:      id,
			Ciphers: GenCiphers(ep.Keys, ep.ReturnPubs),
			Nonces:  ep.Nonces,
		},
		Port:  port,
		Bytes: payload,
		Onion: nop,
	})
}

func (x *Exit) Magic() string { return ExitMagic }

func (x *Exit) Encode(s *zip.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(ExitMagic).Reply(&x.Reply).Uint16(x.Port).Bytes(x.Bytes),
	)
}

func (x *Exit) Decode(s *zip.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ExitLen-magic.Len,
		ExitMagic); check(e) {
		
		return
	}
	s.ReadReply(&x.Reply).ReadUint16(&x.Port).ReadBytes(&x.Bytes)
	return
}

func (x *Exit) Len() int { return ExitLen + x.Bytes.Len() + x.Onion.Len() }

func (x *Exit) Wrap(inner Onion) { x.Onion = inner }

func (x *Exit) Handle(s *zip.Splice, p Onion,
	ng *Engine) (e error) {
	
	// payload is forwarded to a local port and the result is forwarded
	// back with a reverse header.
	var result slice.Bytes
	h := sha256.Single(x.Bytes)
	log.T.S(h)
	log.T.F("%s received exit id %s", ng.GetLocalNodeAddressString(), x.ID)
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
			return
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

type ExitParams struct {
	Port    uint16
	Payload slice.Bytes
	ID      nonce.ID
	Target  *SessionData
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
// The header remains a constant size and each node in the Reverse trims off
// their section at the top, moves the next crypt header to the top and pads the
// remainder with noise, so it always looks like the first hop.
func MakeExit(p ExitParams) Skins {
	headers := GetHeaders(p.Target, p.S, p.KS)
	return Skins{}.
		RoutingHeader(headers.Forward).
		Exit(p.ID, p.Port, p.Payload, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendExit(port uint16, msg slice.Bytes, id nonce.ID,
	target *SessionData, hook Callback) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	// s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := MakeExit(ExitParams{port, msg, id, se[len(se)-1], c, ng.KeySet})
	// log.D.S(ng.GetLocalNodeAddress().String()+" sending out exit onion", o)
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}
