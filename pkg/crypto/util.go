package crypto

// GenerateTestKeyPair generates a key pair.
func GenerateTestKeyPair() (sp *Prv, sP *Pub, e error) {
	if sp, e = GeneratePrvKey(); fails(e) {
		return
	}
	sP = DerivePub(sp)
	return
}

// GenerateTestKeyPairs generates two public/private key pairs.
func GenerateTestKeyPairs() (sp, rp *Prv, sP, rP *Pub, e error) {
	sp, sP, e = GenerateTestKeyPair()
	rp, rP, e = GenerateTestKeyPair()
	return
}
