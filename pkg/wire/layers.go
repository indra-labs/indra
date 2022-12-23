package wire

import (
	"net"

	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirmation"
	"github.com/Indra-Labs/indra/pkg/wire/exit"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	"github.com/Indra-Labs/indra/pkg/wire/message"
	"github.com/Indra-Labs/indra/pkg/wire/purchase"
	"github.com/Indra-Labs/indra/pkg/wire/reply"
	"github.com/Indra-Labs/indra/pkg/wire/response"
	"github.com/Indra-Labs/indra/pkg/wire/session"
	"github.com/Indra-Labs/indra/pkg/wire/token"
)

type OnionSkins []types.Onion

func (o OnionSkins) Message(to *address.Sender, from *prv.Key) OnionSkins {
	return append(o, &message.Type{To: to, From: from})
}
func (o OnionSkins) Confirmation(id nonce.ID) OnionSkins {
	return append(o, &confirmation.Type{ID: id})
}
func (o OnionSkins) Forward(ip net.IP) OnionSkins {
	return append(o, &forward.Type{IP: ip})
}
func (o OnionSkins) Exit(port uint16, ciphers [3]sha256.Hash,
	payload slice.Bytes) OnionSkins {

	return append(o, &exit.Type{Port: port, Ciphers: ciphers, Bytes: payload})
}
func (o OnionSkins) Return(ip net.IP) OnionSkins {
	return append(o, &reply.Type{IP: ip})
}
func (o OnionSkins) Cipher(hdr, pld *prv.Key) OnionSkins {
	return append(o, &cipher.Type{Header: hdr, Payload: pld})
}
func (o OnionSkins) Purchase(nBytes uint64, ciphers [3]sha256.Hash) OnionSkins {
	return append(o, &purchase.Type{NBytes: nBytes, Ciphers: ciphers})
}
func (o OnionSkins) Session(fwd, rtn *pub.Key) OnionSkins {
	return append(o, &session.Type{
		HeaderKey: fwd, PayloadKey: rtn,
	})
}
func (o OnionSkins) Response(res slice.Bytes) OnionSkins {
	return append(o, response.Response(res))
}
func (o OnionSkins) Token(tok sha256.Hash) OnionSkins {
	return append(o, token.Type(tok))
}

// Assemble inserts the slice of Onion s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o OnionSkins) Assemble() (on types.Onion) {
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
