package rpc

import (
	"github.com/multiformats/go-multiaddr"
	"math/rand"
	"time"
)

var (
	config = rpcConfig{
		key:           &nullRPCPrivateKey,
		listenPort:    NullPort,
		peerWhitelist: []RPCPublicKey{},
		ipWhitelist:   []multiaddr.Multiaddr{},
	}
)

type rpcConfig struct {
	key           *RPCPrivateKey
	listenPort    uint16
	peerWhitelist []RPCPublicKey
	ipWhitelist   []multiaddr.Multiaddr
	unixPath      string
}

func (c *rpcConfig) newKey() {

	var err error

	if c.key, err = NewPrivateKey(); check(err) {
		panic(err)
	}
}

func (c *rpcConfig) setKey(key string) {
	c.key.Decode(key)
}

func (c *rpcConfig) isNullKey() bool {
	return c.key.IsZero()
}

func (c *rpcConfig) setPort(port uint16) {
	c.listenPort = port
}

func (c *rpcConfig) isNullPort() bool {
	return c.listenPort == NullPort
}

func (c *rpcConfig) setRandomPort() uint16 {
	rand.Seed(time.Now().Unix())

	c.listenPort = uint16(rand.Intn(45534) + 10000)

	return c.listenPort
}

func (c *rpcConfig) setUnixPath(path string) {
	c.unixPath = path
}

func (conf *rpcConfig) isEnabled() bool {
	return !conf.key.IsZero()
}
