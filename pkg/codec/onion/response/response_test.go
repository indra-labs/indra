package response

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/onion/end"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"testing"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/tests"
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
		NewResponse(id, port, msg, 0),
		end.NewEnd(),
	})
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
