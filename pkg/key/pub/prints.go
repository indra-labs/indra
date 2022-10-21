package pub

const (
	Len         = 8
	ReceiverLen = 12
)

type (
	Print    []byte
	Receiver []byte
)

// New empty public Print fingerprint.
func New() Print { return make(Print, Len) }

// NewReceiver makes a new empty Receiver public key fingerprint.
func NewReceiver() Receiver { return make(Receiver, ReceiverLen) }

// GetPrints creates a slice of fingerprints from a set of public keys.
func GetPrints(keys ...*Key) (fps []Print) {
	for i := range keys {
		fps = append(fps, keys[i].ToBytes().Fingerprint())
	}
	return
}
