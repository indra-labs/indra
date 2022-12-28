package client

import (
	"net/netip"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/node"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/qu"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Client struct {
	*netip.AddrPort
	*node.Node
	node.Nodes
	*address.SendCache
	*address.ReceiveCache
	Circuits
	Sessions
	qu.C
}

func New(tpt ifc.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes) (c *Client, e error) {

	hdrPub := pub.Derive(hdrPrv)
	var n *node.Node
	n, _ = node.New(no.AddrPort, hdrPub, nil, hdrPrv, nil, tpt)
	c = &Client{
		Node:  n,
		Nodes: nodes,
		C:     qu.T(),
	}
	return
}

func (c *Client) Start() {
out:
	for {
		select {
		case <-c.C.Wait():
			c.Cleanup()
			break out
		case msg := <-c.Node.Receive():
			// process received message
			_ = msg

		}
	}
}

func (c *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (c *Client) Shutdown() {
	c.C.Q()
}
