package relay

import (
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ifc"
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

type Relay struct {
	prv.Key
	PubKey pub.Key
	*node.Node
	node.Nodes
	ifc.Transport
	qu.C
}

func New(ip net.IP, tpt ifc.Transport) (r *Relay) {
	n, _ := node.New(ip, tpt)
	r = &Relay{Node: n,
		Nodes:     node.NewNodes(),
		Transport: tpt,
		C:         qu.T()}
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
