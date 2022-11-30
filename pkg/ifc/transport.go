package ifc

type Transport interface {
	Send(b Message)
	Receive() <-chan Message
}

type Message []byte

func ToMessage(b []byte) (msg Message) { return b }
func (msg Message) ToBytes() []byte    { return msg }
