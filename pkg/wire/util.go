package wire

import (
	"github.com/indra-labs/indra/pkg/key/ecdh"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
)

func GenCiphers(prvs [3]*prv.Key, pubs [3]*pub.Key) (ciphers [3]sha256.Hash) {
	for i := range prvs {
		ciphers[2-i] = ecdh.Compute(prvs[i], pubs[i])
	}
	return
}

func Gen3Nonces() (n [3]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GenNonces(count int) (n []nonce.IV) {
	n = make([]nonce.IV, count)
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GenPingNonces() (n [4]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}
