package pub

const (
	PrintLen = 8
)

type (
	Print []byte
)

// GetPrints creates a slice of fingerprints from a set of public keys.
func GetPrints(keys ...*Key) (fps []Print) {
	for i := range keys {
		fps = append(fps, keys[i].ToBytes().Fingerprint())
	}
	return
}
