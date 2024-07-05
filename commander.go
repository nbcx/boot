package cobra

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
	Ctx() context.Context
	SetCtx(context.Context)
	SetContext(ctx context.Context) // note: 同上，后续移除
	GetSilenceErrors() bool
	GetCommandCalledAs() *CommandCalledAs
	GetSilenceUsage() bool
	GetPersistentPreRunE() func(cmd Commander, args []string) error
	GetPersistentPreRun() func(cmd Commander, args []string)
	GetHelpFunc() func(Commander, []string)
	GetValidArgs() []string
	GetHidden() bool
	GetLong() string
	GetExample() string
	GetCommands() []Commander

	GetPersistentPostRunE() func(cmd Commander, args []string) error
	GetPersistentPostRun() func(cmd Commander, args []string)
	ErrPrefix() string
	GetArgs() PositionalArgs
	GetCommandsMaxUseLen() int
	GetCommandsMaxCommandPathLen() int
	GetCommandsMaxNameLen() int
	GetFlagErrorFunc() func(Commander, error) error
	OutOrStdout() io.Writer
	SetCompletionCommandGroupID(id string)
	GenPowerShellCompletionWithDesc(w io.Writer) error
	GenFishCompletion(w io.Writer, includeDesc bool) error
	GenBashCompletionV2(w io.Writer, includeDesc bool) error
	GetTraverseChildren() bool
	RemoveCommand(cmds ...Commander)
	GetDisableFlagParsing() bool
	LocalNonPersistentFlags() *flag.FlagSet
	GetAliases() []string
	UseLine() string
	GetDisableAutoGenTag() bool
	SetDisableAutoGenTag(d bool)

	CommandPath() string
	Name() string
	SetGlobalNormalizationFunc(n func(f *flag.FlagSet, name string) flag.NormalizedName)
	IsAvailableCommand() bool
	HasFlags() bool
	HasPersistentFlags() bool
	IsAdditionalHelpTopicCommand() bool
	Root() Commander
	Usage() error
	Find(args []string) (Commander, []string, error)
	Commands() []Commander
	CheckCommandGroups()
	InitDefaultHelpFlag()
	InitDefaultVersionFlag()
	Help() error
	ExecuteC() (cmd Commander, err error)
	HelpFunc() func(Commander, []string)
	UsageString() string
	Execute(a []string) (err error)
	VisitParents(fn func(Commander))
	GetSuggestFor() []string
	HasAlias(s string) bool

	VersionTemplate() string
	HelpTemplate() string
	UsageTemplate() string

	FlagErrorFunc() (f func(Commander, error) error)
	Traverse(args []string) (Commander, []string, error)
	UsageFunc() (f func(Commander) error)
	OutOrStderr() io.Writer
	PersistentFlags() *flag.FlagSet
	CalledAs() string
	GenPowerShellCompletion(w io.Writer) error
	GenZshCompletionNoDesc(w io.Writer) error
	GenZshCompletion(w io.Writer) error

	ErrOrStderr() io.Writer
	ParseFlags(args []string) error
	GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective)
	GetArgAliases() []string
	InheritedFlags() *flag.FlagSet
	NonInheritedFlags() *flag.FlagSet
	enforceFlagGroupsForCompletion()
	Flag(name string) (flag *flag.Flag)
	HasSubCommands() bool
	HasParent() bool
	findSuggestions(arg string) string
	SetArgs(a []string)
	SetErr(newErr io.Writer)
	SetOut(newOut io.Writer)
	SetOutput(output io.Writer)
	ExecuteContextC(ctx context.Context) (Commander, error)
	Context() context.Context
	MarkFlagsRequiredTogether(flagNames ...string)
	MarkFlagsOneRequired(flagNames ...string)
	MarkFlagsMutuallyExclusive(flagNames ...string)
	InitDefaultHelpCmd()
	Runnable() bool

	getOut(def io.Writer) io.Writer
	getErr(def io.Writer) io.Writer
	getIn(def io.Reader) io.Reader
	hasNameOrAliasPrefix(prefix string) bool
	findNext(next string) Commander
	argsMinusFirstX(args []string, x string) []string
	mergePersistentFlags()
	getCompletions(args []string) (Commander, []string, ShellCompDirective, error)

	Print(i ...interface{})
	Println(i ...interface{})
	Printf(format string, i ...interface{})
	PrintErr(i ...interface{})
	PrintErrln(i ...interface{})
	PrintErrf(format string, i ...interface{})
}

// type DefaultCommand struct {
// 	Command
// }

func (p *Command) GetParent() Commander {
	return p.parent
}

func (p *Command) SetParent(c Commander) {
	p.parent = c
}

func (p *Command) GetUse() string {
	return p.Use
}

func (p *Command) GetGroupID() string {
	return p.GroupID
}

func (p *Command) SetGroupID(groupID string) {
	p.GroupID = groupID
}

func (p *Command) GetFlags() *flag.FlagSet {
	return p.pflags
}

func (p *Command) GetHelpCommand() Commander {
	return p.helpCommand
}
func (p *Command) GetShort() string {
	return p.Short
}

func (p *Command) GetPersistentPostRunE() func(cmd Commander, args []string) error {
	return p.PersistentPostRunE
}

func (p *Command) GetPersistentPostRun() func(cmd Commander, args []string) {
	return p.PersistentPostRun
}

func (p *Command) Ctx() context.Context {
	return p.ctx
}

func (p *Command) SetCtx(ctx context.Context) {
	p.ctx = ctx
}

func (p *Command) GetSilenceErrors() bool {
	return p.SilenceErrors
}

func (p *Command) GetSilenceUsage() bool {
	return p.SilenceUsage
}

func (p *Command) GetCommandCalledAs() *CommandCalledAs {
	return &p.commandCalledAs
}

func (p *Command) GetPersistentPreRunE() func(cmd Commander, args []string) error {
	return p.PersistentPreRunE
}

func (p *Command) GetPersistentPreRun() func(cmd Commander, args []string) {
	return p.PersistentPreRun
}

func (p *Command) GetSuggestFor() []string {
	return p.SuggestFor
}

func (p *Command) GetArgs() PositionalArgs {
	return p.Args
}

func (p *Command) GetCommandsMaxUseLen() int {
	return p.commandsMaxUseLen
}
func (p *Command) GetCommandsMaxCommandPathLen() int {
	return p.commandsMaxCommandPathLen
}
func (p *Command) GetCommandsMaxNameLen() int {
	return p.commandsMaxNameLen
}
func (p *Command) GetHelpFunc() func(Commander, []string) {
	return p.helpFunc
}

func (p *Command) GetFlagErrorFunc() func(Commander, error) error {
	return p.flagErrorFunc
}

func (p *Command) GetTraverseChildren() bool {
	return p.TraverseChildren
}

func (p *Command) GetDisableFlagParsing() bool {
	return p.DisableFlagParsing
}

func (p *Command) GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
	return p.ValidArgsFunction
}

func (p *Command) GetArgAliases() []string {
	return p.ArgAliases
}

func (p *Command) GetValidArgs() []string {
	return p.ValidArgs
}

func (p *Command) GetAliases() []string {
	return p.Aliases
}
func (p *Command) GetHidden() bool {
	return p.Hidden
}

func (p *Command) GetLong() string {
	return p.Long
}

func (p *Command) GetDisableAutoGenTag() bool {
	return p.DisableAutoGenTag
}

func (p *Command) SetDisableAutoGenTag(d bool) {
	p.DisableAutoGenTag = d
}
func (p *Command) GetExample() string {
	return p.Example
}

func (p *Command) GetCommands() []Commander {
	return nil
}

//	func (p *Command) SetCompletionCommandGroupID(id string) {
//		p.completionCommandGroupID = id
//	}
//
// SuggestFor []string
func (p *Command) test() {
	// p.Runnable
}
