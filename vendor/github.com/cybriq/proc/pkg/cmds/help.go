package cmds

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/util"
)

// Help is a default top level command that subsequent
func Help() (h *Command) {
	h = &Command{
		Path: nil,
		Name: "help",
		Description: "Print help information, optionally multiple keywords" +
			" can be given and will be searched to generate output",
		Documentation: "Uses partial matching, if result is ambiguous, prints general top " +
			"level\nhelp and the list of partial match terms.\n\n" +
			"If single term exactly matches a command, the command help will" +
			" be printed.\n\n" +
			"Otherwise, a list showing all items, their paths, and " +
			"description are shown.\n\n" +
			"Use their full path and the full " +
			"documentation for the item will be shown.\n\n" +
			"Note that in all cases, options are only recognised after their\n" +
			"related subcommand.",
		Entrypoint: func(c *Command, args []string) (err error) {
			if args == nil {
				// no args given, just print top level general help
				return
			}
			foundCommands := &[]*Command{}
			fops := make(map[string]config.Option)
			foundOptions := &fops
			foundCommandWhole := false
			foundOptionWhole := false
			c.ForEach(func(cm *Command, depth int) bool {
				for i := range args {
					// check for match of current command name
					if strings.Contains(util.Norm(cm.Name), util.Norm(args[i])) {
						if util.Norm(cm.Name) == util.Norm(args[i]) {
							if len(args) == 1 {
								foundCommandWhole = true
								*foundCommands = append(*foundCommands, cm)
								break
							}
						}
						*foundCommands = append(*foundCommands, cm)
					}
					// check for matches on configs
					for ops := range cm.Configs {
						// log.I.Ln(ops, cm.Name, Norm(ops), Norm(args[i]))
						if strings.Contains(util.Norm(ops), util.Norm(args[i])) {
							// in the case of specifying a command and an option
							// and the option is from the command, and there is
							// only two args, and the option is fully named, not
							// just partial matched, clear found options and
							// break to return one command one option,
							// which later is recognised to indicate show detail
							if len(args) == 2 && len(*foundCommands) == 1 &&
								util.Norm(ops) == util.Norm(args[i]) {
								if cm.Configs[ops].Path().Equal(cm.Path) {
									*foundOptions = make(map[string]config.Option)
									(*foundOptions)[ops] = cm.Configs[ops]
									foundOptionWhole = true
									return false
								}
							} else {
								(*foundOptions)[ops] = cm.Configs[ops]
							}
						}
					}
				}
				return true
			}, 0, 0, c)
			var out string
			switch {
			case foundCommandWhole && len(args) == 1:
				cm := (*foundCommands)[0]
				// Print command help information
				var outs []string
				for i := range cm.Commands {
					outs = append(outs, cm.Commands[i].Name)
				}
				sort.Strings(outs)
				// out += fmt.Sprintf("\n%s - %s\n\n", cm.Path, cm.Description)
				out += fmt.Sprintf("\n%s - %s\n\n", c.Name, c.Description)
				out += fmt.Sprintf(
					"Help information for command '%s':\n\n",
					args[0])
				out += fmt.Sprintf("Documentation:\n\n%s\n\n",
					IndentTextBlock(cm.Documentation, 1))
				if len(cm.Commands) > 0 {
					out += fmt.Sprintf("Available subcommands:\n\n")
					for i := range outs {
						def := ""
						if len(cm.Default) > 0 {
							if outs[i] == cm.Default[len(cm.Default)-1] {
								def = " *"
							}
						}
						out += fmt.Sprintf("\t%s%s\n", outs[i], def)
					}
					if len(cm.Default) > 0 {
						out += "\n* indicates default if no subcommand given\n\n"
					} else {
						out += "\n"
					}
					out += fmt.Sprintf("for more information on subcommands:\n\n")
					out += fmt.Sprintf("\t%s help <subcommand>\n\n", os.Args[0])
				}
				if len(cm.Configs) > 0 {
					out += "Available configuration options on this command:\n\n"
					out += fmt.Sprintf("\t-%s %v - %s (default: '%s')\n",
						"flag", "[alias1 alias2]", "description", "default")
					out += "\t\t(prefix '-' can also be '--', value can follow after space or with '=' and no space)\n\n"
					var opts []string
					for i := range c.Configs {
						opts = append(opts, i)
					}
					sort.Strings(opts)
					for i := range opts {
						aliases := c.Configs[opts[i]].Meta().Aliases()
						for j := range aliases {
							aliases[j] = strings.ToLower(aliases[j])
						}
						var al string
						if len(aliases) > 0 {
							al = fmt.Sprint(aliases, " ")
						}
						out += fmt.Sprintf("\t-%s %v\n\t\t%s (default: '%s')\n", strings.ToLower(opts[i]),
							al,
							c.Configs[opts[i]].Meta().Description(),
							c.Configs[opts[i]].Meta().Default())
					}
					out += fmt.Sprintf(
						"\nUse 'help %s <option>' to get details on option.\n",
						cm.Name)
				}
			case len(*foundOptions) == 1 &&
				(len(*foundCommands) == 0 ||
					foundOptionWhole):

				// For this case there is only one option, and one match, so we
				// will print the full details as it is unambiguous.
				for i := range *foundOptions {
					// there is only one but a range statement grabs it
					op := (*foundOptions)[i]
					om := op.Meta()
					path := op.Path().TrimPrefix().String()
					if len(path) > 0 {
						path = path + " "
					}
					var out string
					out += fmt.Sprintf("\n%s - %s\n\n", c.Name, c.Description)
					search := strings.Join(args, " ")
					if len(args) > 1 {
						out += fmt.Sprintf(
							"Help information for search terms '%s':\n\n", search)
					} else {
						out += fmt.Sprintf("Help information for option '%s'\n\n",
							i)
					}
					if len(path) > 1 {
						out += fmt.Sprintf("Command Path:\n\n\t%s\n\n", path)
					}
					out += fmt.Sprintf("%s [-%s]\n\n", i, strings.ToLower(i))
					out += fmt.Sprintf("\t%s\n\n", om.Description())
					out += fmt.Sprintf("Default:\n\n\t%s %s--%s=%s\n\n",
						os.Args[0], path, strings.ToLower(i), om.Default())
					out += fmt.Sprintf("Documentation:\n\n%s\n\n",
						IndentTextBlock(om.Documentation(), 1))
					fmt.Print(out)
				}
			case len(*foundCommands) > 0 || len(*foundOptions) > 0:
				// if the text was not a command and one of its options, just
				// show all partial matches of both commands and options in
				// summary with their relevant paths
				var out string
				out += fmt.Sprintf("\n%s - %s\n\n", c.Name, c.Description)
				plural := ""
				search := strings.Join(args, " ")
				if len(args) > 1 {
					plural = "s"
				}
				out += fmt.Sprintf(
					"Help information for search term%s '%s':\n\n",
					plural, search)
				if len(*foundCommands)+len(*foundOptions) > 1 {
					out += "Multiple matches found:\n\n"
				}
				if len(*foundCommands) > 0 {
					plural := ""
					if len(*foundCommands) > 1 {
						plural = "s"
					}
					out += fmt.Sprintf("Command%s\n\n", plural)
					for i := range *foundCommands {
						cm := (*foundCommands)[i]
						out += fmt.Sprintf("\t%v %s\n\t\t%s - %s\n\n", os.Args[0],
							strings.ToLower(cm.Name), cm.Name, cm.Description)
					}
				}
				if len(*foundOptions) > 0 {
					out += fmt.Sprintf("Options:\n\n")
					for i := range *foundOptions {
						op := (*foundOptions)[i]
						om := op.Meta()
						path := op.Path().TrimPrefix().String()
						if len(path) > 0 {
							path = path + " "
						}
						out += fmt.Sprintf("\t%s %v-%s=%s\n\t\t%s - %s\n\n",
							os.Args[0], path, strings.ToLower(i), om.Default(),
							i, om.Description())
					}
				}
				fmt.Print(out)
			default:
				// Print top level help information
				var outs []string
				for i := range c.Commands {
					outs = append(outs, c.Commands[i].Name)
				}
				sort.Strings(outs)
				out += fmt.Sprintf("\n%s - %s\n\n", c.Name, c.Description)
				if len(c.Commands) > 0 {
					plural := ""
					if len(c.Commands) > 1 {
						plural = "s"
					}
					out += fmt.Sprintf("Available subcommand%s:\n\n", plural)
					for i := range outs {
						def := ""
						if len(c.Default) > 0 {
							if outs[i] == c.Default[len(c.Default)-1] {
								def = " *"
							}
						}
						out += fmt.Sprintf("\t%s%s\n", outs[i], def)
					}
					if len(c.Default) > 0 {
						out += "\n* indicates default if no subcommand given\n\n"
					} else {
						out += "\n"
					}
				}
				out += fmt.Sprintf("For more information:\n\n")
				out += fmt.Sprintf("\t%s help <subcommand>\n\n", os.Args[0])
				out += "\tUse 'help help' to learn more about the help function.\n\n"
				out += "Available configuration options at top level:\n\n"
				out += fmt.Sprintf("\t-%s %v - %s (default: '%s')\n",
					"flag", "[alias1 alias2]", "description", "default")
				out += "\t\t(prefix '-' can also be '--', value can follow after space or with '=' and no space)\n\n"
				var opts []string
				for i := range c.Configs {
					opts = append(opts, i)
				}
				sort.Strings(opts)
				for i := range opts {
					aliases := c.Configs[opts[i]].Meta().Aliases()
					for j := range aliases {
						aliases[j] = strings.ToLower(aliases[j])
					}
					var al string
					if len(aliases) > 0 {
						al = fmt.Sprint(aliases, " ")
					}
					out += fmt.Sprintf("\t-%s %v\n\t\t%s (default: '%s')\n", strings.ToLower(opts[i]),
						al,
						c.Configs[opts[i]].Meta().Description(),
						c.Configs[opts[i]].Meta().Default())
				}
				out += "\nUse 'help <option>' to get details on option.\n"

			}
			fmt.Print(out)
			return
		},
		Parent:   nil,
		Commands: nil,
		Configs:  nil,
		Default:  nil,
		Mutex:    sync.Mutex{},
	}
	return
}

// IndentTextBlock adds an arbitrary number of tabs to the front of a text
// block that is presumed to already be flowed to ~80 columns.
func IndentTextBlock(s string, tabs int) (o string) {
	s = strings.TrimSpace(s)
	split := strings.Split(s, strings.Repeat("\n", tabs))
	for i := range split {
		split[i] = "\t" + split[i]
	}
	return strings.Join(split, "\n")
}
