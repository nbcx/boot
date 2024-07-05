package boot

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	var rootCmdArgs []string
	root := &Command{
		Use: "root",
		// Args: ExactArgs(2),
		Run: func(_ Commander, args []string) { rootCmdArgs = args },
	}
	aCmd := &Command{Use: "a", Args: NoArgs, Run: func(cmd Commander, args []string) { fmt.Println("a....") }}
	bCmd := &Command{Use: "b", Args: NoArgs, Run: emptyRun}
	root.Add(aCmd, bCmd)

	// buf := new(bytes.Buffer)
	// root.SetOut(buf)
	// root.SetErr(buf)
	root.SetArgs([]string{"a"})

	err := root.ExecuteX()

	fmt.Println(err, rootCmdArgs)

}
