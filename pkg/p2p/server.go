package p2p

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/cfg"
	"git-indra.lan/indra-labs/indra/pkg/p2p/metrics"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"time"

	"github.com/multiformats/go-multiaddr"

	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/interrupt"
	"git-indra.lan/indra-labs/indra/pkg/p2p/introducer"
)

var (
	userAgent = "/indra:" + indra.SemVer + "/"
)

var (
	privKey         crypto.PrivKey
	p2pHost         host.Host
	seedAddresses   []multiaddr.Multiaddr
	listenAddresses []multiaddr.Multiaddr
	netParams       *cfg.Params
)

func init() {
	seedAddresses = []multiaddr.Multiaddr{}
	listenAddresses = []multiaddr.Multiaddr{}
}

func run() {

	log.I.Ln("starting p2p server")

	log.I.Ln("host id:")
	log.I.Ln("-", p2pHost.ID())

	log.I.Ln("p2p listeners:")
	log.I.Ln("-", p2pHost.Addrs())

	// Here we create a context with cancel and add it to the interrupt handler
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(context.Background())

	interrupt.AddHandler(cancel)

	introducer.Bootstrap(ctx, p2pHost, seedAddresses)

	metrics.SetInterval(30 * time.Second)

	metrics.HostStatus(ctx, p2pHost)

	isReadyChan <- true
}

func Run() {

	//storage.Update(func(txn *badger.Txn) error {
	//	txn.Delete([]byte(storeKeyKey))
	//	return nil
	//})

	configure()

	var err error

	p2pHost, err = libp2p.New(
		libp2p.Identity(privKey),
		libp2p.UserAgent(userAgent),
		libp2p.ListenAddrs(listenAddresses...),
	)

	if check(err) {
		return
	}

	run()
}

func Shutdown() (err error) {

	log.I.Ln("shutting down p2p server")

	if err = p2pHost.Close(); check(err) {
		// continue
	}

	log.I.Ln("- p2p server shutdown complete")

	return
}
