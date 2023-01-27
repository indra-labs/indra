package client

import (
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (cl *Client) SendExit(port uint16, message slice.Bytes,
	target *traffic.Session, hook func(b slice.Bytes)) {

	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(traffic.Sessions, len(hops))
	s[2] = target
	se := cl.Select(hops, s)
	var c traffic.Circuit
	copy(c[:], se)
	o := onion.SendExit(port, message, se[len(se)-1], c, cl.KeySet)
	cl.SendOnion(c[0].AddrPort, o, hook)
}
