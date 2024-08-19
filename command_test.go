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

package boot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/nbcx/flag"
)

func emptyRun(Commander, []string) error { return nil }

func executeCommand(root *Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandWithContext(ctx context.Context, root *Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	log.SetOut(buf)
	log.SetErr(buf)
	root.SetArgs(args...)

	root.SetContext(ctx)
	// err = root.ExecuteContext(ctx)
	err = Execute(root)
	return buf.String(), err
}

func executeCommandC(root *Command, args ...string) (c Commander, output string, err error) {
	buf := new(bytes.Buffer)
	log.SetOut(buf)
	log.SetErr(buf)
	root.SetArgs(args...)

	c, err = ExecuteC(root)

	return c, buf.String(), err
}

func executeCommandWithContextC(ctx context.Context, root *Command, args ...string) (c Commander, output string, err error) {
	buf := new(bytes.Buffer)
	log.SetOut(buf)
	log.SetErr(buf)
	root.SetArgs(args...)

	root.SetContext(ctx)
	c, err = ExecuteC(root)

	return c, buf.String(), err
}

func resetCommandLineFlagSet() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func checkStringContains(t *testing.T, got, expected string) {
	if !strings.Contains(got, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
	}
}

func checkStringOmits(t *testing.T, got, expected string) {
	if strings.Contains(got, expected) {
		t.Errorf("Expected to not contain: \n %v\nGot: %v", expected, got)
	}
}

const onetwo = "one two"

func TestSingleCommand(t *testing.T) {
	var rootCmdArgs []string
	rootCmd := &Command{
		Use:  "root",
		Args: ExactArgs(2),
		RunE: func(_ Commander, args []string) error { rootCmdArgs = args; return nil },
	}
	aCmd := &Command{Use: "a", Args: NoArgs, RunE: emptyRun}
	bCmd := &Command{Use: "b", Args: NoArgs, RunE: emptyRun}
	rootCmd.Add(aCmd, bCmd)

	output, err := executeCommand(rootCmd, "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(rootCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

func TestChildCommand(t *testing.T) {
	var child1CmdArgs []string
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	child1Cmd := &Command{
		Use:  "child1",
		Args: ExactArgs(2),
		RunE: func(_ Commander, args []string) error {
			child1CmdArgs = args
			return nil
		},
	}
	child2Cmd := &Command{Use: "child2", Args: NoArgs, RunE: emptyRun}
	rootCmd.Add(child1Cmd, child2Cmd)

	output, err := executeCommand(rootCmd, "child1", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(child1CmdArgs, " ")
	if got != onetwo {
		t.Errorf("child1CmdArgs expected: %q, got: %q", onetwo, got)
	}
}

func TestCallCommandWithoutSubcommands(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	_, err := executeCommand(rootCmd)
	if err != nil {
		t.Errorf("Calling command without subcommands should not have error: %v", err)
	}
}

func TestRootExecuteUnknownCommand(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, _ := executeCommand(rootCmd, "unknown")

	expected := "Error: unknown command \"unknown\" for \"root\"\nRun 'root --help' for usage.\n"

	if output != expected {
		t.Errorf("Expected:\n %q\nGot:\n %q\n", expected, output)
	}
}

func TestSubcommandExecuteC(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)

	c, output, err := executeCommandC(rootCmd, "child")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if name(c) != "child" {
		t.Errorf(`invalid command returned from ExecuteC: expected "child"', got: %q`, name(c))
	}
}

func TestExecuteContext(t *testing.T) {
	ctx := context.TODO()

	ctxRun := func(cmd Commander, args []string) error {
		if cmd.Context() != ctx {
			t.Errorf("Command %q must have context when called with ExecuteContext", cmd.GetUse())
		}

		return nil
	}

	rootCmd := &Command{Use: "root", RunE: ctxRun, PreRunE: ctxRun}
	childCmd := &Command{Use: "child", RunE: ctxRun, PreRunE: ctxRun}
	granchildCmd := &Command{Use: "grandchild", RunE: ctxRun, PreRunE: ctxRun}

	childCmd.Add(granchildCmd)
	rootCmd.Add(childCmd)

	if _, err := executeCommandWithContext(ctx, rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, err := executeCommandWithContext(ctx, rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, err := executeCommandWithContext(ctx, rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

func TestExecuteContextC(t *testing.T) {
	ctx := context.TODO()

	ctxRunE := func(cmd Commander, args []string) error {
		if cmd.Context() != ctx {
			t.Errorf("Command %q must have context when called with ExecuteContext", cmd.GetUse())
		}
		return nil
	}

	rootCmd := &Command{Use: "root", RunE: ctxRunE, PreRunE: ctxRunE}
	childCmd := &Command{Use: "child", RunE: ctxRunE, PreRunE: ctxRunE}
	granChildCmd := &Command{Use: "grandchild", RunE: ctxRunE, PreRunE: ctxRunE}

	childCmd.Add(granChildCmd)
	rootCmd.Add(childCmd)

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

func TestExecute_NoContext(t *testing.T) {
	run := func(cmd Commander, _ []string) {
		if cmd.Context() != context.Background() {
			t.Errorf("Command %s must have background context", cmd.GetUse())
		}
	}
	runE := func(cmd Commander, args []string) error {
		run(cmd, args)
		return nil
	}

	rootCmd := &Command{Use: "root", RunE: runE, PreRunE: runE}
	childCmd := &Command{Use: "child", RunE: runE, PreRunE: runE}
	granchildCmd := &Command{Use: "grandchild", RunE: runE, PreRunE: runE}

	childCmd.Add(granchildCmd)
	rootCmd.Add(childCmd)

	if _, err := executeCommand(rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, err := executeCommand(rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, err := executeCommand(rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

func TestRootUnknownCommandSilenced(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, _ := executeCommand(rootCmd, "unknown")
	if output != "" {
		t.Errorf("Expected blank output, because of silenced usage.\nGot:\n %q\n", output)
	}
}

func TestCommandAlias(t *testing.T) {
	var timesCmdArgs []string
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	echoCmd := &Command{
		Use:     "echo",
		Aliases: []string{"say", "tell"},
		Args:    NoArgs,
		RunE:    emptyRun,
	}
	timesCmd := &Command{
		Use:  "times",
		Args: ExactArgs(2),
		RunE: func(_ Commander, args []string) error {
			timesCmdArgs = args
			return nil
		},
	}
	echoCmd.Add(timesCmd)
	rootCmd.Add(echoCmd)

	output, err := executeCommand(rootCmd, "tell", "times", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(timesCmdArgs, " ")
	if got != onetwo {
		t.Errorf("timesCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

func TestEnablePrefixMatching(t *testing.T) {
	EnablePrefixMatching = true

	var aCmdArgs []string
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	aCmd := &Command{
		Use:  "aCmd",
		Args: ExactArgs(2),
		RunE: func(_ Commander, args []string) error {
			aCmdArgs = args
			return nil
		},
	}
	bCmd := &Command{Use: "bCmd", Args: NoArgs, RunE: emptyRun}
	rootCmd.Add(aCmd, bCmd)

	output, err := executeCommand(rootCmd, "a", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(aCmdArgs, " ")
	if got != onetwo {
		t.Errorf("aCmdArgs expected: %q, got: %q", onetwo, got)
	}

	EnablePrefixMatching = defaultPrefixMatching
}

func TestAliasPrefixMatching(t *testing.T) {
	EnablePrefixMatching = true

	var timesCmdArgs []string
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	echoCmd := &Command{
		Use:     "echo",
		Aliases: []string{"say", "tell"},
		Args:    NoArgs,
		RunE:    emptyRun,
	}
	timesCmd := &Command{
		Use:  "times",
		Args: ExactArgs(2),
		RunE: func(_ Commander, args []string) error {
			timesCmdArgs = args
			return nil
		},
	}
	echoCmd.Add(timesCmd)
	rootCmd.Add(echoCmd)

	output, err := executeCommand(rootCmd, "sa", "times", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(timesCmdArgs, " ")
	if got != onetwo {
		t.Errorf("timesCmdArgs expected: %v, got: %v", onetwo, got)
	}

	EnablePrefixMatching = defaultPrefixMatching
}

// TestPlugin checks usage as plugin for another command such as kubectl.  The
// executable is `kubectl-plugin`, but we run it as `kubectl plugin`. The help
// text should reflect the way we run the command.
func TestPlugin(t *testing.T) {
	cmd := &Command{
		Use:  "kubectl-plugin",
		Args: NoArgs,
		Annotations: map[string]string{
			CommandDisplayNameAnnotation: "kubectl plugin",
		},
		RunE: emptyRun,
	}

	cmdHelp, err := executeCommand(cmd, "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, cmdHelp, "kubectl plugin [flags]")
	checkStringContains(t, cmdHelp, "help for kubectl plugin")
}

// TestPlugin checks usage as plugin with sub commands.
func TestPluginWithSubCommands(t *testing.T) {
	rootCmd := &Command{
		Use:  "kubectl-plugin",
		Args: NoArgs,
		Annotations: map[string]string{
			CommandDisplayNameAnnotation: "kubectl plugin",
		},
	}

	subCmd := &Command{Use: "sub [flags]", Args: NoArgs, RunE: emptyRun}
	rootCmd.Add(subCmd)

	rootHelp, err := executeCommand(rootCmd, "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, rootHelp, "kubectl plugin [command]")
	checkStringContains(t, rootHelp, "help for kubectl plugin")
	checkStringContains(t, rootHelp, "kubectl plugin [command] --help")

	childHelp, err := executeCommand(rootCmd, "sub", "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, childHelp, "kubectl plugin sub [flags]")
	checkStringContains(t, childHelp, "help for sub")

	helpHelp, err := executeCommand(rootCmd, "help", "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, helpHelp, "kubectl plugin help [path to command]")
	checkStringContains(t, helpHelp, "kubectl plugin help [command]")
}

// TestChildSameName checks the correct behaviour of cobra in cases,
// when an application with name "foo" and with subcommand "foo"
// is executed with args "foo foo".
func TestChildSameName(t *testing.T) {
	var fooCmdArgs []string
	rootCmd := &Command{Use: "foo", Args: NoArgs, RunE: emptyRun}
	fooCmd := &Command{
		Use:  "foo",
		Args: ExactArgs(2),
		RunE: func(_ Commander, args []string) error {
			fooCmdArgs = args
			return nil
		},
	}
	barCmd := &Command{Use: "bar", Args: NoArgs, RunE: emptyRun}
	rootCmd.Add(fooCmd, barCmd)

	output, err := executeCommand(rootCmd, "foo", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(fooCmdArgs, " ")
	if got != onetwo {
		t.Errorf("fooCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

// TestGrandChildSameName checks the correct behaviour of cobra in cases,
// when user has a root command and a grand child
// with the same name.
func TestGrandChildSameName(t *testing.T) {
	var fooCmdArgs []string
	rootCmd := &Command{Use: "foo", Args: NoArgs, RunE: emptyRun}
	barCmd := &Command{Use: "bar", Args: NoArgs, RunE: emptyRun}
	fooCmd := &Command{
		Use:  "foo",
		Args: ExactArgs(2),
		RunE: func(_ Commander, args []string) error {
			fooCmdArgs = args
			return nil
		},
	}
	barCmd.Add(fooCmd)
	rootCmd.Add(barCmd)

	output, err := executeCommand(rootCmd, "bar", "foo", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(fooCmdArgs, " ")
	if got != onetwo {
		t.Errorf("fooCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

func TestFlagLong(t *testing.T) {
	var cArgs []string
	c := &Command{
		Use:  "c",
		Args: ArbitraryArgs,
		RunE: func(_ Commander, args []string) error {
			cArgs = args
			return nil
		},
	}

	var intFlagValue int
	var stringFlagValue string
	Flags(c).IntVar(&intFlagValue, "intf", -1, "")
	Flags(c).StringVar(&stringFlagValue, "sf", "", "")

	output, err := executeCommand(c, "--intf=7", "--sf=abc", "one", "--", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if ArgsLenAtDash(c) != 1 {
		t.Errorf("Expected ArgsLenAtDash: %v but got %v", 1, ArgsLenAtDash(c))
	}
	if intFlagValue != 7 {
		t.Errorf("Expected intFlagValue: %v, got %v", 7, intFlagValue)
	}
	if stringFlagValue != "abc" {
		t.Errorf("Expected stringFlagValue: %q, got %q", "abc", stringFlagValue)
	}

	got := strings.Join(cArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

func TestFlagShort(t *testing.T) {
	var cArgs []string
	c := &Command{
		Use:  "c",
		Args: ArbitraryArgs,
		RunE: func(_ Commander, args []string) error {
			cArgs = args
			return nil
		},
	}

	var intFlagValue int
	var stringFlagValue string
	Flags(c).IntVarP(&intFlagValue, "intf", "i", -1, "")
	Flags(c).StringVarP(&stringFlagValue, "sf", "s", "", "")

	output, err := executeCommand(c, "-i", "7", "-sabc", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
	if stringFlagValue != "abc" {
		t.Errorf("Expected stringFlagValue: %q, got %q", "abc", stringFlagValue)
	}

	got := strings.Join(cArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

func TestChildFlag(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)

	var intFlagValue int
	Flags(childCmd).IntVarP(&intFlagValue, "intf", "i", -1, "")

	output, err := executeCommand(rootCmd, "child", "-i7")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
}

func TestChildFlagWithParentLocalFlag(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)

	var intFlagValue int
	Flags(rootCmd).StringP("sf", "s", "", "")
	Flags(childCmd).IntVarP(&intFlagValue, "intf", "i", -1, "")

	_, err := executeCommand(rootCmd, "child", "-i7", "-sabc")
	if err == nil {
		t.Errorf("Invalid flag should generate error")
	}

	checkStringContains(t, err.Error(), "unknown shorthand")

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
}

func TestFlagInvalidInput(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	Flags(rootCmd).IntP("intf", "i", -1, "")

	_, err := executeCommand(rootCmd, "-iabc")
	if err == nil {
		t.Errorf("Invalid flag value should generate error")
	}

	checkStringContains(t, err.Error(), "invalid syntax")
}

func TestFlagBeforeCommand(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)

	var flagValue int
	Flags(childCmd).IntVarP(&flagValue, "intf", "i", -1, "")

	// With short flag.
	_, err := executeCommand(rootCmd, "-i7", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if flagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, flagValue)
	}

	// With long flag.
	_, err = executeCommand(rootCmd, "--intf=8", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if flagValue != 8 {
		t.Errorf("Expected flag value: %v, got %v", 9, flagValue)
	}
}

func TestStripFlags(t *testing.T) {
	tests := []struct {
		input  []string
		output []string
	}{
		{
			[]string{"foo", "bar"},
			[]string{"foo", "bar"},
		},
		{
			[]string{"foo", "--str", "-s"},
			[]string{"foo"},
		},
		{
			[]string{"-s", "foo", "--str", "bar"},
			[]string{},
		},
		{
			[]string{"-i10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-i=10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"--int=100", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-ib", "echo", "-sfoo", "baz"},
			[]string{"echo", "baz"},
		},
		{
			[]string{"-i=baz", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--int=baz", "-sbar", "-i", "foo", "blah"},
			[]string{"blah"},
		},
		{
			[]string{"--bool", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"-b", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--persist", "bar"},
			[]string{"bar"},
		},
		{
			[]string{"-p", "bar"},
			[]string{"bar"},
		},
	}

	c := &Command{Use: "c", RunE: emptyRun}
	PersistentFlags(c).BoolP("persist", "p", false, "")
	Flags(c).IntP("int", "i", -1, "")
	Flags(c).StringP("str", "s", "", "")
	Flags(c).BoolP("bool", "b", false, "")

	for i, test := range tests {
		got := stripFlags(test.input, c)
		if !reflect.DeepEqual(test.output, got) {
			t.Errorf("(%v) Expected: %v, got: %v", i, test.output, got)
		}
	}
}

func TestDisableFlagParsing(t *testing.T) {
	var cArgs []string
	c := &Command{
		Use:                "c",
		DisableFlagParsing: true,
		RunE: func(_ Commander, args []string) error {
			cArgs = args
			return nil
		},
	}

	args := []string{"cmd", "-v", "-race", "-file", "foo.go"}
	output, err := executeCommand(c, args...)
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(args, cArgs) {
		t.Errorf("Expected: %v, got: %v", args, cArgs)
	}
}

func TestPersistentFlagsOnSameCommand(t *testing.T) {
	var rootCmdArgs []string
	rootCmd := &Command{
		Use:  "root",
		Args: ArbitraryArgs,
		RunE: func(_ Commander, args []string) error {
			rootCmdArgs = args
			return nil
		},
	}

	var flagValue int
	PersistentFlags(rootCmd).IntVarP(&flagValue, "intf", "i", -1, "")

	output, err := executeCommand(rootCmd, "-i7", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(rootCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got %q", onetwo, got)
	}
	if flagValue != 7 {
		t.Errorf("flagValue expected: %v, got %v", 7, flagValue)
	}
}

// TestEmptyInputs checks,
// if flags correctly parsed with blank strings in args.
func TestEmptyInputs(t *testing.T) {
	c := &Command{Use: "c", RunE: emptyRun}

	var flagValue int
	Flags(c).IntVarP(&flagValue, "intf", "i", -1, "")

	output, err := executeCommand(c, "", "-i7", "")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if flagValue != 7 {
		t.Errorf("flagValue expected: %v, got %v", 7, flagValue)
	}
}

func TestChildFlagShadowsParentPersistentFlag(t *testing.T) {
	parent := &Command{Use: "parent", RunE: emptyRun}
	child := &Command{Use: "child", RunE: emptyRun}

	PersistentFlags(parent).Bool("boolf", false, "")
	PersistentFlags(parent).Int("intf", -1, "")
	Flags(child).String("strf", "", "")
	Flags(child).Int("intf", -1, "")

	parent.Add(child)

	childInherited := InheritedFlags(child)
	childLocal := LocalFlags(child)

	if childLocal.Lookup("strf") == nil {
		t.Error(`LocalFlags expected to contain "strf", got "nil"`)
	}
	if childInherited.Lookup("boolf") == nil {
		t.Error(`InheritedFlags expected to contain "boolf", got "nil"`)
	}

	if childInherited.Lookup("intf") != nil {
		t.Errorf(`InheritedFlags should not contain shadowed flag "intf"`)
	}
	if childLocal.Lookup("intf") == nil {
		t.Error(`LocalFlags expected to contain "intf", got "nil"`)
	}
}

func TestPersistentFlagsOnChild(t *testing.T) {
	var childCmdArgs []string
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{
		Use:  "child",
		Args: ArbitraryArgs,
		RunE: func(_ Commander, args []string) error {
			childCmdArgs = args
			return nil
		},
	}
	rootCmd.Add(childCmd)

	var parentFlagValue int
	var childFlagValue int
	PersistentFlags(rootCmd).IntVarP(&parentFlagValue, "parentf", "p", -1, "")
	Flags(childCmd).IntVarP(&childFlagValue, "childf", "c", -1, "")

	output, err := executeCommand(rootCmd, "child", "-c7", "-p8", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(childCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
	if parentFlagValue != 8 {
		t.Errorf("parentFlagValue expected: %v, got %v", 8, parentFlagValue)
	}
	if childFlagValue != 7 {
		t.Errorf("childFlagValue expected: %v, got %v", 7, childFlagValue)
	}
}

func TestRequiredFlags(t *testing.T) {
	c := &Command{Use: "c", RunE: emptyRun}
	Flags(c).String("foo1", "", "")
	assertNoErr(t, MarkFlagRequired(c, "foo1"))
	Flags(c).String("foo2", "", "")
	assertNoErr(t, MarkFlagRequired(c, "foo2"))
	Flags(c).String("bar", "", "")

	expected := fmt.Sprintf("required flag(s) %q, %q not set", "foo1", "foo2")

	_, err := executeCommand(c)
	got := err.Error()

	if got != expected {
		t.Errorf("Expected error: %q, got: %q", expected, got)
	}
}

func TestPersistentRequiredFlags(t *testing.T) {
	parent := &Command{Use: "parent", RunE: emptyRun}
	PersistentFlags(parent).String("foo1", "", "")
	assertNoErr(t, MarkPersistentFlagRequired(parent, "foo1"))
	PersistentFlags(parent).String("foo2", "", "")
	assertNoErr(t, MarkPersistentFlagRequired(parent, "foo2"))
	Flags(parent).String("foo3", "", "")

	child := &Command{Use: "child", RunE: emptyRun}
	Flags(child).String("bar1", "", "")
	assertNoErr(t, MarkFlagRequired(child, "bar1"))
	Flags(child).String("bar2", "", "")
	assertNoErr(t, MarkFlagRequired(child, "bar2"))
	Flags(child).String("bar3", "", "")

	parent.Add(child)

	expected := fmt.Sprintf("required flag(s) %q, %q, %q, %q not set", "bar1", "bar2", "foo1", "foo2")

	_, err := executeCommand(parent, "child")
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestPersistentRequiredFlagsWithDisableFlagParsing(t *testing.T) {
	// Make sure a required persistent flag does not break
	// commands that disable flag parsing

	parent := &Command{Use: "parent", RunE: emptyRun}
	PersistentFlags(parent).Bool("foo", false, "")
	flag := PersistentFlags(parent).Lookup("foo")
	assertNoErr(t, MarkPersistentFlagRequired(parent, "foo"))

	child := &Command{Use: "child", RunE: emptyRun}
	child.DisableFlagParsing = true

	parent.Add(child)

	if _, err := executeCommand(parent, "--foo", "child"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Reset the flag or else it will remember the state from the previous command
	flag.Changed = false
	if _, err := executeCommand(parent, "child", "--foo"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Reset the flag or else it will remember the state from the previous command
	flag.Changed = false
	if _, err := executeCommand(parent, "child"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestInitHelpFlagMergesFlags(t *testing.T) {
	usage := "custom flag"
	rootCmd := &Command{Use: "root"}
	PersistentFlags(rootCmd).Bool("help", false, "custom flag")
	childCmd := &Command{Use: "child"}
	rootCmd.Add(childCmd)

	InitDefaultHelpFlag(childCmd)
	got := Flags(childCmd).Lookup("help").Usage
	if got != usage {
		t.Errorf("Expected the help flag from the root command with usage: %v\nGot the default with usage: %v", usage, got)
	}
}

func TestHelpCommandExecuted(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description", RunE: emptyRun}
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
}

func TestHelpCommandExecutedOnChild(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", Long: "Long description", RunE: emptyRun}
	rootCmd.Add(childCmd)

	output, err := executeCommand(rootCmd, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

func TestHelpCommandExecutedOnChildWithFlagThatShadowsParentFlag(t *testing.T) {
	parent := &Command{Use: "parent", RunE: emptyRun}
	child := &Command{Use: "child", RunE: emptyRun}
	parent.Add(child)

	PersistentFlags(parent).Bool("foo", false, "parent foo usage")
	PersistentFlags(parent).Bool("bar", false, "parent bar usage")
	Flags(child).Bool("foo", false, "child foo usage") // This shadows parent's foo flag
	Flags(child).Bool("baz", false, "child baz usage")

	got, err := executeCommand(parent, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `Usage:
  parent child [flags]

Flags:
      --baz    child baz usage
      --foo    child foo usage
  -h, --help   help for child

Global Flags:
      --bar   parent bar usage
`

	if got != expected {
		t.Errorf("Help text mismatch.\nExpected:\n%s\n\nGot:\n%s\n", expected, got)
	}
}

func TestSetHelpCommand(t *testing.T) {
	c := &Command{Use: "c", RunE: emptyRun}
	c.Add(&Command{Use: "empty", RunE: emptyRun})

	expected := "WORKS"
	c.SetHelpCommand(&Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
	Simply type ` + name(c) + ` help [path to command] for full details.`,
		RunE: func(c Commander, _ []string) error {
			log.Print(expected)
			return nil
		},
	})

	got, err := executeCommand(c, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if got != expected {
		t.Errorf("Expected to contain %q, got %q", expected, got)
	}
}

func TestHelpFlagExecuted(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
}

func TestHelpFlagExecutedOnChild(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", Long: "Long description", RunE: emptyRun}
	rootCmd.Add(childCmd)

	output, err := executeCommand(rootCmd, "child", "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

// TestHelpFlagInHelp checks,
// if '--help' flag is shown in help for child (executing `parent help child`),
// that has no other flags.
// Related to https://github.com/spf13/cobra/issues/302.
func TestHelpFlagInHelp(t *testing.T) {
	parentCmd := &Command{Use: "parent", RunE: emptyRun}

	childCmd := &Command{Use: "child", RunE: emptyRun}
	parentCmd.Add(childCmd)

	output, err := executeCommand(parentCmd, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "[flags]")
}

func TestFlagsInUsage(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "[flags]")
}

func TestHelpExecutedOnNonRunnableChild(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", Long: "Long description"}
	rootCmd.Add(childCmd)

	output, err := executeCommand(rootCmd, "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

func TestVersionFlagExecuted(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

func TestVersionFlagExecutedWithNoName(t *testing.T) {
	rootCmd := &Command{Version: "1.0.0", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "version 1.0.0")
}

func TestShortAndLongVersionFlagInHelp(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "-v, --version")
}

func TestLongVersionFlagOnlyInHelpWhenShortPredefined(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	Flags(rootCmd).StringP("foo", "v", "", "not a version flag")

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringOmits(t, output, "-v, --version")
	checkStringContains(t, output, "--version")
}

func TestShorthandVersionFlagExecuted(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "-v", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

func TestVersionTemplate(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.SetVersionTemplate(`customized version: {{.Version}}`)

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "customized version: 1.0.0")
}

func TestShorthandVersionTemplate(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.SetVersionTemplate(`customized version: {{.Version}}`)

	output, err := executeCommand(rootCmd, "-v", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "customized version: 1.0.0")
}

func TestRootErrPrefixExecutedOnSubcommand(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	// rootCmd.SetErrPrefix("root error prefix:")
	rootCmd.Add(&Command{Use: "sub", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "sub", "--unknown-flag")
	if err == nil {
		t.Errorf("Expected error")
	}

	checkStringContains(t, output, "root error prefix: unknown flag: --unknown-flag")
}

func TestRootAndSubErrPrefix(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	subCmd := &Command{Use: "sub", RunE: emptyRun}
	rootCmd.Add(subCmd)
	// rootCmd.SetErrPrefix("root error prefix:")
	// subCmd.SetErrPrefix("sub error prefix:")

	if output, err := executeCommand(rootCmd, "--unknown-root-flag"); err == nil {
		t.Errorf("Expected error")
	} else {
		checkStringContains(t, output, "root error prefix: unknown flag: --unknown-root-flag")
	}

	if output, err := executeCommand(rootCmd, "sub", "--unknown-sub-flag"); err == nil {
		t.Errorf("Expected error")
	} else {
		checkStringContains(t, output, "sub error prefix: unknown flag: --unknown-sub-flag")
	}
}

func TestVersionFlagExecutedOnSubcommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0"}
	rootCmd.Add(&Command{Use: "sub", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "--version", "sub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

func TestShorthandVersionFlagExecutedOnSubcommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0"}
	rootCmd.Add(&Command{Use: "sub", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "-v", "sub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

func TestVersionFlagOnlyAddedToRoot(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.Add(&Command{Use: "sub", RunE: emptyRun})

	_, err := executeCommand(rootCmd, "sub", "--version")
	if err == nil {
		t.Errorf("Expected error")
	}

	checkStringContains(t, err.Error(), "unknown flag: --version")
}

func TestShortVersionFlagOnlyAddedToRoot(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.Add(&Command{Use: "sub", RunE: emptyRun})

	_, err := executeCommand(rootCmd, "sub", "-v")
	if err == nil {
		t.Errorf("Expected error")
	}

	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

func TestVersionFlagOnlyExistsIfVersionNonEmpty(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}

	_, err := executeCommand(rootCmd, "--version")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown flag: --version")
}

func TestShorthandVersionFlagOnlyExistsIfVersionNonEmpty(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

func TestShorthandVersionFlagOnlyAddedIfShorthandNotDefined(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun, Version: "1.2.3"}
	Flags(rootCmd).StringP("notversion", "v", "", "not a version flag")

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	check(t, Flags(rootCmd).ShorthandLookup("v").Name, "notversion")
	checkStringContains(t, err.Error(), "flag needs an argument: 'v' in -v")
}

func TestShorthandVersionFlagOnlyAddedIfVersionNotDefined(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun, Version: "1.2.3"}
	Flags(rootCmd).Bool("version", false, "a different kind of version flag")

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

func TestUsageIsNotPrintedTwice(t *testing.T) {
	var cmd = &Command{Use: "root"}
	var sub = &Command{Use: "sub"}
	cmd.Add(sub)

	output, _ := executeCommand(cmd, "")
	if strings.Count(output, "Usage:") != 1 {
		t.Error("Usage output is not printed exactly once")
	}
}

func TestVisitParents(t *testing.T) {
	c := &Command{Use: "app"}
	sub := &Command{Use: "sub"}
	dsub := &Command{Use: "dsub"}
	sub.Add(dsub)
	c.Add(sub)

	total := 0
	add := func(x Commander) {
		total++
	}
	VisitParents(sub, add)
	if total != 1 {
		t.Errorf("Should have visited 1 parent but visited %d", total)
	}

	total = 0
	VisitParents(dsub, add)
	if total != 2 {
		t.Errorf("Should have visited 2 parents but visited %d", total)
	}

	total = 0
	VisitParents(c, add)
	if total != 0 {
		t.Errorf("Should have visited no parents but visited %d", total)
	}
}

func TestSuggestions(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	timesCmd := &Command{
		Use:        "times",
		SuggestFor: []string{"counts"},
		RunE:       emptyRun,
	}
	rootCmd.Add(timesCmd)

	templateWithSuggestions := "Error: unknown command \"%s\" for \"root\"\n\nDid you mean this?\n\t%s\n\nRun 'root --help' for usage.\n"
	templateWithoutSuggestions := "Error: unknown command \"%s\" for \"root\"\nRun 'root --help' for usage.\n"

	tests := map[string]string{
		"time":     "times",
		"tiems":    "times",
		"tims":     "times",
		"timeS":    "times",
		"rimes":    "times",
		"ti":       "times",
		"t":        "times",
		"timely":   "times",
		"ri":       "",
		"timezone": "",
		"foo":      "",
		"counts":   "times",
	}

	for typo, suggestion := range tests {
		for _, suggestionsDisabled := range []bool{true, false} {
			rootCmd.DisableSuggestions = suggestionsDisabled

			var expected string
			output, _ := executeCommand(rootCmd, typo)

			if suggestion == "" || suggestionsDisabled {
				expected = fmt.Sprintf(templateWithoutSuggestions, typo)
			} else {
				expected = fmt.Sprintf(templateWithSuggestions, typo, suggestion)
			}

			if output != expected {
				t.Errorf("Unexpected response.\nExpected:\n %q\nGot:\n %q\n", expected, output)
			}
		}
	}
}

func TestCaseInsensitive(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", RunE: emptyRun, Aliases: []string{"alternative"}}
	granchildCmd := &Command{Use: "GRANDCHILD", RunE: emptyRun, Aliases: []string{"ALIAS"}}

	childCmd.Add(granchildCmd)
	rootCmd.Add(childCmd)

	tests := []struct {
		args                []string
		failWithoutEnabling bool
	}{
		{
			args:                []string{"child"},
			failWithoutEnabling: false,
		},
		{
			args:                []string{"CHILD"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"chILD"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"CHIld"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"alternative"},
			failWithoutEnabling: false,
		},
		{
			args:                []string{"ALTERNATIVE"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"ALTernatIVE"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"alternatiVE"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"child", "GRANDCHILD"},
			failWithoutEnabling: false,
		},
		{
			args:                []string{"child", "grandchild"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"CHIld", "GRANdchild"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"alternative", "ALIAS"},
			failWithoutEnabling: false,
		},
		{
			args:                []string{"alternative", "alias"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"CHILD", "alias"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"CHIld", "aliAS"},
			failWithoutEnabling: true,
		},
	}

	for _, test := range tests {
		for _, enableCaseInsensitivity := range []bool{true, false} {
			EnableCaseInsensitive = enableCaseInsensitivity

			output, err := executeCommand(rootCmd, test.args...)
			expectedFailure := test.failWithoutEnabling && !enableCaseInsensitivity

			if !expectedFailure && output != "" {
				t.Errorf("Unexpected output: %v", output)
			}
			if !expectedFailure && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	}

	EnableCaseInsensitive = defaultCaseInsensitive
}

// This test make sure we keep backwards-compatibility with respect
// to command names case sensitivity behavior.
func TestCaseSensitivityBackwardCompatibility(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", RunE: emptyRun}

	rootCmd.Add(childCmd)
	_, err := executeCommand(rootCmd, strings.ToUpper(childCmd.Use))
	if err == nil {
		t.Error("Expected error on calling a command in upper case while command names are case sensitive. Got nil.")
	}

}

func TestRemoveCommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	childCmd := &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)
	RemoveCommand(rootCmd, childCmd)

	_, err := executeCommand(rootCmd, "child")
	if err == nil {
		t.Error("Expected error on calling removed command. Got nil.")
	}
}

func TestReplaceCommandWithRemove(t *testing.T) {
	childUsed := 0
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	child1Cmd := &Command{
		Use: "child",
		RunE: func(Commander, []string) error {
			childUsed = 1
			return nil
		},
	}
	child2Cmd := &Command{
		Use: "child",
		RunE: func(Commander, []string) error {
			childUsed = 2
			return nil
		},
	}
	rootCmd.Add(child1Cmd)
	RemoveCommand(rootCmd, child1Cmd)
	rootCmd.Add(child2Cmd)

	output, err := executeCommand(rootCmd, "child")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if childUsed == 1 {
		t.Error("Removed command shouldn't be called")
	}
	if childUsed != 2 {
		t.Error("Replacing command should have been called but didn't")
	}
}

func TestDeprecatedCommand(t *testing.T) {
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	deprecatedCmd := &Command{
		Use:        "deprecated",
		Deprecated: "This command is deprecated",
		RunE:       emptyRun,
	}
	rootCmd.Add(deprecatedCmd)

	output, err := executeCommand(rootCmd, "deprecated")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, deprecatedCmd.Deprecated)
}

func TestHooks(t *testing.T) {
	var (
		persPreArgs  string
		preArgs      string
		runArgs      string
		postArgs     string
		persPostArgs string
	)

	c := &Command{
		Use: "c",
		PersistentPreRun: func(_ Commander, args []string) {
			persPreArgs = strings.Join(args, " ")
		},
		PreRunE: func(_ Commander, args []string) error {
			preArgs = strings.Join(args, " ")
			return nil
		},
		RunE: func(_ Commander, args []string) error {
			runArgs = strings.Join(args, " ")
			return nil
		},
		PostRunE: func(_ Commander, args []string) error {
			postArgs = strings.Join(args, " ")
			return nil
		},
		PersistentPostRun: func(_ Commander, args []string) {
			persPostArgs = strings.Join(args, " ")
		},
	}

	output, err := executeCommand(c, "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, v := range []struct {
		name string
		got  string
	}{
		{"persPreArgs", persPreArgs},
		{"preArgs", preArgs},
		{"runArgs", runArgs},
		{"postArgs", postArgs},
		{"persPostArgs", persPostArgs},
	} {
		if v.got != onetwo {
			t.Errorf("Expected %s %q, got %q", v.name, onetwo, v.got)
		}
	}
}

func TestPersistentHooks(t *testing.T) {
	EnableTraverseRunHooks = true
	testPersistentHooks(t, []string{
		"parent PersistentPreRun",
		"child PersistentPreRun",
		"child PreRun",
		"child Run",
		"child PostRun",
		"child PersistentPostRun",
		"parent PersistentPostRun",
	})

	EnableTraverseRunHooks = false
	testPersistentHooks(t, []string{
		"child PersistentPreRun",
		"child PreRun",
		"child Run",
		"child PostRun",
		"child PersistentPostRun",
	})
}

func testPersistentHooks(t *testing.T, expectedHookRunOrder []string) {
	var hookRunOrder []string

	validateHook := func(args []string, hookName string) {
		hookRunOrder = append(hookRunOrder, hookName)
		got := strings.Join(args, " ")
		if onetwo != got {
			t.Errorf("Expected %s %q, got %q", hookName, onetwo, got)
		}
	}

	parentCmd := &Command{
		Use: "parent",
		PersistentPreRun: func(_ Commander, args []string) {
			validateHook(args, "parent PersistentPreRun")
		},
		PreRunE: func(_ Commander, args []string) error {
			validateHook(args, "parent PreRun")
			return nil
		},
		RunE: func(_ Commander, args []string) error {
			validateHook(args, "parent Run")
			return nil
		},
		PostRunE: func(_ Commander, args []string) error {
			validateHook(args, "parent PostRun")
			return nil
		},
		PersistentPostRun: func(_ Commander, args []string) {
			validateHook(args, "parent PersistentPostRun")
		},
	}

	childCmd := &Command{
		Use: "child",
		PersistentPreRun: func(_ Commander, args []string) {
			validateHook(args, "child PersistentPreRun")
		},
		PreRunE: func(_ Commander, args []string) error {
			validateHook(args, "child PreRun")
			return nil
		},
		RunE: func(_ Commander, args []string) error {
			validateHook(args, "child Run")
			return nil
		},
		PostRunE: func(_ Commander, args []string) error {
			validateHook(args, "child PostRun")
			return nil
		},
		PersistentPostRun: func(_ Commander, args []string) {
			validateHook(args, "child PersistentPostRun")
		},
	}
	parentCmd.Add(childCmd)

	output, err := executeCommand(parentCmd, "child", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for idx, exp := range expectedHookRunOrder {
		if len(hookRunOrder) > idx {
			if act := hookRunOrder[idx]; act != exp {
				t.Errorf("Expected %q at %d, got %q", exp, idx, act)
			}
		} else {
			t.Errorf("Expected %q at %d, got nothing", exp, idx)
		}
	}
}

// Related to https://github.com/spf13/cobra/issues/521.
func TestGlobalNormFuncPropagation(t *testing.T) {
	normFunc := func(f *flag.FlagSet, name string) flag.NormalizedName {
		return flag.NormalizedName(name)
	}

	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)

	SetGlobalNormalizationFunc(rootCmd, normFunc)
	if reflect.ValueOf(normFunc).Pointer() != reflect.ValueOf(rootCmd.GlobalNormalizationFunc()).Pointer() {
		t.Error("rootCmd seems to have a wrong normalization function")
	}

	if reflect.ValueOf(normFunc).Pointer() != reflect.ValueOf(childCmd.GlobalNormalizationFunc()).Pointer() {
		t.Error("childCmd should have had the normalization function of rootCmd")
	}
}

// Related to https://github.com/spf13/cobra/issues/521.
func TestNormPassedOnLocal(t *testing.T) {
	toUpper := func(f *flag.FlagSet, name string) flag.NormalizedName {
		return flag.NormalizedName(strings.ToUpper(name))
	}

	c := &Command{}
	Flags(c).Bool("flagname", true, "this is a dummy flag")
	SetGlobalNormalizationFunc(c, toUpper)
	if LocalFlags(c).Lookup("flagname") != LocalFlags(c).Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to Local flag set")
	}
}

// Related to https://github.com/spf13/cobra/issues/521.
func TestNormPassedOnInherited(t *testing.T) {
	toUpper := func(f *flag.FlagSet, name string) flag.NormalizedName {
		return flag.NormalizedName(strings.ToUpper(name))
	}

	c := &Command{}
	SetGlobalNormalizationFunc(c, toUpper)

	child1 := &Command{}
	c.Add(child1)

	PersistentFlags(c).Bool("flagname", true, "")

	child2 := &Command{}
	c.Add(child2)

	inherited := InheritedFlags(child1)
	if inherited.Lookup("flagname") == nil || inherited.Lookup("flagname") != inherited.Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to inherited flag set in command added before flag")
	}

	inherited = InheritedFlags(child2)
	if inherited.Lookup("flagname") == nil || inherited.Lookup("flagname") != inherited.Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to inherited flag set in command added after flag")
	}
}

// Related to https://github.com/spf13/cobra/issues/521.
func TestConsistentNormalizedName(t *testing.T) {
	toUpper := func(f *flag.FlagSet, name string) flag.NormalizedName {
		return flag.NormalizedName(strings.ToUpper(name))
	}
	n := func(f *flag.FlagSet, name string) flag.NormalizedName {
		return flag.NormalizedName(name)
	}

	c := &Command{}
	Flags(c).Bool("flagname", true, "")
	SetGlobalNormalizationFunc(c, toUpper)
	SetGlobalNormalizationFunc(c, n)

	if LocalFlags(c).Lookup("flagname") == LocalFlags(c).Lookup("FLAGNAME") {
		t.Error("Normalizing flag names should not result in duplicate flags")
	}
}

func TestFlagOnPflagCommandLine(t *testing.T) {
	flagName := "flagOnCommandLine"
	flag.String(flagName, "", "about my flag")

	c := &Command{Use: "c", RunE: emptyRun}
	c.Add(&Command{Use: "child", RunE: emptyRun})

	output, _ := executeCommand(c, "--help")
	checkStringContains(t, output, flagName)

	resetCommandLineFlagSet()
}

// TestHiddenCommandExecutes checks,
// if hidden commands run as intended.
func TestHiddenCommandExecutes(t *testing.T) {
	executed := false
	c := &Command{
		Use:    "c",
		Hidden: true,
		RunE: func(Commander, []string) error {
			executed = true
			return nil
		},
	}

	output, err := executeCommand(c)
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !executed {
		t.Error("Hidden command should have been executed")
	}
}

// test to ensure hidden commands do not show up in usage/help text
func TestHiddenCommandIsHidden(t *testing.T) {
	c := &Command{Use: "c", Hidden: true, RunE: emptyRun}
	if IsAvailableCommand(c) {
		t.Errorf("Hidden command should be unavailable")
	}
}

func TestCommandsAreSorted(t *testing.T) {
	EnableCommandSorting = true

	originalNames := []string{"middle", "zlast", "afirst"}
	expectedNames := []string{"afirst", "middle", "zlast"}

	var rootCmd = &Command{Use: "root"}

	for _, name := range originalNames {
		rootCmd.Add(&Command{Use: name})
	}

	for i, c := range rootCmd.Commands() {
		got := name(c)
		if expectedNames[i] != got {
			t.Errorf("Expected: %s, got: %s", expectedNames[i], got)
		}
	}

	EnableCommandSorting = defaultCommandSorting
}

func TestEnableCommandSortingIsDisabled(t *testing.T) {
	EnableCommandSorting = false

	originalNames := []string{"middle", "zlast", "afirst"}

	var rootCmd = &Command{Use: "root"}

	for _, name := range originalNames {
		rootCmd.Add(&Command{Use: name})
	}

	for i, c := range rootCmd.Commands() {
		got := name(c)
		if originalNames[i] != got {
			t.Errorf("expected: %s, got: %s", originalNames[i], got)
		}
	}

	EnableCommandSorting = defaultCommandSorting
}

func TestUsageWithGroup(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddGroup(&Group{ID: "group1", Title: "group1"})
	rootCmd.AddGroup(&Group{ID: "group2", Title: "group2"})

	rootCmd.Add(&Command{Use: "cmd1", GroupID: "group1", RunE: emptyRun})
	rootCmd.Add(&Command{Use: "cmd2", GroupID: "group2", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// help should be ungrouped here
	checkStringContains(t, output, "\nAdditional Commands:\n  help")
	checkStringContains(t, output, "\ngroup1\n  cmd1")
	checkStringContains(t, output, "\ngroup2\n  cmd2")
}

func TestUsageHelpGroup(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddGroup(&Group{ID: "group", Title: "group"})
	rootCmd.Add(&Command{Use: "xxx", GroupID: "group", RunE: emptyRun})
	rootCmd.SetHelpCommandGroupID("group")

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// now help should be grouped under "group"
	checkStringOmits(t, output, "\nAdditional Commands:\n  help")
	checkStringContains(t, output, "\ngroup\n  help")
}

func TestUsageCompletionGroup(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}

	rootCmd.AddGroup(&Group{ID: "group", Title: "group"})
	rootCmd.AddGroup(&Group{ID: "help", Title: "help"})

	rootCmd.Add(&Command{Use: "xxx", GroupID: "group", RunE: emptyRun})
	rootCmd.SetHelpCommandGroupID("help")
	rootCmd.SetCompletionCommandGroupID("group")

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// now completion should be grouped under "group"
	checkStringOmits(t, output, "\nAdditional Commands:\n  completion")
	checkStringContains(t, output, "\ngroup\n  completion")
}

func TestUngroupedCommand(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}

	rootCmd.AddGroup(&Group{ID: "group", Title: "group"})
	rootCmd.AddGroup(&Group{ID: "help", Title: "help"})

	rootCmd.Add(&Command{Use: "xxx", GroupID: "group", RunE: emptyRun})
	rootCmd.SetHelpCommandGroupID("help")
	rootCmd.SetCompletionCommandGroupID("group")

	// Add a command without a group
	rootCmd.Add(&Command{Use: "yyy", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// The yyy command should be in the additional command "group"
	checkStringContains(t, output, "\nAdditional Commands:\n  yyy")
}

func TestAddGroup(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}

	rootCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	rootCmd.Add(&Command{Use: "cmd", GroupID: "group", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "\nTest group\n  cmd")
}

func TestWrongGroupFirstLevel(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}

	rootCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	// Use the wrong group ID
	rootCmd.Add(&Command{Use: "cmd", GroupID: "wrong", RunE: emptyRun})

	defer func() {
		if recover() == nil {
			t.Errorf("The code should have panicked due to a missing group")
		}
	}()
	_, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWrongGroupNestedLevel(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}
	var childCmd = &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)

	childCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	// Use the wrong group ID
	childCmd.Add(&Command{Use: "cmd", GroupID: "wrong", RunE: emptyRun})

	defer func() {
		if recover() == nil {
			t.Errorf("The code should have panicked due to a missing group")
		}
	}()
	_, err := executeCommand(rootCmd, "child", "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWrongGroupForHelp(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}
	var childCmd = &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)

	rootCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	// Use the wrong group ID
	rootCmd.SetHelpCommandGroupID("wrong")

	defer func() {
		if recover() == nil {
			t.Errorf("The code should have panicked due to a missing group")
		}
	}()
	_, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWrongGroupForCompletion(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", RunE: emptyRun}
	var childCmd = &Command{Use: "child", RunE: emptyRun}
	rootCmd.Add(childCmd)

	rootCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	// Use the wrong group ID
	rootCmd.SetCompletionCommandGroupID("wrong")

	defer func() {
		if recover() == nil {
			t.Errorf("The code should have panicked due to a missing group")
		}
	}()
	_, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestSetOutput(t *testing.T) {
	// c := &Command{}
	log.SetOutput(nil)
	if out := log.OutOrStdout(); out != os.Stdout {
		t.Errorf("Expected setting output to nil to revert back to stdout")
	}
}

func TestSetOut(t *testing.T) {
	// c := &Command{}
	log.SetOut(nil)
	if out := log.OutOrStdout(); out != os.Stdout {
		t.Errorf("Expected setting output to nil to revert back to stdout")
	}
}

func TestSetErr(t *testing.T) {
	// c := &Command{}
	log.SetErr(nil)
	if out := log.ErrOrStderr(); out != os.Stderr {
		t.Errorf("Expected setting error to nil to revert back to stderr")
	}
}

func TestSetIn(t *testing.T) {
	// c := &Command{}
	log.SetIn(nil)
	if out := log.InOrStdin(); out != os.Stdin {
		t.Errorf("Expected setting input to nil to revert back to stdin")
	}
}

func TestUsageStringRedirected(t *testing.T) {
	c := &Command{}
	// c.usageFunc = func(cmd Commander) error {
	// 	log.Print("[stdout1]")
	// 	log.PrintErr("[stderr2]")
	// 	log.Print("[stdout3]")
	// 	return nil
	// }

	expected := "[stdout1][stderr2][stdout3]"
	if got := UsageString(c); got != expected {
		t.Errorf("Expected usage string to consider both stdout and stderr")
	}
}

func TestCommandPrintRedirection(t *testing.T) {
	errBuff, outBuff := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	root := &Command{
		RunE: func(cmd Commander, args []string) error {

			log.PrintErr("PrintErr")
			log.PrintErrLn("PrintErr", "line")
			log.PrintErrF("PrintEr%s", "r")

			log.Print("Print")
			log.Println("Print", "line")
			log.Printf("Prin%s", "t")
			return nil
		},
	}

	log.SetErr(errBuff)
	log.SetOut(outBuff)

	if err := Execute(root); err != nil {
		t.Error(err)
	}

	gotErrBytes, err := io.ReadAll(errBuff)
	if err != nil {
		t.Error(err)
	}

	gotOutBytes, err := io.ReadAll(outBuff)
	if err != nil {
		t.Error(err)
	}

	if wantErr := []byte("PrintErrPrintErr line\nPrintErr"); !bytes.Equal(gotErrBytes, wantErr) {
		t.Errorf("got: '%s' want: '%s'", gotErrBytes, wantErr)
	}

	if wantOut := []byte("PrintPrint line\nPrint"); !bytes.Equal(gotOutBytes, wantOut) {
		t.Errorf("got: '%s' want: '%s'", gotOutBytes, wantOut)
	}
}

func TestFlagErrorFunc(t *testing.T) {
	c := &Command{Use: "c", RunE: emptyRun}

	expectedFmt := "This is expected: %v"
	c.SetFlagErrorFunc(func(_ Commander, err error) error {
		return fmt.Errorf(expectedFmt, err)
	})

	_, err := executeCommand(c, "--unknown-flag")

	got := err.Error()
	expected := fmt.Sprintf(expectedFmt, "unknown flag: --unknown-flag")
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestFlagErrorFuncHelp(t *testing.T) {
	c := &Command{Use: "c", RunE: emptyRun}
	PersistentFlags(c).Bool("help", false, "help for c")
	c.SetFlagErrorFunc(func(_ Commander, err error) error {
		return fmt.Errorf("wrap error: %w", err)
	})

	out, err := executeCommand(c, "--help")
	if err != nil {
		t.Errorf("--help should not fail: %v", err)
	}

	expected := `Usage:
  c [flags]

Flags:
      --help   help for c
`
	if out != expected {
		t.Errorf("Expected: %v, got: %v", expected, out)
	}

	out, err = executeCommand(c, "-h")
	if err != nil {
		t.Errorf("-h should not fail: %v", err)
	}

	if out != expected {
		t.Errorf("Expected: %v, got: %v", expected, out)
	}
}

// TestSortedFlags checks,
// if cmd.LocalFlags() is unsorted when cmd.Flags().SortFlags set to false.
// Related to https://github.com/spf13/cobra/issues/404.
func TestSortedFlags(t *testing.T) {
	c := &Command{}
	Flags(c).SortFlags = false
	names := []string{"C", "B", "A", "D"}
	for _, name := range names {
		Flags(c).Bool(name, false, "")
	}

	i := 0
	LocalFlags(c).VisitAll(func(f *flag.Flag) {
		if i == len(names) {
			return
		}
		if stringInSlice(f.Name, names) {
			if names[i] != f.Name {
				t.Errorf("Incorrect order. Expected %v, got %v", names[i], f.Name)
			}
			i++
		}
	})
}

// TestMergeCommandLineToFlags checks,
// if pflag.CommandLine is correctly merged to c.Flags() after first call
// of c.mergePersistentFlags.
// Related to https://github.com/spf13/cobra/issues/443.
func TestMergeCommandLineToFlags(t *testing.T) {
	flag.Bool("boolflag", false, "")
	c := &Command{Use: "c", RunE: emptyRun}
	mergePersistentFlags(c)
	if Flags(c).Lookup("boolflag") == nil {
		t.Fatal("Expecting to have flag from CommandLine in c.Flags()")
	}

	resetCommandLineFlagSet()
}

// TestUseDeprecatedFlags checks,
// if cobra.Execute() prints a message, if a deprecated flag is used.
// Related to https://github.com/spf13/cobra/issues/463.
func TestUseDeprecatedFlags(t *testing.T) {
	c := &Command{Use: "c", RunE: emptyRun}
	Flags(c).BoolP("deprecated", "d", false, "deprecated flag")
	assertNoErr(t, Flags(c).MarkDeprecated("deprecated", "This flag is deprecated"))

	output, err := executeCommand(c, "c", "-d")
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	checkStringContains(t, output, "This flag is deprecated")
}

func TestTraverseWithParentFlags(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}
	Flags(rootCmd).String("str", "", "")
	Flags(rootCmd).BoolP("bool", "b", false, "")

	childCmd := &Command{Use: "child"}
	Flags(childCmd).Int("int", -1, "")

	rootCmd.Add(childCmd)

	c, args, err := Traverse(rootCmd, []string{"-b", "--str", "ok", "child", "--int"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 1 && args[0] != "--add" {
		t.Errorf("Wrong args: %v", args)
	}
	if name(c) != name(childCmd) {
		t.Errorf("Expected Root: %q, got: %q", name(childCmd), name(c))
	}
}

func TestTraverseNoParentFlags(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}
	Flags(rootCmd).String("foo", "", "foo things")

	childCmd := &Command{Use: "child"}
	Flags(childCmd).String("str", "", "")
	rootCmd.Add(childCmd)

	c, args, err := Traverse(rootCmd, []string{"child"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 0 {
		t.Errorf("Wrong args %v", args)
	}
	if name(c) != name(childCmd) {
		t.Errorf("Expected Root: %q, got: %q", name(childCmd), name(c))
	}
}

func TestTraverseWithBadParentFlags(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}

	childCmd := &Command{Use: "child"}
	Flags(childCmd).String("str", "", "")
	rootCmd.Add(childCmd)

	expected := "unknown flag: --str"

	c, _, err := Traverse(rootCmd, []string{"--str", "ok", "child"})
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error, %q, got %q", expected, err)
	}
	if c != nil {
		t.Errorf("Expected nil command")
	}
}

func TestTraverseWithBadChildFlag(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}
	Flags(rootCmd).String("str", "", "")

	childCmd := &Command{Use: "child"}
	rootCmd.Add(childCmd)

	// Expect no error because the last commands args shouldn't be parsed in
	// Traverse.
	c, args, err := Traverse(rootCmd, []string{"child", "--str"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 1 && args[0] != "--str" {
		t.Errorf("Wrong args: %v", args)
	}
	if name(c) != name(childCmd) {
		t.Errorf("Expected command %q, got: %q", name(childCmd), name(c))
	}
}

func TestTraverseWithTwoSubcommands(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}

	subCmd := &Command{Use: "sub", TraverseChildren: true}
	rootCmd.Add(subCmd)

	subsubCmd := &Command{
		Use: "subsub",
	}
	subCmd.Add(subsubCmd)

	c, _, err := Traverse(rootCmd, []string{"sub", "subsub"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if name(c) != name(subsubCmd) {
		t.Fatalf("Expected Root: %q, got %q", name(subsubCmd), name(c))
	}
}

// TestUpdateName checks if name(c) updates on changed c.Use.
// Related to https://github.com/spf13/cobra/pull/422#discussion_r143918343.
func TestUpdateName(t *testing.T) {
	c := &Command{Use: "name xyz"}
	originalName := name(c)

	c.Use = "changedName abc"
	if originalName == name(c) || name(c) != "changedName" {
		t.Error("name(c) should be updated on changed c.Use")
	}
}

type calledAsTestcase struct {
	args []string
	call string
	want string
	epm  bool
}

func (tc *calledAsTestcase) test(t *testing.T) {
	defer func(ov bool) { EnablePrefixMatching = ov }(EnablePrefixMatching)
	EnablePrefixMatching = tc.epm

	var called Commander
	run := func(c Commander, _ []string) error {
		t.Logf("called: %q", name(c))
		called = c
		return nil
	}

	parent := &Command{Use: "parent", RunE: run}
	child1 := &Command{Use: "child1", RunE: run, Aliases: []string{"this"}}
	child2 := &Command{Use: "child2", RunE: run, Aliases: []string{"that"}}

	parent.Add(child1)
	parent.Add(child2)
	parent.SetArgs(tc.args...)

	output := new(bytes.Buffer)
	log.SetOut(output)
	log.SetErr(output)

	_ = Execute(parent)

	if called == nil {
		if tc.call != "" {
			t.Errorf("missing expected call to Root: %s", tc.call)
		}
		return
	}

	if name(called) != tc.call {
		t.Errorf("called command == %q; Wanted %q", name(called), tc.call)
	} else if got := CalledAs(called); got != tc.want {
		t.Errorf("%s.CalledAs() == %q; Wanted: %q", tc.call, got, tc.want)
	}
}

func TestCalledAs(t *testing.T) {
	tests := map[string]calledAsTestcase{
		"find/no-args":            {nil, "parent", "parent", false},
		"find/real-name":          {[]string{"child1"}, "child1", "child1", false},
		"find/full-alias":         {[]string{"that"}, "child2", "that", false},
		"find/part-no-prefix":     {[]string{"thi"}, "", "", false},
		"find/part-alias":         {[]string{"thi"}, "child1", "this", true},
		"find/conflict":           {[]string{"th"}, "", "", true},
		"traverse/no-args":        {nil, "parent", "parent", false},
		"traverse/real-name":      {[]string{"child1"}, "child1", "child1", false},
		"traverse/full-alias":     {[]string{"that"}, "child2", "that", false},
		"traverse/part-no-prefix": {[]string{"thi"}, "", "", false},
		"traverse/part-alias":     {[]string{"thi"}, "child1", "this", true},
		"traverse/conflict":       {[]string{"th"}, "", "", true},
	}

	for name, tc := range tests {
		t.Run(name, tc.test)
	}
}

func TestFParseErrWhitelistBackwardCompatibility(t *testing.T) {
	c := &Command{Use: "c", RunE: emptyRun}
	Flags(c).BoolP("boola", "a", false, "a boolean flag")

	output, err := executeCommand(c, "c", "-a", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

func TestFParseErrWhitelistSameCommand(t *testing.T) {
	c := &Command{
		Use:  "c",
		RunE: emptyRun,
		FParseErrWhitelist: FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	Flags(c).BoolP("boola", "a", false, "a boolean flag")

	_, err := executeCommand(c, "c", "-a", "--unknown", "flag")
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestFParseErrWhitelistParentCommand(t *testing.T) {
	root := &Command{
		Use:  "root",
		RunE: emptyRun,
		FParseErrWhitelist: FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	c := &Command{
		Use:  "child",
		RunE: emptyRun,
	}
	Flags(c).BoolP("boola", "a", false, "a boolean flag")

	root.Add(c)

	output, err := executeCommand(root, "child", "-a", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

func TestFParseErrWhitelistChildCommand(t *testing.T) {
	root := &Command{
		Use:  "root",
		RunE: emptyRun,
	}

	c := &Command{
		Use:  "child",
		RunE: emptyRun,
		FParseErrWhitelist: FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	Flags(c).BoolP("boola", "a", false, "a boolean flag")

	root.Add(c)

	_, err := executeCommand(root, "child", "-a", "--unknown", "flag")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
}

func TestFParseErrWhitelistSiblingCommand(t *testing.T) {
	root := &Command{
		Use:  "root",
		RunE: emptyRun,
	}

	c := &Command{
		Use:  "child",
		RunE: emptyRun,
		FParseErrWhitelist: FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	Flags(c).BoolP("boola", "a", false, "a boolean flag")

	s := &Command{
		Use:  "sibling",
		RunE: emptyRun,
	}
	Flags(s).BoolP("boolb", "b", false, "a boolean flag")

	root.Add(c)
	root.Add(s)

	output, err := executeCommand(root, "sibling", "-b", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

func TestSetContext(t *testing.T) {
	type key struct{}
	val := "foobar"
	root := &Command{
		Use: "root",
		RunE: func(cmd Commander, args []string) error {
			key := cmd.Context().Value(key{})
			got, ok := key.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
			return nil
		},
	}

	ctx := context.WithValue(context.Background(), key{}, val)
	root.SetContext(ctx)
	err := Execute(root)
	if err != nil {
		t.Error(err)
	}
}

func TestSetContextPreRun(t *testing.T) {
	type key struct{}
	val := "barr"
	root := &Command{
		Use: "root",
		PreRunE: func(cmd Commander, args []string) error {
			ctx := context.WithValue(cmd.Context(), key{}, val)
			cmd.SetContext(ctx)
			return nil
		},
		RunE: func(cmd Commander, args []string) error {
			val := cmd.Context().Value(key{})
			got, ok := val.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
			return nil
		},
	}
	err := Execute(root)
	if err != nil {
		t.Error(err)
	}
}

func TestSetContextPreRunOverwrite(t *testing.T) {
	type key struct{}
	val := "blah"
	root := &Command{
		Use: "root",
		RunE: func(cmd Commander, args []string) error {
			key := cmd.Context().Value(key{})
			_, ok := key.(string)
			if ok {
				t.Error("key found in context when not expected")
			}
			return nil
		},
	}
	ctx := context.WithValue(context.Background(), key{}, val)
	root.SetContext(ctx)
	err := Execute(root)
	// err := root.ExecuteContext(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestSetContextPersistentPreRun(t *testing.T) {
	type key struct{}
	val := "barbar"
	root := &Command{
		Use: "root",
		PersistentPreRun: func(cmd Commander, args []string) {
			ctx := context.WithValue(cmd.Context(), key{}, val)
			cmd.SetContext(ctx)
		},
	}
	child := &Command{
		Use: "child",
		RunE: func(cmd Commander, args []string) error {
			key := cmd.Context().Value(key{})
			got, ok := key.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
			return nil
		},
	}
	root.Add(child)
	root.SetArgs("child")
	err := Execute(root)
	if err != nil {
		t.Error(err)
	}
}

const VersionFlag = "--version"
const HelpFlag = "--help"

func TestNoRootRunCommandExecutedWithVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Long: "Long description"}
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, err := executeCommand(rootCmd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringContains(t, output, VersionFlag)
}

func TestNoRootRunCommandExecutedWithoutVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description"}
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, err := executeCommand(rootCmd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringOmits(t, output, VersionFlag)
}

func TestHelpCommandExecutedWithVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Long: "Long description", RunE: emptyRun}
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringContains(t, output, VersionFlag)
}

func TestHelpCommandExecutedWithoutVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description", RunE: emptyRun}
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringOmits(t, output, VersionFlag)
}

func TestHelpflagCommandExecutedWithVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Long: "Long description", RunE: emptyRun}
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, err := executeCommand(rootCmd, HelpFlag)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringContains(t, output, VersionFlag)
}

func TestHelpflagCommandExecutedWithoutVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description", RunE: emptyRun}
	rootCmd.Add(&Command{Use: "child", RunE: emptyRun})

	output, err := executeCommand(rootCmd, HelpFlag)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringOmits(t, output, VersionFlag)
}

func TestFind(t *testing.T) {
	var foo, bar string
	root := &Command{
		Use: "root",
	}
	PersistentFlags(root).StringVarP(&foo, "foo", "f", "", "")
	PersistentFlags(root).StringVarP(&bar, "bar", "b", "something", "")

	child := &Command{
		Use: "child",
	}
	root.Add(child)

	testCases := []struct {
		args              []string
		expectedFoundArgs []string
	}{
		{
			[]string{"child"},
			[]string{},
		},
		{
			[]string{"child", "child"},
			[]string{"child"},
		},
		{
			[]string{"child", "foo", "child", "bar", "child", "baz", "child"},
			[]string{"foo", "child", "bar", "child", "baz", "child"},
		},
		{
			[]string{"-f", "child", "child"},
			[]string{"-f", "child"},
		},
		{
			[]string{"child", "-f", "child"},
			[]string{"-f", "child"},
		},
		{
			[]string{"-b", "child", "child"},
			[]string{"-b", "child"},
		},
		{
			[]string{"child", "-b", "child"},
			[]string{"-b", "child"},
		},
		{
			[]string{"child", "-b"},
			[]string{"-b"},
		},
		{
			[]string{"-b", "-f", "child", "child"},
			[]string{"-b", "-f", "child"},
		},
		{
			[]string{"-f", "child", "-b", "something", "child"},
			[]string{"-f", "child", "-b", "something"},
		},
		{
			[]string{"-f", "child", "child", "-b"},
			[]string{"-f", "child", "-b"},
		},
		{
			[]string{"-f=child", "-b=something", "child"},
			[]string{"-f=child", "-b=something"},
		},
		{
			[]string{"--foo", "child", "--bar", "something", "child"},
			[]string{"--foo", "child", "--bar", "something"},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.args), func(t *testing.T) {
			cmd, foundArgs, err := Find(root, tc.args)
			if err != nil {
				t.Fatal(err)
			}

			if cmd != child {
				t.Fatal("Expected cmd to be child, but it was not")
			}

			if !reflect.DeepEqual(tc.expectedFoundArgs, foundArgs) {
				t.Fatalf("Wrong args\nExpected: %v\nGot: %v", tc.expectedFoundArgs, foundArgs)
			}
		})
	}
}

func TestUnknownFlagShouldReturnSameErrorRegardlessOfArgPosition(t *testing.T) {
	testCases := [][]string{
		// {"--unknown", "--namespace", "foo", "child", "--bar"}, // FIXME: This test case fails, returning the error `unknown command "foo" for "root"` instead of the expected error `unknown flag: --unknown`
		{"--namespace", "foo", "--unknown", "child", "--bar"},
		{"--namespace", "foo", "child", "--unknown", "--bar"},
		{"--namespace", "foo", "child", "--bar", "--unknown"},

		{"--unknown", "--namespace=foo", "child", "--bar"},
		{"--namespace=foo", "--unknown", "child", "--bar"},
		{"--namespace=foo", "child", "--unknown", "--bar"},
		{"--namespace=foo", "child", "--bar", "--unknown"},

		{"--unknown", "--namespace=foo", "child", "--bar=true"},
		{"--namespace=foo", "--unknown", "child", "--bar=true"},
		{"--namespace=foo", "child", "--unknown", "--bar=true"},
		{"--namespace=foo", "child", "--bar=true", "--unknown"},
	}

	root := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	PersistentFlags(root).String("namespace", "", "a string flag")

	c := &Command{
		Use:  "child",
		RunE: emptyRun,
	}
	Flags(c).Bool("bar", false, "a boolean flag")

	root.Add(c)

	for _, tc := range testCases {
		t.Run(strings.Join(tc, " "), func(t *testing.T) {
			output, err := executeCommand(root, tc...)
			if err == nil {
				t.Error("expected unknown flag error")
			}
			checkStringContains(t, output, "unknown flag: --unknown")
		})
	}
}
