//go:build failingtests

package services

import (
	"context"
	"git.indra-labs.org/dev/ind/pkg/codec"
	intro "git.indra-labs.org/dev/ind/pkg/codec/ad/intro"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/exit"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/getbalance"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/engine/node"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	"git.indra-labs.org/dev/ind/pkg/engine/transport"
	headers2 "git.indra-labs.org/dev/ind/pkg/headers"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"testing"
	"time"
)

func TestOnions_HiddenService(t *testing.T) {
	ci.TraceIfNotCI()
	var e error
	n3 := crypto.Gen3Nonces()
	id := nonce.NewID()
	pr, ks, _ := crypto.NewSigner()
	var prvs crypto.Privs
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs crypto.Pubs
	for i := range pubs {
		pubs[i] = crypto.DerivePub(prvs[i])
	}
	ctx := context.Background()
	var k1, k2 *crypto.Keys
	if k1, e = crypto.GenerateKeys(); fails(e) {
		return
	}
	if k2, e = crypto.GenerateKeys(); fails(e) {
		return
	}
	var circ sessions.Circuit
	for i := range circ {
		tpt := transport.NewSimDuplex(10, ctx)
		adr := slice.GenerateRandomAddrPortIPv4()
		nod, _ := node.NewNode(adr, k1, tpt, 50000)
		ss := sessions.NewSessionData(nonce.NewID(), nod,
			1<<16, nil, nil, 1)
		circ[i] = ss
	}
	atpt := transport.NewSimDuplex(10, ctx)
	aaddr := slice.GenerateRandomAddrPortIPv4()
	anode, _ := node.NewNode(aaddr, k1, atpt, 50000)
	alice := sessions.NewSessionData(nonce.NewID(), anode,
		1<<16, nil, nil, 1)
	btpt := transport.NewSimDuplex(10, ctx)
	baddr := slice.GenerateRandomAddrPortIPv4()
	bnode, _ := node.NewNode(baddr, k2, btpt, 50000)
	bob := sessions.NewSessionData(nonce.NewID(), bnode,
		1<<16, nil, nil, 1)
	_, KS, _ := crypto.NewSigner()
	headers := headers2.GetHeaders(alice, bob, circ, KS)
	ep := &exit.ExitPoint{
		Routing: &exit.Routing{
			Sessions: headers.ExitPoint().Routing.Sessions,
			Keys:     prvs,
			Nonces:   n3,
		},
		ReturnPubs: pubs,
	}
	log.D.S("ep", ep)
	in := intro.New(id, pr, slice.GenerateRandomAddrPortIPv6(),
		20000, 80, time.Now().Add(time.Hour))
	on := ont.Assemble([]ont.Onion{New(in, ep)})
	s := ont.Encode(on)
	s.SetCursor(0)
	var onc codec.Codec
	if onc = reg.Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	var ex *HiddenService
	var ok bool
	if ex, ok = onc.(*HiddenService); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	for i := range ex.Ciphers {
		if ex.Ciphers[i] != on.(*getbalance.GetBalance).Ciphers[i] {
			t.Errorf("cipher %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	for i := range ex.Nonces {
		if ex.Nonces[i] != n3[i] {
			t.Errorf("nonce %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
}
