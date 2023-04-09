package engine

import (
	"fmt"
	"net"
	"net/netip"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"github.com/davecgh/go-spew/spew"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/rudp"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestRCP(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	var one, two *Keys
	var lis *RCP
	if one, two, e = Generate2Keys(); fails(e) {
		t.FailNow()
	}
	_ = two
	quit := qu.T()
	var ks *signer.KeySet
	_, ks, e = signer.New()
	listenerAddress := "127.0.0.1"
	if lis, e = NewListenerRCP(one, listenerAddress, 64, 1382, ks,
		quit); fails(e) {
		t.FailNow()
	}
	var outbound *RCP
	go func() {
		log.D.Ln("listening")
		b := <-lis.Receive()
		log.I.S("bytes", b)
		fmt.Println("bytes", spew.Sdump(b))
		time.Sleep(time.Second)
	}()
	time.Sleep(time.Second)
	outbound, e = NewOutboundRCP(two, lis.uConn.LocalAddr().String(), one.Pub,
		"127.0.0.1:0", 64, 1382, ks, quit)
	var msg []byte
	if msg, _, e = tests.GenMessage(256, "b0r0k"); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	if e = outbound.Send(msg); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	lis.Stop()
	outbound.Stop()
	<-quit
}

func TestRCPGeneral(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	laddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
	var e error
	var lConn *net.UDPConn
	if lConn, e = net.ListenUDP("udp4", laddr); fails(e) {
		t.FailNow()
	}
	var rAddrPort netip.AddrPort
	rAddrPort, e = netip.ParseAddrPort(lConn.LocalAddr().String())
	raddr := &net.UDPAddr{IP: net.ParseIP(rAddrPort.Addr().String()),
		Port: int(rAddrPort.Port())}
	log.D.S("lConn", lConn.LocalAddr().String())
	var rConn *net.UDPConn
	if rConn, e = net.DialUDP("udp4", laddr, raddr); fails(e) {
		t.FailNow()
	}
	sc := rudp.NewConn(rConn, rudp.New())
	// log.D.S("rConn", rConn)
	var rc *rudp.Conn
	listener := rudp.NewListener(lConn)
	go func() {
		for {
			log.D.Ln("starting listener")
			data := make([]byte, rudp.MAX_PACKAGE)
			log.D.Ln("accepting rudp")
			rc, e = listener.AcceptRudp()
			log.D.Ln("reading from rudp")
			n, err := rc.Read(data)
			log.D.S("received", n, err, data[:n])
		}
	}()
	time.Sleep(time.Second)
	_, msg, _ := tests.GenMessage(256, "alpha")
	_ = msg
	var n int
	if n, e = sc.Write(msg[:]); fails(e) {
		t.FailNow()
	}
	log.D.S("wrote", msg[:n])
	time.Sleep(time.Second)
}
