package wire

import (
	"fmt"
	"net"
	"reflect"

	"github.com/Indra-Labs/indra/pkg/ifc"
)

type MessageMagic byte

type Message interface{}

const (
	ReturnMessageMagic MessageMagic = iota
	RelayMessageMagic
	ExitMessage
)

func Serialize(message interface{}) (b ifc.Message, e error) {
	switch m := message.(type) {
	case ReturnMessage:
		b = m.Serialize()
	default:
		e = fmt.Errorf("unknown type %v", reflect.TypeOf(m))
	}
	return
}

func Deserialize(b ifc.Message) (out Message, e error) {
	msgMagic := MessageMagic(b[0])
	switch msgMagic {
	case ReturnMessageMagic:
		out, e = DeserializeReturnMessage(b)
	}
	return
}

func ToReturnMessage(msg Message) (rm *ReturnMessage, e error) {
	switch r := msg.(type) {
	case *ReturnMessage:
		rm = r
	default:
		e = fmt.Errorf("incorrect type returned %v expected %v",
			reflect.TypeOf(r), reflect.TypeOf(&ReturnMessage{}))
	}
	return
}

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
// contains another message, the last one informing the client of the node that
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
	totalLen := 1 + ipLen + len(rm.Message) + 1
	o = make(ifc.Message, totalLen)
	o[0] = byte(ReturnMessageMagic)
	o[1] = byte(ipLen)
	copy(o[2:ipLen+2], rm.IP)
	copy(o[ipLen+2:], rm.Message)
	return
}

func DeserializeReturnMessage(msg ifc.Message) (rm *ReturnMessage, e error) {
	ipLen := int(msg[1])
	rm = &ReturnMessage{
		IP:      net.IP(msg[2 : 2+ipLen]),
		Message: msg[2+ipLen:],
	}
	return
}
