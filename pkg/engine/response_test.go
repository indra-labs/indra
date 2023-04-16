package engine

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestOnionSkins_Response(t *testing.T) {
	var e error
	id := nonce.NewID()
	var msg slice.Bytes
	var hash sha256.Hash
	if msg, hash, e = tests.GenMessage(10000, ""); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	port := uint16(cryptorand.IntN(65536))
	on := Skins{}.
		Response(id, msg, port).
		End().Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc coding.Codec
	if onc = Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	var rs *Response
	var ok bool
	if rs, ok = onc.(*Response); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	plH := sha256.Single(rs.Bytes)
	if plH != hash {
		t.Errorf("exit message did not unwrap correctly")
		t.FailNow()
	}
	if rs.ID != id {
		t.Error("ID did not decode correctly")
		t.FailNow()
	}
	if rs.Port != port {
		t.Error("port did not decode correctly")
		t.FailNow()
	}
}
