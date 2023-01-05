package app

import (
	"github.com/Indra-Labs/indra"
	cmds2 "github.com/Indra-Labs/indra/pkg/cmds"
	log2 "github.com/Indra-Labs/indra/pkg/log"
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

func New(cmd *cmds2.Command, args []string) (a *App, e error) {
	// Add the default configuration items for datadir/configfile
	cmds2.GetConfigBase(cmd.Configs, cmd.Name, false)
	// Add the help function
	cmd.AddCommand(cmds2.Help())
	a = &App{Command: cmd}
	// We first parse the CLI args, in case config file location has been
	// specified
	if a.launch, _, e = a.Command.ParseCLIArgs(args); check(e) {
		return
	}
	if e = cmd.LoadConfig(); log.E.Chk(e) {
		return
	}
	if a.Command, e = cmds2.Init(cmd, nil); check(e) {
		return
	}
	a.Envs = cmd.GetEnvs()
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
