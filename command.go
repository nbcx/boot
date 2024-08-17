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
	flag "github.com/nbcx/flag"
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

// Command is just that, a command for your application.
// E.g.  'go run ...' - 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Command struct {
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

	// FParseErrWhitelist flag parse errors to be ignored
	FParseErrWhitelist FParseErrWhitelist

	// CompletionOptions is a set of options to control the handling of shell completion
	CompletionOptions CompletionOptions

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
func (c *Command) GetFParseErrWhitelist() FParseErrWhitelist {
	return c.FParseErrWhitelist
}

// SetFParseErrWhitelist implements Commander.
func (c *Command) SetFParseErrWhitelist(fp FParseErrWhitelist) {
	c.FParseErrWhitelist = fp
}

// GetGlobNormFunc implements Commander.
func (c *Command) GetGlobNormFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
}

// SetGlobNormFunc implements Commander.
func (c *Command) SetGlobNormFunc(f func(f *flag.FlagSet, name string) flag.NormalizedName) {
	c.globNormFunc = f
}

// SetVersionTemplate sets version template to be used. Application can use it to set custom template.
func (c *Command) SetVersionTemplate(s string) {
	c.versionTemplate = s
}

// Runnable determines if the command is itself runnable.
func (c *Command) Runnable() bool {
	return true
}

// Parent returns a commands parent command.
func (c *Command) Parent() Commander {
	return c.parent
}

func (c *Command) GetUse() string {
	return c.Use
}

func (c *Command) GetGroupID() string {
	return c.GroupID
}

func (c *Command) SetGroupID(groupID string) {
	c.GroupID = groupID
}

func (c *Command) GetShort() string {
	return c.Short
}

func (c *Command) GetPersistentPostRunE() func(cmd Commander, args []string) error {
	return c.PersistentPostRunE
}

func (c *Command) GetPersistentPostRun() func(cmd Commander, args []string) {
	return c.PersistentPostRun
}

func (c *Command) GetSilenceErrors() bool {
	return c.SilenceErrors
}

func (c *Command) GetSilenceUsage() bool {
	return c.SilenceUsage
}

func (c *Command) GetSuggestFor() []string {
	return c.SuggestFor
}

func (c *Command) GetPositionalArgs() PositionalArgs {
	return c.Args
}

func (c *Command) GetTraverseChildren() bool {
	return c.TraverseChildren
}

func (c *Command) GetDisableFlagParsing() bool {
	return c.DisableFlagParsing
}

func (c *Command) GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
	return c.ValidArgsFunction
}

func (c *Command) GetArgAliases() []string {
	return c.ArgAliases
}

func (c *Command) GetValidArgs() []string {
	return c.ValidArgs
}

func (c *Command) GetAliases() []string {
	return c.Aliases
}
func (c *Command) GetHidden() bool {
	return c.Hidden
}

func (c *Command) GetLong() string {
	return c.Long
}

func (c *Command) GetDisableAutoGenTag() bool {
	return c.DisableAutoGenTag
}

func (c *Command) SetDisableAutoGenTag(d bool) {
	c.DisableAutoGenTag = d
}

func (c *Command) GetExample() string {
	return c.Example
}

func (c *Command) PreExec(args []string) error {
	if c.PreRunE != nil {
		return c.PreRunE(c, args)
	}
	return nil
}

func (c *Command) Exec(args []string) error {
	if c.RunE != nil {
		return c.RunE(c, args)
	}
	return nil
}
func (c *Command) PostExec(args []string) error {
	return nil
}

func (c *Command) GetVersion() string {
	return c.Version
}

func (c *Command) GetDeprecated() string {
	return c.Deprecated
}

func (c *Command) GetDisableFlagsInUseLine() bool {
	return c.DisableFlagsInUseLine
}

func (c *Command) GetAnnotations() map[string]string {
	return c.Annotations
}

func (c *Command) GetDisableSuggestions() bool {
	return c.DisableSuggestions
}

func (c *Command) GetCompletionOptions() *CompletionOptions {
	return &c.CompletionOptions
}

func (c *Command) Add(v ...Commander) {
	Bind(c, v...)
}

func (c *Command) Execute() error {
	return Execute(c)
}
