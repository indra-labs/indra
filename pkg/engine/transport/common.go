package transport

import (
	"fmt"
	"git.indra-labs.org/dev/ind"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/options"
	"github.com/gookit/color"
	badger "github.com/indra-labs/go-ds-badger3"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

const (
	// LocalhostZeroIPv4TCP is the default localhost to bind to any address.
	// Used in tests.
	LocalhostZeroIPv4TCP = "/ip4/127.0.0.1/tcp/0"

	// LocalhostZeroIPv6TCP is the default localhost to bind to any address.
	// Used in tests.
	LocalhostZeroIPv6TCP = "/ip6/::1/tcp/0"

	// LocalhostZeroIPv4QUIC - Don't use. Buffer problems on linux and fails on CI.
	// LocalhostZeroIPv4QUIC = "/ip4/127.0.0.1/udp/0/quic"

	// DefaultMTU is the default maximum size for a packet.
	DefaultMTU = 1382

	// ConnBufs is the number of buffers to use in message dispatch channels.
	ConnBufs = 8192

	// IndraLibP2PID is the indra protocol identifier.
	IndraLibP2PID = "/indra/relay/" + indra.SemVer
)

var (
	DefaultUserAgent = "/indra:" + indra.SemVer + "/"

	blue  = color.Blue.Sprint
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// SetUserAgent changes the user agent. Note that this will only have an effect
// before a new listener is created.
func SetUserAgent(s string) {
	DefaultUserAgent = "/indra " + indra.SemVer + " [" + s + "]/"
}

// GetHostFirstMultiaddr returns the multiaddr string encoding of a host.Host's
// network listener.
func GetHostFirstMultiaddr(ha host.Host) string {
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s",
		ha.ID().String()))
	addr := ha.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

// GetHostOnlyFirstMultiaddr returns the multiaddr string without the p2p key.
func GetHostOnlyFirstMultiaddr(ha host.Host) string {
	addr := ha.Addrs()[0]
	return addr.String()
}

// GetHostMultiaddrs returns the multiaddr strings encoding of a host.Host's
// network listener.
//
// This includes (the repeated) p2p key sections of the peer identity key.
func GetHostMultiaddrs(ha host.Host) (addrs []string) {
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s",
		ha.ID().String()))
	for _, v := range ha.Addrs() {
		addrs = append(addrs, v.Encapsulate(hostAddr).String())
	}
	return
}

// GetHostOnlyMultiaddrs returns the multiaddr string without the p2p key.
func GetHostOnlyMultiaddrs(ha host.Host) (addrs []string) {
	for _, v := range ha.Addrs() {
		addrs = append(addrs, v.String())
	}
	return
}

// BadgerStore creates a new badger database backed persistence engine for keys
// and values used in the peer information database, basically keeps a
// collection of all received advertisments on the network keyed to their public
// key.
//
// This database includes a data structure that maps and records history related
// to:
//
//   - keys of hidden connections - having several past versions can help avoid
//     needing to find an introducer again.
//   - keys of relay to relay and client to relay connections receiver public
//     keys.
//
// This database must be encrypted as the data stored in it includes the
// libp2p.Host private key.
func BadgerStore(dataPath string,
	cipher []byte) (store *badger.Datastore,
	closer func()) {

	if cipher == nil {
		panic("DANGER: database is being created without encryption")
	}

	log.T.Ln("dataPath", dataPath)
	o := options.Default(dataPath, cipher[:])
	opts := &badger.Options{Options: *o}
	store, err := badger.NewDatastore(dataPath, opts)
	if fails(err) {
		return nil, func() {}
	}
	closer = func() {
		store.Close()
	}
	return store, closer
}
