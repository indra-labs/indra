package intro

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestLayer_Validate(t *testing.T) {
	addr := slice.GenerateRandomAddrPortIPv4()
	var e error
	var idPrv *prv.Key
	if idPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	im := New(idPrv, addr)
	log.I.S(im)
	if !im.Validate() {
		t.Error("failed to validate")
	}
}
