package transport

import "github.com/libp2p/go-libp2p/core/peer"

func (l *Listener) Publish(p peer.ID, key string, val interface{}) error {
	return l.Host.Peerstore().Put(p, key, val)
}

func (l *Listener) FindPeerRecord(p peer.ID, key string) (val interface{}, e error) {
	return l.Host.Peerstore().Get(p, key)
}
