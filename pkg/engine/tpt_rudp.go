package engine

import (
	"net/netip"
	
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type RUDP struct {
	endpoint *netip.AddrPort
	in, out  ByteChan
}

func (k RUDP) Send(b slice.Bytes) {
	k.out <- b
}

func (k RUDP) Receive() <-chan slice.Bytes {
	return k.in
}
