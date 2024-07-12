package boot

import (
	"fmt"
	"testing"
)

type b struct {
	Command
}

func (p *b) GetUse() string {
	return "b"
}

func (p *b) Run(args []string) error {
	fmt.Println("b   >>>", args)
	return nil
}

func TestMain(t *testing.T) {
	var rootCmdArgs []string
	root := &Command{
		Use:  "root",
		Args: RangeArgs(0, 2),
		RunE: func(_ Commander, args []string) error {
			rootCmdArgs = args
			fmt.Println("root >>", args)
			return nil
		},
	}
	aCmd := &Command{Use: "a", Args: NoArgs, RunE: func(cmd Commander, args []string) error { fmt.Println("a...."); return nil }}
	bCmd := &b{Command: Command{Args: RangeArgs(0, 2)}}
	root.Add(aCmd, bCmd)

	// buf := new(bytes.Buffer)
	// root.SetOut(buf)
	// root.SetErr(buf)
	// root.SetArgs("b", "jj", "cc")
	// root.SetArgs("one", "two")
	// root.SetArgs("--help")
	// root.SetArgs("b", "--help")

	// aCmd.SetArgs("a")
	err := Execute(root)
	if err != nil {
		fmt.Println("execute x", err, rootCmdArgs)
	}

}
