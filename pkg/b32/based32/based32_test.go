package based32

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
)

const (
	seed    = 1234567890
	numKeys = 32
)

func TestCodec(t *testing.T) {
	
	// Generate 10 pseudorandom 64 bit values. We do this here rather than
	// pre-generating this separately as ultimately it is the same thing, the
	// same seed produces the same series of pseudorandom values, and the hashes
	// of these values are deterministic.
	rand.Seed(seed)
	seeds := make([]uint64, numKeys)
	for i := range seeds {
		
		seeds[i] = rand.Uint64()
	}
	
	// Convert the uint64 values to 8 byte long slices for the hash function.
	seedBytes := make([][]byte, numKeys)
	for i := range seedBytes {
		
		seedBytes[i] = make([]byte, 8)
		binary.LittleEndian.PutUint64(seedBytes[i], seeds[i])
	}
	
	// Generate hashes from the seeds
	hashedSeeds := make([][]byte, numKeys)
	
	// Uncomment lines relating to this variable to regenerate expected data
	// that will log to console during test run.
	generated := "\nexpected := []string{\n"
	
	for i := range hashedSeeds {
		
		hashed := sha256.Single(seedBytes[i])
		hashedSeeds[i] = hashed[:]
		
		generated += fmt.Sprintf("\t\"%x\",\n", hashedSeeds[i])
	}
	
	generated += "}\n"
	// t.Log(generated)
	
	expected := []string{
		"ee94d6cef460b180c995b2f8672e53006aced15fe4d5cc0da332d041feaa1514",
		"0f92907a4d76ece96b042e2cbd60e2378039cc92c95ac99f73e8eacbdd38a7d3",
		"dc0892182d0f4dd8643d6e1c29442cf96c2d0a0a985b747a747d96a4f87e06dc",
		"fa1066bd91acc3a16eb08d6c5ef5893ff8a0d01525bb30cd6be66cea34f3b4b6",
		"7eef96527a625f6489e1ca37377184daaa7d4ceb3cafc091f34fdc0357101fab",
		"5ea29a714835e54ae1fd5549e10436a2619d1b8ae909468d3700903ae871c8c0",
		"41070be84762fc76a36c0f3506c3dc90e78fc12ac5f3cd3e38c6e73c6d6ff427",
		"2e19378670d2dd76d89f9e29d28213b0f2e0dd673ad6b9c5ab27b34772ca30f3",
		"343134a858ca19cc988f30a2503729dcb83a544e2cc7eb3ca637110759afe782",
		"70548744c390460b47a035dcbc7a72534172fa7ec1260659830bc587ea78ce18",
		"13adbec37cbbe311fca9c9d37a884cad590ab362615cbf0ea275ab4d29c77a8d",
		"ff145ee2c983438b15d3111365e45a8f4c7390e0d2d3e750036bb6b97dc72f96",
		"f9fea53c3eac4866e637e11afe1766f22a168f9e99e8998d4d5c4cd885a99811",
		"5b9ca521047cc06261acc6b3dcb7c6ac340a0b384a464987c7a45ff5c2283707",
		"7f0451e8c9a294335238839159fc2ee850ac21b234444fef8af2088b2661103a",
		"24fe6c69e5217befdf0325f52e35673f1cb5f674592fd82c612931ebaa22c37e",
		"d89275ca53104332d20acd14d3112a08684be50f4947c730ece6b3443c444a5f",
		"02674760e23fb0c5780e2514c2aeffa797207b2db97f4abf7208ed396d0d48b3",
		"da477ff2ef2f9194bb21ca766038b120e2068fcb0662c4f63e39eeb68c9c1631",
		"9435716f250de2d33fe4c76d143d31ffa7e1d536f64456625a5b52d7c5bb1ff1",
		"b79033f579221800651b767612ece7f8b08f4a52565f72ef1ceca707c8d0ffb1",
		"bd451c36d6487842378951ca94725699ccb28fecab1851ea50e8073a68e1ee44",
		"be94236d0204998274ed5ae3ea7198b7f839f3642b04c83b35e37a48ba13b186",
		"017d82fb33d0f1f0873a18d8dafa9b85b35ec70af1715d3f9d3d204532b3660e",
		"97047d8ec8f6f49ea7152e6626e1c7e8e32c2e9dc6a60b6c1030b654772883a2",
		"2634e9a3bf48d55eab32623b14b323ea4d3603e4c5fce573bfd7ebae33e69eaf",
		"f8bc405edbaa4423f7b272649d79495c5cd0dbd39cb60484e9c3f6b828b320fc",
		"d8a2f7aa2021e3c77cd04df8b60330c5b79d3cc5cdd156e86fb3a0fb34b0685d",
		"5c381d4c470c99d7beb596a359be35fd9bb455b088031c45368b9928ce66a774",
		"83b77abed4c677e169802de0c4b6176230fe4e673fa29b801fbdbde34d1e47e7",
		"12b40270e989ddc550f74a2a66f6092903fe0ec075df2826148fa9080aa933b3",
		"4db6259bb154e007bfe5be06a641bb3a797b4deaa9447d2f6d4deeed3f6ad07a",
	}
	
	for i := range hashedSeeds {
		
		if expected[i] != hex.EncodeToString(hashedSeeds[i]) {
			
			t.Log("failed", i, "expected", expected[1], "found", hashedSeeds)
			t.FailNow()
		}
	}
	
	encodedStr := []string{
		"thff4kw73strqxfjke0seew2vqx4nk3tljdtnqd5vedqs074g23fwfq",
		"v8e9yr6f4mwe6ttqshze0tqugmcqwwvjty44jvlw05w4j7a8zneeksz",
		"nwq3ysc9585mkry84hpc22y9nukctg2p2v9kar6w37edf8c06ed7sru",
		"hapqe4ajxkv8gtwkzxkchh43yll3gxsz5jmkvxdd0nxe635xrtk9zqe",
		"elwl9jj0f397eyfu89rwdm3snd25l2vav72lsy37d8acqeu4l3qqsrl",
		"f029xn3fq672jhpl425ncgyx63xr8gm3t5sj35dxuqfqwhgw8yvq9cn",
		"dqswzlgga30ca4rds8n2pkrmjgw0r7p9tzl8nf78rrww0rddl600me5",
		"shpjduxwrfd6akcn70zn55zzwc09cxavuaddww94vnmx3mjefqdjt5j",
		"56rzd9gtr9pnnyc3uc2y5ph98wtswj5fckv06eu5cm3zp6eretjkdqg",
		"ec9fp6ycwgyvz685q6ae0r6wff5zuh60mqjvpjesv9utpmyd2k8kxxa",
		"gf6m0kr0ja7xy0u48yax75gfjk4jz4nvfs4e0cw5f66knffcaagmvdw",
		"0l3ghhzexp58zc46vg3xe0yt285cuusurfd8e6sqd4mdwtacuhanry0",
		"nulaffu86kysehxxls34lshvmez5950n6v73xvdf4wyeky948z964ua",
		"4deeffpq37vqcnp4nrt8h9hc6krgzst8p9yvjv8c7j9lawz0cn952qs",
		"elsg50gex3fgv6j8zpezk0u9m59ptppkg6ygnl03teq3zlslxjynhwa",
		"gj0umrfu5shhm7lqvjl2t34vul3ed0kw3vjlkpvvy5nr6a2ytpham9a",
		"0vfyaw22vgyxvkjptx3f5c39gyxsjl9pay503esanntx3pug39f7jm6",
		"spxw3mquglmp3tcpcj3fs4wl7newgrm9kuh7j4lwgyw6wtdp5ph9mhj",
		"hdywlljauher99my898vcpckyswyp50evrx938k8cu7ad5v8f0dy0mc",
		"62r2ut0y5x795elunrk69pax8l60cw4xmmyg4nztfd4947fpnw9mejv",
		"2meqvl40y3psqr9rdm8vyhvulutpr622ft97uh0rnk2wp7g6rlmraew",
		"w7528pk6ey8ss3h39gu49rj26vuev50aj43s5022r5qwwngu8hwzxkn",
		"jlfggmdqgzfnqn5a4dw86n3nzmlsw0nvs4sfjpmxh3h5j96z0te4n0s",
		"5qhmqhmx0g0ruy88gvd3kh6nwzmxhk8ptchzhfln57jq3fj9rjctaxt",
		"6tsglvwerm0f848z5hxvfhpcl5wxtpwnhr2vzmvzqctv496h9j6ftjl",
		"gnrf6drhayd2h4txf3rk99ny04y6dsrunzleetnhlt7ht3nu602lezx",
		"0utcsz7mw4ygglhkfexf8tef9w9e5xm6wwtvpyya8pldwpgkvs0xfsm",
		"nv29aa2yqs783mu6pxl3dsrxrzm08fuchxaz4hgd7e6p7e5kq8jghqc",
		"4wrs82vguxfn4a7kkt2xkd7xh7ehdz4kzyqx8z9x69ej2xwgj02ga8f",
		"6pmw7476nr80ctfsqk7p39kza3rpljwvul69xuqr77mmcl804xvv9sl",
		"gftgqnsaxyam32s7a9z5ehkpy5s8lswcp6a72pxzj86jzq24yemxmdp",
		"dxmvfvmk92wqpaluklqdfjphva8j76da255glf0d4x7amfldtgzkp4a",
	}
	
	// encoded := "\nencodedStr := []string{\n"
	
	// Convert hashes to our base32 encoding format
	for i := range hashedSeeds {
		
		// Note that we are slicing off a number of bytes at the end according
		// to the sequence number to get different check byte lengths from a
		// uniform original data. As such, this will be accounted for in the
		// check by truncating the same amount in the check (times two, for the
		// hex encoding of the string).
		encode, err := Codec.Encode(hashedSeeds[i][:len(hashedSeeds[i])-i%5])
		if err != nil {
			t.Fatal(err)
		}
		if encode != encodedStr[i] {
			t.Errorf(
				"Decode failed, expected item %d '%s' got '%s'",
				i, encodedStr[i], encode,
			)
		}
		// encoded += "\t\"" + encode + "\",\n"
	}
	
	// encoded += "}\n"
	// t.Log(encoded)
	
	// Next, decode the encodedStr above, which should be the output of the
	// original generated seeds, with the index mod 5 truncations performed on
	// each as was done to generate them.
	
	for i := range encodedStr {
		
		res, err := Codec.Decode(encodedStr[i])
		if err != nil {
			t.Fatalf("error: '%v'", err)
		}
		elen := len(expected[i])
		etrimlen := 2 * (i % 5)
		expectedHex := expected[i][:elen-etrimlen]
		resHex := fmt.Sprintf("%x", res)
		if resHex != expectedHex {
			t.Fatalf(
				"got: '%s' expected: '%s'",
				resHex,
				expectedHex,
			)
		}
	}
}
