package app

import (
	"git.indra-labs.org/dev/ind"
	"os"
	"strings"
	"testing"

	"git.indra-labs.org/dev/ind/pkg/proc/cmds"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

func TestNew(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}

	args1 := "/random/path/to/server_binary --cafile ~/some/cafile --LC=cn node -addrindex --BD 48h30s"
	args1s := strings.Split(args1, " ")
	var a *App
	var e error
	if a, e = New(cmds.GetExampleCommands(), args1s); log.E.Chk(e) {
		t.FailNow()
	}
	if e = a.Launch(); check(e) {
		t.FailNow()
	}
	if e = os.RemoveAll(a.Command.Configs["DataDir"].
		Expanded()); check(e) {

		t.FailNow()
	}

}
