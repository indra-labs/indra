package ifc

type Transport interface {
	Send(b Message)
	Receive() <-chan Message
}

type Message []byte
