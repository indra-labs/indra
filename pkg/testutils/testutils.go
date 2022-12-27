package testutils

import (
	"crypto/rand"
	"fmt"
	"net/netip"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func GenerateTestMessage(msgSize int) (msg []byte, hash sha256.Hash, e error) {
	msg = make([]byte, msgSize)
	var n int
	if n, e = rand.Read(msg); check(e) && n != msgSize {
		return
	}
	copy(msg, "payload")
	hash = sha256.Single(msg)
	return
}

func GenerateTestKeyPairs() (sp, rp *prv.Key, sP, rP *pub.Key, e error) {
	if sp, e = prv.GenerateKey(); check(e) {
		return
	}
	sP = pub.Derive(sp)
	if rp, e = prv.GenerateKey(); check(e) {
		return
	}
	rP = pub.Derive(rp)
	return
}

func GenerateRandomAddrPortIPv4() (ap *netip.AddrPort) {
	a := netip.AddrPort{}
	b := make([]byte, 7)
	_, e := rand.Read(b)
	if check(e) {
		log.E.Ln(e)
	}
	port := slice.DecodeUint16(b[5:7])
	str := fmt.Sprintf("%d.%d.%d.%d:%d", b[1], b[2], b[3], b[4], port)
	a, e = netip.ParseAddrPort(str)
	return &a
}
