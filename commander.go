package boot

import (
	"bytes"
	"context"
	"sort"

	flag "github.com/nbcx/flag"
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
func (d *Default) GetGlobNormFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return d.globNormFunc
}

func (d *Default) SetCommands(v ...Commander) {
	d.commands = v
}

// GetFParseErrWhitelist implements Commander.
func (d *Default) GetFParseErrWhitelist() FParseErrWhitelist { return FParseErrWhitelist{} }

// SetFParseErrWhitelist implements Commander.
func (d *Default) SetFParseErrWhitelist(fp FParseErrWhitelist) {}

// SetGlobNormFunc implements Commander.
func (d *Default) SetGlobNormFunc(f func(f *flag.FlagSet, name string) flag.NormalizedName) {
	d.globNormFunc = f
}

// GetPFlags implements Commander.
func (d *Default) GetPFlags() *flag.FlagSet { return d.pFlags }

// GetIFlags implements Commander.
func (d *Default) GetIFlags() *flag.FlagSet { return d.iFlags }

// GetLFlags implements Commander.
func (d *Default) GetLFlags() *flag.FlagSet { return d.lFlags }

// GetParentsPFlags implements Commander.
func (d *Default) GetParentsPFlags() *flag.FlagSet { return d.parentsPFlags }

// SetIFlags implements Commander.
func (d *Default) SetIFlags(i *flag.FlagSet) { d.iFlags = i }

// SetLFlags implements Commander.
func (d *Default) SetLFlags(l *flag.FlagSet) { d.lFlags = l }

// SetPFlags implements Commander.
func (d *Default) SetPFlags(l *flag.FlagSet) { d.pFlags = l }

// SetParentsPFlags implements Commander.
func (d *Default) SetParentsPFlags(pf *flag.FlagSet) { d.parentsPFlags = pf }

// Context returns underlying command context. If command was executed
// with ExecuteContext or the context was set with SetContext, the
// previously set context will be returned. Otherwise, nil is returned.
//
// Notice that a call to Execute and ExecuteC will replace a nil context of
// a command with a context.Background, so a background context will be
// returned by Context after one of these functions has been called.
func (d *Default) Context() context.Context { return d.ctx }

// SetContext sets context for the command. This context will be overwritten by
// Command.ExecuteContext or Command.ExecuteContextC.
func (d *Default) SetContext(ctx context.Context) { d.ctx = ctx }

// SetArgs sets arguments for the command. It is set to os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (d *Default) SetArgs(a ...string) { d.args = a }

// func (c *Default) SetCommands(v ...Commander) { c.commands = v }

func (d *Default) GetArgs() []string                               { return d.args }
func (d *Default) SetFlagErrorFunc(f func(Commander, error) error) { d.flagErrorFunc = f }                                  // SetFlagErrorFunc sets a function to generate an error when flag parsing fails.
func (d *Default) SetHelpCommand(cmd Commander)                    { d.helpCommand = cmd }                                  // SetHelpCommand sets help command.
func (d *Default) Groups() []*Group                                { return d.commandGroups }                               // Groups returns a slice of child command groups.
func (d *Default) Runnable() bool                                  { return true }                                          // Runnable determines if the command is itself runnable.
func (d *Default) AddGroup(groups ...*Group)                       { d.commandGroups = append(d.commandGroups, groups...) } // AddGroup adds one or more command groups to this parent command.
func (d *Default) Parent() Commander                               { return d.parent }                                      // Parent returns a commands parent command.
func (d *Default) SetParent(c Commander)                           { d.parent = c }
func (d *Default) GetGroupID() string                              { return "" }
func (d *Default) SetGroupID(groupID string)                       {}
func (d *Default) GetFlags() *flag.FlagSet                         { return d.flags }
func (d *Default) SetFlags(f *flag.FlagSet)                        { d.flags = f }
func (d *Default) GetHelpCommand() Commander                       { return d.helpCommand }
func (d *Default) GetShort() string                                { return "" }
func (d *Default) PersistentPostExec(args []string) error          { return nil }
func (d *Default) GetSilenceErrors() bool                          { return false }
func (d *Default) GetSilenceUsage() bool                           { return false }
func (d *Default) GetCommandCalledAs() *CommandCalledAs            { return &d.commandCalledAs }
func (d *Default) PersistentPreExec(args []string) error           { return nil }
func (d *Default) GetSuggestFor() []string                         { return nil }
func (d *Default) GetPositionalArgs() PositionalArgs               { return nil }
func (d *Default) GetCommandsMaxUseLen() int                       { return d.commandsMaxUseLen }
func (d *Default) GetCommandsMaxCommandPathLen() int               { return d.commandsMaxCommandPathLen }
func (d *Default) GetCommandsMaxNameLen() int                      { return d.commandsMaxNameLen }
func (d *Default) SetCommandsMaxUseLen(v int)                      { d.commandsMaxUseLen = v }
func (d *Default) SetCommandsMaxCommandPathLen(v int)              { d.commandsMaxCommandPathLen = v }
func (d *Default) SetCommandsMaxNameLen(v int)                     { d.commandsMaxNameLen = v }

func (d *Default) GetFlagErrorFunc() func(Commander, error) error { return d.flagErrorFunc }
func (d *Default) GetTraverseChildren() bool                      { return false }
func (d *Default) GetDisableFlagParsing() bool                    { return false }
func (d *Default) GetArgAliases() []string                        { return nil }
func (d *Default) GetValidArgs() []string                         { return nil }
func (d *Default) GetAliases() []string                           { return nil }
func (d *Default) GetHidden() bool                                { return false }
func (d *Default) GetLong() string                                { return "" }
func (d *Default) GetDisableAutoGenTag() bool                     { return false }
func (d *Default) SetDisableAutoGenTag(v bool)                    {}
func (d *Default) GetExample() string                             { return "" }
func (d *Default) GetCommands() []Commander                       { return d.commands }
func (d *Default) PreExec(args []string) error                    { return nil }

// func (d *Default) Exec(args []string) error                        { return nil } // todo: 这个考虑不默认实现
func (d *Default) PostExec(args []string) error   { return nil }
func (d *Default) getHelpCommandGroupID() string  { return d.helpCommandGroupID }
func (d *Default) GetVersion() string             { return "" }
func (d *Default) GetDeprecated() string          { return "" }
func (d *Default) GetDisableFlagsInUseLine() bool { return false }
func (d *Default) GetDisableSuggestions() bool    { return false }

// func (d *Default) GetUse() string                                  { return "" } // todo: 这个考虑不默认实现
func (d *Default) GetAnnotations() map[string]string        { return nil }
func (d *Default) GetCommandGroups() []*Group               { return nil }
func (d *Default) GetCompletionOptions() *CompletionOptions { return nil }
func (d *Default) GetSuggestionsMinimumDistance() int       { return 2 }
func (d *Default) SetSuggestionsMinimumDistance(v int)      {}
func (d *Default) GetCompletionCommandGroupID() string      { return d.completionCommandGroupID }
func (d *Default) SetFlagErrorBuf(b *bytes.Buffer)          { d.flagErrorBuf = b }
func (d *Default) GetFlagErrorBuf() *bytes.Buffer           { return d.flagErrorBuf }
func (d *Default) GetValidArgsFunction() func(cmd Commander, args []string, toComplete string) ([]string, ShellCompDirective) {
	return nil
}

// GlobalNormalizationFunc returns the global normalization function or nil if it doesn't exist.
func (d *Default) GlobalNormalizationFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return d.globNormFunc
}

// SetHelpCommandGroupID sets the group id of the help command.
func (d *Default) SetHelpCommandGroupID(groupID string) {
	if d.helpCommand != nil {
		d.helpCommand.SetGroupID(groupID)
	}
	// helpCommandGroupID is used if no helpCommand is defined by the user
	d.helpCommandGroupID = groupID
}

// SetCompletionCommandGroupID sets the group id of the completion command.
func (d *Default) SetCompletionCommandGroupID(groupID string) {
	// completionCommandGroupID is used if no completion command is defined by the user
	// Base(c).SetCompletionCommandGroupID(groupID) // todo: wait do
}

// CommandPathPadding return padding for the command path.
func (d *Default) CommandPathPadding() int {
	if d.parent == nil || minCommandPathPadding > d.parent.GetCommandsMaxCommandPathLen() {
		return minCommandPathPadding
	}
	return d.parent.GetCommandsMaxCommandPathLen()
}

// ErrPrefix return error message prefix for the command
func (d *Default) ErrPrefix() string { return "Error:" }

// ResetCommands delete parent, subcommand and help command from c.
func (d *Default) ResetCommands() {
	d.parent = nil
	d.commands = nil
	d.helpCommand = nil
	d.parentsPFlags = nil
}

// Sorts commands by their names.
type commandSorterByName []Commander

func (c commandSorterByName) Len() int           { return len(c) }
func (c commandSorterByName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c commandSorterByName) Less(i, j int) bool { return name(c[i]) < name(c[j]) }

// Commands returns a sorted slice of child commands.
func (d *Default) Commands() []Commander {
	// do not sort commands if it already sorted or sorting was disabled
	if EnableCommandSorting && !d.commandsAreSorted {
		sort.Sort(commandSorterByName(d.commands))
		d.commandsAreSorted = true
	}
	return d.commands
}

func (d *Default) SetCommandsAreSorted(v bool) { d.commandsAreSorted = v }

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
func (d *Default) ContainsGroup(groupID string) bool {
	for _, x := range d.commandGroups {
		if x.ID == groupID {
			return true
		}
	}
	return false
}
