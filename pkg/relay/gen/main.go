package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"text/template"
	
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func main() {
	typesList := []string{"balance", "confirm", "crypt", "delay", "dxresponse",
		"exit", "forward", "getbalance", "reverse", "response", "session"}
	sort.Strings(typesList)
	tpl := `package relay

//go:generate go run ./pkg/relay/gen/main.go

import (
	"fmt"
	
	"github.com/davecgh/go-spew/spew"
	
	"git-indra.lan/indra-labs/indra/pkg/onion/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/delay"
	"git-indra.lan/indra-labs/indra/pkg/onion/dxresponse"
	"git-indra.lan/indra-labs/indra/pkg/onion/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/onion/response"
	"git-indra.lan/indra-labs/indra/pkg/onion/reverse"
	"git-indra.lan/indra-labs/indra/pkg/onion/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func Peel(b slice.Bytes, c *slice.Cursor) (on types.Onion, e error) {
	switch b[*c:c.Inc(magicbytes.Len)].String() {
{{range .}}case {{.}}.MagicString:
		on = &{{.}}.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	{{end}}default:
		e = fmt.Errorf("message magic not found")
		log.T.C(func() string {
			return fmt.Sprintln(e) + spew.Sdump(b.ToBytes())
		})
		return
	}
	return
}
`
	t, e := template.New("peel").Parse(tpl)
	if check(e) {
		panic(e)
	}
	const filename = "pkg/relay/peel.go"
	f, err := os.Create(filename)
	if check(err) {
		panic(err)
	}
	if e = t.Execute(f, typesList); check(e) {
		panic(e)
	}
	if e = runCmd("go", "fmt", filename); e != nil {
		os.Exit(1)
	}
	
}

func errPrintln(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

func runCmd(cmd ...string) (err error) {
	c := exec.Command(cmd[0], cmd[1:]...)
	var output []byte
	output, err = c.CombinedOutput()
	if err == nil && string(output) != "" {
		errPrintln(string(output))
	}
	return
}

func runCmdWithOutput(cmd ...string) (out string, err error) {
	c := exec.Command(cmd[0], cmd[1:]...)
	var output []byte
	output, err = c.CombinedOutput()
	// if err == nil && string(output) != "" {
	// 	errPrintln(string(output))
	// }
	out = string(output)
	return
}
