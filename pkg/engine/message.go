package engine

import (
	"reflect"
	
	"github.com/davecgh/go-spew/spew"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MessageMagic    = "ms"
	ReplyCiphersLen = 2*RoutingHeaderLen + 6*sha256.Len + 6*nonce.IVLen
	MessageLen      = magic.Len + 2*nonce.IDLen + 2*RoutingHeaderLen +
		ReplyCiphersLen
)

func MessagePrototype() Onion       { return &Message{} }
func init()                         { Register(MessageMagic, MessagePrototype) }
func (x *Message) Magic() string    { return MessageMagic }
func (x *Message) Len() int         { return MessageLen + x.Payload.Len() }
func (x *Message) Wrap(inner Onion) {}

type Message struct {
	Forwards        [2]*SessionData
	Address         *pub.Key
	ID, Re          nonce.ID
	Forward, Return *ReplyHeader
	Payload         slice.Bytes
}

func (o Skins) Message(msg *Message, ks *signer.KeySet) Skins {
	return append(o.
		ForwardCrypt(msg.Forwards[0], ks.Next(), nonce.New()).
		ForwardCrypt(msg.Forwards[1], ks.Next(), nonce.New()),
		msg)
}

func (x *Message) Encode(s *Splice) (e error) {
	log.T.F("encoding %s %x %x %v %s", reflect.TypeOf(x),
		x.ID, x.Re, x.Address, spew.Sdump(x.Forward, x.Return,
			x.Payload.ToBytes()),
	)
	s.RoutingHeader(x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.Magic(MessageMagic).
		Pubkey(x.Address).
		ID(x.ID).ID(x.Re).
		RoutingHeader(x.Return.RoutingHeaderBytes).
		Ciphers(x.Return.Ciphers).
		Nonces(x.Return.Nonces).
		Bytes(x.Payload)
	for i := range x.Forward.Ciphers {
		blk := ciph.BlockFromHash(x.Forward.Ciphers[i])
		log.D.F("encrypting %s", x.Forward.Ciphers[i].String())
		ciph.Encipher(blk, x.Forward.Nonces[i], s.GetFrom(start))
	}
	return
}

func (x *Message) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), MessageLen-magic.Len,
		MessageMagic); fails(e) {
		return
	}
	x.Return = &ReplyHeader{}
	s.ReadPubkey(&x.Address).
		ReadID(&x.ID).ReadID(&x.Re).
		ReadRoutingHeader(&x.Return.RoutingHeaderBytes).
		ReadCiphers(&x.Return.Ciphers).
		ReadNonces(&x.Return.Nonces).
		ReadBytes(&x.Payload)
	return
}

func (x *Message) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	// Forward payload out to service port.
	_, e = ng.PendingResponses.ProcessAndDelete(x.ID, x, s.GetAll())
	return
}

func (ng *Engine) SendMessage(mp *Message, hook Callback) (id nonce.ID) {
	// Add another two hops for security against unmasking.
	preHops := []byte{0, 1}
	oo := ng.SelectHops(preHops, mp.Forwards[:], "sendmessage")
	mp.Forwards = [2]*SessionData{oo[0], oo[1]}
	o := Skins{}.Message(mp, ng.KeySet)
	res := ng.PostAcctOnion(o)
	log.D.Ln("sending out message onion")
	ng.SendWithOneHook(mp.Forwards[0].Node.AddrPort, res, hook,
		ng.PendingResponses)
	return res.ID
}

func (x *Message) Account(res *SendData, sm *SessionManager, s *SessionData, last bool) (skip bool, sd *SessionData) {
	
	res.ID = x.ID
	res.Billable = append(res.Billable, s.ID)
	skip = true
	return
}
