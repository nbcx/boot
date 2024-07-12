package boot

import (
	"bytes"
	"context"
	"sort"

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
	// GetPersistentPreRunE() func(cmd Commander, args []string) error
	PersistentPreExec(args []string) error
	Exec(args []string) error // Typically the actual work function. Most commands will only implement this.
	PreExec(args []string) error
	PostExec(args []string) error
	PersistentPostExec(args []string) error

	Context() context.Context
	SetContext(ctx context.Context)

	ErrPrefix() string
	GetPositionalArgs() PositionalArgs
	GetCommandsMaxUseLen() int
	GetCommandsMaxCommandPathLen() int
	GetCommandsMaxNameLen() int
	SetCommandsMaxUseLen(v int)
	SetCommandsMaxCommandPathLen(v int)
	SetCommandsMaxNameLen(v int)
	GetTraverseChildren() bool
	GetDisableFlagParsing() bool
	// SetCommands(...Commander)

	GetAliases() []string
	GetDisableAutoGenTag() bool
	SetDisableAutoGenTag(d bool)
	GetVersion() string
	GetAnnotations() map[string]string
	GetDisableSuggestions() bool
	GetSuggestionsMinimumDistance() int
	GetDeprecated() string
	SetCommandsAreSorted(v bool)
	SetCommands(...Commander)

	// 层级关系
	SetParent(Commander)
	Parent() Commander
	GetGroupID() string
	SetGroupID(groupID string)

	// PersistentFlags() *flag.FlagSet
	GetArgs() []string
	// Add(cmds ...Commander)
	// ResetAdd(cmds ...Commander)
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

func (c *Default) SetCommands(v ...Commander) {
	c.commands = v
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
func (c *Default) SetArgs(a ...string) { c.args = a }

// func (c *Default) SetCommands(v ...Commander) { c.commands = v }

func (c *Default) GetArgs() []string                               { return c.args }
func (c *Default) SetFlagErrorFunc(f func(Commander, error) error) { c.flagErrorFunc = f }                                  // SetFlagErrorFunc sets a function to generate an error when flag parsing fails.
func (c *Default) SetHelpCommand(cmd Commander)                    { c.helpCommand = cmd }                                  // SetHelpCommand sets help command.
func (c *Default) Groups() []*Group                                { return c.commandGroups }                               // Groups returns a slice of child command groups.
func (c *Default) Runnable() bool                                  { return true }                                          // Runnable determines if the command is itself runnable.
func (c *Default) AddGroup(groups ...*Group)                       { c.commandGroups = append(c.commandGroups, groups...) } // AddGroup adds one or more command groups to this parent command.
func (d *Default) Parent() Commander                               { return d.parent }                                      // Parent returns a commands parent command.
func (p *Default) SetParent(c Commander)                           { p.parent = c }
func (p *Default) GetGroupID() string                              { return "" }
func (p *Default) SetGroupID(groupID string)                       {}
func (p *Default) GetFlags() *flag.FlagSet                         { return p.flags }
func (p *Default) SetFlags(f *flag.FlagSet)                        { p.flags = f }
func (p *Default) GetHelpCommand() Commander                       { return p.helpCommand }
func (p *Default) GetShort() string                                { return "" }
func (p *Default) PersistentPostExec(args []string) error          { return nil }
func (p *Default) GetSilenceErrors() bool                          { return false }
func (p *Default) GetSilenceUsage() bool                           { return false }
func (p *Default) GetCommandCalledAs() *CommandCalledAs            { return &p.commandCalledAs }
func (p *Default) PersistentPreExec(args []string) error           { return nil }
func (p *Default) GetSuggestFor() []string                         { return nil }
func (p *Default) GetPositionalArgs() PositionalArgs               { return nil }
func (p *Default) GetCommandsMaxUseLen() int                       { return p.commandsMaxUseLen }
func (p *Default) GetCommandsMaxCommandPathLen() int               { return p.commandsMaxCommandPathLen }
func (p *Default) GetCommandsMaxNameLen() int                      { return p.commandsMaxNameLen }
func (p *Default) SetCommandsMaxUseLen(v int)                      { p.commandsMaxUseLen = v }
func (p *Default) SetCommandsMaxCommandPathLen(v int)              { p.commandsMaxCommandPathLen = v }
func (p *Default) SetCommandsMaxNameLen(v int)                     { p.commandsMaxNameLen = v }

func (p *Default) GetFlagErrorFunc() func(Commander, error) error { return p.flagErrorFunc }
func (p *Default) GetTraverseChildren() bool                      { return false }
func (p *Default) GetDisableFlagParsing() bool                    { return false }
func (p *Default) GetArgAliases() []string                        { return nil }
func (p *Default) GetValidArgs() []string                         { return nil }
func (p *Default) GetAliases() []string                           { return nil }
func (p *Default) GetHidden() bool                                { return false }
func (p *Default) GetLong() string                                { return "" }
func (p *Default) GetDisableAutoGenTag() bool                     { return false }
func (p *Default) SetDisableAutoGenTag(d bool)                    {}
func (p *Default) GetExample() string                             { return "" }
func (p *Default) GetCommands() []Commander                       { return p.commands }
func (p *Default) PreExec(args []string) error                    { return nil }

// func (p *Default) Exec(args []string) error                        { return nil } // todo: 这个考虑不默认实现
func (p *Default) PostExec(args []string) error   { return nil }
func (p *Default) getHelpCommandGroupID() string  { return p.helpCommandGroupID }
func (p *Default) GetVersion() string             { return "" }
func (p *Default) GetDeprecated() string          { return "" }
func (p *Default) GetDisableFlagsInUseLine() bool { return false }
func (p *Default) GetDisableSuggestions() bool    { return false }

// func (p *Default) GetUse() string                                  { return "" } // todo: 这个考虑不默认实现
func (p *Default) GetAnnotations() map[string]string        { return nil }
func (p *Default) GetCommandGroups() []*Group               { return nil }
func (p *Default) GetCompletionOptions() *CompletionOptions { return nil }
func (p *Default) GetSuggestionsMinimumDistance() int       { return 2 }
func (p *Default) SetSuggestionsMinimumDistance(v int)      {}
func (p *Default) GetCompletionCommandGroupID() string      { return p.completionCommandGroupID }
func (p *Default) SetFlagErrorBuf(b *bytes.Buffer)          { p.flagErrorBuf = b }
func (p *Default) GetFlagErrorBuf() *bytes.Buffer           { return p.flagErrorBuf }
func (p *Default) GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
	return nil
}

// GlobalNormalizationFunc returns the global normalization function or nil if it doesn't exist.
func (c *Default) GlobalNormalizationFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
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
	// Base(c).SetCompletionCommandGroupID(groupID) // todo: wait do
}

// CommandPathPadding return padding for the command path.
func (c *Default) CommandPathPadding() int {
	if c.parent == nil || minCommandPathPadding > c.parent.GetCommandsMaxCommandPathLen() {
		return minCommandPathPadding
	}
	return c.parent.GetCommandsMaxCommandPathLen()
}

// ErrPrefix return error message prefix for the command
func (c *Default) ErrPrefix() string { return "Error:" }

// ResetCommands delete parent, subcommand and help command from c.
func (c *Default) ResetCommands() {
	c.parent = nil
	c.commands = nil
	c.helpCommand = nil
	c.parentsPFlags = nil
}

// Sorts commands by their names.
type commandSorterByName []Commander

func (c commandSorterByName) Len() int           { return len(c) }
func (c commandSorterByName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c commandSorterByName) Less(i, j int) bool { return name(c[i]) < name(c[j]) }

// Commands returns a sorted slice of child commands.
func (c *Default) Commands() []Commander {
	// do not sort commands if it already sorted or sorting was disabled
	if EnableCommandSorting && !c.commandsAreSorted {
		sort.Sort(commandSorterByName(c.commands))
		c.commandsAreSorted = true
	}
	return c.commands
}

func (c *Default) SetCommandsAreSorted(v bool) { c.commandsAreSorted = v }

// func (c *Default) ResetAdd(cmds ...Commander) {
// 	c.commands = cmds
// 	// recompute all lengths
// 	c.commandsMaxUseLen = 0
// 	c.commandsMaxCommandPathLen = 0
// 	c.commandsMaxNameLen = 0
// 	for _, command := range c.commands {
// 		usageLen := len(command.GetUse())
// 		if usageLen > c.commandsMaxUseLen {
// 			c.commandsMaxUseLen = usageLen
// 		}
// 		commandPathLen := len(CommandPath(command))
// 		if commandPathLen > c.commandsMaxCommandPathLen {
// 			c.commandsMaxCommandPathLen = commandPathLen
// 		}
// 		nameLen := len(name(command))
// 		if nameLen > c.commandsMaxNameLen {
// 			c.commandsMaxNameLen = nameLen
// 		}
// 	}
// }

// ContainsGroup return if groupID exists in the list of command groups.
func (c *Default) ContainsGroup(groupID string) bool {
	for _, x := range c.commandGroups {
		if x.ID == groupID {
			return true
		}
	}
	return false
}
