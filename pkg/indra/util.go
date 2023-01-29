package indra

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
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
	copy(rb[*cur:cur.Inc(crypt.ReverseHeaderLen)], header[:crypt.ReverseHeaderLen])
	copy(rb[crypt.ReverseHeaderLen:], res)
	start := *cur
	for i := range ciphers {
		blk := ciph.BlockFromHash(ciphers[i])
		ciph.Encipher(blk, nonces[2-i], rb[start:])
	}
	return
}
