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
	"sort"

	flag "github.com/spf13/pflag"
)

func ParseName(c Commander) string {
	return name(c)
}

// Root is just that, a command for your application.
// E.g.  'go run ...' - 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Default struct {
	// groups for subcommands
	commandGroups []*Group

	// args is actual args parsed from flags.
	args []string
	// flagErrorBuf contains all error messages from pflag.
	flagErrorBuf *bytes.Buffer

	// flags is full set of flags.
	flags *flag.FlagSet

	// pFlags contains persistent flags.
	pFlags *flag.FlagSet
	// lFlags contains local flags.
	// This field does not represent internal state, it's used as a cache to optimise LocalFlags function call
	lFlags *flag.FlagSet
	// iFlags contains inherited flags.
	// This field does not represent internal state, it's used as a cache to optimise InheritedFlags function call
	iFlags *flag.FlagSet
	// parentsPFlags is all persistent flags of cmd's parents.
	parentsPFlags *flag.FlagSet
	// globNormFunc is the global normalization function
	// that we can use on every pflag set and children commands
	globNormFunc func(f *flag.FlagSet, name string) flag.NormalizedName

	// usageFunc is usage func defined by user.
	// usageFunc func(Commander) error
	// usageTemplate is usage template defined by user.
	// usageTemplate string
	// flagErrorFunc is func defined by user and it's called when the parsing of
	// flags returns an error.
	flagErrorFunc func(Commander, error) error

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
}

// GetGlobNormFunc implements Commander.
func (c *Default) GetGlobNormFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
}

// GetFParseErrWhitelist implements Commander.
func (c *Default) GetFParseErrWhitelist() FParseErrWhitelist { return FParseErrWhitelist{} }

// SetFParseErrWhitelist implements Commander.
func (c *Default) SetFParseErrWhitelist(fp FParseErrWhitelist) {}

// SetGlobNormFunc implements Commander.
func (c *Default) SetGlobNormFunc(f func(f *flag.FlagSet, name string) flag.NormalizedName) {
	c.globNormFunc = f
}

// GetPFlags implements Commander.
func (c *Default) GetPFlags() *flag.FlagSet { return c.pFlags }

// GetIFlags implements Commander.
func (c *Default) GetIFlags() *flag.FlagSet { return c.iFlags }

// GetLFlags implements Commander.
func (c *Default) GetLFlags() *flag.FlagSet { return c.lFlags }

// GetParentsPFlags implements Commander.
func (c *Default) GetParentsPFlags() *flag.FlagSet { return c.parentsPFlags }

// SetIFlags implements Commander.
func (c *Default) SetIFlags(i *flag.FlagSet) { c.iFlags = i }

// SetLFlags implements Commander.
func (c *Default) SetLFlags(l *flag.FlagSet) { c.lFlags = l }

// SetPFlags implements Commander.
func (c *Default) SetPFlags(l *flag.FlagSet) { c.pFlags = l }

// SetParentsPFlags implements Commander.
func (c *Default) SetParentsPFlags(pf *flag.FlagSet) { c.parentsPFlags = pf }

// Context returns underlying command context. If command was executed
// with ExecuteContext or the context was set with SetContext, the
// previously set context will be returned. Otherwise, nil is returned.
//
// Notice that a call to Execute and ExecuteC will replace a nil context of
// a command with a context.Background, so a background context will be
// returned by Context after one of these functions has been called.
func (c *Default) Context() context.Context { return c.ctx }

// SetContext sets context for the command. This context will be overwritten by
// Command.ExecuteContext or Command.ExecuteContextC.
func (c *Default) SetContext(ctx context.Context) { c.ctx = ctx }

// SetArgs sets arguments for the command. It is set to os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (c *Default) SetArgs(a []string) { c.args = a }

// SetFlagErrorFunc sets a function to generate an error when flag parsing
// fails.
func (c *Default) SetFlagErrorFunc(f func(Commander, error) error) {
	c.flagErrorFunc = f
}

// SetHelpCommand sets help command.
func (c *Default) SetHelpCommand(cmd Commander) {
	c.helpCommand = cmd
}

// SetHelpCommandGroupID sets the group id of the help command.
func (c *Default) SetHelpCommandGroupID(groupID string) {
	if c.helpCommand != nil {
		c.helpCommand.SetGroupID(groupID)
	}
	// helpCommandGroupID is used if no helpCommand is defined by the user
	c.helpCommandGroupID = groupID
}

// SetCompletionCommandGroupID sets the group id of the completion command.
func (c *Default) SetCompletionCommandGroupID(groupID string) {
	// completionCommandGroupID is used if no completion command is defined by the user
	Base(c).SetCompletionCommandGroupID(groupID)
}

// SetVersionTemplate sets version template to be used. Application can use it to set custom template.
func (c *Default) SetVersionTemplate(s string) {
	c.versionTemplate = s
}

// UsagePadding return padding for the usage.
func (c *Default) UsagePadding() int {
	if c.parent == nil || minUsagePadding > c.parent.GetCommandsMaxUseLen() {
		return minUsagePadding
	}
	return c.parent.GetCommandsMaxUseLen()
}

// CommandPathPadding return padding for the command path.
func (c *Default) CommandPathPadding() int {
	if c.parent == nil || minCommandPathPadding > c.parent.GetCommandsMaxCommandPathLen() {
		return minCommandPathPadding
	}
	return c.parent.GetCommandsMaxCommandPathLen()
}

// ErrPrefix return error message prefix for the command
func (c *Default) ErrPrefix() string {
	if c.errPrefix != "" {
		return c.errPrefix
	}
	//todo: wait do
	if HasParent(c) {
		return c.parent.ErrPrefix()
	}
	return "Error:"
}

// ResetCommands delete parent, subcommand and help command from c.
func (c *Default) ResetCommands() {
	c.parent = nil
	c.commands = nil
	c.helpCommand = nil
	c.parentsPFlags = nil
}

// Commands returns a sorted slice of child commands.
func (c *Default) Commands() []Commander {
	// do not sort commands if it already sorted or sorting was disabled
	if EnableCommandSorting && !c.commandsAreSorted {
		sort.Sort(commandSorterByName(c.commands))
		c.commandsAreSorted = true
	}
	return c.commands
}

// Add adds one or more commands to this parent command.
func (c *Default) Add(cmds ...Commander) {
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

func (c *Default) ResetAdd(cmds ...Commander) {
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

// Groups returns a slice of child command groups.
func (c *Default) Groups() []*Group {
	return c.commandGroups
}

// ContainsGroup return if groupID exists in the list of command groups.
func (c *Default) ContainsGroup(groupID string) bool {
	for _, x := range c.commandGroups {
		if x.ID == groupID {
			return true
		}
	}
	return false
}

// AddGroup adds one or more command groups to this parent command.
func (c *Default) AddGroup(groups ...*Group) {
	c.commandGroups = append(c.commandGroups, groups...)
}
func (s *Simple) GetCommandGroups() []*Group { return s.commandGroups }

// Runnable determines if the command is itself runnable.
func (c *Default) Runnable() bool {
	return true
}

// GlobalNormalizationFunc returns the global normalization function or nil if it doesn't exist.
func (c *Default) GlobalNormalizationFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
}

// Parent returns a commands parent command.
func (c *Default) Parent() Commander {
	return c.parent
}

//////// new add ////////////////

func (p *Default) SetParent(c Commander) {
	p.parent = c
}

func (p *Default) GetGroupID() string        { return "" }
func (p *Default) SetGroupID(groupID string) {}

func (p *Default) GetFlags() *flag.FlagSet {
	return p.flags
}

func (p *Default) SetFlags(f *flag.FlagSet) {
	p.flags = f
}

func (p *Default) GetHelpCommand() Commander {
	return p.helpCommand
}

func (p *Default) GetShort() string {
	return ""
}

func (p *Default) GetPersistentPostRunE() func(cmd Commander, args []string) error {
	return nil
}

func (p *Default) GetPersistentPostRun() func(cmd Commander, args []string) {
	return nil
}

func (p *Default) GetSilenceErrors() bool { return false }

func (p *Default) GetSilenceUsage() bool {
	return false
}

func (p *Default) GetCommandCalledAs() *CommandCalledAs {
	return &p.commandCalledAs
}

func (p *Default) GetPersistentPreRunE() func(cmd Commander, args []string) error {
	return nil
}

func (p *Default) GetPersistentPreRun() func(cmd Commander, args []string) {
	return nil
}

func (p *Default) GetSuggestFor() []string {
	return nil
}

func (p *Default) GetArgs() PositionalArgs { return nil }

func (p *Default) GetCommandsMaxUseLen() int {
	return p.commandsMaxUseLen
}

func (p *Default) GetCommandsMaxCommandPathLen() int {
	return p.commandsMaxCommandPathLen
}

func (p *Default) GetCommandsMaxNameLen() int {
	return p.commandsMaxNameLen
}

func (p *Default) GetFlagErrorFunc() func(Commander, error) error {
	return p.flagErrorFunc
}

func (p *Default) GetTraverseChildren() bool {
	return false
}

func (p *Default) GetDisableFlagParsing() bool              { return false }
func (p *Default) GetArgAliases() []string                  { return nil }
func (p *Default) GetValidArgs() []string                   { return nil }
func (p *Default) GetAliases() []string                     { return nil }
func (p *Default) GetHidden() bool                          { return false }
func (p *Default) GetLong() string                          { return "" }
func (p *Default) GetDisableAutoGenTag() bool               { return false }
func (p *Default) SetDisableAutoGenTag(d bool)              {}
func (p *Default) GetExample() string                       { return "" }
func (p *Default) GetCommands() []Commander                 { return p.commands }
func (p *Default) PreRun(args []string) error               { return nil }
func (p *Default) Run(args []string) error                  { return nil } // todo: 这个考虑不默认实现
func (p *Default) PostRun(args []string) error              { return nil }
func (p *Default) getHelpCommandGroupID() string            { return p.helpCommandGroupID }
func (p *Default) GetVersion() string                       { return "" }
func (p *Default) GetDeprecated() string                    { return "" }
func (p *Default) GetDisableFlagsInUseLine() bool           { return false }
func (p *Default) GetDisableSuggestions() bool              { return false }
func (p *Default) GetUse() string                           { return "" } // todo: 这个考虑不默认实现
func (p *Default) GetAnnotations() map[string]string        { return nil }
func (p *Default) GetCommandGroups() []*Group               { return nil }
func (p *Default) GetCompletionOptions() *CompletionOptions { return nil }
func (p *Default) GetSuggestionsMinimumDistance() int       { return 1 }
func (p *Default) SetSuggestionsMinimumDistance(v int)      {}
func (p *Default) GetCompletionCommandGroupID() string      { return p.completionCommandGroupID }
func (p *Default) SetFlagErrorBuf(b *bytes.Buffer)          { p.flagErrorBuf = b }
func (p *Default) GetFlagErrorBuf() *bytes.Buffer           { return p.flagErrorBuf }
func (p *Default) GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
	return nil
}

// UsageString returns usage string.
func (c *Default) UsageString() string {
	// Storing normal writers
	tmpOutput := log.outWriter
	tmpErr := log.errWriter

	bb := new(bytes.Buffer)
	log.outWriter = bb
	log.errWriter = bb

	mergePersistentFlags(c)
	err := tmpl(log.OutOrStderr(), UsageTemplate(c), c)
	if err != nil {
		log.PrintErrLn(err)
	}
	CheckErr(err)

	// todo: 这里的IO还需要理一理

	// Setting things back to normal
	log.outWriter = tmpOutput
	log.errWriter = tmpErr

	return bb.String()
}

type Simple struct {
	Default

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

	// Long is the long message shown in the 'help <this-command>' output.
	// Long string
	Long string

	// Short is the short description shown in the 'help' output.
	Short string

	// DisableFlagsInUseLine will disable the addition of [flags] to the usage
	// line of a command when printing help or generating docs
	DisableFlagsInUseLine bool

	// Hidden defines, if this command is hidden and should NOT show up in the list of available commands.
	Hidden bool

	// Deprecated defines, if this command is deprecated and should print this string when used.
	Deprecated string

	// Version defines the version for this command. If this value is non-empty and the command does not
	// define a "version" flag, a "version" boolean flag will be added to the command and, if specified,
	// will print content of the "Version" variable. A shorthand "v" flag will also be added if the
	// command does not define one.
	Version string

	// Aliases is an array of aliases that can be used instead of the first word in Use.
	Aliases []string

	// SuggestFor is an array of command names for which this command will be suggested -
	// similar to aliases but only suggests.
	SuggestFor []string

	// Example is examples of how to use the command.
	Example string

	// TraverseChildren parses flags on all parents before executing child command.
	TraverseChildren bool

	// SilenceErrors is an option to quiet errors down stream.
	SilenceErrors bool

	// The group id under which this subcommand is grouped in the 'help' output of its parent.
	GroupID string

	// ValidArgs is list of all valid non-flag arguments that are accepted in shell completions
	ValidArgs []string

	// SuggestionsMinimumDistance defines minimum levenshtein distance to display suggestions.
	// Must be > 0.
	SuggestionsMinimumDistance int

	// Expected arguments
	Args PositionalArgs

	// ArgAliases is List of aliases for ValidArgs.
	// These are not suggested to the user in the shell completion,
	// but accepted if entered manually.
	ArgAliases []string

	// SilenceUsage is an option to silence usage when an error occurs.
	SilenceUsage bool

	// DisableFlagParsing disables the flag parsing.
	// If this is true all flags will be passed to the command as arguments.
	DisableFlagParsing bool

	// DisableAutoGenTag defines, if gen tag ("Auto generated by spf13/cobra...")
	// will be printed by generating docs for this command.
	DisableAutoGenTag bool

	// DisableSuggestions disables the suggestions based on Levenshtein distance
	// that go along with 'unknown command' messages.
	DisableSuggestions bool

	// ValidArgsFunction is an optional function that provides valid non-flag arguments for shell completion.
	// It is a dynamic version of using ValidArgs.
	// Only one of ValidArgs and ValidArgsFunction can be used for a command.
	ValidArgsFunction func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective)

	// BashCompletionFunction is custom bash functions used by the legacy bash autocompletion generator.
	// For portability with other shells, it is recommended to instead use ValidArgsFunction
	BashCompletionFunction string

	// Annotations are key/value pairs that can be used by applications to identify or
	// group commands or set special options.
	Annotations map[string]string

	// FParseErrWhitelist flag parse errors to be ignored
	FParseErrWhitelist FParseErrWhitelist

	// CompletionOptions is a set of options to control the handling of shell completion
	CompletionOptions CompletionOptions

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
}

func (s *Simple) GetUse() string {
	return s.Use
}

func (s *Simple) GetLong() string {
	return s.Long
}

func (s *Simple) GetHidden() bool {
	return s.Hidden
}

func (s *Simple) GetDeprecated() string {
	return s.Deprecated
}

func (s *Simple) GetVersion() string {
	return s.Version
}

func (s *Simple) GetAliases() []string {
	return s.Aliases
}

func (s *Simple) GetPersistentPostRunE() func(cmd Commander, args []string) error {
	return s.PersistentPostRunE
}

func (s *Simple) GetPersistentPostRun() func(cmd Commander, args []string) {
	return s.PersistentPostRun
}

func (s *Simple) PreRun(args []string) error {
	if s.PreRunE != nil {
		return s.PreRunE(s, args)
	}
	return nil
}

func (s *Simple) Run(args []string) error {
	if s.RunE != nil {
		return s.RunE(s, args)
	}
	return nil
}

func (s *Simple) GetPersistentPreRunE() func(cmd Commander, args []string) error {
	return s.PersistentPreRunE
}

func (s *Simple) GetPersistentPreRun() func(cmd Commander, args []string) {
	return s.PersistentPreRun
}

func (s *Simple) GetSuggestFor() []string {
	return s.SuggestFor
}

func (s *Simple) GetExample() string {
	return s.Example
}

func (s *Simple) GetTraverseChildren() bool {
	return s.TraverseChildren
}

func (s *Simple) GetSilenceErrors() bool {
	return s.SilenceErrors
}

func (s *Simple) GetGroupID() string {
	return s.GroupID
}

func (s *Simple) SetGroupID(groupID string) {
	s.GroupID = groupID
}

func (s *Simple) GetValidArgs() []string {
	return s.ValidArgs
}

func (s *Simple) GetSuggestionsMinimumDistance() int  { return s.SuggestionsMinimumDistance }
func (s *Simple) SetSuggestionsMinimumDistance(v int) { s.SuggestionsMinimumDistance = v }
func (s *Simple) GetArgs() PositionalArgs             { return s.Args }

func (s *Simple) GetArgAliases() []string {
	return s.ArgAliases
}

func (s *Simple) GetSilenceUsage() bool {
	return s.SilenceUsage
}

func (s *Simple) GetDisableAutoGenTag() bool               { return s.DisableAutoGenTag }
func (s *Simple) SetDisableAutoGenTag(d bool)              { s.DisableAutoGenTag = d }
func (s *Simple) GetDisableFlagParsing() bool              { return s.DisableFlagParsing }
func (s *Simple) GetDisableFlagsInUseLine() bool           { return s.DisableFlagsInUseLine }
func (s *Simple) GetDisableSuggestions() bool              { return s.DisableSuggestions }
func (s *Simple) GetAnnotations() map[string]string        { return s.Annotations }
func (s *Simple) GetCompletionOptions() *CompletionOptions { return &s.CompletionOptions }

// GetFParseErrWhitelist implements Commander.
func (s *Simple) GetFParseErrWhitelist() FParseErrWhitelist {
	return s.FParseErrWhitelist
}

// SetFParseErrWhitelist implements Commander.
func (s *Simple) SetFParseErrWhitelist(fp FParseErrWhitelist) {
	s.FParseErrWhitelist = fp
}
