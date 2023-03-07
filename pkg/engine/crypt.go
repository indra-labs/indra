package engine

import (
	"crypto/cipher"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/session"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	CryptMagic       = "cr"
	CryptLen         = MagicLen + nonce.IVLen + cloak.Len + pub.KeyLen
	ReverseLayerLen  = ReverseLen + CryptLen + MagicLen
	ReverseHeaderLen = 3 * ReverseLayerLen
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

func (x *Crypt) Encode(s *octet.Splice) (e error) {
	s.Magic(CryptMagic).
		IV(x.Nonce).Cloak(x.ToHeaderPub).Pubkey(pub.Derive(x.From))
	if e = x.Onion.Encode(s); check(e) {
		return
	}
	// Then we can encrypt the message segment
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.ToHeaderPub); check(e) {
		panic(e)
	}
	start := s.GetCursor()
	end := s.Len()
	switch {
	case x.Depth == 0:
	case x.Depth > 0:
		end = start + (x.Depth-1)*ReverseLayerLen
	default:
		panic("incorrect value for crypt sequence")
	}
	ciph.Encipher(blk, x.Nonce, s.GetRange(start, end))
	if end != s.Len() {
		if blk = ciph.GetBlock(x.From, x.ToPayloadPub); check(e) {
			panic(e)
		}
		ciph.Encipher(blk, x.Nonce, s.GetRange(end, -1))
	}
	return e
}

func (x *Crypt) Decode(s *octet.Splice) (e error) {
	if e = TooShort(s.Remaining(), CryptLen-MagicLen, CryptMagic); check(e) {
		return
	}
	s.ReadIV(&x.Nonce).ReadCloak(&x.Cloak).ReadPubkey(&x.FromPub)
	return
}

func (x *Crypt) Len() int {
	return CryptLen + x.Onion.Len()
}

func (x *Crypt) Wrap(inner Onion) { x.Onion = inner }

func (x *Crypt) Handle(s *octet.Splice, p Onion,
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
		if string(s.GetRange(s.GetCursor(), -1)[:MagicLen]) != session.
			MagicString {
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
func (x *Crypt) Decrypt(prk *prv.Key, s *octet.Splice) {
	ciph.Encipher(ciph.GetBlock(prk, x.FromPub), x.Nonce,
		s.GetRange(s.GetCursor(), -1))
}
