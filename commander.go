package boot

import (
	"bytes"
	"context"

	flag "github.com/spf13/pflag"
)

type Commander interface {
	// 参数
	GetUse() string
	GetHelpCommand() Commander
	GetShort() string
	GetSilenceErrors() bool
	GetSilenceUsage() bool
	GetValidArgs() []string
	GetHidden() bool
	GetLong() string
	GetExample() string
	GetCommandCalledAs() *CommandCalledAs

	// run
	GetPersistentPreRunE() func(cmd Commander, args []string) error
	GetPersistentPreRun() func(cmd Commander, args []string)
	Run(args []string) error // Typically the actual work function. Most commands will only implement this.
	PreRun(args []string) error
	PostRun(args []string) error

	Context() context.Context
	SetContext(ctx context.Context)
	GetPersistentPostRunE() func(cmd Commander, args []string) error
	GetPersistentPostRun() func(cmd Commander, args []string)
	ErrPrefix() string
	GetArgs() PositionalArgs
	GetCommandsMaxUseLen() int
	GetCommandsMaxCommandPathLen() int
	GetCommandsMaxNameLen() int
	GetTraverseChildren() bool
	GetDisableFlagParsing() bool

	GetAliases() []string
	GetDisableAutoGenTag() bool
	SetDisableAutoGenTag(d bool)
	GetVersion() string
	GetAnnotations() map[string]string
	GetDisableSuggestions() bool
	GetSuggestionsMinimumDistance() int
	GetDeprecated() string

	// 层级关系
	SetParent(Commander)
	Parent() Commander
	GetGroupID() string
	SetGroupID(groupID string)
	Self() Commander
	SetSelf(Commander)

	// PersistentFlags() *flag.FlagSet
	Add(cmds ...Commander)
	ResetAdd(cmds ...Commander)
	GetCompletionOptions() *CompletionOptions
	GetCompletionCommandGroupID() string
	SetCompletionCommandGroupID(v string)

	GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective)
	GetArgAliases() []string

	Runnable() bool
	GetCommandGroups() []*Group
	getHelpCommandGroupID() string
	Commands() []Commander

	// Flags
	GetFlags() *flag.FlagSet
	SetFlags(*flag.FlagSet)
	GetPFlags() *flag.FlagSet
	SetPFlags(*flag.FlagSet)
	GetLFlags() *flag.FlagSet
	SetLFlags(*flag.FlagSet)
	GetIFlags() *flag.FlagSet
	SetIFlags(*flag.FlagSet)
	GetParentsPFlags() *flag.FlagSet
	SetParentsPFlags(*flag.FlagSet)
	SetGlobNormFunc(f func(f *flag.FlagSet, name string) flag.NormalizedName)
	GetGlobNormFunc() func(f *flag.FlagSet, name string) flag.NormalizedName
	GetDisableFlagsInUseLine() bool
	GetFParseErrWhitelist() FParseErrWhitelist
	SetFParseErrWhitelist(FParseErrWhitelist)
	GetFlagErrorFunc() func(Commander, error) error
	SetFlagErrorBuf(*bytes.Buffer)
	GetFlagErrorBuf() *bytes.Buffer
	GetSuggestFor() []string
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

func (p *Root) GetDisableSuggestions() bool {
	return p.DisableSuggestions
}

func (p *Root) GetCompletionOptions() *CompletionOptions {
	return &p.CompletionOptions
}
