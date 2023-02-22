package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
)

// GetBalance sends out a request in a similar way to SendExit except the node
// being queried can be any of the 5.
func GetBalance(id, confID nonce.ID, client *Session,
	s Circuit, ks *signer.KeySet) Skins {
	
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	n := GenNonces(6)
	var retNonces [3]nonce.IV
	copy(retNonces[:], n[3:])
	var pubs [3]*pub.Key
	pubs[0] = s[3].PayloadPub
	pubs[1] = s[4].PayloadPub
	pubs[2] = client.PayloadPub
	return Skins{}.
		ReverseCrypt(s[0], ks.Next(), n[0], 3).
		ReverseCrypt(s[1], ks.Next(), n[1], 2).
		ReverseCrypt(s[2], ks.Next(), n[2], 1).
		GetBalance(id, confID, prvs, pubs, retNonces).
		ReverseCrypt(s[3], prvs[0], n[3], 0).
		ReverseCrypt(s[4], prvs[1], n[4], 0).
		ReverseCrypt(client, prvs[2], n[5], 0)
}
