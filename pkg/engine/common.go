package engine

import (
	"fmt"
	"testing"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/ecdh"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/relay/types"
	"git-indra.lan/indra-labs/indra/pkg/transport"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	MagicLen    = 2
	ErrTooShort = "'%s' message  minimum size: %d got: %d"
)

func TooShort(got, found int, magic string) (e error) {
	if got >= found {
		return
	}
	e = fmt.Errorf(ErrTooShort, magic, got, found)
	return
	
}

func BudgeUp(s *octet.Splice) (o *octet.Splice) {
	o = s
	start := o.GetCursor()
	copy(o.GetRange(-1, -1), s.GetRange(start, -1))
	copy(s.GetRange(o.Len()-start, -1), slice.NoisePad(start))
	return
}

func FormatReply(header, res slice.Bytes, ciphers [3]sha256.Hash,
	nonces [3]nonce.IV) (rb *octet.Splice) {
	
	rl := ReverseHeaderLen
	rb = octet.New(rl + len(res))
	copy(rb.GetRange(-1, rl), header[:rl])
	copy(rb.GetRange(ReverseHeaderLen, -1), res)
	for i := range ciphers {
		blk := ciph.BlockFromHash(ciphers[i])
		ciph.Encipher(blk, nonces[2-i], rb.GetRange(ReverseHeaderLen, -1))
	}
	return
}

func GenCiphers(prvs [3]*prv.Key, pubs [3]*pub.Key) (ciphers [3]sha256.Hash) {
	for i := range prvs {
		ciphers[2-i] = ecdh.Compute(prvs[i], pubs[i])
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
	tpts := make([]types.Transport, nTotal)
	ss := make(Sessions, nTotal-1)
	for i := range tpts {
		tpts[i] = transport.NewSim(nTotal)
	}
	for i := range nodes {
		var idPrv *prv.Key
		if idPrv, e = prv.GenerateKey(); check(e) {
			return
		}
		idPub := pub.Derive(idPrv)
		addr := slice.GenerateRandomAddrPortIPv4()
		var local bool
		if i == 0 {
			local = true
		}
		nodes[i], _ = New(addr, idPub, idPrv, tpts[i], 50000, local)
		if cl[i], e = NewEngine(Params{
			tpts[i],
			idPrv,
			nodes[i],
			nil,
			nReturnSessions},
		); check(e) {
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

func GetTwoPrvKeys(t *testing.T) (prv1, prv2 *prv.Key) {
	var e error
	if prv1, e = prv.GenerateKey(); check(e) {
		t.FailNow()
	}
	if prv2, e = prv.GenerateKey(); check(e) {
		t.FailNow()
	}
	return
}

func GetCipherSet(t *testing.T) (prvs [3]*prv.Key, pubs [3]*pub.Key) {
	for i := range prvs {
		prv1, prv2 := GetTwoPrvKeys(t)
		prvs[i] = prv1
		pubs[i] = pub.Derive(prv2)
	}
	return
}

func Gen3Nonces() (n [3]nonce.IV) {
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
