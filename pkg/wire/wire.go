package wire

import (
	"net"

	"github.com/Indra-Labs/indra/pkg/ifc"
)

type RelayMsg struct {
}

type ExitMsg struct {
}

// ReturnMessage is just an IP address and a wrapper for another return message.
// These are wrapped in three onion layers, the final one being a secret
// provided by the client to identify the private key of the message, and the
// hop in the path that has successfully been relayed that was sent out that it
// relates to.
//
// These messages are used to trace the progress of messages through a circuit.
// If they do not return, either the hop itself, or one of the three relays used
// in the ReturnMessage onion are not working. These relays will also be used in
// parallel in other circuits and so by this in a process of elimination the
// dead relays can be eliminated by the ones that are known and responding.
//
// In a standard path diagnostic onion, one of these contains another, which
// contains another message, the last one informing the node of the node that
// succeeded in the path trace.
//
// The last layer of the onion contains the public key of the hop this
// return was sent from.
type ReturnMessage struct {
	net.IP
	ifc.Message
}

func (rm *ReturnMessage) Serialize() (o ifc.Message) {
	ipLen := len(rm.IP)
	totalLen := 1 + ipLen + len(rm.Message)
	o = make(ifc.Message, totalLen)
	o[0] = byte(ipLen)
	copy(o[1:ipLen+1], rm.IP)
	copy(o[ipLen+1:], rm.Message)
	return
}

func Deserialize(msg ifc.Message) (rm *ReturnMessage) {
	ipLen := int(msg[0])
	rm = &ReturnMessage{
		IP:      net.IP(msg[1 : 1+ipLen]),
		Message: msg[1+ipLen:],
	}
	return
}
