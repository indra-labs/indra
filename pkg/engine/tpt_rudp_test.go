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
	var sAddrPort netip.AddrPort
	sAddrPort, e = netip.ParseAddrPort(lConn.LocalAddr().String())
	sAddr := &net.UDPAddr{IP: net.ParseIP(sAddrPort.Addr().String()),
		Port: int(sAddrPort.Port())}
	log.D.S("lConn", lConn.LocalAddr().String())
	// log.D.S("rConn", rConn)
	var rc *rudp.Conn
	listener := rudp.NewListener(lConn)
	go func() {
		for {
			data := make([]byte, 1382)
			log.D.Ln("ready rudp", lConn.LocalAddr().String())
			rc, e = listener.AcceptRudp()
			log.D.Ln("reading from rudp", lConn.LocalAddr().String())
			n, _ := rc.Read(data)
			log.D.S("received "+rc.LocalAddr().String()+" from "+
				rc.RemoteAddr().String(), data[:n])
		}
	}()
	time.Sleep(time.Second)
	for i := 0; i < 8; i++ {
		var sConn *net.UDPConn
		if sConn, e = net.DialUDP("udp4", laddr, sAddr); fails(e) {
			t.FailNow()
		}
		sc := rudp.NewConn(sConn, rudp.New())
		msg, _, _ := tests.GenMessage(256, "")
		var n int
		if n, e = sc.Write(msg[:]); fails(e) {
			t.FailNow()
		}
		log.D.S("wrote", msg[:n])
		time.Sleep(time.Second / 4)
	}
	time.Sleep(time.Second)
}
