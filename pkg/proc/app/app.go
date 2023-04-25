package app

import (
	"git-indra.lan/indra-labs/indra"
	cmds2 "git-indra.lan/indra-labs/indra/pkg/proc/cmds"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type App struct {
	*cmds2.Command
	launch  *cmds2.Command
	runArgs []string
	cmds2.Envs
}

func New(c *cmds2.Command, args []string) (a *App, e error) {
	log2.App = c.Name
	// AddIntro the default configuration items for datadir/configfile
	log.T.Ln("test")
	cmds2.GetConfigBase(c.Configs, c.Name, false)
	log.T.Ln("test")
	// AddIntro the help function
	c.AddCommand(cmds2.Help())
	a = &App{Command: c}
	log.T.Ln("test")
	if a.Command, e = cmds2.Init(c, nil); check(e) {
		return
	}
	log.T.Ln("test")
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
