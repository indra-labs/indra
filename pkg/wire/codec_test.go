package wire

import (
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirmation"
	log2 "github.com/cybriq/proc/pkg/log"
)

func TestOnionSkins_Cipher(t *testing.T) {
	log2.CodeLoc = true
	var e error
	hdrP, pldP := GetTwoPrvKeys(t)
	hdr, pld := pub.Derive(hdrP), pub.Derive(pldP)
	log.I.S(hdr, pld)
	n := nonce.NewID()
	log.I.S(n)
	on := OnionSkins{}.
		Cipher(hdr, pld).
		Confirmation(n).
		Assemble()
	onb := EncodeOnion(on)
	var sc slice.Cursor
	c := &sc
	var onc types.Onion
	if onc, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	log.I.S(onc.(*cipher.OnionSkin))
	var oncn types.Onion
	if oncn, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	log.I.S(oncn.(*confirmation.OnionSkin))
}

func TestOnionSkins_Confirmation(t *testing.T) {
	log2.CodeLoc = true
	var e error
	n := nonce.NewID()
	log.I.S(n)
	on := OnionSkins{}.
		Confirmation(n).
		Assemble()
	onb := EncodeOnion(on)
	var sc slice.Cursor
	c := &sc
	var oncn types.Onion
	if oncn, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	log.I.S(oncn.(*confirmation.OnionSkin))
}

func TestOnionSkins_Exit(t *testing.T) {

}

func TestOnionSkins_Forward(t *testing.T) {

}

func TestOnionSkins_Message(t *testing.T) {

}

func TestOnionSkins_Purchase(t *testing.T) {

}

func TestOnionSkins_Reply(t *testing.T) {

}

func TestOnionSkins_Response(t *testing.T) {

}

func TestOnionSkins_Session(t *testing.T) {

}

func TestOnionSkins_Token(t *testing.T) {

}

func GetTwoPrvKeys(t *testing.T) (prv1, prv2 *prv.Key) {
	var e error
	if prv1, e = prv.GenerateKey(); check(e) {
		t.FailNow()
	}
	if prv2, e = prv.GenerateKey(); check(e) {
		t.FailNow()
	}
	return
}
