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

func NewOnion() (o OnionSkins) { return o }

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

func (o OnionSkins) Assemble() (on Onion) {
	for i := range o {
		_ = i
	}
	return
}

/*
func (o OnionSkins) () OnionSkins {
	return append(o, &{})
}
*/
