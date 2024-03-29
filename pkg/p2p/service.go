package p2p

import (
	"context"
	"crypto/rand"
	"git.indra-labs.org/dev/ind/pkg/cfg"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"git.indra-labs.org/dev/ind/pkg/engine/transport"
	"git.indra-labs.org/dev/ind/pkg/interrupt"
	"git.indra-labs.org/dev/ind/pkg/p2p/metrics"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/spf13/viper"
	"time"

	"github.com/multiformats/go-multiaddr"

	"git.indra-labs.org/dev/ind"
	crypto2 "git.indra-labs.org/dev/ind/pkg/crypto"
)

var (
	userAgent       = "/indra:" + indra.SemVer + "/"
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

// Run is the main entrypoint for the seed p2p service.
func Run() {

	// storage.Update(func(txn *badger.Txn) error {
	//	txn.Delete([]byte(storeKeyKey))
	//	return nil
	// })

	configure()

	var e error

	netParams = cfg.SelectNetworkParams(viper.GetString("network"))
	dataPath := viper.GetString("data-dir")
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
	secret := sha256.New()
	rand.Read(secret[:])
	store, closer := transport.BadgerStore(dataPath, secret[:])
	if store == nil {
		panic("could not open database")
	}
	list, e = transport.NewListener(netParams.GetSeedsMultiAddrStrings(),
		la, keys, store, closer, ctx, transport.DefaultMTU, cancel)
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
