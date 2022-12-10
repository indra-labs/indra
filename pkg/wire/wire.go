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
	ForwardMagic MessageMagic = iota
	ExitMagic
	ReturnMagic
)

func Serialize(message interface{}) (b ifc.Message, e error) {
	switch m := message.(type) {
	case Forward:
		b = m.Serialize()
	default:
		e = fmt.Errorf("unknown type %v", reflect.TypeOf(m))
	}
	return
}

func Deserialize(b ifc.Message) (out Message, e error) {
	msgMagic := MessageMagic(b[0])
	switch msgMagic {
	case ForwardMagic:
		out, e = DeserializeReturnMessage(b)
	}
	return
}

func ToReturnMessage(msg Message) (rm *Forward, e error) {
	switch r := msg.(type) {
	case *Forward:
		rm = r
	default:
		e = fmt.Errorf("incorrect type returned %v expected %v",
			reflect.TypeOf(r), reflect.TypeOf(&Forward{}))
	}
	return
}

// Forward is just an IP address and a wrapper for another message.
type Forward struct {
	net.IP
	ifc.Message
}

func (rm *Forward) Serialize() (o ifc.Message) {
	ipLen := len(rm.IP)
	totalLen := 1 + ipLen + len(rm.Message) + 1
	o = make(ifc.Message, totalLen)
	o[0] = byte(ForwardMagic)
	o[1] = byte(ipLen)
	copy(o[2:ipLen+2], rm.IP)
	copy(o[ipLen+2:], rm.Message)
	return
}

func DeserializeReturnMessage(msg ifc.Message) (rm *Forward, e error) {
	ipLen := int(msg[1])
	rm = &Forward{
		IP:      net.IP(msg[2 : 2+ipLen]),
		Message: msg[2+ipLen:],
	}
	return
}

type Exit struct {
}

type Return struct {
}
