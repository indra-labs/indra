package ci

import (
	indra "git.indra-labs.org/dev/ind"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

func TraceIfNot() {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
}
