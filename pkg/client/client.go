package client

import (
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/onion"
	"github.com/Indra-Labs/indra/pkg/transport"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/qu"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Client struct {
	net.IP
	*node.Node
	node.Nodes
	onion.Circuits
	transport.Dispatcher
	qu.C
}

func New() (c *Client) {
	c = &Client{Node: node.New(nil),
		Nodes:      node.NewNodes(),
		Dispatcher: make(transport.Dispatcher),
		C:          qu.T()}
	return
}

func (c *Client) Start() {
out:
	for {
		select {
		case <-c.C.Wait():
			c.Cleanup()
			break out
		case <-c.Node.Receive():
			// process received message
		}
	}
}

func (c *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (c *Client) Shutdown() {
	c.C.Q()
}
