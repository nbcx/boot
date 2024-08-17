package boot

import (
	"bytes"
	"fmt"
	"strings"
)

type HelpCmd struct {
	Command
}

func (p *HelpCmd) Run(args []string) error {
	cmd, _, e := Find(Base(p), args)
	if cmd == nil || e != nil {
		log.Printf("Unknown help topic %#q\n", args)
		CheckErr(Usage(Base(cmd)))
	} else {
		InitDefaultHelpFlag(cmd)    // make possible 'help' flag to be shown
		InitDefaultVersionFlag(cmd) // make possible 'version' flag to be shown
		CheckErr(Help(cmd))
	}

	return nil
}

func (p *HelpCmd) GetUse() string {
	return "help [command]"
}

func NewHelpCmd(cmd Commander) *HelpCmd {
	return &HelpCmd{
		Command: Command{
			Short: "Help about any command",
			Long: `Help provides help for any command in the application.
Simply type ` + displayName(cmd) + ` help [path to command] for full details.`,
			GroupID: cmd.getHelpCommandGroupID(),
			ValidArgsFunction: func(c Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
				var completions []string
				cmd, _, e := Find(Base(c), args)
				if e != nil {
					return nil, ShellCompDirectiveNoFileComp
				}
				if cmd == nil {
					// Root help command.
					cmd = Base(c)
				}
				for _, subCmd := range cmd.Commands() {
					if IsAvailableCommand(subCmd) || subCmd == cmd.GetHelpCommand() {
						if strings.HasPrefix(name(subCmd), toComplete) {
							completions = append(completions, fmt.Sprintf("%s\t%s", name(subCmd), subCmd.GetShort()))
						}
					}
				}
				return completions, ShellCompDirectiveNoFileComp
			},
		},
	}
}

// UsageTemplate returns usage template for the command.
func UsageTemplate(c Commander) string {
	// if c.usageTemplate != "" {
	// 	return c.usageTemplate
	// }
	// if c.HasParent() {
	// 	return UsageTemplate(c.Parent())
	// }
	return `Usage:{{if .Runnable}}
  {{. | UseLine}}{{end}}{{if . | HasAvailableSubCommands}}
  {{. | CommandPath}} [command]{{end}}{{if gt (len .GetAliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if . | HasExample}}

Examples:
{{.Example}}{{end}}{{if . | HasAvailableSubCommands}}{{$cmds := .GetCommands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or (. | IsAvailableCommand) (eq . | Name "help"))}}
  {{rpad (. | Name) (. | NamePadding) }} {{.GetShort}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or (. | IsAvailableCommand) (eq . | Name "help")))}}
  {{rpad (. | Name) (. | NamePadding) }} {{.GetShort}}{{end}}{{end}}{{end}}{{if not . | AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or (. | IsAvailableCommand) (eq (. | Name) "help")))}}
  {{rpad (. | Name) (. | NamePadding) }} {{.GetShort}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if . | HasAvailableLocalFlags}}

Flags:
{{. | LocalFlagUsages | trimTrailingWhitespaces}}{{end}}{{if . | HasAvailableInheritedFlags}}

Global Flags:
{{. | InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if . | HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad (. | CommandPath) (. | CommandPathPadding)}} {{.GetShort}}{{end}}{{end}}{{end}}{{if . | HasAvailableSubCommands}}

Use "{{. | CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

// HelpTemplate return help template for the command.
func HelpTemplate(c Commander) string {
	// 	if c.helpTemplate != "" {
	// 		return c.helpTemplate
	// 	}

	// if c.HasParent() {
	// 	return HelpTemplate(c.Parent())
	// }
	str := c.GetLong()
	if str == "" {
		str = c.GetShort()
	}
	str = trimRightSpace(str)
	if c.Runnable() || HasSubCommands(c) {
		str += UsageString(c)
	}
	return str
	// return `{{with (or .GetLong .GetShort)}}{{. | trimTrailingWhitespaces}}

	// {{end}}{{if or .Runnable .HasSubCommands}}{{ .| $UsageString}}{{end}}`
}

// UsageString returns usage string.
func UsageString(c Commander) string {
	// Storing normal writers
	// tmpOutput := log.outWriter
	// tmpErr := log.errWriter

	bb := new(bytes.Buffer)
	// log.outWriter = bb
	// log.errWriter = bb

	// usageFunc := func(c Commander) error {
	// 	mergePersistentFlags(c)
	// 	err := tmpl(log.OutOrStderr(), UsageTemplate(c), c)
	// 	if err != nil {
	// 		log.PrintErrLn(err)
	// 	}
	// 	return err
	// }
	mergePersistentFlags(c)
	err := tmpl(bb, UsageTemplate(c), c)
	if err != nil {
		log.PrintErrLn(err)
	}
	// return err
	CheckErr(err)

	// Setting things back to normal
	// log.outWriter = tmpOutput
	// log.errWriter = tmpErr

	return bb.String()
	// return fmt.Sprintf("UsageString: %v", c.GetUse())
}

// Help puts out the help for the command.
// Used when a user calls help [command].
// Can be defined by user by overriding HelpFunc.
func Help(c Commander) error {
	HelpFunc(c, []string{})
	return nil
}
