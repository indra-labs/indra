package client

import (
	"net"

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
	net.IP
	Prv *prv.Key
	Pub *pub.Key
	*node.Node
	node.Nodes
	*address.SendCache
	*address.ReceiveCache
	Circuits
	Sessions
	ifc.Transport
	qu.C
}

func New(tpt ifc.Transport, nodes node.Nodes) (c *Client, e error) {
	var p *prv.Key
	if p, e = prv.GenerateKey(); check(e) {
		return
	}
	pubKey := pub.Derive(p)
	var n *node.Node
	n, _ = node.New(nil, pubKey, nil, tpt)
	c = &Client{
		Node:      n,
		Nodes:     nodes,
		Transport: tpt,
		C:         qu.T(),
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
