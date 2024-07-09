package boot

import (
	"fmt"
	"strconv"
	"strings"
)

type CompleteCmd struct {
	Default
}

func (cmd *CompleteCmd) GetUse() string {
	return fmt.Sprintf("%s [command-line]", ShellCompRequestCmd)
}

func (cmd *CompleteCmd) Run(args []string) error {
	finalCmd, completions, directive, err := getCompletions(cmd, args)
	if err != nil {
		CompErrorln(err.Error())
		// Keep going for multiple reasons:
		// 1- There could be some valid completions even though there was an error
		// 2- Even without completions, we need to print the directive
	}

	noDescriptions := cmd.CalledAs() == ShellCompNoDescRequestCmd
	if !noDescriptions {
		if doDescriptions, err := strconv.ParseBool(getEnvConfig(cmd, configEnvVarSuffixDescriptions)); err == nil {
			noDescriptions = !doDescriptions
		}
	}
	noActiveHelp := GetActiveHelpConfig(finalCmd) == activeHelpGlobalDisable
	out := log.OutOrStdout()
	for _, comp := range completions {
		if noActiveHelp && strings.HasPrefix(comp, activeHelpMarker) {
			// Remove all activeHelp entries if it's disabled.
			continue
		}
		if noDescriptions {
			// Remove any description that may be included following a tab character.
			comp = strings.SplitN(comp, "\t", 2)[0]
		}

		// Make sure we only write the first line to the output.
		// This is needed if a description contains a linebreak.
		// Otherwise the shell scripts will interpret the other lines as new flags
		// and could therefore provide a wrong completion.
		comp = strings.SplitN(comp, "\n", 2)[0]

		// Finally trim the completion.  This is especially important to get rid
		// of a trailing tab when there are no description following it.
		// For example, a sub-command without a description should not be completed
		// with a tab at the end (or else zsh will show a -- following it
		// although there is no description).
		comp = strings.TrimSpace(comp)

		// Print each possible completion to the output for the completion script to consume.
		fmt.Fprintln(out, comp)
	}

	// As the last printout, print the completion directive for the completion script to parse.
	// The directive integer must be that last character following a single colon (:).
	// The completion script expects :<directive>
	fmt.Fprintf(out, ":%d\n", directive)

	// Print some helpful info to stderr for the user to understand.
	// Output from stderr must be ignored by the completion script.
	fmt.Fprintf(log.ErrOrStderr(), "Completion ended with directive: %s\n", directive.string())
	return nil
}

type BashCompleteCmd struct {
	Default
}

func (cmd *BashCompleteCmd) Run(args []string) error {
	return GenBashCompletionV2(Base(cmd), log.OutOrStdout(), !cmd.CompletionOptions.DisableDescriptions)
}
func (cmd *BashCompleteCmd) GetUse() string {
	return "bash"
}

func NewBashCompleteCmd(cmd Commander, shortDesc string) *BashCompleteCmd {
	return &BashCompleteCmd{
		Default{
			Short: fmt.Sprintf(shortDesc, "bash"),
			Long: fmt.Sprintf(`Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(%[1]s completion bash)

To load completions for every new session, execute once:

#### Linux:

	%[1]s completion bash > /etc/bash_completion.d/%[1]s

#### macOS:

	%[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

You will need to start a new shell for this setup to take effect.
`, name(cmd)),
			Args:                  NoArgs,
			DisableFlagsInUseLine: true,
			ValidArgsFunction:     NoFileCompletions,
		},
	}
}

type ZshCompleteCmd struct {
	Default
	noDesc bool
}

func (p *ZshCompleteCmd) GetUse() string {
	return "zsh"
}

func NewZshCompleteCmd(cmd Commander, shortDesc string, noDesc bool) *ZshCompleteCmd {
	return &ZshCompleteCmd{
		Default: Default{
			Short: fmt.Sprintf(shortDesc, "zsh"),
			Long: fmt.Sprintf(`Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(%[1]s completion zsh)

To load completions for every new session, execute once:

#### Linux:

	%[1]s completion zsh > "${fpath[1]}/_%[1]s"

#### macOS:

	%[1]s completion zsh > $(brew --prefix)/share/zsh/site-functions/_%[1]s

You will need to start a new shell for this setup to take effect.
`, name(Base(cmd))),
			Args:              NoArgs,
			ValidArgsFunction: NoFileCompletions,
		},
		noDesc: noDesc,
	}
}

func (p *ZshCompleteCmd) Run(args []string) error {
	out := log.OutOrStdout()
	if p.noDesc {
		return GenZshCompletionNoDesc(Base(p), out)
	}
	return GenZshCompletion(Base(p), out)
}

type FishCompleteCmd struct {
	Default
	noDesc bool
}

func (p *FishCompleteCmd) GetUse() string {
	return "fish"
}

func (p *FishCompleteCmd) Run(args []string) error {
	out := log.OutOrStdout()
	return GenFishCompletion(Base(p), out, !p.noDesc)
}

func NewFishCompleteCmd(cmd Commander, shortDesc string, noDesc bool) *FishCompleteCmd {
	return &FishCompleteCmd{
		Default: Default{
			Short: fmt.Sprintf(shortDesc, "fish"),
			Long: fmt.Sprintf(`Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	%[1]s completion fish | source

To load completions for every new session, execute once:

	%[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

You will need to start a new shell for this setup to take effect.
`, name(Base(cmd))),
			Args:              NoArgs,
			ValidArgsFunction: NoFileCompletions,
		},
		noDesc: noDesc,
	}
}

type PowershellCompleteCmd struct {
	Default
	noDesc bool
}

func (cmd *PowershellCompleteCmd) Run(args []string) error {
	out := log.OutOrStdout()
	if cmd.noDesc {
		return GenPowerShellCompletion(Base(cmd), out)
	}
	return GenPowerShellCompletionWithDesc(Base(cmd), out)
}
func (cmd *PowershellCompleteCmd) GetUse() string {
	return "powershell"
}
func NewPowershellCompleteCmd(cmd Commander, shortDesc string, noDesc bool) *PowershellCompleteCmd {
	return &PowershellCompleteCmd{
		Default: Default{
			Short: fmt.Sprintf(shortDesc, "powershell"),
			Long: fmt.Sprintf(`Generate the autocompletion script for powershell.

To load completions in your current shell session:

	%[1]s completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.
`, name(Base(cmd))),
			Args:              NoArgs,
			ValidArgsFunction: NoFileCompletions,
		},
		noDesc: noDesc,
	}
}

type HelpCmd struct {
	Default
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
		Default: Default{
			Short: "Help about any command",
			Long: `Help provides help for any command in the application.
Simply type ` + displayName(cmd) + ` help [path to command] for full details.`,
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
			GroupID: cmd.getHelpCommandGroupID(),
		},
	}
}
