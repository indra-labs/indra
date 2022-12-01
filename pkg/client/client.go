package client

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
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
	Prv *prv.Key
	Pub *pub.Key
	*node.Node
	node.Nodes
	*address.SendCache
	*address.ReceiveCache
	onion.Circuits
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
	n, _ = node.New(nil, pubKey, tpt)
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

const CircuitLen = 5
const CircuitExit = 2
const ReturnLen = 3

func (c *Client) GeneratePath(length, exit int) (ci *onion.Circuit, e error) {
	if len(c.Sessions) < 5 {
		e = fmt.Errorf("insufficient Sessions to form a circuit, "+
			"5 required, %d available", len(c.Sessions))
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
		Hops: s[:length],
		Exit: exit,
	}
	c.Circuits = c.Circuits.Add(ci)
	return
}

func (c *Client) GenerateCircuit() (ci *onion.Circuit, e error) {
	return c.GeneratePath(CircuitLen, CircuitExit)
}

func (c *Client) GenerateReturn() (ci *onion.Circuit, e error) {
	return c.GeneratePath(ReturnLen, CircuitExit)
}
