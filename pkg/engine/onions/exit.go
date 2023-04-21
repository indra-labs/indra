package onions

import (
	"reflect"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ExitMagic = "exit"
	ExitLen   = magic.Len + slice.Uint16Len + 3*sha256.Len +
		slice.Uint32Len + nonce.IVLen*3 + nonce.IDLen
)

type Exit struct {
	ID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces
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

func exitGen() coding.Codec           { return &Exit{} }
func init()                           { Register(ExitMagic, exitGen) }
func (x *Exit) Magic() string         { return ExitMagic }
func (x *Exit) Len() int              { return ExitLen + x.Bytes.Len() + x.Onion.Len() }
func (x *Exit) Wrap(inner Onion)      { x.Onion = inner }
func (x *Exit) GetOnion() interface{} { return x }

func (x *Exit) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Ciphers, x.Nonces, x.Port, x.Bytes.ToBytes(),
	)
	return x.Onion.Encode(s.
		Magic(ExitMagic).
		ID(x.ID).Ciphers(x.Ciphers).Nonces(x.Nonces).
		Uint16(x.Port).
		Bytes(x.Bytes),
	)
}

func (x *Exit) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ExitLen-magic.Len,
		ExitMagic); fails(e) {
		
		return
	}
	s.
		ReadID(&x.ID).ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces).
		ReadUint16(&x.Port).ReadBytes(&x.Bytes)
	return
}

func (x *Exit) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	// payload is forwarded to a local port and the result is forwarded
	// back with a reverse header.
	var result slice.Bytes
	h := sha256.Single(x.Bytes)
	log.T.S(h)
	log.T.F("%s received exit id %s", ng.Mgr().
		GetLocalNodeAddressString(), x.ID)
	if e = ng.Mgr().SendFromLocalNode(x.Port, x.Bytes); fails(e) {
		return
	}
	timer := time.NewTicker(time.Second * 5)
	select {
	case result = <-ng.Mgr().ReceiveToLocalNode(x.Port):
	case <-timer.C:
	}
	// We need to wrap the result in a message crypt.
	res := Encode(&Response{
		ID:    x.ID,
		Port:  x.Port,
		Load:  byte(ng.GetLoad()),
		Bytes: result,
	})
	rb := FormatReply(GetRoutingHeaderFromCursor(s),
		x.Ciphers, x.Nonces, res.GetAll())
	switch on := p.(type) {
	case *Crypt:
		sess := ng.Mgr().FindSessionByHeader(on.ToPriv)
		if sess == nil {
			return
		}
		sess.Node.Lock()
		for i := range sess.Node.Services {
			if x.Port != sess.Node.Services[i].Port {
				sess.Node.Unlock()
				continue
			}
			in := sess.Node.Services[i].RelayRate * s.Len() / 2
			out := sess.Node.Services[i].RelayRate * rb.Len() / 2
			sess.Node.Unlock()
			ng.Mgr().DecSession(sess.ID, in+out, false, "exit")
			break
		}
	}
	ng.HandleMessage(rb, x)
	return
}

type ExitParams struct {
	Port       uint16
	Payload    slice.Bytes
	ID         nonce.ID
	Alice, Bob *sessions.Data
	S          sessions.Circuit
	KS         *crypto.KeySet
}

func (x *Exit) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	s.Node.Lock()
	for j := range s.Node.Services {
		if s.Node.Services[j].Port != x.Port {
			s.Node.Unlock()
			continue
		}
		s.Node.Unlock()
		res.Port = x.Port
		res.PostAcct = append(res.PostAcct,
			func() {
				sm.DecSession(s.ID,
					s.Node.Services[j].RelayRate*len(res.B)/2, true, "exit")
			})
		break
	}
	res.Billable = append(res.Billable, s.ID)
	res.ID = x.ID
	skip = true
	return
}
