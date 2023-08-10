package log_test

import (
	"errors"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"testing"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

func TestGetLogger(t *testing.T) {
	log.I.Ln("info")
	fails(errors.New("dummy error"))
}
