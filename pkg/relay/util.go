package relay

import (
	"fmt"
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/ecdh"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/crypt"
	"git-indra.lan/indra-labs/indra/pkg/relay/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func BudgeUp(b slice.Bytes, start slice.Cursor) (o slice.Bytes) {
	o = b
	copy(o, o[start:])
	copy(o[len(o)-int(start):], slice.NoisePad(int(start)))
	return
}

func FormatReply(header, res slice.Bytes, ciphers [3]sha256.Hash,
	nonces [3]nonce.IV) (rb slice.Bytes) {
	
	rb = make(slice.Bytes, crypt.ReverseHeaderLen+len(res))
	cur := slice.NewCursor()
	copy(rb[*cur:cur.Inc(crypt.ReverseHeaderLen)],
		header[:crypt.ReverseHeaderLen])
	copy(rb[crypt.ReverseHeaderLen:], res)
	start := *cur
	for i := range ciphers {
		blk := ciph.BlockFromHash(ciphers[i])
		ciph.Encipher(blk, nonces[2-i], rb[start:])
	}
	return
}

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

func GenPingNonces() (n [6]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func recLog(on types.Onion, b slice.Bytes, cl *Engine) func() string {
	return func() string {
		return cl.GetLocalNodeAddress().String() +
			" received " +
			fmt.Sprint(reflect.TypeOf(on)) + "\n" +
			""
		// spew.Sdump(b.ToBytes())
	}
}
