package app

import (
	"os"
	"strings"
	"testing"

	"github.com/cybriq/proc/pkg/cmds"
)

func TestNew(t *testing.T) {
	args1 := "/random/path/to/server_binary --cafile ~/some/cafile --LC=cn node -addrindex --BD 48h30s"
	args1s := strings.Split(args1, " ")
	var a *App
	var err error
	if a, err = New(cmds.GetExampleCommands(), args1s); log.E.Chk(err) {
		t.FailNow()
	}
	if err = a.Launch(); log.E.Chk(err) {
		t.FailNow()
	}
	if err = os.RemoveAll(a.Command.Configs["DataDir"].
		Expanded()); log.E.Chk(err) {

		t.FailNow()
	}

}
