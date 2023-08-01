package transport

import (
	"bufio"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/engine/tpt"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	"sync"
)

// Conn is a net.Conn implementation with the additional features required by
// Indra.
type Conn struct {

	// Conn is the actual network connection, which is a ReaderWriterCloser.
	network.Conn

	// MTU is the size of packet segments handled on the connection.
	MTU int

	// RemoteKey is the receiver public key messages should be encrypted to.
	//
	// todo: this is also handled by the dispatcher for key changes etc?
	RemoteKey *crypto.Pub

	// MultiAddr is the multiaddr.Multiaddr of the other side of the
	// connection.
	MultiAddr multiaddr.Multiaddr

	// Host is the libp2p host implementing the Conn.
	Host host.Host

	// rw is the read-write interface to the Conn.
	//
	// todo: isn't the Conn supposed to do this also???
	rw *bufio.ReadWriter

	// Transport is the duplex channel that is given to calling code to
	// dispatch messages through the Conn.
	Transport *DuplexByteChan

	// Mutex to prevent concurrent read/write of shared data.
	sync.Mutex

	// C can be closed to shut down the connection, and closes the Conn.
	qu.C
}

// GetMTU returns the Maximum Transmission Unit (MTU) of the Conn.
func (c *Conn) GetMTU() int {
	c.Lock()
	defer c.Unlock()
	return c.MTU
}

// GetRecv returns the Transport that is functioning as receiver, used to
// receive messages.
func (c *Conn) GetRecv() tpt.Transport { return c.Transport.Receiver }

// GetRemoteKey returns the current remote receiver public key we want to
// encrypt to (with ECDH).
func (c *Conn) GetRemoteKey() (remoteKey *crypto.Pub) {
	c.Lock()
	defer c.Unlock()
	return c.RemoteKey
}

// GetSend returns the Transport that is functioning as sender, used to send
// messages.
func (c *Conn) GetSend() tpt.Transport { return c.Transport.Sender }

// SetMTU defines the size of the packets messages will be segmented into.
func (c *Conn) SetMTU(mtu int) {
	c.Lock()
	c.MTU = mtu
	c.Unlock()
}

// SetRemoteKey changes the key that should be used with ECDH to generate
// message encryption secrets. This will be called in response to the other side
// sending a key change message.
func (c *Conn) SetRemoteKey(remoteKey *crypto.Pub) {
	c.Lock()
	c.RemoteKey = remoteKey
	c.Unlock()
}
