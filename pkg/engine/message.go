package engine

import (
	"reflect"
	
	"github.com/davecgh/go-spew/spew"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MessageMagic    = "ms"
	ReplyCiphersLen = 2*RoutingHeaderLen + 6*sha256.Len + 6*nonce.IVLen
	MessageLen      = magic.Len + 2*nonce.IDLen + 2*RoutingHeaderLen +
		ReplyCiphersLen
)

func MessageGen() coding.Codec           { return &Message{} }
func init()                              { Register(MessageMagic, MessageGen) }
func (x *Message) Magic() string         { return MessageMagic }
func (x *Message) Len() int              { return MessageLen + x.Payload.Len() }
func (x *Message) Wrap(inner Onion)      {}
func (x *Message) GetOnion() interface{} { return x }

type Message struct {
	Forwards        [2]*SessionData
	Address         *crypto.Pub
	ID, Re          nonce.ID
	Forward, Return *ReplyHeader
	Payload         slice.Bytes
}

func (x *Message) Encode(s *splice.Splice) (e error) {
	log.T.F("encoding %s %x %x %v %s", reflect.TypeOf(x),
		x.ID, x.Re, x.Address, spew.Sdump(x.Forward, x.Return,
			x.Payload.ToBytes()),
	)
	WriteRoutingHeader(s, x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.Magic(MessageMagic).
		Pubkey(x.Address).
		ID(x.ID).ID(x.Re)
	WriteRoutingHeader(s, x.Return.RoutingHeaderBytes).
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

func (x *Message) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), MessageLen-magic.Len,
		MessageMagic); fails(e) {
		return
	}
	x.Return = &ReplyHeader{}
	s.ReadPubkey(&x.Address).
		ReadID(&x.ID).ReadID(&x.Re)
	
	ReadRoutingHeader(s, &x.Return.RoutingHeaderBytes).
		ReadCiphers(&x.Return.Ciphers).
		ReadNonces(&x.Return.Nonces).
		ReadBytes(&x.Payload)
	return
}

func (x *Message) Handle(s *splice.Splice, p Onion,
	ni interface{}) (e error) {
	
	ng := ni.(*Engine)
	// Forward payload out to service port.
	_, e = ng.PendingResponses.ProcessAndDelete(x.ID, x, s.GetAll())
	return
}

func (x *Message) Account(res *Data, sm *SessionManager, s *SessionData, last bool) (skip bool, sd *SessionData) {
	
	res.ID = x.ID
	res.Billable = append(res.Billable, s.ID)
	skip = true
	return
}
