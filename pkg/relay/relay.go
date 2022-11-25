package relay

import (
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/transport"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/qu"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Relay struct {
	*node.Node
	node.Nodes
	transport.Dispatcher
	qu.C
}

func New(ip net.IP) (r *Relay) {
	r = &Relay{Node: node.New(ip),
		Nodes:      node.NewNodes(),
		Dispatcher: make(transport.Dispatcher),
		C:          qu.T()}
	return
}

func (c *Relay) Start() {
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

func (c *Relay) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (c *Relay) Shutdown() {
	c.C.Q()
}
