package response

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/end"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"testing"

	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"git.indra-labs.org/dev/ind/pkg/util/cryptorand"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"git.indra-labs.org/dev/ind/pkg/util/tests"
)

func TestOnions_Response(t *testing.T) {
	var e error
	id := nonce.NewID()
	var msg slice.Bytes
	var hash sha256.Hash
	if msg, hash, e = tests.GenMessage(10000, ""); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	port := uint16(cryptorand.IntN(65536))
	on := ont.Assemble([]ont.Onion{
		New(id, port, msg, 0),
		end.NewEnd(),
	})
	s := codec.Encode(on)
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
		t.Error("Keys did not decode correctly")
		t.FailNow()
	}
	if rs.Port != port {
		t.Error("port did not decode correctly")
		t.FailNow()
	}
}
