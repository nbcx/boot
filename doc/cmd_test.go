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

package doc

import (
	"strings"
	"testing"

	"github.com/nbcx/boot"
)

func emptyRun(boot.Commander, []string) error { return nil }

func init() {
	rootCmd.PersistentFlags().StringP("rootflag", "r", "two", "")
	rootCmd.PersistentFlags().StringP("strtwo", "t", "two", "help message for parent flag strtwo")

	echoCmd.PersistentFlags().StringP("strone", "s", "one", "help message for flag strone")
	echoCmd.PersistentFlags().BoolP("persistentbool", "p", false, "help message for flag persistentbool")
	boot.Flags(echoCmd).IntP("intone", "i", 123, "help message for flag intone")
	boot.Flags(echoCmd).BoolP("boolone", "b", true, "help message for flag boolone")

	timesCmd.PersistentFlags().StringP("strtwo", "t", "2", "help message for child flag strtwo")
	boot.Flags(timesCmd).IntP("inttwo", "j", 234, "help message for flag inttwo")
	boot.Flags(timesCmd).BoolP("booltwo", "c", false, "help message for flag booltwo")

	printCmd.PersistentFlags().StringP("strthree", "s", "three", "help message for flag strthree")
	boot.Flags(printCmd).IntP("intthree", "i", 345, "help message for flag intthree")
	boot.Flags(printCmd).BoolP("boolthree", "b", true, "help message for flag boolthree")

	echoCmd.Add(timesCmd, echoSubCmd, deprecatedCmd)
	rootCmd.Add(printCmd, echoCmd, dummyCmd)
}

var rootCmd = &boot.Root{
	Use:   "root",
	Short: "Root short description",
	Long:  "Root long description",
	RunE:  emptyRun,
}

var echoCmd = &boot.Root{
	Use:     "echo [string to echo]",
	Aliases: []string{"say"},
	Short:   "Echo anything to the screen",
	Long:    "an utterly useless command for testing",
	Example: "Just run cobra-test echo",
}

var echoSubCmd = &boot.Root{
	Use:   "echosub [string to print]",
	Short: "second sub command for echo",
	Long:  "an absolutely utterly useless command for testing gendocs!.",
	RunE:  emptyRun,
}

var timesCmd = &boot.Root{
	Use:        "times [# times] [string to echo]",
	SuggestFor: []string{"counts"},
	Short:      "Echo anything to the screen more times",
	Long:       `a slightly useless command for testing.`,
	RunE:       emptyRun,
}

var deprecatedCmd = &boot.Root{
	Use:        "deprecated [can't do anything here]",
	Short:      "A command which is deprecated",
	Long:       `an absolutely utterly useless command for testing deprecation!.`,
	Deprecated: "Please use echo instead",
}

var printCmd = &boot.Root{
	Use:   "print [string to print]",
	Short: "Print anything to the screen",
	Long:  `an absolutely utterly useless command for testing.`,
}

var dummyCmd = &boot.Root{
	Use:   "dummy [action]",
	Short: "Performs a dummy action",
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
