package node

import (
	"net"

	"github.com/Indra-Labs/indra/pkg/key/address"
)

type Node struct {
	net.IP
	In  *address.SendCache
	Out *address.ReceiveCache
}

func New(ip net.IP) *Node {
	return &Node{
		IP:  ip,
		In:  address.NewSendCache(),
		Out: address.NewReceiveCache(),
	}
}
