package engine

import (
	"context"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/multiformats/go-multiaddr"

	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/engine/node"
	"git-indra.lan/indra-labs/indra/pkg/engine/transport"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func TestEngine_Dispatcher(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	log.D.Ln(os.PathSeparator)
	_ = e
	const nTotal = 26
	ctx, cancel := context.WithCancel(context.Background())
	var listeners []*transport.Listener
	var keys []*crypto.Keys
	var nodes []*node.Node
	var engines []*Engine
	var seed string
	for i := 0; i < nTotal; i++ {
		var k *crypto.Keys
		if k, e = crypto.GenerateKeys(); fails(e) {
			t.FailNow()
		}
		keys = append(keys, k)
		var l *transport.Listener
		dataPath, err := os.MkdirTemp(os.TempDir(), "badger")
		if err != nil {
			t.FailNow()
		}
		log.T.Ln(
			"dataPath", dataPath,
		)
		if l, e = transport.NewListener(seed, transport.LocalhostZeroIPv4TCP,
			dataPath, k, ctx, transport.DefaultMTU); fails(e) {
			t.FailNow()
		}
		sa := transport.GetHostAddress(l.Host)
		if i == 0 {
			seed = sa
		}
		listeners = append(listeners, l)
		var addr netip.AddrPort
		var ma multiaddr.Multiaddr
		if ma, e = multiaddr.NewMultiaddr(sa); fails(e) {
			t.FailNow()
		}

		var ip, port string
		if ip, e = ma.ValueForProtocol(multiaddr.P_IP4); fails(e) {
			// we specified ipv4 previously.
			t.FailNow()
		}
		if port, e = ma.ValueForProtocol(multiaddr.P_TCP); fails(e) {
			t.FailNow()
		}
		if addr, e = netip.ParseAddrPort(ip + ":" + port); fails(e) {
			t.FailNow()
		}
		var nod *node.Node
		if nod, _ = node.NewNode(&addr, k, nil, 50000); fails(e) {
			t.FailNow()
		}
		nodes = append(nodes, nod)
		var eng *Engine
		if eng, e = NewEngine(Params{
			Listener: l,
			Keys:     k,
			Node:     nod,
		}); fails(e) {
			t.FailNow()
		}
		engines = append(engines, eng)
	}
	time.Sleep(time.Second * 2)
	cancel()
}
