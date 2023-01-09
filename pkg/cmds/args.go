package cmds

import (
	"fmt"
	"strings"

	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/opts/meta"
	"github.com/indra-labs/indra/pkg/util"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// ParseCLIArgs reads a command line argument slice (presumably from os.Args),
// identifies the command to run and
//
// Rules for constructing CLI args:
//
//   - Commands are identified by name, and must appear in their hierarchic
//     order to invoke subcommands. They are matched as normalised to lower
//     case.
//
//   - Options can be preceded by "--" or "-", and the full name, or the
//     alias, normalised to lower case for matching, and if there is an "=" after
//     it, the value is after this, otherwise, the next element in the args is the
//     value, except booleans, which default to true unless set to false. For these,
//     the prefix "no" or similar indicates the semantics of the true value.
//
//   - Options only match when preceded by their relevant Command, except for
//     the root Command, and these options must precede any other command
//     options.
//
//   - If no command is selected, the root Command.Default is selected. This
//     can optionally be used for subcommands as well, though it is unlikely
//     needed, if found, the Default of the tip of the Command branch selected
//     by the CLI if there is one, otherwise the Command itself.
func (c *Command) ParseCLIArgs(a []string) (run *Command, runArgs []string,
	e error) {

	var args []string
	var cursor int
	for i := range a {
		if len(a[i]) > 0 {
			args = append(args, a[i])
			cursor++
		}
	}
	var segments [][]string
	commands := Commands{c}
	var depth, last int
	cur := c
	cursor = 0
	// First pass matches Command names in order to slice up the sections
	// where relevant config items will be found.
	for ; cursor < len(args); cursor++ {
		for i := range cur.Commands {
			if util.NormEq(args[cursor], cur.Commands[i].Name) {
				// the command to run is the last, so update
				// each new command found
				run = cur.Commands[i]
				commands = append(commands, run)
				depth++
				cur = cur.Commands[i]
				segments = append(segments, args[last:cursor])
				last = cursor
				break
			}
		}
	}
	// append the remainder to the last segment
	var tmp []string
	for _, item := range args[last:cursor] {
		if len(item) > 0 {
			tmp = append(tmp, item)
		}
	}
	segments = append(segments, tmp)
	// The segments that have been cut from args will now provide the root
	// level command name, and all subsequent items until the next segment
	// should be names found in the configs map.
	for i := range segments {
		if len(segments[i]) < 1 {
			continue
		}
		iArgs := segments[i][1:]
		cmd := commands[i]
		// the final command can accept arbitrary arguments,
		// that are passed into the entrypoint
		runArgs = iArgs
		if util.Norm(commands[i].Name) == "help" {
			break
		}
		for cursor = 0; cursor < len(iArgs); {
			inc := 1
			arg := iArgs[cursor]
			if len(arg) == 0 {
				cursor++
				continue
			}
			if !strings.HasPrefix(arg, "-") {
				e = fmt.Errorf("argument %s missing '-', "+
					"context %s, most likely misspelled "+
					"subcommand", arg, iArgs)
				return
			}
			arg = arg[1:]
			if strings.HasPrefix(arg, "-") {
				arg = arg[1:]
			}
			if strings.Contains(arg, "=") {
				if e = valueInArg(cmd, arg); check(e) {
					return
				}
				cursor += inc
				continue
			}
			found := false
			for cfgName := range cmd.Configs {
				found, cursor, inc, e =
					lookAhead(cmd, cfgName, arg, iArgs,
						found, cursor, inc)
			}
			if !found {
				e = fmt.Errorf(
					"option not found: '%s' context %v",
					arg, segments[i])
				return
			}
			cursor += inc
		}
	}
	// if no Command was found, return the default. If there is no default,
	// the top level Command will be returned
	if len(c.Default) > 0 && len(segments) < 2 {
		run = c
		def := c.Default
		var lastFound int
		for i := range def {
			for _, sc := range run.Commands {
				if sc.Name == def[i] {
					lastFound = i
					run = sc
				}
			}
		}
		if lastFound != len(def)-1 {
			e = fmt.Errorf("default command %v not found at %s",
				c.Default, def)
		}
	}
	return
}

func lookAhead(cmd *Command, cfgName, arg string, iArgs []string,
	found bool, cursor, inc int) (fnd bool, curs, ino int, e error) {

	aliases := cmd.Configs[cfgName].Meta().Aliases()
	names := append(
		[]string{cfgName}, aliases...)
	for _, name := range names {
		if util.Norm(name) != util.Norm(arg) {
			continue
		}
		// check for booleans, which can only be
		// followed by true or false
		if cmd.Configs[cfgName].Type() != meta.Bool {
			if len(iArgs)-1 <= cursor {
				continue
			}
			e = cmd.Configs[cfgName].FromString(iArgs[cursor+1])
			if e != nil {
				inc++
				found = true
				break
			}
		}
		// If we find a truth value in the next arg, assign it.
		if len(iArgs)-1 > cursor {
			e = cmd.Configs[cfgName].FromString(iArgs[cursor+1])
			if e == nil {
				inc++
				found = true
				break
			}
		}
		cur := cmd.Configs[cfgName].Meta().Default()
		cmd.Configs[cfgName].FromString(cur)
		v := !cmd.Configs[cfgName].Value().Bool()
		cmd.Configs[cfgName].FromString(fmt.Sprint(v))
		found = true
		break
	}
	return found, cursor, inc, e
}

func valueInArg(cmd *Command, arg string) (e error) {
	split := strings.Split(arg, "=")
	if len(split) > 2 {
		split = append(split[:1],
			strings.Join(split[1:], "="))
	}
	for cfgName := range cmd.Configs {
		aliases := cmd.Configs[cfgName].Meta().
			Aliases()
		names := append(
			[]string{cfgName}, aliases...)
		for _, name := range names {
			if util.Norm(name) !=
				util.Norm(split[0]) {
				continue
			}

			e = cmd.Configs[cfgName].
				FromString(split[1])
			if log.E.Chk(e) {
				return
			}

		}
	}
	return
}
