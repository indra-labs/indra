package rpc

import (
	"github.com/multiformats/go-multiaddr"
	"math/rand"
	"time"
)

type rpcConfig struct {
	Key            *RPCPrivateKey
	ListenPort     uint16
	Peer_Whitelist []RPCPublicKey
	IP_Whitelist   []multiaddr.Multiaddr
	UnixPath       string
}

func (c *rpcConfig) NewKey() {

	var err error

	if c.Key, err = NewPrivateKey(); check(err) {
		panic(err)
	}
}

func (c *rpcConfig) SetKey(key string) {
	c.Key.Decode(key)
}

func (c *rpcConfig) IsNullKey() bool {
	return c.Key.IsZero()
}

func (c *rpcConfig) SetPort(port uint16) {
	c.ListenPort = port
}

func (c *rpcConfig) IsNullPort() bool {
	return c.ListenPort == NullPort
}

func (c *rpcConfig) SetRandomPort() uint16 {
	rand.Seed(time.Now().Unix())

	c.ListenPort = uint16(rand.Intn(45534) + 10000)

	return c.ListenPort
}

func (c *rpcConfig) SetUnixPath(path string) {
	c.UnixPath = path
}

func (conf *rpcConfig) IsEnabled() bool {
	return !conf.Key.IsZero()
}
