package onions

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"testing"
)

func TestPeerAd(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	pr, ks, _ := crypto.NewSigner()
	id := nonce.NewID()
	// in := NewPeer(id, pr, time.Now().Add(time.Hour))
	var prvs crypto.Privs
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs crypto.Pubs
	for i := range pubs {
		pubs[i] = crypto.DerivePub(prvs[i])
	}
	pa := NewPeer(id, pr, 20000)
	s := splice.New(pa.Len())
	if e = pa.Encode(s); fails(e) {
		t.Fatalf("did not encode")
	}
	log.D.S(s.GetAll().ToBytes())
	s.SetCursor(0)
	var onc coding.Codec
	if onc = Recognise(s); onc == nil {
		t.Fatalf("did not unwrap")
	}
	if e = onc.Decode(s); fails(e) {
		t.Fatalf("did not decode")
	}
	log.D.S(onc)
	var peer *PeerAd
	var ok bool
	if peer, ok = onc.(*PeerAd); !ok {
		t.Fatal("did not unwrap expected type")
	}
	if !peer.Validate() {
		t.Fatalf("received PeerAd did not validate")
	}
}
