package app

import (
	"os"
	"strings"
	"testing"

	"github.com/Indra-Labs/indra/pkg/cmds"
	log2 "github.com/Indra-Labs/indra/pkg/log"
)

func TestNew(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	log2.CodeLoc = true
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
