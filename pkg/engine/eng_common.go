package engine

import (
	"net"
	"net/netip"
	"testing"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

type RoutingHeaderBytes [RoutingHeaderLen]byte

type ReplyHeader struct {
	RoutingHeaderBytes
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	Nonces
}

func MakeReplyHeader(ng *Engine) (returnHeader *ReplyHeader) {
	n := GenNonces(3)
	rvKeys := ng.KeySet.Next3()
	hops := []byte{3, 4, 5}
	sessions := make(Sessions, len(hops))
	ng.SelectHops(hops, sessions, "make message reply header")
	rt := &Routing{
		Sessions: [3]*SessionData{sessions[0], sessions[1], sessions[2]},
		Keys:     Privs{rvKeys[0], rvKeys[1], rvKeys[2]},
		Nonces:   Nonces{n[0], n[1], n[2]},
	}
	rh := Skins{}.RoutingHeader(rt)
	rHdr := Encode(rh.Assemble())
	rHdr.SetCursor(0)
	ep := ExitPoint{
		Routing: rt,
		ReturnPubs: Pubs{
			crypto.DerivePub(sessions[0].Payload.Prv),
			crypto.DerivePub(sessions[1].Payload.Prv),
			crypto.DerivePub(sessions[2].Payload.Prv),
		},
	}
	returnHeader = &ReplyHeader{
		RoutingHeaderBytes: rHdr.GetRoutingHeaderFromCursor(),
		Ciphers:            GenCiphers(ep.Routing.Keys, ep.ReturnPubs),
		Nonces:             ep.Routing.Nonces,
	}
	return
}

type RoutingLayer struct {
	*Reverse
	*Crypt
}

type RoutingHeader struct {
	Layers [3]RoutingLayer
}

func BudgeUp(s *Splice) (o *Splice) {
	o = s
	start := o.GetCursor()
	copy(o.GetAll(), s.GetFrom(start))
	copy(s.GetFrom(o.Len()-start), slice.NoisePad(start))
	return
}

func FormatReply(header RoutingHeaderBytes, ciphers Ciphers,
	nonces Nonces, res slice.Bytes) (rb *Splice) {
	
	rl := RoutingHeaderLen
	rb = NewSplice(rl + len(res))
	copy(rb.GetUntil(rl), header[:rl])
	copy(rb.GetFrom(rl), res)
	for i := range ciphers {
		blk := ciph.BlockFromHash(ciphers[i])
		ciph.Encipher(blk, nonces[i], rb.GetFrom(rl))
	}
	return
}

func GenCiphers(prvs Privs, pubs Pubs) (ciphers Ciphers) {
	for i := range prvs {
		ciphers[i] = crypto.ComputeSharedSecret(prvs[i], pubs[i])
		log.T.Ln("cipher", i, ciphers[i])
	}
	return
}

func GenNonces(count int) (n []nonce.IV) {
	n = make([]nonce.IV, count)
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func createNMockCircuits(inclSessions bool, nCircuits int,
	nReturnSessions int) (cl []*Engine, e error) {
	
	nTotal := 1 + nCircuits*5
	cl = make([]*Engine, nTotal)
	nodes := make([]*Node, nTotal)
	tpts := make([]Transport, nTotal)
	ss := make(Sessions, nTotal-1)
	for i := range tpts {
		tpts[i] = NewSim(nTotal)
	}
	for i := range nodes {
		var idPrv *crypto.Prv
		if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
			return
		}
		addr := slice.GenerateRandomAddrPortIPv4()
		var local bool
		if i == 0 {
			local = true
		}
		nodes[i], _ = NewNode(addr, idPrv, tpts[i], tpts[i], 50000, local)
		if cl[i], e = NewEngine(Params{
			tpts[i],
			tpts[i],
			idPrv,
			nodes[i],
			nil,
			nReturnSessions},
		); fails(e) {
			return
		}
		cl[i].SetLocalNodeAddress(nodes[i].AddrPort)
		cl[i].SetLocalNode(nodes[i])
		if inclSessions {
			// Create a session for all but the first.
			if i > 0 {
				ss[i-1] = NewSessionData(nonce.NewID(), nodes[i],
					1<<16, nil, nil, byte((i-1)/nCircuits))
				// AddIntro session to node, so it will be able to relay if it
				// gets a message with the key.
				cl[i].AddSession(ss[i-1])
				// we need a copy for the node so the balance adjustments don't
				// double up.
				s := *ss[i-1]
				cl[0].AddSession(&s)
			}
		}
	}
	// Add all the nodes to each other, so they can pass messages.
	for i := range cl {
		for j := range nodes {
			if i == j {
				continue
			}
			cl[i].AddNodes(nodes[j])
		}
	}
	return
}

func CreateNMockCircuits(nCirc int, nReturns int) (cl []*Engine, e error) {
	return createNMockCircuits(false, nCirc, nReturns)
}

func CreateNMockCircuitsWithSessions(nCirc int, nReturns int) (cl []*Engine,
	e error) {
	
	return createNMockCircuits(true, nCirc, nReturns)
}

func GetTwoPrvKeys(t *testing.T) (prv1, prv2 *crypto.Prv) {
	var e error
	if prv1, e = crypto.GeneratePrvKey(); fails(e) {
		t.FailNow()
	}
	if prv2, e = crypto.GeneratePrvKey(); fails(e) {
		t.FailNow()
	}
	return
}

func GetCipherSet(t *testing.T) (prvs Privs, pubs Pubs) {
	for i := range prvs {
		prv1, prv2 := GetTwoPrvKeys(t)
		prvs[i] = prv1
		pubs[i] = crypto.DerivePub(prv2)
	}
	return
}

func Gen3Nonces() (n Nonces) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GenPingNonces() (n [6]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func StandardCircuit() []byte { return []byte{0, 1, 2, 3, 4, 5} }

func GetNetworkFromAddrPort(addr string) (nw string, u *net.UDPAddr,
	e error) {
	
	nw = "udp"
	var ap netip.AddrPort
	if ap, e = netip.ParseAddrPort(addr); fails(e) {
		return
	}
	u = &net.UDPAddr{IP: net.ParseIP(ap.Addr().String()), Port: int(ap.Port())}
	if u.IP.To4() != nil {
		nw = "udp4"
	}
	return
}
