package boot

import (
	"context"
	"io"

	flag "github.com/spf13/pflag"
)

type Commander interface {
	SetParent(Commander)
	GetParent() Commander
	Parent() Commander // note: 同上，后续移除
	GetUse() string
	GetGroupID() string
	SetGroupID(groupID string)
	GetFlags() *flag.FlagSet
	GetHelpCommand() Commander
	GetShort() string

	GetSilenceErrors() bool
	GetCommandCalledAs() *CommandCalledAs
	GetSilenceUsage() bool
	GetHelpFunc() func(Commander, []string)
	GetValidArgs() []string
	GetHidden() bool
	GetLong() string
	GetExample() string
	GetCommands() []Commander

	// run
	GetPersistentPreRunE() func(cmd Commander, args []string) error
	GetPersistentPreRun() func(cmd Commander, args []string)

	Run(args []string) error // Typically the actual work function. Most commands will only implement this.
	PreRun(args []string) error
	PostRun(args []string) error

	Context() context.Context
	SetContext(ctx context.Context) // note: 同上，后续移除
	GetPersistentPostRunE() func(cmd Commander, args []string) error
	GetPersistentPostRun() func(cmd Commander, args []string)
	ErrPrefix() string
	GetArgs() PositionalArgs
	GetCommandsMaxUseLen() int
	GetCommandsMaxCommandPathLen() int
	GetCommandsMaxNameLen() int
	GetFlagErrorFunc() func(Commander, error) error
	OutOrStdout() io.Writer
	GenPowerShellCompletionWithDesc(w io.Writer) error
	GenFishCompletion(w io.Writer, includeDesc bool) error
	GenBashCompletionV2(w io.Writer, includeDesc bool) error
	GetTraverseChildren() bool
	RemoveCommand(cmds ...Commander)
	GetDisableFlagParsing() bool
	LocalNonPersistentFlags() *flag.FlagSet
	GetAliases() []string
	// UseLine() string
	GetDisableAutoGenTag() bool
	SetDisableAutoGenTag(d bool)
	GetVersion() string
	GetAnnotations() map[string]string

	// CommandPath() string
	// Name() string
	SetGlobalNormalizationFunc(n func(f *flag.FlagSet, name string) flag.NormalizedName)
	// IsAvailableCommand() bool
	HasFlags() bool
	HasPersistentFlags() bool
	IsAdditionalHelpTopicCommand() bool
	Base() Commander
	Usage() error
	// Find(args []string) (Commander, []string, error)
	Commands() []Commander
	// CheckCommandGroups()
	// InitDefaultHelpFlag()
	// InitDefaultVersionFlag()
	Help() error
	// ExecuteC() (cmd Commander, err error)
	HelpFunc() func(Commander, []string)
	UsageString() string
	// Execute(a []string) (err error)
	// VisitParents(fn func(Commander))
	GetSuggestFor() []string
	HasAlias(s string) bool

	VersionTemplate() string
	HelpTemplate() string
	UsageTemplate() string

	FlagErrorFunc() (f func(Commander, error) error)
	// Traverse(args []string) (Commander, []string, error)
	UsageFunc() (f func(Commander) error)
	OutOrStderr() io.Writer
	PersistentFlags() *flag.FlagSet
	CalledAs() string
	GenPowerShellCompletion(w io.Writer) error
	GenZshCompletionNoDesc(w io.Writer) error
	GenZshCompletion(w io.Writer) error
	GetSuggestionsMinimumDistance() int
	Add(cmds ...Commander)
	GetCompletionOptions() *CompletionOptions
	GetCompletionCommandGroupID() string
	SetCompletionCommandGroupID(v string)

	ErrOrStderr() io.Writer
	ParseFlags(args []string) error
	GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective)
	GetArgAliases() []string
	InheritedFlags() *flag.FlagSet
	NonInheritedFlags() *flag.FlagSet

	Flag(name string) (flag *flag.Flag)
	HasSubCommands() bool
	HasParent() bool

	SetArgs(a []string)
	SetErr(newErr io.Writer)
	SetOut(newOut io.Writer)
	SetOutput(output io.Writer)
	// ExecuteContextC(ctx context.Context) (Commander, error)
	// MarkFlagsRequiredTogether(flagNames ...string)
	// MarkFlagsOneRequired(flagNames ...string)
	// MarkFlagsMutuallyExclusive(flagNames ...string)
	// InitDefaultHelpCmd()
	Runnable() bool
	// ValidateArgs(args []string) error
	GetDeprecated() string
	GetDisableFlagsInUseLine() bool
	// GetUse() string
	GetCommandGroups() []*Group

	enforceFlagGroupsForCompletion()
	// findSuggestions(arg string) string
	getOut(def io.Writer) io.Writer
	getErr(def io.Writer) io.Writer
	getIn(def io.Reader) io.Reader
	// hasNameOrAliasPrefix(prefix string) bool
	// findNext(next string) Commander
	argsMinusFirstX(args []string, x string) []string
	mergePersistentFlags()
	getCompletions(args []string) (Commander, []string, ShellCompDirective, error)
	// displayName() string
	getHelpCommandGroupID() string
	GetDisableSuggestions() bool
	SetSuggestionsMinimumDistance(v int)

	Print(i ...interface{})
	Println(i ...interface{})
	Printf(format string, i ...interface{})
	PrintErr(i ...interface{})
	PrintErrLn(i ...interface{})
	PrintErrF(format string, i ...interface{})
}

func (p *Root) GetParent() Commander {
	return p.parent
}

func (p *Root) SetParent(c Commander) {
	p.parent = c
}

func (p *Root) GetUse() string {
	return p.Use
}

func (p *Root) GetGroupID() string {
	return p.GroupID
}

func (p *Root) SetGroupID(groupID string) {
	p.GroupID = groupID
}

func (p *Root) GetFlags() *flag.FlagSet {
	return p.pflags
}

func (p *Root) GetHelpCommand() Commander {
	return p.helpCommand
}

func (p *Root) GetShort() string {
	return p.Short
}

func (p *Root) GetPersistentPostRunE() func(cmd Commander, args []string) error {
	return p.PersistentPostRunE
}

func (p *Root) GetPersistentPostRun() func(cmd Commander, args []string) {
	return p.PersistentPostRun
}

func (p *Root) GetSilenceErrors() bool {
	return p.SilenceErrors
}

func (p *Root) GetSilenceUsage() bool {
	return p.SilenceUsage
}

func (p *Root) GetCommandCalledAs() *CommandCalledAs {
	return &p.commandCalledAs
}

func (p *Root) GetPersistentPreRunE() func(cmd Commander, args []string) error {
	return p.PersistentPreRunE
}

func (p *Root) GetPersistentPreRun() func(cmd Commander, args []string) {
	return p.PersistentPreRun
}

func (p *Root) GetSuggestFor() []string {
	return p.SuggestFor
}

func (p *Root) GetArgs() PositionalArgs {
	return p.Args
}

func (p *Root) GetCommandsMaxUseLen() int {
	return p.commandsMaxUseLen
}
func (p *Root) GetCommandsMaxCommandPathLen() int {
	return p.commandsMaxCommandPathLen
}
func (p *Root) GetCommandsMaxNameLen() int {
	return p.commandsMaxNameLen
}
func (p *Root) GetHelpFunc() func(Commander, []string) {
	return p.helpFunc
}

func (p *Root) GetFlagErrorFunc() func(Commander, error) error {
	return p.flagErrorFunc
}

func (p *Root) GetTraverseChildren() bool {
	return p.TraverseChildren
}

func (p *Root) GetDisableFlagParsing() bool {
	return p.DisableFlagParsing
}

func (p *Root) GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
	return p.ValidArgsFunction
}

func (p *Root) GetArgAliases() []string {
	return p.ArgAliases
}

func (p *Root) GetValidArgs() []string {
	return p.ValidArgs
}

func (p *Root) GetAliases() []string {
	return p.Aliases
}
func (p *Root) GetHidden() bool {
	return p.Hidden
}

func (p *Root) GetLong() string {
	return p.Long
}

func (p *Root) GetDisableAutoGenTag() bool {
	return p.DisableAutoGenTag
}

func (p *Root) SetDisableAutoGenTag(d bool) {
	p.DisableAutoGenTag = d
}
func (p *Root) GetExample() string {
	return p.Example
}

func (p *Root) GetCommands() []Commander {
	return p.commands
}

func (p *Root) PreRun(args []string) error {
	if p.PreRunE != nil {
		return p.PreRunE(p, args)
	}
	return nil
}

func (p *Root) Run(args []string) error {
	if p.RunE != nil {
		return p.RunE(p, args)
	}
	return nil
}
func (p *Root) PostRun(args []string) error {
	return nil
}

func (p *Root) getHelpCommandGroupID() string {
	return p.helpCommandGroupID
}

func (p *Root) GetVersion() string {
	return p.Version
}

func (p *Root) GetDeprecated() string {
	return p.Deprecated
}

func (p *Root) GetDisableFlagsInUseLine() bool {
	return p.DisableFlagsInUseLine
}

func (p *Root) GetAnnotations() map[string]string {
	return p.Annotations
}

func (p *Root) GetCommandGroups() []*Group {
	return p.commandgroups
}
func (p *Root) GetDisableSuggestions() bool {
	return p.DisableSuggestions
}

func (p *Root) GetSuggestionsMinimumDistance() int {
	return p.SuggestionsMinimumDistance
}

func (p *Root) SetSuggestionsMinimumDistance(v int) {
	p.SuggestionsMinimumDistance = v
}

func (p *Root) GetCompletionOptions() *CompletionOptions {
	return &p.CompletionOptions
}

func (p *Root) GetCompletionCommandGroupID() string {
	return p.completionCommandGroupID
}

// SuggestFor []string
func (p *Root) test() {
	p.GetUse()
}
