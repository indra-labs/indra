module github.com/Indra-Labs/indra

go 1.19

require (
	github.com/cybriq/proc v0.2.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0
	github.com/minio/sha256-simd v1.0.0
	github.com/templexxx/reedsolomon v1.1.3
	github.com/zeebo/blake3 v0.2.3
	gopkg.in/src-d/go-git.v4 v4.13.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/gookit/color v1.5.1 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/templexxx/cpu v0.0.1 // indirect
	github.com/templexxx/xorsimd v0.1.1 // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/crypto v0.0.0-20221005025214-4161e89ecf1b // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

replace crypto/sha256 => github.com/minio/sha256-simd v1.0.0
