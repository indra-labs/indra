package indra

import (
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) SendExit(port uint16, message slice.Bytes, id nonce.ID,
	target *traffic.Session, hook func(id nonce.ID, b slice.Bytes)) {

	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(traffic.Sessions, len(hops))
	s[2] = target
	se := en.Select(hops, s)
	var c traffic.Circuit
	copy(c[:], se)
	o := onion.SendExit(port, message, id, se[len(se)-1], c, en.KeySet)
	en.SendOnion(c[0].AddrPort, o, hook)
}
