//go:build ignore

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

type handlemessage struct {
	Name   string
	Extern bool
}

type handlemessages []handlemessage

func (p handlemessages) Len() int           { return len(p) }
func (p handlemessages) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p handlemessages) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func main() {
	typesList := []string{"balance", "confirm", "crypt", "delay", "dxresponse",
		"exit", "forward", "getbalance", "reverse", "response", "session"}
	sort.Strings(typesList)
	tpl := `package relay

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
	_ = f.Close()
	if e = runCmd("go", "fmt", filename); check(e) {
		os.Exit(1)
	}
	typesList2 := handlemessages{
		// {"parp", false},
		{"balance", false},
		{"confirm", true},
		{"crypt", true},
		{"delay", true},
		{"exit", true},
		{"forward", true},
		{"getbalance", true},
		{"reverse", false},
		{"response", true},
		{"session", true},
	}
	sort.Sort(typesList2)
	
	tpl = `package relay

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/onion/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/delay"
	"git-indra.lan/indra-labs/indra/pkg/onion/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/response"
	"git-indra.lan/indra-labs/indra/pkg/onion/reverse"
	"git-indra.lan/indra-labs/indra/pkg/onion/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) handleMessage(b slice.Bytes, prev types.Onion) {
	log.T.F("%v handling received message", eng.GetLocalNodeAddress())
	var on1 types.Onion
	var e error
	c := slice.NewCursor()
	if on1, e = Peel(b, c); check(e) {
		return
	}
	switch on := on1.(type) {
	{{range .}}case *{{.Name}}.Layer:
		{{if .Extern}}if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		{{end}}log.T.C(recLog(on, b, eng))
		eng.{{.Name}}(on, b, c, prev)
	{{end}}default:
		log.I.S("unrecognised packet", b)
	}
}
`
	t, e = template.New("peel").Parse(tpl)
	if check(e) {
		panic(e)
	}
	const handlemessage = "pkg/relay/handlemessage.go"
	f, e = os.Create(handlemessage)
	if check(e) {
		panic(e)
	}
	if e = t.Execute(f, typesList2); check(e) {
		panic(e)
	}
	if e = runCmd("go", "fmt", handlemessage); check(e) {
		panic(e)
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
