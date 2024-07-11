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

// const (
// 	FlagSetByCobraAnnotation     = "cobra_annotation_flag_set_by_cobra"
// 	CommandDisplayNameAnnotation = "cobra_annotation_command_display_name"
// )

// FParseErrWhitelist configures Flag parse errors to be ignored
// type FParseErrWhitelist flag.ParseErrorsWhitelist

// Group Structure to manage groups for commands
// type Group struct {
// 	ID    string
// 	Title string
// }

// type CommandCalledAs struct {
// 	name   string
// 	called bool
// }

func ParseName(c Commander) string {
	return name(c)
}

// Root is just that, a command for your application.
// E.g.  'go run ...' - 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Default struct {
	// Use is the one-line usage message.
	// Recommended syntax is as follows:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	// Use string

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
func (c *Default) GetFParseErrWhitelist() FParseErrWhitelist {
	return c.FParseErrWhitelist
}

// GetGlobNormFunc implements Commander.
func (c *Default) GetGlobNormFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
}

// SetFParseErrWhitelist implements Commander.
func (c *Default) SetFParseErrWhitelist(fp FParseErrWhitelist) {
	c.FParseErrWhitelist = fp
}

// SetGlobNormFunc implements Commander.
func (c *Default) SetGlobNormFunc(f func(f *flag.FlagSet, name string) flag.NormalizedName) {
	c.globNormFunc = f
}

// GetPFlags implements Commander.
func (c *Default) GetPFlags() *flag.FlagSet {
	return c.pFlags
}

// GetIFlags implements Commander.
func (c *Default) GetIFlags() *flag.FlagSet {
	return c.iFlags
}

// GetLFlags implements Commander.
func (c *Default) GetLFlags() *flag.FlagSet {
	return c.lFlags
}

// GetParentsPFlags implements Commander.
func (c *Default) GetParentsPFlags() *flag.FlagSet {
	return c.parentsPFlags
}

// SetIFlags implements Commander.
func (c *Default) SetIFlags(i *flag.FlagSet) {
	c.iFlags = i
}

// SetLFlags implements Commander.
func (c *Default) SetLFlags(l *flag.FlagSet) {
	c.lFlags = l
}

// SetPFlags implements Commander.
func (c *Default) SetPFlags(l *flag.FlagSet) {
	c.pFlags = l
}

// SetParentsPFlags implements Commander.
func (c *Default) SetParentsPFlags(pf *flag.FlagSet) {
	c.parentsPFlags = pf
}

// Context returns underlying command context. If command was executed
// with ExecuteContext or the context was set with SetContext, the
// previously set context will be returned. Otherwise, nil is returned.
//
// Notice that a call to Execute and ExecuteC will replace a nil context of
// a command with a context.Background, so a background context will be
// returned by Context after one of these functions has been called.
func (c *Default) Context() context.Context {
	return c.ctx
}

// SetContext sets context for the command. This context will be overwritten by
// Command.ExecuteContext or Command.ExecuteContextC.
func (c *Default) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetArgs sets arguments for the command. It is set to os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (c *Default) SetArgs(a []string) {
	c.args = a
}

// SetUsageFunc sets usage function. Usage can be defined by application.
// func (c *Default) SetUsageFunc(f func(Commander) error) {
// 	c.usageFunc = f
// }

// SetUsageTemplate sets usage template. Can be defined by Application.
// func (c *Default) SetUsageTemplate(s string) {
// 	c.usageTemplate = s
// }

// SetFlagErrorFunc sets a function to generate an error when flag parsing
// fails.
func (c *Default) SetFlagErrorFunc(f func(Commander, error) error) {
	c.flagErrorFunc = f
}

// SetHelpFunc sets help function. Can be defined by Application.
// func (c *Default) SetHelpFunc(f func(Commander, []string)) {
// 	c.helpFunc = f
// }

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

// SetHelpTemplate sets help template to be used. Application can use it to set custom template.
func (c *Default) SetHelpTemplate(s string) {
	c.helpTemplate = s
}

// SetVersionTemplate sets version template to be used. Application can use it to set custom template.
func (c *Default) SetVersionTemplate(s string) {
	c.versionTemplate = s
}

// SetErrPrefix sets error message prefix to be used. Application can use it to set custom prefix.
func (c *Default) SetErrPrefix(s string) {
	c.errPrefix = s
}

// SetGlobalNormalizationFunc sets a normalization function to all flag sets and also to child commands.
// The user should not have a cyclic dependency on commands.
// func (c *Default) SetGlobalNormalizationFunc(n func(f *flag.FlagSet, name string) flag.NormalizedName) {
// 	c.Flags().SetNormalizeFunc(n)
// 	c.PersistentFlags().SetNormalizeFunc(n)
// 	c.globNormFunc = n

// 	for _, command := range c.commands {
// 		command.SetGlobalNormalizationFunc(n)
// 	}
// }

// FlagErrorFunc returns either the function set by SetFlagErrorFunc for this
// command or a parent, or it returns a function which returns the original
// error.
func (c *Default) FlagErrorFunc() (f func(Commander, error) error) {
	if c.flagErrorFunc != nil {
		return c.flagErrorFunc
	}

	if c.HasParent() {
		return FlagErrorFunc(c.parent)
	}
	return func(c Commander, err error) error {
		return err
	}
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

	if c.HasParent() {
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

// CalledAs returns the command name or alias that was used to invoke
// this command or an empty string if the command has not been called.
func (c *Default) CalledAs() string {
	if c.commandCalledAs.called {
		return c.commandCalledAs.name
	}
	return ""
}

// HasExample determines if the command has example.
func (c *Default) HasExample() bool {
	return len(c.Example) > 0
}

// Runnable determines if the command is itself runnable.
func (c *Default) Runnable() bool {
	return true
}

// HasSubCommands determines if the command has children commands.
func (c *Default) HasSubCommands() bool {
	return len(c.commands) > 0
}

// HasParent determines if the command is a child command.
func (c *Default) HasParent() bool {
	return c.parent != nil
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

func (p *Default) GetGroupID() string {
	return p.GroupID
}

func (p *Default) SetGroupID(groupID string) {
	p.GroupID = groupID
}

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
	return p.Short
}

func (p *Default) GetPersistentPostRunE() func(cmd Commander, args []string) error {
	return p.PersistentPostRunE
}

func (p *Default) GetPersistentPostRun() func(cmd Commander, args []string) {
	return p.PersistentPostRun
}

func (p *Default) GetSilenceErrors() bool {
	return p.SilenceErrors
}

func (p *Default) GetSilenceUsage() bool {
	return p.SilenceUsage
}

func (p *Default) GetCommandCalledAs() *CommandCalledAs {
	return &p.commandCalledAs
}

func (p *Default) GetPersistentPreRunE() func(cmd Commander, args []string) error {
	return p.PersistentPreRunE
}

func (p *Default) GetPersistentPreRun() func(cmd Commander, args []string) {
	return p.PersistentPreRun
}

func (p *Default) GetSuggestFor() []string {
	return p.SuggestFor
}

func (p *Default) GetArgs() PositionalArgs {
	return p.Args
}

func (p *Default) GetCommandsMaxUseLen() int {
	return p.commandsMaxUseLen
}
func (p *Default) GetCommandsMaxCommandPathLen() int {
	return p.commandsMaxCommandPathLen
}
func (p *Default) GetCommandsMaxNameLen() int {
	return p.commandsMaxNameLen
}
func (p *Default) GetHelpFunc() func(Commander, []string) {
	return p.helpFunc
}

func (p *Default) GetFlagErrorFunc() func(Commander, error) error {
	return p.flagErrorFunc
}

func (p *Default) GetTraverseChildren() bool {
	return p.TraverseChildren
}

func (p *Default) GetDisableFlagParsing() bool {
	return p.DisableFlagParsing
}

func (p *Default) GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
	return p.ValidArgsFunction
}

func (p *Default) GetArgAliases() []string {
	return p.ArgAliases
}

func (p *Default) GetValidArgs() []string {
	return p.ValidArgs
}

func (p *Default) GetAliases() []string {
	return p.Aliases
}
func (p *Default) GetHidden() bool {
	return p.Hidden
}

func (p *Default) GetLong() string {
	return p.Long
}

func (p *Default) GetDisableAutoGenTag() bool {
	return p.DisableAutoGenTag
}

func (p *Default) SetDisableAutoGenTag(d bool) {
	p.DisableAutoGenTag = d
}
func (p *Default) GetExample() string {
	return p.Example
}

func (p *Default) GetCommands() []Commander {
	return p.commands
}

func (p *Default) PreRun(args []string) error {
	if p.PreRunE != nil {
		return p.PreRunE(p, args)
	}
	return nil
}

func (p *Default) Run(args []string) error {
	if p.RunE != nil {
		return p.RunE(p, args)
	}
	return nil
}

func (p *Default) PostRun(args []string) error {
	return nil
}

func (p *Default) getHelpCommandGroupID() string {
	return p.helpCommandGroupID
}

func (p *Default) GetVersion() string {
	return p.Version
}

func (p *Default) GetDeprecated() string {
	return p.Deprecated
}

func (p *Default) GetDisableFlagsInUseLine() bool {
	return p.DisableFlagsInUseLine
}

func (p *Default) GetUse() string {
	return ""
}

func (p *Default) GetAnnotations() map[string]string {
	return p.Annotations
}

func (p *Default) GetCommandGroups() []*Group {
	return p.commandGroups
}

func (p *Default) GetDisableSuggestions() bool {
	return p.DisableSuggestions
}

func (p *Default) GetSuggestionsMinimumDistance() int {
	return p.SuggestionsMinimumDistance
}

func (p *Default) SetSuggestionsMinimumDistance(v int) {
	p.SuggestionsMinimumDistance = v
}

func (p *Default) GetCompletionOptions() *CompletionOptions {
	return &p.CompletionOptions
}

func (p *Default) GetCompletionCommandGroupID() string {
	return p.completionCommandGroupID
}

func (p *Default) SetFlagErrorBuf(b *bytes.Buffer) {
	p.flagErrorBuf = b
}

func (p *Default) GetFlagErrorBuf() *bytes.Buffer {
	return p.flagErrorBuf
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
