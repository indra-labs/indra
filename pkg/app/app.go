package app

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/cmds"
	log2 "github.com/indra-labs/indra/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type App struct {
	*cmds.Command
	launch  *cmds.Command
	runArgs []string
	cmds.Envs
}

func New(c *cmds.Command, args []string) (a *App, e error) {
	log2.App = c.Name
	// Add the default configuration items for datadir/configfile
	cmds.GetConfigBase(c.Configs, c.Name, false)
	// Add the help function
	c.AddCommand(cmds.Help())
	a = &App{Command: c}
	if a.Command, e = cmds.Init(c, nil); check(e) {
		return
	}
	// We first parse the CLI args, in case config file location has been
	// specified
	if a.launch, _, e = a.Command.ParseCLIArgs(args); check(e) {
		return
	}
	if e = c.LoadConfig(); log.E.Chk(e) {
		return
	}
	a.Envs = c.GetEnvs()
	if e = a.Envs.LoadFromEnvironment(); check(e) {
		return
	}
	// This is done again, to ensure the effect of CLI args take precedence
	if a.launch, a.runArgs, e = a.Command.ParseCLIArgs(args); check(e) {
		return
	}
	return
}

func (a *App) Launch() (e error) {
	e = a.launch.Entrypoint(a.launch, a.runArgs)
	log.E.Chk(e)
	return
}
