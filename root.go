// Copyright 2013-2023 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// package boot is a commander providing a simple interface to create powerful modern CLI interfaces.
// In addition to providing an interface, Cobra simultaneously provides a controller to organize your application code.
package boot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	FlagSetByCobraAnnotation     = "cobra_annotation_flag_set_by_cobra"
	CommandDisplayNameAnnotation = "cobra_annotation_command_display_name"
)

// FParseErrWhitelist configures Flag parse errors to be ignored
type FParseErrWhitelist flag.ParseErrorsWhitelist

// Group Structure to manage groups for commands
type Group struct {
	ID    string
	Title string
}

type CommandCalledAs struct {
	name   string
	called bool
}

// Root is just that, a command for your application.
// E.g.  'go run ...' - 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Root struct {
	// Use is the one-line usage message.
	// Recommended syntax is as follows:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	Use string

	// Aliases is an array of aliases that can be used instead of the first word in Use.
	Aliases []string

	// SuggestFor is an array of command names for which this command will be suggested -
	// similar to aliases but only suggests.
	SuggestFor []string

	// Short is the short description shown in the 'help' output.
	Short string

	// The group id under which this subcommand is grouped in the 'help' output of its parent.
	GroupID string

	// Long is the long message shown in the 'help <this-command>' output.
	Long string

	// Example is examples of how to use the command.
	Example string

	// ValidArgs is list of all valid non-flag arguments that are accepted in shell completions
	ValidArgs []string
	// ValidArgsFunction is an optional function that provides valid non-flag arguments for shell completion.
	// It is a dynamic version of using ValidArgs.
	// Only one of ValidArgs and ValidArgsFunction can be used for a command.
	ValidArgsFunction func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective)

	// Expected arguments
	Args PositionalArgs

	// ArgAliases is List of aliases for ValidArgs.
	// These are not suggested to the user in the shell completion,
	// but accepted if entered manually.
	ArgAliases []string

	// BashCompletionFunction is custom bash functions used by the legacy bash autocompletion generator.
	// For portability with other shells, it is recommended to instead use ValidArgsFunction
	BashCompletionFunction string

	// Deprecated defines, if this command is deprecated and should print this string when used.
	Deprecated string

	// Annotations are key/value pairs that can be used by applications to identify or
	// group commands or set special options.
	Annotations map[string]string

	// Version defines the version for this command. If this value is non-empty and the command does not
	// define a "version" flag, a "version" boolean flag will be added to the command and, if specified,
	// will print content of the "Version" variable. A shorthand "v" flag will also be added if the
	// command does not define one.
	Version string

	// The *Run functions are executed in the following order:
	//   * PersistentPreRun()
	//   * PreRun()
	//   * Run()
	//   * PostRun()
	//   * PersistentPostRun()
	// All functions get the same args, the arguments after the command name.
	// The *PreRun and *PostRun functions will only be executed if the Run function of the current
	// command has been declared.
	//
	// PersistentPreRun: children of this command will inherit and execute.
	PersistentPreRun func(cmd Commander, args []string)
	// PersistentPreRunE: PersistentPreRun but returns an error.
	PersistentPreRunE func(cmd Commander, args []string) error
	// PreRun: children of this command will not inherit.
	// PreRun func(cmd Commander, args []string)
	// PreRunE: PreRun but returns an error.
	PreRunE func(cmd Commander, args []string) error
	// Run: Typically the actual work function. Most commands will only implement this.
	// Run func(cmd Commander, args []string)
	// RunE: Run but returns an error.
	// RunE func(cmd Commander, args []string) error
	RunE func(cmd Commander, args []string) error

	// PostRun: run after the Run command.
	// PostRun func(cmd Commander, args []string)
	// PostRunE: PostRun but returns an error.
	PostRunE func(cmd Commander, args []string) error
	// PersistentPostRun: children of this command will inherit and execute after PostRun.
	PersistentPostRun func(cmd Commander, args []string)
	// PersistentPostRunE: PersistentPostRun but returns an error.
	PersistentPostRunE func(cmd Commander, args []string) error

	// groups for subcommands
	commandgroups []*Group

	// args is actual args parsed from flags.
	args []string
	// flagErrorBuf contains all error messages from pflag.
	flagErrorBuf *bytes.Buffer
	// flags is full set of flags.
	flags *flag.FlagSet
	// pflags contains persistent flags.
	pflags *flag.FlagSet
	// lflags contains local flags.
	// This field does not represent internal state, it's used as a cache to optimise LocalFlags function call
	lflags *flag.FlagSet
	// iflags contains inherited flags.
	// This field does not represent internal state, it's used as a cache to optimise InheritedFlags function call
	iflags *flag.FlagSet
	// parentsPflags is all persistent flags of cmd's parents.
	parentsPflags *flag.FlagSet
	// globNormFunc is the global normalization function
	// that we can use on every pflag set and children commands
	globNormFunc func(f *flag.FlagSet, name string) flag.NormalizedName

	// usageFunc is usage func defined by user.
	usageFunc func(Commander) error
	// usageTemplate is usage template defined by user.
	usageTemplate string
	// flagErrorFunc is func defined by user and it's called when the parsing of
	// flags returns an error.
	flagErrorFunc func(Commander, error) error
	// helpTemplate is help template defined by user.
	helpTemplate string
	// helpFunc is help func defined by user.
	helpFunc func(Commander, []string)
	// helpCommand is command with usage 'help'. If it's not defined by user,
	// cobra uses default help command.
	helpCommand Commander
	// helpCommandGroupID is the group id for the helpCommand
	helpCommandGroupID string

	// completionCommandGroupID is the group id for the completion command
	completionCommandGroupID string

	// versionTemplate is the version template defined by user.
	versionTemplate string

	// errPrefix is the error message prefix defined by user.
	errPrefix string

	// FParseErrWhitelist flag parse errors to be ignored
	FParseErrWhitelist FParseErrWhitelist

	// CompletionOptions is a set of options to control the handling of shell completion
	CompletionOptions CompletionOptions

	// commandsAreSorted defines, if command slice are sorted or not.
	commandsAreSorted bool
	// commandCalledAs is the name or alias value used to call this command.
	commandCalledAs CommandCalledAs

	ctx context.Context

	// commands is the list of commands supported by this program.
	commands []Commander
	// parent is a parent command for this command.
	parent Commander
	// Max lengths of commands' string lengths for use in padding.
	commandsMaxUseLen         int
	commandsMaxCommandPathLen int
	commandsMaxNameLen        int

	// TraverseChildren parses flags on all parents before executing child command.
	TraverseChildren bool

	// Hidden defines, if this command is hidden and should NOT show up in the list of available commands.
	Hidden bool

	// SilenceErrors is an option to quiet errors down stream.
	SilenceErrors bool

	// SilenceUsage is an option to silence usage when an error occurs.
	SilenceUsage bool

	// DisableFlagParsing disables the flag parsing.
	// If this is true all flags will be passed to the command as arguments.
	DisableFlagParsing bool

	// DisableAutoGenTag defines, if gen tag ("Auto generated by spf13/cobra...")
	// will be printed by generating docs for this command.
	DisableAutoGenTag bool

	// DisableFlagsInUseLine will disable the addition of [flags] to the usage
	// line of a command when printing help or generating docs
	DisableFlagsInUseLine bool

	// DisableSuggestions disables the suggestions based on Levenshtein distance
	// that go along with 'unknown command' messages.
	DisableSuggestions bool

	// SuggestionsMinimumDistance defines minimum levenshtein distance to display suggestions.
	// Must be > 0.
	SuggestionsMinimumDistance int
}

// GetFParseErrWhitelist implements Commander.
func (c *Root) GetFParseErrWhitelist() FParseErrWhitelist {
	return c.FParseErrWhitelist
}

// SetFParseErrWhitelist implements Commander.
func (c *Root) SetFParseErrWhitelist(fp FParseErrWhitelist) {
	c.FParseErrWhitelist = fp
}

// GetGlobNormFunc implements Commander.
func (c *Root) GetGlobNormFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
}

// SetGlobNormFunc implements Commander.
func (c *Root) SetGlobNormFunc(f func(f *flag.FlagSet, name string) flag.NormalizedName) {
	c.globNormFunc = f
}

// GetIFlags implements Commander.
func (c *Root) GetIFlags() *flag.FlagSet {
	return c.iflags
}

// GetLFlags implements Commander.
func (c *Root) GetLFlags() *flag.FlagSet {
	return c.lflags
}

// GetPFlags implements Commander.
func (c *Root) GetPFlags() *flag.FlagSet {
	return c.pflags
}

// GetParentsPFlags implements Commander.
func (c *Root) GetParentsPFlags() *flag.FlagSet {
	return c.parentsPflags

}

// SetIFlags implements Commander.
func (c *Root) SetIFlags(i *flag.FlagSet) {
	c.iflags = i
}

// SetLFlags implements Commander.
func (c *Root) SetLFlags(l *flag.FlagSet) {
	c.lflags = l
}

// SetPFlags implements Commander.
func (c *Root) SetPFlags(l *flag.FlagSet) {
	c.pflags = l
}

// SetParentsPFlags implements Commander.
func (c *Root) SetParentsPFlags(pf *flag.FlagSet) {
	c.parentsPflags = pf
}

// Context returns underlying command context. If command was executed
// with ExecuteContext or the context was set with SetContext, the
// previously set context will be returned. Otherwise, nil is returned.
//
// Notice that a call to Execute and ExecuteC will replace a nil context of
// a command with a context.Background, so a background context will be
// returned by Context after one of these functions has been called.
func (c *Root) Context() context.Context {
	return c.ctx
}

// SetContext sets context for the command. This context will be overwritten by
// Command.ExecuteContext or Command.ExecuteContextC.
func (c *Root) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetArgs sets arguments for the command. It is set to os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (c *Root) SetArgs(a ...string) {
	c.args = a
}

// SetUsageFunc sets usage function. Usage can be defined by application.
func (c *Root) SetUsageFunc(f func(Commander) error) {
	c.usageFunc = f
}

// SetUsageTemplate sets usage template. Can be defined by Application.
func (c *Root) SetUsageTemplate(s string) {
	c.usageTemplate = s
}

// SetFlagErrorFunc sets a function to generate an error when flag parsing
// fails.
func (c *Root) SetFlagErrorFunc(f func(Commander, error) error) {
	c.flagErrorFunc = f
}

// SetHelpFunc sets help function. Can be defined by Application.
func (c *Root) SetHelpFunc(f func(Commander, []string)) {
	c.helpFunc = f
}

// SetHelpCommand sets help command.
func (c *Root) SetHelpCommand(cmd Commander) {
	c.helpCommand = cmd
}

// SetHelpCommandGroupID sets the group id of the help command.
func (c *Root) SetHelpCommandGroupID(groupID string) {
	if c.helpCommand != nil {
		c.helpCommand.SetGroupID(groupID)
	}
	// helpCommandGroupID is used if no helpCommand is defined by the user
	c.helpCommandGroupID = groupID
}

// SetCompletionCommandGroupID sets the group id of the completion command.
func (c *Root) SetCompletionCommandGroupID(groupID string) {
	// completionCommandGroupID is used if no completion command is defined by the user
	Base(c).SetCompletionCommandGroupID(groupID)
}

// SetHelpTemplate sets help template to be used. Application can use it to set custom template.
func (c *Root) SetHelpTemplate(s string) {
	c.helpTemplate = s
}

// SetVersionTemplate sets version template to be used. Application can use it to set custom template.
func (c *Root) SetVersionTemplate(s string) {
	c.versionTemplate = s
}

// SetErrPrefix sets error message prefix to be used. Application can use it to set custom prefix.
func (c *Root) SetErrPrefix(s string) {
	c.errPrefix = s
}

// SetGlobalNormalizationFunc sets a normalization function to all flag sets and also to child commands.
// The user should not have a cyclic dependency on commands.
func SetGlobalNormalizationFunc(c Commander, n func(f *flag.FlagSet, name string) flag.NormalizedName) {
	Flags(c).SetNormalizeFunc(n)
	PersistentFlags(c).SetNormalizeFunc(n)
	c.SetGlobNormFunc(n)

	for _, command := range c.Commands() {
		SetGlobalNormalizationFunc(command, n)
	}
}

// Usage puts out the usage for the command.
// Used when a user provides invalid input.
// Can be defined by user by overriding UsageFunc.
func Usage(c Commander) error {
	// return c.UsageFunc()(c)

	mergePersistentFlags(c)
	err := tmpl(log.OutOrStderr(), UsageTemplate(c), c)
	if err != nil {
		log.PrintErrLn(err)
	}
	return err
}

// HelpFunc returns either the function set by SetHelpFunc for this command
// or a parent, or it returns a function with default help behavior.
func HelpFunc(c Commander, a []string) {
	mergePersistentFlags(c)
	// The help should be sent to stdout
	// See https://github.com/spf13/cobra/issues/1002
	err := tmpl(log.OutOrStdout(), HelpTemplate(c), c)
	if err != nil {
		fmt.Println(">>>>>", err)
		log.PrintErrLn(err)
	}

	// log.Print(HelpTemplate(c))
}

// FlagErrorFunc returns either the function set by SetFlagErrorFunc for this
// command or a parent, or it returns a function which returns the original
// error.
func FlagErrorFunc(c Commander) (f func(Commander, error) error) {
	if c.GetFlagErrorFunc() != nil {
		return c.GetFlagErrorFunc()
	}

	if HasParent(c) {
		return c.Parent().GetFlagErrorFunc()
	}
	return func(c Commander, err error) error {
		return err
	}
}

var minUsagePadding = 25

// UsagePadding return padding for the usage.
func (c *Root) UsagePadding() int {
	if c.parent == nil || minUsagePadding > c.parent.GetCommandsMaxUseLen() {
		return minUsagePadding
	}
	return c.parent.GetCommandsMaxUseLen()
}

var minCommandPathPadding = 11

// CommandPathPadding return padding for the command path.
func CommandPathPadding(c Commander) int {
	if c.Parent() == nil || minCommandPathPadding > c.Parent().GetCommandsMaxCommandPathLen() {
		return minCommandPathPadding
	}
	return c.Parent().GetCommandsMaxCommandPathLen()
}

var minNamePadding = 11

// NamePadding returns padding for the name.
func NamePadding(c Commander) int {
	if c.Parent() == nil || minNamePadding > c.Parent().GetCommandsMaxNameLen() {
		return minNamePadding
	}
	return c.Parent().GetCommandsMaxNameLen()
}

// VersionTemplate return version template for the command.
func VersionTemplate(c Commander) string {
	// 	if c.versionTemplate != "" {
	// 		return c.versionTemplate
	// 	}

	// 	if c.HasParent() {
	// 		return c.parent.VersionTemplate()
	// 	}
	// 	return `{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
	// `

	return "VersionTemplate"
}

// ErrPrefix return error message prefix for the command
func (c *Root) ErrPrefix() string {
	if c.errPrefix != "" {
		return c.errPrefix
	}

	if HasParent(c) {
		return c.parent.ErrPrefix()
	}
	return "Error:"
}

func hasNoOptDefVal(name string, fs *flag.FlagSet) bool {
	flag := fs.Lookup(name)
	if flag == nil {
		return false
	}
	return flag.NoOptDefVal != ""
}

func shortHasNoOptDefVal(name string, fs *flag.FlagSet) bool {
	if len(name) == 0 {
		return false
	}

	flag := fs.ShorthandLookup(name[:1])
	if flag == nil {
		return false
	}
	return flag.NoOptDefVal != ""
}

func stripFlags(args []string, c Commander) []string {
	if len(args) == 0 {
		return args
	}
	mergePersistentFlags(c)

	commands := []string{}
	flags := Flags(c)

Loop:
	for len(args) > 0 {
		s := args[0]
		args = args[1:]
		switch {
		case s == "--":
			// "--" terminates the flags
			break Loop
		case strings.HasPrefix(s, "--") && !strings.Contains(s, "=") && !hasNoOptDefVal(s[2:], flags):
			// If '--flag arg' then
			// delete arg from args.
			fallthrough // (do the same as below)
		case strings.HasPrefix(s, "-") && !strings.Contains(s, "=") && len(s) == 2 && !shortHasNoOptDefVal(s[1:], flags):
			// If '-f arg' then
			// delete 'arg' from args or break the loop if len(args) <= 1.
			if len(args) <= 1 {
				break Loop
			} else {
				args = args[1:]
				continue
			}
		case s != "" && !strings.HasPrefix(s, "-"):
			commands = append(commands, s)
		}
	}

	return commands
}

// argsMinusFirstX removes only the first x from args.  Otherwise, commands that look like
// openshift admin policy add-role-to-user admin my-user, lose the admin argument (arg[4]).
// Special care needs to be taken not to remove a flag value.
func argsMinusFirstX(c Commander, args []string, x string) []string {
	if len(args) == 0 {
		return args
	}
	mergePersistentFlags(c)
	flags := Flags(c)

Loop:
	for pos := 0; pos < len(args); pos++ {
		s := args[pos]
		switch {
		case s == "--":
			// -- means we have reached the end of the parseable args. Break out of the loop now.
			break Loop
		case strings.HasPrefix(s, "--") && !strings.Contains(s, "=") && !hasNoOptDefVal(s[2:], flags):
			fallthrough
		case strings.HasPrefix(s, "-") && !strings.Contains(s, "=") && len(s) == 2 && !shortHasNoOptDefVal(s[1:], flags):
			// This is a flag without a default value, and an equal sign is not used. Increment pos in order to skip
			// over the next arg, because that is the value of this flag.
			pos++
			continue
		case !strings.HasPrefix(s, "-"):
			// This is not a flag or a flag value. Check to see if it matches what we're looking for, and if so,
			// return the args, excluding the one at this position.
			if s == x {
				ret := make([]string, 0, len(args)-1)
				ret = append(ret, args[:pos]...)
				ret = append(ret, args[pos+1:]...)
				return ret
			}
		}
	}
	return args
}

func isFlagArg(arg string) bool {
	return ((len(arg) >= 3 && arg[0:2] == "--") ||
		(len(arg) >= 2 && arg[0] == '-' && arg[1] != '-'))
}

// Find the target command given the args and command tree
// Meant to be run on the highest node. Only searches down.
func Find(c Commander, args []string) (Commander, []string, error) {
	var innerFind func(Commander, []string) (Commander, []string)

	innerFind = func(c Commander, innerArgs []string) (Commander, []string) {
		argsWOflags := stripFlags(innerArgs, c)
		if len(argsWOflags) == 0 {
			return c, innerArgs
		}
		nextSubCmd := argsWOflags[0]

		cmd := findNext(c, nextSubCmd)
		if cmd != nil {
			return innerFind(cmd, argsMinusFirstX(c, innerArgs, nextSubCmd))
		}
		return c, innerArgs
	}

	commandFound, a := innerFind(c, args)
	if commandFound.GetArgs() == nil {
		return commandFound, a, legacyArgs(commandFound, stripFlags(a, commandFound))
	}
	return commandFound, a, nil
}

func findSuggestions(c Commander, arg string) string {
	if c.GetDisableSuggestions() {
		return ""
	}
	// if c.GetSuggestionsMinimumDistance() <= 0 {
	// 	c.SetSuggestionsMinimumDistance(2)
	// }
	var sb strings.Builder
	if suggestions := SuggestionsFor(c, arg); len(suggestions) > 0 {
		sb.WriteString("\n\nDid you mean this?\n")
		for _, s := range suggestions {
			_, _ = fmt.Fprintf(&sb, "\t%v\n", s)
		}
	}
	return sb.String()
}

func findNext(c Commander, next string) Commander {
	matches := make([]Commander, 0)
	for _, cmd := range c.Commands() {
		if commandNameMatches(name(cmd), next) || HasAlias(cmd, next) {
			cmd.GetCommandCalledAs().name = next
			return cmd
		}
		if EnablePrefixMatching && hasNameOrAliasPrefix(cmd, next) {
			matches = append(matches, cmd)
		}
	}

	if len(matches) == 1 {
		// Temporarily disable gosec G602, which produces a false positive.
		// See https://github.com/securego/gosec/issues/1005.
		return matches[0] // #nosec G602
	}

	return nil
}

// Traverse the command tree to find the command, and parse args for
// each parent.
func Traverse(c Commander, args []string) (Commander, []string, error) {
	flags := []string{}
	inFlag := false

	for i, arg := range args {
		switch {
		// A long flag with a space separated value
		case strings.HasPrefix(arg, "--") && !strings.Contains(arg, "="):
			// TODO: this isn't quite right, we should really check ahead for 'true' or 'false'
			inFlag = !hasNoOptDefVal(arg[2:], Flags(c))
			flags = append(flags, arg)
			continue
		// A short flag with a space separated value
		case strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") && len(arg) == 2 && !shortHasNoOptDefVal(arg[1:], Flags(c)):
			inFlag = true
			flags = append(flags, arg)
			continue
		// The value for a flag
		case inFlag:
			inFlag = false
			flags = append(flags, arg)
			continue
		// A flag without a value, or with an `=` separated value
		case isFlagArg(arg):
			flags = append(flags, arg)
			continue
		}

		cmd := findNext(c, arg)
		if cmd == nil {
			return c, args, nil
		}

		if err := ParseFlags(c, flags); err != nil {
			return nil, args, err
		}
		return Traverse(cmd, args[i+1:])
	}
	return c, args, nil
}

// SuggestionsFor provides suggestions for the typedName.
func SuggestionsFor(c Commander, typedName string) []string {
	suggestions := []string{}
	for _, cmd := range c.Commands() {
		if IsAvailableCommand(cmd) {
			levenshteinDistance := ld(typedName, name(cmd), true)
			suggestByLevenshtein := levenshteinDistance <= c.GetSuggestionsMinimumDistance()
			suggestByPrefix := strings.HasPrefix(strings.ToLower(name(cmd)), strings.ToLower(typedName))
			if suggestByLevenshtein || suggestByPrefix {
				suggestions = append(suggestions, name(cmd))
			}
			for _, explicitSuggestion := range cmd.GetSuggestFor() {
				if strings.EqualFold(typedName, explicitSuggestion) {
					suggestions = append(suggestions, name(cmd))
				}
			}
		}
	}
	return suggestions
}

// VisitParents visits all parents of the command and invokes fn on each parent.
func VisitParents(c Commander, fn func(Commander)) {
	if HasParent(c) {
		fn(c.Parent())
		VisitParents(c.Parent(), fn)
	}
}

// Base finds root command.
func Base(c Commander) Commander {
	if HasParent(c) {
		return Base(c.Parent())
	}
	return c
}

// ArgsLenAtDash will return the length of c.Flags().Args at the moment
// when a -- was found during args parsing.
func (c *Root) ArgsLenAtDash() int {
	return Flags(c).ArgsLenAtDash()
}

func Execute(c Commander, a []string) (err error) {
	if c == nil {
		return fmt.Errorf("called Execute() on a nil Command")
	}

	if len(c.GetDeprecated()) > 0 {
		log.Printf("Command %q is deprecated, %s\n", name(c), c.GetDeprecated())
	}

	// initialize help and version flag at the last point possible to allow for user
	// overriding
	InitDefaultHelpFlag(c)
	InitDefaultVersionFlag(c)

	err = ParseFlags(c, a)
	if err != nil {
		return FlagErrorFunc(c)(c, err)
	}

	// If help is called, regardless of other flags, return we want help.
	// Also say we need help if the command isn't runnable.
	helpVal, err := Flags(c).GetBool("help")
	if err != nil {
		// should be impossible to get here as we always declare a help
		// flag in InitDefaultHelpFlag()
		log.Println("\"help\" flag declared as non-bool. Please correct your code")
		return err
	}

	if helpVal {
		return flag.ErrHelp
	}

	// for back-compat, only add version flag behavior if version is defined
	if c.GetVersion() != "" {
		versionVal, err := Flags(c).GetBool("version")
		if err != nil {
			log.Println("\"version\" flag declared as non-bool. Please correct your code")
			return err
		}
		if versionVal {
			err := tmpl(log.OutOrStdout(), VersionTemplate(c), c)
			if err != nil {
				log.Println(err)
			}
			return err
		}
	}

	if !c.Runnable() {
		return flag.ErrHelp
	}

	// todo: wait
	// c.preRun()
	// defer c.postRun()

	argWoFlags := Flags(c).Args()
	if c.GetDisableFlagParsing() {
		argWoFlags = a
	}

	if err := ValidateArgs(c, argWoFlags); err != nil {
		return err
	}

	parents := make([]Commander, 0, 5)
	var pc Commander
	for pc = c; pc != nil; pc = pc.Parent() {
		if EnableTraverseRunHooks {
			// When EnableTraverseRunHooks is set:
			// - Execute all persistent pre-runs from the root parent till this command.
			// - Execute all persistent post-runs from this command till the root parent.
			parents = append([]Commander{pc}, parents...)
		} else {
			// Otherwise, execute only the first found persistent hook.
			parents = append(parents, pc)
		}
	}
	for _, p := range parents {
		if p.GetPersistentPreRunE() != nil {
			if err := p.GetPersistentPreRunE()(c, argWoFlags); err != nil {
				return err
			}
			if !EnableTraverseRunHooks {
				break
			}
		} else if p.GetPersistentPreRun() != nil {
			p.GetPersistentPreRun()(c, argWoFlags)
			if !EnableTraverseRunHooks {
				break
			}
		}
	}

	if err := c.PreRun(argWoFlags); err != nil {
		return err
	}

	if err := ValidateRequiredFlags(c); err != nil {
		return err
	}
	if err := ValidateFlagGroups(c); err != nil {
		return err
	}

	if err := c.Run(argWoFlags); err != nil {
		return err
	}

	if err := c.PostRun(argWoFlags); err != nil {
		return err
	}
	var p Commander
	for p = c; p != nil; p = p.Parent() {
		if p.GetPersistentPostRunE() != nil {
			if err := p.GetPersistentPostRunE()(c, argWoFlags); err != nil {
				return err
			}
			if !EnableTraverseRunHooks {
				break
			}
		} else if p.GetPersistentPostRun() != nil {
			p.GetPersistentPostRun()(c, argWoFlags)
			if !EnableTraverseRunHooks {
				break
			}
		}
	}

	return nil
}

func (c *Root) preRun() {
	for _, x := range initializers {
		x()
	}
}

func (c *Root) postRun() {
	for _, x := range finalizers {
		x()
	}
}

// Execute uses the args (os.Args[1:] by default)
// and run through the command tree finding appropriate matches
// for commands and then corresponding flags.
func (c *Root) Execute() error { // todo: 原Execute
	_, err := c.ExecuteC()
	return err
}

// ExecuteC executes the command.
func (c *Root) ExecuteC() (cmd Commander, err error) {
	if c.ctx == nil {
		c.ctx = context.Background()
	}

	// windows hook
	if preExecHookFn != nil {
		preExecHookFn(c)
	}

	// initialize help at the last point to allow for user overriding
	InitDefaultHelpCmd(c)
	// initialize completion at the last point to allow for user overriding
	InitDefaultCompletionCmd(c)

	// Now that all commands have been created, let's make sure all groups
	// are properly created also
	CheckCommandGroups(c)

	args := c.args

	// Workaround FAIL with "go test -v" or "cobra.test -test.v", see #155
	if c.args == nil && filepath.Base(os.Args[0]) != "cobra.test" {
		args = os.Args[1:]
	}

	// initialize the hidden command to be used for shell completion
	c.initCompleteCmd(args)

	var flags []string
	if c.TraverseChildren {
		cmd, flags, err = Traverse(c, args)
	} else {
		cmd, flags, err = Find(c, args)
	}
	if err != nil {
		var dc Commander
		// If found parse to a subcommand and then failed, talk about the subcommand
		if cmd != nil {
			dc = cmd
		} else {
			dc = c
		}
		if !dc.GetSilenceErrors() {
			log.PrintErrLn(c.ErrPrefix(), err.Error())
			log.PrintErrF("Run '%v --help' for usage.\n", CommandPath(c))
		}
		return dc, err
	}
	as := cmd.GetCommandCalledAs()
	as.called = true
	if as.name == "" {
		as.name = name(cmd)
	}

	// We have to pass global context to children command
	// if context is present on the parent command.
	if cmd.Context() == nil {
		cmd.SetContext(c.ctx)
	}

	err = Execute(cmd, flags)
	if err != nil {
		// Always show help if requested, even if SilenceErrors is in
		// effect
		if errors.Is(err, flag.ErrHelp) {
			HelpFunc(cmd, args)
			return cmd, nil
		}

		// If root command has SilenceErrors flagged,
		// all subcommands should respect it
		if !cmd.GetSilenceErrors() && !c.GetSilenceErrors() {
			log.PrintErrLn(cmd.ErrPrefix(), err.Error())
		}

		// If root command has SilenceUsage flagged,
		// all subcommands should respect it
		if !cmd.GetSilenceUsage() && !c.GetSilenceUsage() {
			log.Println(UsageString(cmd))
		}
	}
	return cmd, err
}

func ValidateArgs(c Commander, args []string) error {
	cArgs := c.GetArgs()
	if cArgs == nil {
		return ArbitraryArgs(c, args)
	}
	return cArgs(c, args)
}

// ValidateRequiredFlags validates all required flags are present and returns an error otherwise
func ValidateRequiredFlags(c Commander) error {
	if c.GetDisableFlagParsing() {
		// if c.DisableFlagParsing {
		return nil
	}

	flags := Flags(c)
	missingFlagNames := []string{}
	flags.VisitAll(func(pflag *flag.Flag) {
		requiredAnnotation, found := pflag.Annotations[BashCompOneRequiredFlag]
		if !found {
			return
		}
		if (requiredAnnotation[0] == "true") && !pflag.Changed {
			missingFlagNames = append(missingFlagNames, pflag.Name)
		}
	})

	if len(missingFlagNames) > 0 {
		return fmt.Errorf(`required flag(s) "%s" not set`, strings.Join(missingFlagNames, `", "`))
	}
	return nil
}

// checkCommandGroups checks if a command has been added to a group that does not exists.
// If so, we panic because it indicates a coding error that should be corrected.
func CheckCommandGroups(c Commander) {
	for _, sub := range c.Commands() {
		// if Group is not defined let the developer know right away
		if sub.GetGroupID() != "" && !ContainsGroup(c, sub.GetGroupID()) {
			panic(fmt.Sprintf("group id '%s' is not defined for subcommand '%s'", sub.GetGroupID(), CommandPath(sub)))
		}

		CheckCommandGroups(sub)
	}
}

// InitDefaultHelpFlag adds default help flag to c.
// It is called automatically by executing the c or by calling help and usage.
// If c already has help flag, it will do nothing.
func InitDefaultHelpFlag(c Commander) {
	mergePersistentFlags(c)
	if Flags(c).Lookup("help") == nil {
		usage := "help for "
		name := displayName(c)
		if name == "" {
			usage += "this command"
		} else {
			usage += name
		}
		Flags(c).BoolP("help", "h", false, usage)
		_ = Flags(c).SetAnnotation("help", FlagSetByCobraAnnotation, []string{"true"})
	}
}

// InitDefaultVersionFlag adds default version flag to c.
// It is called automatically by executing the c.
// If c already has a version flag, it will do nothing.
// If c.Version is empty, it will do nothing.
func InitDefaultVersionFlag(c Commander) {
	if c.GetVersion() == "" {
		return
	}

	mergePersistentFlags(c)
	if Flags(c).Lookup("version") == nil {
		usage := "version for "
		if name(c) == "" {
			usage += "this command"
		} else {
			usage += name(c)
		}
		if Flags(c).ShorthandLookup("v") == nil {
			Flags(c).BoolP("version", "v", false, usage)
		} else {
			Flags(c).Bool("version", false, usage)
		}
		_ = Flags(c).SetAnnotation("version", FlagSetByCobraAnnotation, []string{"true"})
	}
}

// InitDefaultHelpCmd adds default help command to c.
// It is called automatically by executing the c or by calling help and usage.
// If c already has help command or c has no subcommands, it will do nothing.
func InitDefaultHelpCmd(c Commander) {
	if !HasSubCommands(c) {
		return
	}

	// if c.helpCommand == nil {
	// 	// 		c.helpCommand = &Root{
	// 	// 			Use:   "help [command]",
	// 	// 			Short: "Help about any command",
	// 	// 			Long: `Help provides help for any command in the application.
	// 	// Simply type ` + c.displayName() + ` help [path to command] for full details.`,
	// 	// 			ValidArgsFunction: func(c Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
	// 	// 				var completions []string
	// 	// 				cmd, _, e := c.Base().Find(args)
	// 	// 				if e != nil {
	// 	// 					return nil, ShellCompDirectiveNoFileComp
	// 	// 				}
	// 	// 				if cmd == nil {
	// 	// 					// Root help command.
	// 	// 					cmd = c.Base()
	// 	// 				}
	// 	// 				for _, subCmd := range cmd.Commands() {
	// 	// 					if subCmd.IsAvailableCommand() || subCmd == cmd.GetHelpCommand() {
	// 	// 						if strings.HasPrefix(subCmd.Name(), toComplete) {
	// 	// 							completions = append(completions, fmt.Sprintf("%s\t%s", subCmd.Name(), subCmd.GetShort()))
	// 	// 						}
	// 	// 					}
	// 	// 				}
	// 	// 				return completions, ShellCompDirectiveNoFileComp
	// 	// 			},
	// 	// 			Run: func(c Commander, args []string) {
	// 	// 				cmd, _, e := c.Base().Find(args)
	// 	// 				if cmd == nil || e != nil {
	// 	// 					c.Printf("Unknown help topic %#q\n", args)
	// 	// 					CheckErr(c.Base().Usage())
	// 	// 				} else {
	// 	// 					cmd.InitDefaultHelpFlag()    // make possible 'help' flag to be shown
	// 	// 					cmd.InitDefaultVersionFlag() // make possible 'version' flag to be shown
	// 	// 					CheckErr(cmd.Help())
	// 	// 				}
	// 	// 			},
	// 	// 			GroupID: c.helpCommandGroupID,
	// 	// 		}
	// 	c.helpCommand = NewHelpCmd(c)
	// }
	// c.RemoveCommand(c.helpCommand)
	c.Add(NewHelpCmd(c))
}

// ResetCommands delete parent, subcommand and help command from c.
func (c *Root) ResetCommands() {
	c.parent = nil
	c.commands = nil
	c.helpCommand = nil
	c.parentsPflags = nil
}

// Sorts commands by their names.
type commandSorterByName []Commander

func (c commandSorterByName) Len() int           { return len(c) }
func (c commandSorterByName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c commandSorterByName) Less(i, j int) bool { return name(c[i]) < name(c[j]) }

// Commands returns a sorted slice of child commands.
func (c *Root) Commands() []Commander {
	// do not sort commands if it already sorted or sorting was disabled
	if EnableCommandSorting && !c.commandsAreSorted {
		sort.Sort(commandSorterByName(c.commands))
		c.commandsAreSorted = true
	}
	return c.commands
}

// Add adds one or more commands to this parent command.
func (c *Root) Add(cmds ...Commander) {
	for i, x := range cmds {
		if cmds[i] == c {
			panic("Command can't be a child of itself")
		}
		cmds[i].SetParent(c)
		// cmds[i].parent = c
		// update max lengths
		usageLen := len(x.GetUse())
		if usageLen > c.commandsMaxUseLen {
			c.commandsMaxUseLen = usageLen
		}
		commandPathLen := len(CommandPath(x))
		if commandPathLen > c.commandsMaxCommandPathLen {
			c.commandsMaxCommandPathLen = commandPathLen
		}
		nameLen := len(name(x))
		if nameLen > c.commandsMaxNameLen {
			c.commandsMaxNameLen = nameLen
		}
		// If global normalization function exists, update all children
		if c.globNormFunc != nil {
			SetGlobalNormalizationFunc(x, c.globNormFunc)
		}
		c.commands = append(c.commands, x)
		c.commandsAreSorted = false
	}
}

// Groups returns a slice of child command groups.
func (c *Root) Groups() []*Group {
	return c.commandgroups
}

// AllChildCommandsHaveGroup returns if all subcommands are assigned to a group
func AllChildCommandsHaveGroup(c Commander) bool {
	for _, sub := range c.Commands() {
		if (IsAvailableCommand(sub) || sub == c.GetHelpCommand()) && sub.GetGroupID() == "" {
			return false
		}
	}
	return true
}

// ContainsGroup return if groupID exists in the list of command groups.
func ContainsGroup(c Commander, groupID string) bool {
	for _, x := range c.GetCommandGroups() {
		if x.ID == groupID {
			return true
		}
	}
	return false
}

// AddGroup adds one or more command groups to this parent command.
func (c *Root) AddGroup(groups ...*Group) {
	c.commandgroups = append(c.commandgroups, groups...)
}

// RemoveCommand removes one or more commands from a parent command.
func RemoveCommand(c Commander, cmds ...Commander) {
	commands := []Commander{}
main:
	for _, command := range c.Commands() {
		for _, cmd := range cmds {
			if command == cmd {
				// command.parent = nil
				command.SetParent(nil)
				continue main
			}
		}
		commands = append(commands, command)
	}
	c.ResetAdd(commands...)
}

func (c *Root) ResetAdd(cmds ...Commander) {
	c.commands = cmds
	// recompute all lengths
	c.commandsMaxUseLen = 0
	c.commandsMaxCommandPathLen = 0
	c.commandsMaxNameLen = 0
	for _, command := range c.commands {
		usageLen := len(command.GetUse())
		if usageLen > c.commandsMaxUseLen {
			c.commandsMaxUseLen = usageLen
		}
		commandPathLen := len(CommandPath(command))
		if commandPathLen > c.commandsMaxCommandPathLen {
			c.commandsMaxCommandPathLen = commandPathLen
		}
		nameLen := len(name(command))
		if nameLen > c.commandsMaxNameLen {
			c.commandsMaxNameLen = nameLen
		}
	}
}

// CommandPath returns the full path to this command.
func CommandPath(c Commander) string {
	if HasParent(c) {
		return CommandPath(c.Parent()) + " " + name(c)
	}
	return displayName(c)
}

func displayName(c Commander) string {
	annotations := c.GetAnnotations()
	if displayName, ok := annotations[CommandDisplayNameAnnotation]; ok {
		return displayName
	}
	return name(c)
}

// UseLine puts out the full usage for a given command (including parents).
func UseLine(c Commander) string {
	var useLine string
	use := strings.Replace(c.GetUse(), name(c), displayName(c), 1)
	if HasParent(c) {
		useLine = CommandPath(c.Parent()) + " " + use
	} else {
		useLine = use
	}
	if c.GetDisableFlagsInUseLine() {
		return useLine
	}
	if HasAvailableFlags(c) && !strings.Contains(useLine, "[flags]") {
		useLine += " [flags]"
	}
	return useLine
}

// DebugFlags used to determine which flags have been assigned to which commands
// and which persist.
func DebugFlags(c Commander) {
	log.Println("DebugFlags called on", name(c))
	var debugFlags func(Commander)

	debugFlags = func(x Commander) {
		if HasFlags(x) || HasPersistentFlags(x) {
			log.Println(name(x))
		}
		if HasFlags(x) {
			x.GetFlags().VisitAll(func(f *flag.Flag) {
				if HasPersistentFlags(x) && persistentFlag(x, f.Name) != nil {
					log.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [LP]")
				} else {
					log.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [L]")
				}
			})
		}
		if HasPersistentFlags(x) {
			x.GetFlags().VisitAll(func(f *flag.Flag) {
				if HasFlags(x) {
					if x.GetFlags().Lookup(f.Name) == nil {
						log.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
					}
				} else {
					log.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
				}
			})
		}
		log.Println(x.GetFlagErrorBuf())
		if HasSubCommands(x) {
			for _, y := range x.Commands() {
				debugFlags(y)
			}
		}
	}

	debugFlags(c)
}

// Name returns the command's name: the first word in the use line.
func name(c Commander) string {
	name := c.GetUse()
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// HasAlias determines if a given string is an alias of the command.
func HasAlias(c Commander, s string) bool {
	for _, a := range c.GetAliases() {
		if commandNameMatches(a, s) {
			return true
		}
	}
	return false
}

// CalledAs returns the command name or alias that was used to invoke
// this command or an empty string if the command has not been called.
func (c *Root) CalledAs() string {
	if c.commandCalledAs.called {
		return c.commandCalledAs.name
	}
	return ""
}

// hasNameOrAliasPrefix returns true if the Name or any of aliases start
// with prefix
func hasNameOrAliasPrefix(c Commander, prefix string) bool {

	if strings.HasPrefix(name(c), prefix) {
		c.GetCommandCalledAs().name = name(c)
		// c.commandCalledAs.name = name(c)
		return true
	}

	for _, alias := range c.GetAliases() {
		if strings.HasPrefix(alias, prefix) {
			c.GetCommandCalledAs().name = alias
			return true
		}
	}
	return false
}

// NameAndAliases returns a list of the command name and all aliases
func (c *Root) NameAndAliases() string {
	return strings.Join(append([]string{name(c)}, c.Aliases...), ", ")
}

// HasExample determines if the command has example.
func (c *Root) HasExample() bool {
	return len(c.Example) > 0
}

// Runnable determines if the command is itself runnable.
func (c *Root) Runnable() bool {
	return true
}

// HasSubCommands determines if the command has children commands.
func HasSubCommands(c Commander) bool {
	return len(c.Commands()) > 0
}

// IsAvailableCommand determines if a command is available as a non-help command
// (this includes all non deprecated/hidden commands).
func IsAvailableCommand(c Commander) bool {

	if len(c.GetDeprecated()) != 0 || c.GetHidden() {
		return false
	}

	if HasParent(c) && c.Parent().GetHelpCommand() == c {
		return false
	}

	if c.Runnable() || HasAvailableSubCommands(c) {
		return true
	}

	return false
}

// IsAdditionalHelpTopicCommand determines if a command is an additional
// help topic command; additional help topic command is determined by the
// fact that it is NOT runnable/hidden/deprecated, and has no sub commands that
// are runnable/hidden/deprecated.
// Concrete example: https://github.com/spf13/cobra/issues/393#issuecomment-282741924.
func IsAdditionalHelpTopicCommand(c Commander) bool {
	// if a command is runnable, deprecated, or hidden it is not a 'help' command
	if c.Runnable() || len(c.GetDeprecated()) != 0 || c.GetHidden() {
		return false
	}

	// if any non-help sub commands are found, the command is not a 'help' command
	for _, sub := range c.Commands() {
		if !IsAdditionalHelpTopicCommand(sub) {
			return false
		}
	}

	// the command either has no sub commands, or no non-help sub commands
	return true
}

// HasHelpSubCommands determines if a command has any available 'help' sub commands
// that need to be shown in the usage/help default template under 'additional help
// topics'.
func HasHelpSubCommands(c Commander) bool {
	// return true on the first found available 'help' sub command
	for _, sub := range c.Commands() {
		if IsAdditionalHelpTopicCommand(sub) {
			return true
		}
	}

	// the command either has no sub commands, or no available 'help' sub commands
	return false
}

// HasAvailableSubCommands determines if a command has available sub commands that
// need to be shown in the usage/help default template under 'available commands'.
func HasAvailableSubCommands(c Commander) bool {
	// return true on the first found available (non deprecated/help/hidden)
	// sub command
	for _, sub := range c.Commands() {
		if IsAvailableCommand(sub) {
			return true
		}
	}

	// the command either has no sub commands, or no available (non deprecated/help/hidden)
	// sub commands
	return false
}

// HasParent determines if the command is a child command.
func HasParent(c Commander) bool {
	return c.Parent() != nil
}

// GlobalNormalizationFunc returns the global normalization function or nil if it doesn't exist.
func (c *Root) GlobalNormalizationFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
}

// Flags returns the complete FlagSet that applies
// to this command (local and persistent declared here and by all parents).
func Flags(c Commander) *flag.FlagSet {
	if c.GetFlags() == nil {
		// c.flags = flag.NewFlagSet(displayName(c), flag.ContinueOnError)
		c.SetFlags(flag.NewFlagSet(displayName(c), flag.ContinueOnError))

		if c.GetFlagErrorBuf() == nil {
			c.SetFlagErrorBuf(new(bytes.Buffer))
		}
		c.GetFlags().SetOutput(c.GetFlagErrorBuf())
	}

	return c.GetFlags()
}

// LocalNonPersistentFlags are flags specific to this command which will NOT persist to subcommands.
// This function does not modify the flags of the current command, it's purpose is to return the current state.
func LocalNonPersistentFlags(c Commander) *flag.FlagSet {
	persistentFlags := PersistentFlags(c)

	out := flag.NewFlagSet(displayName(c), flag.ContinueOnError)
	LocalFlags(c).VisitAll(func(f *flag.Flag) {
		if persistentFlags.Lookup(f.Name) == nil {
			out.AddFlag(f)
		}
	})
	return out
}

// LocalFlags returns the local FlagSet specifically set in the current command.
// This function does not modify the flags of the current command, it's purpose is to return the current state.
func LocalFlags(c Commander) *flag.FlagSet {
	mergePersistentFlags(c)

	if c.GetLFlags() == nil {
		c.SetLFlags(flag.NewFlagSet(displayName(c), flag.ContinueOnError))

		if c.GetFlagErrorBuf() == nil {
			c.SetFlagErrorBuf(new(bytes.Buffer))
			// c.flagErrorBuf = new(bytes.Buffer)
		}

		c.GetLFlags().SetOutput(c.GetFlagErrorBuf())
	}
	c.GetLFlags().SortFlags = Flags(c).SortFlags
	if c.GetGlobNormFunc() != nil {
		c.GetLFlags().SetNormalizeFunc(c.GetGlobNormFunc())
	}

	addToLocal := func(f *flag.Flag) {
		// Add the flag if it is not a parent PFlag, or it shadows a parent PFlag
		if c.GetLFlags().Lookup(f.Name) == nil && f != c.GetParentsPFlags().Lookup(f.Name) {
			c.GetLFlags().AddFlag(f)
		}
	}
	Flags(c).VisitAll(addToLocal)
	PersistentFlags(c).VisitAll(addToLocal)
	return c.GetLFlags()
}

// InheritedFlags returns all flags which were inherited from parent commands.
// This function does not modify the flags of the current command, it's purpose is to return the current state.
func InheritedFlags(c Commander) *flag.FlagSet {
	mergePersistentFlags(c)

	if c.GetIFlags() == nil {
		c.SetIFlags(flag.NewFlagSet(displayName(c), flag.ContinueOnError))
		if c.GetFlagErrorBuf() == nil {
			c.SetFlagErrorBuf(new(bytes.Buffer))
		}
		c.GetIFlags().SetOutput(c.GetFlagErrorBuf())
	}

	local := LocalFlags(c)
	if c.GetGlobNormFunc() != nil {

		c.GetIFlags().SetNormalizeFunc(c.GetGlobNormFunc())
	}

	c.GetParentsPFlags().VisitAll(func(f *flag.Flag) {
		if c.GetIFlags().Lookup(f.Name) == nil && local.Lookup(f.Name) == nil {
			c.GetIFlags().AddFlag(f)
		}
	})
	return c.GetIFlags()
}

// NonInheritedFlags returns all flags which were not inherited from parent commands.
// This function does not modify the flags of the current command, it's purpose is to return the current state.
func NonInheritedFlags(c Commander) *flag.FlagSet {
	return LocalFlags(c)
}

// PersistentFlags returns the persistent FlagSet specifically set in the current command.
func PersistentFlags(c Commander) *flag.FlagSet {
	if c.GetPFlags() == nil {
		c.SetPFlags(flag.NewFlagSet(displayName(c), flag.ContinueOnError))
		// c.pflags = flag.NewFlagSet(displayName(c), flag.ContinueOnError)
		if c.GetFlagErrorBuf() == nil {
			// if c.flagErrorBuf == nil {
			// c.flagErrorBuf = new(bytes.Buffer)
			c.SetFlagErrorBuf(new(bytes.Buffer))
		}
		c.GetPFlags().SetOutput(c.GetFlagErrorBuf())
	}
	return c.GetPFlags()
}

// todo: 业务调用
// ResetFlags deletes all flags from command.
func (c *Root) ResetFlags() {
	c.flagErrorBuf = new(bytes.Buffer)
	c.flagErrorBuf.Reset()
	c.flags = flag.NewFlagSet(displayName(c), flag.ContinueOnError)
	c.flags.SetOutput(c.flagErrorBuf)
	c.pflags = flag.NewFlagSet(displayName(c), flag.ContinueOnError)
	c.pflags.SetOutput(c.flagErrorBuf)

	c.lflags = nil
	c.iflags = nil
	c.parentsPflags = nil
}

// HasFlags checks if the command contains any flags (local plus persistent from the entire structure).
func HasFlags(c Commander) bool {
	return Flags(c).HasFlags()
}

// HasPersistentFlags checks if the command contains persistent flags.
func HasPersistentFlags(c Commander) bool {
	return PersistentFlags(c).HasFlags()
}

// HasLocalFlags checks if the command has flags specifically declared locally.
func HasLocalFlags(c Commander) bool {
	return LocalFlags(c).HasFlags()
}

// HasInheritedFlags checks if the command has flags inherited from its parent command.
func HasInheritedFlags(c Commander) bool {
	return InheritedFlags(c).HasFlags()
}

// HasAvailableFlags checks if the command contains any flags (local plus persistent from the entire
// structure) which are not hidden or deprecated.
func HasAvailableFlags(c Commander) bool {
	return Flags(c).HasAvailableFlags()
}

// HasAvailablePersistentFlags checks if the command contains persistent flags which are not hidden or deprecated.
func HasAvailablePersistentFlags(c Commander) bool {
	return PersistentFlags(c).HasAvailableFlags()
}

// HasAvailableLocalFlags checks if the command has flags specifically declared locally which are not hidden
// or deprecated.
func HasAvailableLocalFlags(c Commander) bool {
	return LocalFlags(c).HasAvailableFlags()
}

// HasAvailableInheritedFlags checks if the command has flags inherited from its parent command which are
// not hidden or deprecated.
func HasAvailableInheritedFlags(c Commander) bool {
	return InheritedFlags(c).HasAvailableFlags()
}

// Flag climbs up the command tree looking for matching flag.
func Flag(c Commander, name string) (flag *flag.Flag) {
	flag = Flags(c).Lookup(name)

	if flag == nil {
		flag = persistentFlag(c, name)
	}

	return
}

// Recursively find matching persistent flag.
func persistentFlag(c Commander, name string) (flag *flag.Flag) {
	if HasPersistentFlags(c) {
		flag = PersistentFlags(c).Lookup(name)
	}

	if flag == nil {
		updateParentsPflags(c)

		flag = c.GetParentsPFlags().Lookup(name)
	}
	return
}

// ParseFlags parses persistent flag tree and local flags.
func ParseFlags(c Commander, args []string) error {
	if c.GetDisableFlagParsing() {
		return nil
	}

	if c.GetFlagErrorBuf() == nil {
		c.SetFlagErrorBuf(new(bytes.Buffer))
	}
	beforeErrorBufLen := c.GetFlagErrorBuf().Len()
	mergePersistentFlags(c)

	// do it here after merging all flags and just before parse
	Flags(c).ParseErrorsWhitelist = flag.ParseErrorsWhitelist(c.GetFParseErrWhitelist())

	err := Flags(c).Parse(args)
	// Print warnings if they occurred (e.g. deprecated flag messages).
	if c.GetFlagErrorBuf().Len()-beforeErrorBufLen > 0 && err == nil {
		log.Print(c.GetFlagErrorBuf().String())
	}

	return err
}

// Parent returns a commands parent command.
func (c *Root) Parent() Commander {
	return c.parent
}

// mergePersistentFlags merges c.PersistentFlags() to c.Flags()
// and adds missing persistent flags of all parents.
func mergePersistentFlags(c Commander) {
	updateParentsPflags(c)
	Flags(c).AddFlagSet(PersistentFlags(c))
	Flags(c).AddFlagSet(c.GetParentsPFlags())
}

// updateParentsPflags updates c.parentsPflags by adding
// new persistent flags of all parents.
// If c.parentsPflags == nil, it makes new.
func updateParentsPflags(c Commander) {
	if c.GetParentsPFlags() == nil {
		c.SetParentsPFlags(flag.NewFlagSet(displayName(c), flag.ContinueOnError))
		c.GetParentsPFlags().SetOutput(c.GetFlagErrorBuf())
		c.GetParentsPFlags().SortFlags = false
	}

	if c.GetGlobNormFunc() != nil {
		c.GetParentsPFlags().SetNormalizeFunc(c.GetGlobNormFunc())
	}

	PersistentFlags(Base(c)).AddFlagSet(flag.CommandLine)

	VisitParents(c, func(parent Commander) {
		c.GetParentsPFlags().AddFlagSet(PersistentFlags(parent))
	})
}

// commandNameMatches checks if two command names are equal
// taking into account case sensitivity according to
// EnableCaseInsensitive global configuration.
func commandNameMatches(s string, t string) bool {
	if EnableCaseInsensitive {
		return strings.EqualFold(s, t)
	}

	return s == t
}
