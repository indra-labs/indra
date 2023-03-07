package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ExitMagic = "ex"
	ExitLen   = MagicLen + slice.Uint16Len + 3*sha256.Len +
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
	if e = TooShort(s.Remaining(), ExitLen-MagicLen, ExitMagic); check(e) {
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
		res, x.Ciphers, x.Nonces)
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
