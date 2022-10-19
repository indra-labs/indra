package fing

const FingerprintLen = 8

// Fingerprint is a truncated SHA256D hash of the pubkey, indicating the relevant
// key when the full Pubkey will be available, and for easier human recognition.
// Being an array means it can be compared with `==` same as a string.
type Fingerprint [FingerprintLen]byte
