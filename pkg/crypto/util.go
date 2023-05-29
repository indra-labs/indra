package crypto

func GenerateTestKeyPair() (sp *Prv, sP *Pub, e error) {
	if sp, e = GeneratePrvKey(); fails(e) {
		return
	}
	sP = DerivePub(sp)
	return
}

func GenerateTestKeyPairs() (sp, rp *Prv, sP, rP *Pub, e error) {
	sp, sP, e = GenerateTestKeyPair()
	rp, rP, e = GenerateTestKeyPair()
	return
}
