package engine

import (
	"crypto/cipher"
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/splice"
)

const (
	RouteMagic = "ro"
	RouteLen   = magic.Len + crypto.CloakLen + crypto.PubKeyLen + nonce.IVLen +
		nonce.IDLen + 3*sha256.Len + 3*nonce.IVLen
)

func RouteGen() coding.Codec           { return &Route{} }
func init()                            { Register(RouteMagic, RouteGen) }
func (x *Route) Magic() string         { return RouteMagic }
func (x *Route) Len() int              { return RouteLen + x.Onion.Len() }
func (x *Route) Wrap(inner Onion)      { x.Onion = inner }
func (x *Route) GetOnion() interface{} { return x }

type Route struct {
	HiddenService *crypto.Pub
	HiddenCloaked crypto.PubKey
	Sender        *crypto.Prv
	SenderPub     *crypto.Pub
	nonce.IV
	// ------- the rest is encrypted to the HiddenService/Sender keys.
	ID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces
	RoutingHeaderBytes
	Onion
}

func (x *Route) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.HiddenService, x.Sender, x.IV, x.Ciphers, x.Nonces,
		x.RoutingHeaderBytes,
	)
	s.Magic(RouteMagic).
		Cloak(x.HiddenService).
		Pubkey(crypto.DerivePub(x.Sender)).
		IV(x.IV)
	start := s.GetCursor()
	s.ID(x.ID).Ciphers(x.Ciphers).Nonces(x.Nonces)
	if e = x.Onion.Encode(s); fails(e) {
		return
	}
	var blk cipher.Block
	// Encrypt the message!
	if blk = ciph.GetBlock(x.Sender, x.HiddenService, "route"); fails(e) {
		return
	}
	ciph.Encipher(blk, x.IV, s.GetFrom(start))
	return
}

func (x *Route) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), RouteLen-magic.Len,
		RouteMagic); fails(e) {
		return
	}
	s.ReadCloak(&x.HiddenCloaked).
		ReadPubkey(&x.SenderPub).
		ReadIV(&x.IV)
	return
}

// Decrypt decrypts the rest of a message after the Route segment if the
// recipient has the hidden service private key.
func (x *Route) Decrypt(prk *crypto.Prv, s *splice.Splice) {
	ciph.Encipher(ciph.GetBlock(prk, x.SenderPub, "route decrypt"), x.IV,
		s.GetRest())
	// And now we can see the reply field for the return trip.
	s.ReadID(&x.ID).ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces)
	ReadRoutingHeader(s, &x.RoutingHeaderBytes)
}

func (x *Route) Handle(s *splice.Splice, p Onion, ni interface{}) (e error) {
	
	ng := ni.(*Engine)
	log.D.Ln(ng.GetLocalNodeAddressString(), "handling route")
	hc := ng.FindCloakedHiddenService(x.HiddenCloaked)
	if hc == nil {
		log.T.Ln("no matching hidden service key found from cloaked key")
		return
	}
	if x.HiddenService, e = crypto.PubFromBytes((*hc)[:]); fails(e) {
		return
	}
	log.D.Ln("route key", *hc)
	hcl := *hc
	if hh, ok := ng.HiddenRouting.HiddenServices[hcl]; ok {
		log.D.F("we are the hidden service %s - decrypting...",
			hh.CurrentIntros[0].Key.ToBase32Abbreviated())
		// We have the keys to unwrap this one.
		x.Decrypt(hh.Prv, s)
		log.D.Ln(s)
		n := crypto.GenNonces(5)
		rvKeys := ng.KeySet.Next3()
		hops := []byte{3, 4, 5, 0, 1}
		s := make(sessions.Sessions, len(hops))
		ng.SelectHops(hops, s, "route reply header")
		rt := &Routing{
			Sessions: [3]*sessions.Data{s[0], s[1], s[2]},
			Keys:     crypto.Privs{rvKeys[0], rvKeys[1], rvKeys[2]},
			Nonces:   crypto.Nonces{n[0], n[1], n[2]},
		}
		rh := Skins{}.RoutingHeader(rt)
		rHdr := Encode(rh.Assemble())
		rHdr.SetCursor(0)
		ep := ExitPoint{
			Routing: rt,
			ReturnPubs: crypto.Pubs{
				crypto.DerivePub(s[0].Payload.Prv),
				crypto.DerivePub(s[1].Payload.Prv),
				crypto.DerivePub(s[2].Payload.Prv),
			},
		}
		mr := Skins{}.
			ForwardCrypt(s[3], ng.KeySet.Next(), n[3]).
			ForwardCrypt(s[4], ng.KeySet.Next(), n[4]).
			Ready(x.ID, x.HiddenService,
				x.RoutingHeaderBytes,
				GetRoutingHeaderFromCursor(rHdr),
				x.Ciphers,
				crypto.GenCiphers(ep.Keys, ep.ReturnPubs),
				x.Nonces,
				ep.Nonces)
		assembled := mr.Assemble()
		reply := Encode(assembled)
		ng.HandleMessage(reply, x)
	}
	return
}

func (x *Route) Account(res *Data, sm *SessionManager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	
	copy(res.ID[:], x.ID[:])
	res.Billable = append(res.Billable, s.ID)
	return
}
