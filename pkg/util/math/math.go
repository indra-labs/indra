package math

// MinUint32 is a helper function to return the minimum of two uint32s. This avoids a math import and the need to cast
// to floats.
func MinUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
