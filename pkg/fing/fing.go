package fing

const FingerprintLen = 8

// Fingerprint is a truncated SHA256D hash of the pubkey, indicating the relevant
// key when the full Pubkey will be available, and for easier human recognition.
type Fingerprint [FingerprintLen]byte
