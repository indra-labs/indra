package pub

//
// func TestBase32(t *testing.T) {
// 	for i := 0; i < 1000; i++ {
// 		var k *prv.Prv
// 		var e error
// 		if k, e = prv.GenerateKey(); check(e) {
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 		p := Derive(k)
// 		b32 := p.ToBase32()
// 		log.I.Ln(b32)
// 		var kk *Prv
// 		kk, e = FromBase32(b32)
// 		if b32 != kk.ToBase32() {
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 	}
// }
