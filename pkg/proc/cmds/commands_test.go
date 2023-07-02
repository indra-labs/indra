package cmds

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/util/path"
	"testing"

	"github.com/indra-labs/indra/pkg/proc/opts/config"
)

func TestCommand_Foreach(t *testing.T) {
	cm, _ := Init(GetExampleCommands(), nil)
	log.D.Ln("spewing only droptxindex")
	cm.ForEach(func(cmd *Command, _ int) bool {
		if cmd.Name == "droptxindex" {
			log.D.S(cmd)
		}
		return true
	}, 0, 0, cm)
	log.D.Ln("printing name of all commands found on search")
	cm.ForEach(func(cmd *Command, depth int) bool {
		log.D.F("%s%s #(%d)", path.GetIndent(depth), cmd.Path, depth)
		for i := range cmd.Configs {
			log.D.F("%s%s -%s %v #%v (%d)", path.GetIndent(depth),
				cmd.Configs[i].Path(), i, cmd.Configs[i].String(), cmd.Configs[i].Meta().Aliases(), depth)
		}
		return true
	}, 0, 0, cm)
}

func TestCommand_MarshalText(t *testing.T) {
	o, _ := Init(GetExampleCommands(), nil)
	conf, err := o.MarshalText()
	if log.E.Chk(err) {
		t.FailNow()
	}
	log.D.Ln("\n" + string(conf))
}

func TestCommand_UnmarshalText(t *testing.T) {
	o, _ := Init(GetExampleCommands(), nil)
	var conf []byte
	var err error
	conf, err = o.MarshalText()
	if log.E.Chk(err) {
		t.FailNow()
	}
	err = o.UnmarshalText(conf)
	if err != nil {
		t.FailNow()
	}
}

func TestCommand_GetEnvs(t *testing.T) {
	o, _ := Init(GetExampleCommands(), nil)
	envs := o.GetEnvs()
	var out []string
	err := envs.ForEach(func(env string, opt config.Option) error {
		out = append(out, env)
		return nil
	})
	for i := range out { // verifying ordering groups subcommands
		log.D.Ln(out[i])
	}
	if err != nil {
		t.FailNow()
	}
}

var _print = func(a ...any) {
	fmt.Println(a)
}

var _printt = func(a ...any) {
	fmt.Print(a)
}
var disabledPrinter = func(a ...any) {
}

//
//func TestCommand_Help(t *testing.T) {
//	if indra.CI != "false" {
//		_print = disabledPrinter
//		_printt = disabledPrinter
//	}
//	ex := GetExampleCommands()
//	ex.AddCommand(Help())
//	o, _ := Init(ex, nil)
//	o.Commands = append(o.Commands)
//	args1 := "/random/path/to/server_binary help"
//	_print(args1)
//	args1s := strings.Split(args1, " ")
//	run, args, err := o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help loglevel"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help help"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help node"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help rpcconnect"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help kopach rpcconnect"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help node rpcconnect"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help nodeoff"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help user"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	args1 = "/random/path/to/server_binary help file"
//	_print(args1)
//	args1s = strings.Split(args1, " ")
//	run, args, err = o.ParseCLIArgs(args1s)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//	err = run.Entrypoint(o, args)
//	if log.E.Chk(err) {
//		t.FailNow()
//	}
//
//}

//func TestCommand_LogToFile(t *testing.T) {
//	ex := GetExampleCommands()
//	ex.AddCommand(Help())
//	ex, _ = Init(ex, nil)
//	ex.GetOpt(path.From("pod123 loglevel")).FromString("debug")
//	var err error
//	// this will create a place we can write the logs
//	if err = ex.SaveConfig(); log.E.Chk(err) {
//		err = os.RemoveAll(ex.Configs["ConfigFile"].Expanded())
//		if log.E.Chk(err) {
//		}
//		t.FailNow()
//	}
//	lfp := ex.GetOpt(path.From("pod123 logfilepath"))
//	o := ex.GetOpt(path.From("pod123 logtofile"))
//	o.FromString("true")
//	o.FromString("false")
//	var b []byte
//	if b, err = os.ReadFile(lfp.Expanded()); log.E.Chk(err) {
//		t.FailNow()
//	}
//	str := string(b)
//	if !strings.Contains(str, lfp.String()) {
//		t.FailNow()
//	}
//	if err := os.RemoveAll(ex.Configs["DataDir"].Expanded()); log.E.Chk(err) {
//	}
//}
