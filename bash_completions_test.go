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
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func checkOmit(t *testing.T, found, unexpected string) {
	if strings.Contains(found, unexpected) {
		t.Errorf("Got: %q\nBut should not have!\n", unexpected)
	}
}

func check(t *testing.T, found, expected string) {
	if !strings.Contains(found, expected) {
		t.Errorf("Expecting to contain: \n %q\nGot:\n %q\n", expected, found)
	}
}

func checkNumOccurrences(t *testing.T, found, expected string, expectedOccurrences int) {
	numOccurrences := strings.Count(found, expected)
	if numOccurrences != expectedOccurrences {
		t.Errorf("Expecting to contain %d occurrences of: \n %q\nGot %d:\n %q\n", expectedOccurrences, expected, numOccurrences, found)
	}
}

func checkRegex(t *testing.T, found, pattern string) {
	matched, err := regexp.MatchString(pattern, found)
	if err != nil {
		t.Errorf("Error thrown performing MatchString: \n %s\n", err)
	}
	if !matched {
		t.Errorf("Expecting to match: \n %q\nGot:\n %q\n", pattern, found)
	}
}

func runShellCheck(s string) error {
	cmd := exec.Command("shellcheck", "-s", "bash", "-", "-e",
		"SC2034", // PREFIX appears unused. Verify it or export it.
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		_, err := stdin.Write([]byte(s))
		CheckErr(err)

		stdin.Close()
	}()

	return cmd.Run()
}

// World worst custom function, just keep telling you to enter hello!
const bashCompletionFunc = `__root_custom_func() {
	COMPREPLY=( "hello" )
}
`

func TestBashCompletions(t *testing.T) {
	rootCmd := &Root{
		Use:                    "root",
		ArgAliases:             []string{"pods", "nodes", "services", "replicationcontrollers", "po", "no", "svc", "rc"},
		ValidArgs:              []string{"pod", "node", "service", "replicationcontroller"},
		BashCompletionFunction: bashCompletionFunc,
		RunE:                   emptyRun,
	}
	Flags(rootCmd).IntP("introot", "i", -1, "help message for flag introot")
	assertNoErr(t, MarkFlagRequired(rootCmd, "introot"))

	// Filename.
	Flags(rootCmd).String("filename", "", "Enter a filename")
	assertNoErr(t, rootCmd.MarkFlagFilename("filename", "json", "yaml", "yml"))

	// Persistent filename.
	PersistentFlags(rootCmd).String("persistent-filename", "", "Enter a filename")
	assertNoErr(t, MarkPersistentFlagFilename(rootCmd, "persistent-filename"))
	assertNoErr(t, MarkPersistentFlagRequired(rootCmd, "persistent-filename"))

	// Filename extensions.
	Flags(rootCmd).String("filename-ext", "", "Enter a filename (extension limited)")
	assertNoErr(t, rootCmd.MarkFlagFilename("filename-ext"))
	Flags(rootCmd).String("custom", "", "Enter a filename (extension limited)")
	assertNoErr(t, rootCmd.MarkFlagCustom("custom", "__complete_custom"))

	// Subdirectories in a given directory.
	Flags(rootCmd).String("theme", "", "theme to use (located in /themes/THEMENAME/)")
	assertNoErr(t, Flags(rootCmd).SetAnnotation("theme", BashCompSubdirsInDir, []string{"themes"}))

	// For two word flags check
	Flags(rootCmd).StringP("two", "t", "", "this is two word flags")
	Flags(rootCmd).BoolP("two-w-default", "T", false, "this is not two word flags")

	echoCmd := &Root{
		Use:     "echo [string to echo]",
		Aliases: []string{"say"},
		Short:   "Echo anything to the screen",
		Long:    "an utterly useless command for testing.",
		Example: "Just run cobra-test echo",
		RunE:    emptyRun,
	}

	Flags(echoCmd).String("filename", "", "Enter a filename")
	assertNoErr(t, echoCmd.MarkFlagFilename("filename", "json", "yaml", "yml"))
	Flags(echoCmd).String("config", "", "config to use (located in /config/PROFILE/)")
	assertNoErr(t, Flags(echoCmd).SetAnnotation("config", BashCompSubdirsInDir, []string{"config"}))

	printCmd := &Root{
		Use:   "print [string to print]",
		Args:  MinimumNArgs(1),
		Short: "Print anything to the screen",
		Long:  "an absolutely utterly useless command for testing.",
		RunE:  emptyRun,
	}

	deprecatedCmd := &Root{
		Use:        "deprecated [can't do anything here]",
		Args:       NoArgs,
		Short:      "A command which is deprecated",
		Long:       "an absolutely utterly useless command for testing deprecation!.",
		Deprecated: "Please use echo instead",
		RunE:       emptyRun,
	}

	colonCmd := &Root{
		Use:  "cmd:colon",
		RunE: emptyRun,
	}

	timesCmd := &Root{
		Use:        "times [# times] [string to echo]",
		SuggestFor: []string{"counts"},
		Args:       OnlyValidArgs,
		ValidArgs:  []string{"one", "two", "three", "four"},
		Short:      "Echo anything to the screen more times",
		Long:       "a slightly useless command for testing.",
		RunE:       emptyRun,
	}

	echoCmd.Add(timesCmd)
	rootCmd.Add(echoCmd, printCmd, deprecatedCmd, colonCmd)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenBashCompletion(buf))
	output := buf.String()

	check(t, output, "_root")
	check(t, output, "_root_echo")
	check(t, output, "_root_echo_times")
	check(t, output, "_root_print")
	check(t, output, "_root_cmd__colon")

	// check for required flags
	check(t, output, `must_have_one_flag+=("--introot=")`)
	check(t, output, `must_have_one_flag+=("--persistent-filename=")`)
	// check for custom completion function with both qualified and unqualified name
	checkNumOccurrences(t, output, `__custom_func`, 2)      // 1. check existence, 2. invoke
	checkNumOccurrences(t, output, `__root_custom_func`, 3) // 1. check existence, 2. invoke, 3. actual definition
	// check for custom completion function body
	check(t, output, `COMPREPLY=( "hello" )`)
	// check for required nouns
	check(t, output, `must_have_one_noun+=("pod")`)
	// check for noun aliases
	check(t, output, `noun_aliases+=("pods")`)
	check(t, output, `noun_aliases+=("rc")`)
	checkOmit(t, output, `must_have_one_noun+=("pods")`)
	// check for filename extension flags
	check(t, output, `flags_completion+=("_filedir")`)
	// check for filename extension flags
	check(t, output, `must_have_one_noun+=("three")`)
	// check for filename extension flags
	check(t, output, fmt.Sprintf(`flags_completion+=("__%s_handle_filename_extension_flag json|yaml|yml")`, name(rootCmd)))
	// check for filename extension flags in a subcommand
	checkRegex(t, output, fmt.Sprintf(`_root_echo\(\)\n{[^}]*flags_completion\+=\("__%s_handle_filename_extension_flag json\|yaml\|yml"\)`, name(rootCmd)))
	// check for custom flags
	check(t, output, `flags_completion+=("__complete_custom")`)
	// check for subdirs_in_dir flags
	check(t, output, fmt.Sprintf(`flags_completion+=("__%s_handle_subdirs_in_dir_flag themes")`, name(rootCmd)))
	// check for subdirs_in_dir flags in a subcommand
	checkRegex(t, output, fmt.Sprintf(`_root_echo\(\)\n{[^}]*flags_completion\+=\("__%s_handle_subdirs_in_dir_flag config"\)`, name(rootCmd)))

	// check two word flags
	check(t, output, `two_word_flags+=("--two")`)
	check(t, output, `two_word_flags+=("-t")`)
	checkOmit(t, output, `two_word_flags+=("--two-w-default")`)
	checkOmit(t, output, `two_word_flags+=("-T")`)

	// check local nonpersistent flag
	check(t, output, `local_nonpersistent_flags+=("--two")`)
	check(t, output, `local_nonpersistent_flags+=("--two=")`)
	check(t, output, `local_nonpersistent_flags+=("-t")`)
	check(t, output, `local_nonpersistent_flags+=("--two-w-default")`)
	check(t, output, `local_nonpersistent_flags+=("-T")`)

	checkOmit(t, output, name(deprecatedCmd))

	// If available, run shellcheck against the script.
	if err := exec.Command("which", "shellcheck").Run(); err != nil {
		return
	}
	if err := runShellCheck(output); err != nil {
		t.Fatalf("shellcheck failed: %v", err)
	}
}

func TestBashCompletionHiddenFlag(t *testing.T) {
	c := &Root{Use: "c", RunE: emptyRun}

	const flagName = "hiddenFlag"
	Flags(c).Bool(flagName, false, "")
	assertNoErr(t, Flags(c).MarkHidden(flagName))

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletion(buf))
	output := buf.String()

	if strings.Contains(output, flagName) {
		t.Errorf("Expected completion to not include %q flag: Got %v", flagName, output)
	}
}

func TestBashCompletionDeprecatedFlag(t *testing.T) {
	c := &Root{Use: "c", RunE: emptyRun}

	const flagName = "deprecated-flag"
	Flags(c).Bool(flagName, false, "")
	assertNoErr(t, Flags(c).MarkDeprecated(flagName, "use --not-deprecated instead"))

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletion(buf))
	output := buf.String()

	if strings.Contains(output, flagName) {
		t.Errorf("expected completion to not include %q flag: Got %v", flagName, output)
	}
}

func TestBashCompletionTraverseChildren(t *testing.T) {
	c := &Root{Use: "c", RunE: emptyRun, TraverseChildren: true}

	Flags(c).StringP("string-flag", "s", "", "string flag")
	Flags(c).BoolP("bool-flag", "b", false, "bool flag")

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletion(buf))
	output := buf.String()

	// check that local nonpersistent flag are not set since we have TraverseChildren set to true
	checkOmit(t, output, `local_nonpersistent_flags+=("--string-flag")`)
	checkOmit(t, output, `local_nonpersistent_flags+=("--string-flag=")`)
	checkOmit(t, output, `local_nonpersistent_flags+=("-s")`)
	checkOmit(t, output, `local_nonpersistent_flags+=("--bool-flag")`)
	checkOmit(t, output, `local_nonpersistent_flags+=("-b")`)
}

func TestBashCompletionNoActiveHelp(t *testing.T) {
	c := &Root{Use: "c", RunE: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletion(buf))
	output := buf.String()

	// check that active help is being disabled
	activeHelpVar := activeHelpEnvVar(name(c))
	check(t, output, fmt.Sprintf("%s=0", activeHelpVar))
}
