package transport

type Message []byte

type Dispatcher chan Message

func (d Dispatcher) Send(b []byte) {
	d <- b
}

func (d Dispatcher) Receive() <-chan Message {
	return d
}
