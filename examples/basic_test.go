package exec_test

import (
	"fmt"

	exec "github.com/gloo-foo/cmd-exec"
)

// ExamplePlan shows how to preview the exact command line Exec will run, without
// spawning anything. Plain string arguments are the program and its arguments;
// flags that need shell features route the command through a POSIX shell.
func ExamplePlan() {
	name, args, _ := exec.Plan("tr", "a-z", "A-Z")
	fmt.Println(name, args)

	name, args, _ = exec.Plan("ls", exec.ExecWorkingDir("/tmp"))
	fmt.Println(name, args)
	// Output:
	// tr [a-z A-Z]
	// sh [-c cd '/tmp' && 'ls']
}
