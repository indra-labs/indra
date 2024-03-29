package cmds

import (
	"git.indra-labs.org/dev/ind/pkg/util/ci"
	"strings"
	"testing"
)

func TestCommand_ParseCLIArgs(t *testing.T) {
	ci.TraceIfNot()
	ec := GetExampleCommands()
	o, _ := Init(ec, nil)
	args6 := "/random/path/to/server_binary --cafile ~/some/cafile --LC=cn " +
		"node -addrindex false --blocksonly"
	args6s := strings.Split(args6, " ")
	run, _, e := o.ParseCLIArgs(args6s)
	if log.E.Chk(e) {
		t.Error(e)
		t.FailNow()
	}
	args1 := "/random/path/to/server_binary --cafile ~/some/cafile --LC=cn " +
		"node -addrindex --BD=5m"
	args1s := strings.Split(args1, " ")
	run, _, e = o.ParseCLIArgs(args1s)
	if log.E.Chk(e) {
		t.Error(e)
		t.FailNow()
	}
	_, _ = run, e
	args3 := "node -addrindex --BD 48h30s dropaddrindex somegarbage " +
		"--autoports"
	args3s := strings.Split(args3, " ")
	run, _, e = o.ParseCLIArgs(args3s)
	// This one must fail, 'somegarbage' is not a command and has no -/-- prefix
	if e == nil {
		t.Error(e)
		t.FailNow()
	}
	args5 := "/random/path/to/server_binary --cafile ~/some/cafile --LC=cn "
	args5s := strings.Split(args5, " ")
	run, _, e = o.ParseCLIArgs(args5s)
	if log.E.Chk(e) {
		t.Error(e)
		t.FailNow()
	}
	args2 := "/random/path/to/server_binary node -addrindex --BD=48h30s  " +
		"-RPCMaxConcurrentReqs -16 dropaddrindex"
	args2s := strings.Split(args2, " ")
	run, _, e = o.ParseCLIArgs(args2s)
	if log.E.Chk(e) {
		t.Error(e)
		t.FailNow()
	}
}
