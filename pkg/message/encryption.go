package message

type Initial struct {
	PubKey    []byte
	Nonce     []byte
	Message   []byte
	Signature []byte
}

type Subsequent struct {
	Fingerprint []byte
	Nonce       []byte
	Message     []byte
	Signature   []byte
}
