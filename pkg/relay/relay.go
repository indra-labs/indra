package relay

import (
	"net"

	"github.com/cybriq/qu"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/node"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/types"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Relay struct {
	prv.Key
	PubKey pub.Key
	*node.Node
	node.Nodes
	types.Transport
	qu.C
}

func New(ip net.IP, tpt types.Transport) (r *Relay) {
	// r = &Relay{Node: node.New(ip),
	// 	Nodes:     node.NewNodes(),
	// 	Transport: tpt,
	// 	C:         qu.T()}
	return
}

func (c *Relay) Start() {
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

func (c *Relay) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (c *Relay) Shutdown() {
	c.C.Q()
}
