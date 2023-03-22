package engine

import (
	"crypto/cipher"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
)

const (
	CryptMagic       = "cr"
	CryptLen         = magic.Len + nonce.IVLen + cloak.Len + pub.KeyLen
	ReverseCryptLen  = ReverseLen + CryptLen
	RoutingHeaderLen = 3 * ReverseCryptLen
)

type Crypt struct {
	Depth                     int
	ToHeaderPub, ToPayloadPub *pub.Key
	From                      *prv.Key
	Nonce                     nonce.IV
	// The remainder here are for Decode.
	Cloak   cloak.PubKey
	ToPriv  *prv.Key
	FromPub *pub.Key
	Onion
}

func cryptPrototype() Onion { return &Crypt{} }

func init() { Register(CryptMagic, cryptPrototype) }

func (o Skins) Crypt(toHdr, toPld *pub.Key, from *prv.Key, n nonce.IV,
	depth int) Skins {
	
	return append(o, &Crypt{
		Depth:        depth,
		ToHeaderPub:  toHdr,
		ToPayloadPub: toPld,
		From:         from,
		Nonce:        n,
		Onion:        nop,
	})
}

func (x *Crypt) Magic() string { return CryptMagic }

func (x *Crypt) Encode(s *Splice) (e error) {
	// log.T.S("encoding", reflect.TypeOf(x),
	// 	x.Nonce, x.Cloak, pub.Derive(x.From),
	// )
	if x.ToHeaderPub == nil || x.From == nil {
		s.Advance(CryptLen, "crypt")
		return
	}
	s.Magic(CryptMagic).
		IV(x.Nonce).Cloak(x.ToHeaderPub).Pubkey(pub.Derive(x.From))
	// Then we can encrypt the message segment
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.ToHeaderPub); check(e) {
		panic(e)
	}
	start := s.GetCursor()
	end := s.Len()
	// log.T.Ln("start", start, "end", end)
	switch {
	case x.Depth == 0:
	case x.Depth > 0:
		end = start + (x.Depth-1)*ReverseCryptLen
	default:
		panic("incorrect value for crypt sequence")
	}
	if x.Onion != nil {
		if e = x.Onion.Encode(s); check(e) {
			return
		}
	}
	ciph.Encipher(blk, x.Nonce, s.GetRange(start, end))
	if end != s.Len() {
		if blk = ciph.GetBlock(x.From, x.ToPayloadPub); check(e) {
			return
		}
		ciph.Encipher(blk, x.Nonce, s.GetFrom(end))
	}
	return e
}

func (x *Crypt) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), CryptLen-magic.Len, CryptMagic); check(e) {
		return
	}
	s.ReadIV(&x.Nonce).ReadCloak(&x.Cloak).ReadPubkey(&x.FromPub)
	return
}

func (x *Crypt) Len() int {
	return CryptLen + x.Onion.Len()
}

func (x *Crypt) Wrap(inner Onion) { x.Onion = inner }

func (x *Crypt) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	// this is probably an encrypted crypt for us.
	hdr, _, _, identity := ng.FindCloaked(x.Cloak)
	if hdr == nil {
		log.T.Ln("no matching key found from cloaked key")
		return
	}
	x.ToPriv = hdr
	x.Decrypt(hdr, s)
	if identity {
		if string(s.GetCursorToEnd()[:magic.Len]) != SessionMagic {
			log.T.Ln("dropping message due to identity key with" +
				" no following session")
			return e
		}
		ng.HandleMessage(BudgeUp(s), x)
		return e
	}
	ng.HandleMessage(BudgeUp(s), x)
	
	return e
}

// Decrypt requires the prv.Key to be located from the Cloak, using the FromPub
// key to derive the shared secret, and then decrypts the rest of the message.
func (x *Crypt) Decrypt(prk *prv.Key, s *Splice) {
	ciph.Encipher(ciph.GetBlock(prk, x.FromPub), x.Nonce,
		s.GetFrom(s.GetCursor()))
}
