package client

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/onion"
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
	ifc.Transport
	qu.C
}

func New(tpt ifc.Transport) (c *Client) {
	n, _ := node.New(nil, tpt)
	c = &Client{Node: n,
		Nodes:     node.NewNodes(),
		Transport: tpt,
		C:         qu.T()}
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

const CircuitLen = 5

func (c *Client) GenerateCircuit() (ci *onion.Circuit, e error) {
	if len(c.Nodes) < 5 {
		e = fmt.Errorf("insufficient nodes to form a circuit, "+
			"5 required, %d available", len(c.Nodes))
		return
	}
	nodesLen := len(c.Nodes)
	s := make(node.Nodes, nodesLen)
	for i := range s {
		s[i] = c.Nodes[i]
	}
	randBytes := make([]byte, 8)
	// use crypto/rand to seed PRNG to avoid possible timing attacks.
	var n int
	if n, e = crand.Read(randBytes); check(e) && n != 8 {
		return
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(randBytes)))
	rand.Shuffle(nodesLen, func(i, j int) { s[i], s[j] = s[j], s[i] })
	ci = &onion.Circuit{
		ID:   nonce.ID{},
		Hops: s[:CircuitLen],
		Exit: 2,
	}
	return
}
