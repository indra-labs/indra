package wire

import (
	"net"

	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
)

type OnionSkins []Onion

func (o OnionSkins) Message(to *address.Sender, from *prv.Key) OnionSkins {
	return append(o, &Message{To: to, From: from})
}
func (o OnionSkins) Forward(ip net.IP) OnionSkins {
	return append(o, &Forward{IP: ip})
}
func (o OnionSkins) Exit(port uint16, ciphers [3]sha256.Hash,
	payload slice.Bytes) OnionSkins {

	return append(o, &Exit{Port: port, Cipher: ciphers, Bytes: payload})
}
func (o OnionSkins) Return(ip net.IP, fwd, rtn pub.Key) OnionSkins {
	return append(o, &Return{IP: ip, Forward: fwd, Return: rtn})
}
func (o OnionSkins) Cipher(id nonce.ID, key pub.Key) OnionSkins {
	return append(o, &Cipher{ID: id, Key: key.ToBytes()})
}
func (o OnionSkins) Purchase(value uint64) OnionSkins {
	return append(o, &Purchase{Value: value})
}
func (o OnionSkins) Session(fwd, rtn pub.Key) OnionSkins {
	return append(o, &Session{
		ForwardKey: fwd.ToBytes(), ReturnKey: rtn.ToBytes(),
	})
}
func (o OnionSkins) Acknowledgement(id nonce.ID) OnionSkins {
	return append(o, &Acknowledgement{ID: id})
}
func (o OnionSkins) Response(res slice.Bytes) OnionSkins {
	return append(o, Response(res))
}
func (o OnionSkins) Token(tok sha256.Hash) OnionSkins {
	return append(o, Token(tok))
}

// Assemble inserts the slice of Onion s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o OnionSkins) Assemble() (on Onion) {
	// First item is the outer layer.
	on = o[0]
	// Iterate through the remaining layers.
	for _, oc := range o[1:] {
		on.Insert(oc)
		// Next step we are inserting inside the one we just inserted.
		on = oc
	}
	// At the end, the first element contains references to every element
	// inside it.
	return o[0]
}
