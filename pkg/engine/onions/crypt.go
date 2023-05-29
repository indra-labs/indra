package onions

import (
	"crypto/cipher"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

const (
	CryptMagic       = "cryp"
	CryptLen         = magic.Len + nonce.IVLen + crypto.CloakLen + crypto.PubKeyLen
	ReverseCryptLen  = ReverseLen + CryptLen
	RoutingHeaderLen = 3 * ReverseCryptLen
)

type Crypt struct {
	Depth                     int
	ToHeaderPub, ToPayloadPub *crypto.Pub
	From                      *crypto.Prv
	IV                        nonce.IV
	// The remainder here are for Decode.
	Cloak   crypto.PubKey
	ToPriv  *crypto.Prv
	FromPub *crypto.Pub
	Onion
}

func (x *Crypt) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	sd = sm.FindSessionByHeaderPub(x.ToHeaderPub)
	if sd == nil {
		return
	}
	res.Sessions = append(res.Sessions, sd)
	// The last hop needs no accounting as it's us!
	if last {
		res.Ret = sd.Header.Bytes
		res.Billable = append(res.Billable, sd.Header.Bytes)
	}
	return
}

func (x *Crypt) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), CryptLen-magic.Len, CryptMagic); fails(e) {
		return
	}
	s.ReadIV(&x.IV).ReadCloak(&x.Cloak).ReadPubkey(&x.FromPub)
	return
}

// Decrypt requires the prv.Pub to be located from the Cloak, using the FromPub
// key to derive the shared secret, and then decrypts the rest of the message.
func (x *Crypt) Decrypt(prk *crypto.Prv, s *splice.Splice) {
	ciph.Encipher(ciph.GetBlock(prk, x.FromPub, "decrypt crypt header"),
		x.IV, s.GetRest())
}

func (x *Crypt) Encode(s *splice.Splice) (e error) {
	log.T.F("encoding %s %s %x %x", reflect.TypeOf(x),
		x.ToHeaderPub, x.From.ToBytes(), x.IV,
	)
	if x.ToHeaderPub == nil || x.From == nil {
		s.Advance(CryptLen, "crypt")
		return
	}
	s.Magic(CryptMagic).
		IV(x.IV).Cloak(x.ToHeaderPub).Pubkey(crypto.DerivePub(x.From))
	// Then we can encrypt the message segment
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.ToHeaderPub, "crypt header"); fails(e) {
		panic(e)
	}
	start := s.GetCursor()
	end := s.Len()
	switch {
	case x.Depth == 0:
	case x.Depth > 0:
		end = start + (x.Depth-1)*ReverseCryptLen
	default:
		panic("incorrect value for crypt sequence")
	}
	if x.Onion != nil {
		if e = x.Onion.Encode(s); fails(e) {
			return
		}
	}
	ciph.Encipher(blk, x.IV, s.GetRange(start, end))
	if end != s.Len() {
		if blk = ciph.GetBlock(x.From, x.ToPayloadPub,
			"crypt payload"); fails(e) {
			return
		}
		ciph.Encipher(blk, x.IV, s.GetFrom(end))
	}
	return e
}

func (x *Crypt) GetOnion() interface{} { return x }

func (x *Crypt) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	hdr, _, _, identity := ng.Mgr().FindCloaked(x.Cloak)
	if hdr == nil {
		log.T.Ln("no matching key found from cloaked key")
		return
	}
	x.ToPriv = hdr
	x.Decrypt(hdr, s)
	if identity {
		if string(s.GetRest()[:magic.Len]) != SessionMagic {
			log.T.Ln("dropping message due to identity key with" +
				" no following session")
			return e
		}
		ng.HandleMessage(splice.BudgeUp(s), x)
		return e
	}
	ng.HandleMessage(splice.BudgeUp(s), x)
	return e
}

func (x *Crypt) Len() int         { return CryptLen + x.Onion.Len() }
func (x *Crypt) Magic() string    { return CryptMagic }
func (x *Crypt) Wrap(inner Onion) { x.Onion = inner }
func cryptGen() coding.Codec      { return &Crypt{} }
func init()                       { Register(CryptMagic, cryptGen) }
