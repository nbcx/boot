package boot

import (
	"fmt"
	"testing"
)

type b struct {
	Default
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
	root := &Root{
		Use:  "root",
		Args: RangeArgs(0, 2),
		RunE: func(_ Commander, args []string) error {
			rootCmdArgs = args
			fmt.Println("root >>", args)
			return nil
		},
	}
	aCmd := &Root{Use: "a", Args: NoArgs, RunE: func(cmd Commander, args []string) error { fmt.Println("a...."); return nil }}
	bCmd := &b{Default: Default{Args: RangeArgs(0, 2)}} // &Root{Use: "b", Args: NoArgs, RunE: emptyRun}
	root.Add(aCmd, bCmd)

	// buf := new(bytes.Buffer)
	// root.SetOut(buf)
	// root.SetErr(buf)
	// root.SetArgs("b", "jj", "cc")
	// root.SetArgs("one", "two")
	root.SetArgs("--help")
	// root.SetArgs("b", "--help")

	err := root.Execute()
	if err != nil {
		fmt.Println("execute x", err, rootCmdArgs)
	}

}
