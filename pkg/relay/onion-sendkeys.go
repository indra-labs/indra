package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/messages/session"
)

// SendKeys provides a pair of private keys that will be used to generate the
// Purchase header bytes and to generate the ciphers provided in the Purchase
// message to encrypt the Session that is returned.
//
// The OnionSkin key, its cloaked public key counterpart used in the ToHeaderPub
// field of the Purchase message preformed header bytes, but the Ciphers
// provided in the Purchase message, for encrypting the Session to be returned,
// uses the Payload key, along with the public key found in the encrypted crypt
// of the header for the Reverse relay.
//
// This message's last crypt is a Confirmation, which allows the client to know
// that the keys were successfully delivered.
//
// This is the only onion that uses the node identity keys. The payment preimage
// hash must be available or the relay should not forward the remainder of the
// packet.
//
// If hdr/pld cipher keys are nil there must be a HeaderPub available on the
// session for the hop. This allows this function to send keys to any number of
// hops, but the very first SendKeys must have all in order to create the first
// set of sessions. This is by way of indicating to not use the IdentityPub but
// the HeaderPub instead. Not allowing free relay at all prevents spam attacks.
func SendKeys(id nonce.ID, s [5]*session.Layer,
	client *Session, hop []*Node, ks *signer.KeySet) Skins {
	
	n := GenNonces(6)
	sk := Skins{}
	for i := range s {
		if i == 0 {
			sk = sk.Crypt(hop[i].IdentityPub, nil, ks.Next(),
				n[i], 0).Session(s[i])
		} else {
			sk = sk.ForwardSession(hop[i], ks.Next(), n[i], s[i])
		}
	}
	return sk.
		ForwardCrypt(client, ks.Next(), n[5]).
		Confirmation(id, 0)
}
