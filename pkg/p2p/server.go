package p2p

import (
	"context"
	"github.com/indra-labs/indra/pkg/cfg"
	"github.com/indra-labs/indra/pkg/engine/transport"
	"github.com/indra-labs/indra/pkg/interrupt"
	"github.com/indra-labs/indra/pkg/p2p/metrics"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/spf13/viper"
	"time"

	"github.com/multiformats/go-multiaddr"

	"github.com/indra-labs/indra"
	crypto2 "github.com/indra-labs/indra/pkg/crypto"
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

func Run() {

	//storage.Update(func(txn *badger.Txn) error {
	//	txn.Delete([]byte(storeKeyKey))
	//	return nil
	//})

	configure()

	var e error

	netParams = cfg.SelectNetworkParams(viper.GetString("network"))
	dataPath := viper.GetString("")
	var pkr []byte
	if pkr, e = privKey.Raw(); check(e) {
		return
	}
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	interrupt.AddHandler(cancel)
	pkk := crypto2.PrvKeyFromBytes(pkr)
	keys := crypto2.MakeKeys(pkk)
	var l []*transport.Listener
	var la []string
	for i := range listenAddresses {
		la = append(la, listenAddresses[i].String())
	}
	var list *transport.Listener
	list, e = transport.NewListener(netParams.GetSeedsMultiAddrStrings(),
		la, dataPath, keys, ctx,
		transport.DefaultMTU)
	l = append(l, list)
	p2pHost = list.Host
	log.I.Ln("starting p2p server")

	log.I.Ln("host id:")
	log.I.Ln("-", p2pHost.ID())

	log.I.Ln("p2p listeners:")
	log.I.Ln("-", p2pHost.Addrs())

	metrics.SetInterval(30 * time.Second)

	metrics.HostStatus(ctx, p2pHost)

	isReadyChan <- true
}

func Shutdown() (err error) {

	log.I.Ln("shutting down p2p server")

	if p2pHost != nil {
		if err = p2pHost.Close(); check(err) {
			// continue
		}
	}

	log.I.Ln("- p2p server shutdown complete")

	return
}
