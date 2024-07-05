package cobra

type Default struct{}

// Run: Typically the actual work function. Most commands will only implement this.
func (p *Default) Run(cmd Commander, args []string) {

}

type Boot struct{}

func (p *Boot) Exec(cmds ...Commander) error {

	for _, v := range cmds {
		p.Find(v)
	}
	return nil
}

// Find the target command given the args and command tree
// Meant to be run on the highest node. Only searches down.
func (c *Boot) Find(args []string) (Commander, []string, error) {
	var innerFind func(Commander, []string) (Commander, []string)

	innerFind = func(c Commander, innerArgs []string) (Commander, []string) {
		argsWOflags := stripFlags(innerArgs, c)
		if len(argsWOflags) == 0 {
			return c, innerArgs
		}
		nextSubCmd := argsWOflags[0]

		cmd := c.findNext(nextSubCmd)
		if cmd != nil {
			return innerFind(cmd, c.argsMinusFirstX(innerArgs, nextSubCmd))
		}
		return c, innerArgs
	}

	commandFound, a := innerFind(c, args)
	if commandFound.GetArgs() == nil {
		return commandFound, a, legacyArgs(commandFound, stripFlags(a, commandFound))
	}
	return commandFound, a, nil
}

func (c *Boot) findNext(cmd Commander, next string) Commander {
	matches := make([]Commander, 0)
	for _, cmd := range cmd.GetCommands() {
		if commandNameMatches(cmd.Name(), next) || cmd.HasAlias(next) {
			cmd.GetCommandCalledAs().name = next
			return cmd
		}
		if EnablePrefixMatching && cmd.hasNameOrAliasPrefix(next) {
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
