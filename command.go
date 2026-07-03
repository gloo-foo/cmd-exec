package command

import (
	"context"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

// Error is the sentinel error type for this package. Every error the package can
// emit is a constant of this type, so callers match with errors.Is.
type Error string

// Error renders the sentinel's message.
func (e Error) Error() string { return string(e) }

// ErrNoCommand is returned by Exec when no program is given to run.
const ErrNoCommand Error = "exec: no command specified"

// Exec returns a Command that runs an external program as a pipeline stage: the
// input stream is written to the program's stdin and its stdout becomes the
// output stream (exactly the role of a command in a shell pipe).
//
// The program and its arguments are the positional string parameters:
//
//	Exec("tr", "a-z", "A-Z")
//
// Flags (see opt.go and the alias package):
//   - ExecWorkingDir(dir): run in dir
//   - ExecEnvVar("K=V"): add an environment variable (repeatable)
//   - ExecShell(sh) / ExecUseShell: run the command line through a shell
//   - ExecQuiet: discard the program's stderr
//   - ExecIgnoreErrors: succeed even if the program exits non-zero
//
// When no flag needs shell features the program is executed directly; otherwise
// it is run through a POSIX shell so the flags can take effect.
//
// Note: a generic exec runs arbitrary binaries, which the framework otherwise
// discourages in favor of per-tool Subprocess wrappers. It is offered here as an
// explicit, opt-in escape hatch.
func Exec(parameters ...any) gloo.Command[[]byte, []byte] {
	name, args, err := Plan(parameters...)
	if err != nil {
		return errorCommand(err)
	}
	return patterns.Subprocess(patterns.ProcessName(name), processArgs(args)...)
}

// processArgs converts the planned argument strings into the patterns.ProcessArg
// vector patterns.Subprocess consumes.
func processArgs(args []string) []patterns.ProcessArg {
	out := make([]patterns.ProcessArg, len(args))
	for i, a := range args {
		out[i] = patterns.ProcessArg(a)
	}
	return out
}

// errorCommand returns a command that ignores its input and fails with err.
func errorCommand(err error) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.Generate(ctx, func(_ context.Context, _ func([]byte) bool, sendErr func(error)) {
			sendErr(err)
		})
	})
}
