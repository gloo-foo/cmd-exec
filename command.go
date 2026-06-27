package command

import (
	"context"
	"strings"

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
	return patterns.Subprocess(name, args...)
}

// Plan resolves parameters into the concrete program and arguments Exec would
// run, without spawning anything. It is the dry-run / preview counterpart of
// Exec: a caller can log or inspect the exact command line first. It returns
// ErrNoCommand when no program is named.
//
// When no flag needs shell features the program and its arguments are returned
// directly; otherwise the program is wrapped in a POSIX shell invocation
// ("sh -c <command line>") so the flags take effect.
func Plan(parameters ...any) (name string, args []string, err error) {
	params := gloo.NewParameters[gloo.File, flags](parameters...)
	argv := argvOf(params.Positional)
	if len(argv) == 0 {
		return "", nil, ErrNoCommand
	}
	name, args = resolve(argv, params.Flags)
	return name, args, nil
}

// argvOf extracts the program and its arguments from the classified positional
// values. NewParameters[gloo.File, flags] converts every string positional to a
// gloo.File, so the program name and each argument arrive as File values; other
// positional kinds (e.g. an io.Reader) carry no argv token and are skipped.
func argvOf(positional []any) []string {
	argv := make([]string, 0, len(positional))
	for _, p := range positional {
		if f, ok := p.(gloo.File); ok {
			argv = append(argv, string(f))
		}
	}
	return argv
}

// resolve turns argv plus flags into the concrete program and arguments to run.
// Flags that change execution (working dir, env vars, quiet, ignore-errors, or
// an explicit shell) route the command through a shell; otherwise the program
// runs directly.
func resolve(argv []string, f flags) (name string, args []string) {
	needShell := bool(f.useShell) ||
		string(f.shell) != "" ||
		string(f.workingDir) != "" ||
		len(f.envVars) > 0 ||
		bool(f.quiet) ||
		bool(f.ignoreErrors)

	if !needShell {
		return argv[0], argv[1:]
	}

	shell := string(f.shell)
	if shell == "" {
		shell = "sh"
	}
	return shell, []string{"-c", commandLine(argv, f)}
}

// commandLine builds a POSIX shell command line that applies the flag-driven
// behavior to argv. Environment variables are added on top of the inherited
// environment.
func commandLine(argv []string, f flags) string {
	var b strings.Builder

	if dir := string(f.workingDir); dir != "" {
		b.WriteString("cd ")
		b.WriteString(shellQuote(dir))
		b.WriteString(" && ")
	}
	for _, e := range f.envVars {
		b.WriteString(envAssignment(string(e)))
		b.WriteByte(' ')
	}

	parts := make([]string, len(argv))
	for i, a := range argv {
		parts[i] = shellQuote(a)
	}
	b.WriteString(strings.Join(parts, " "))

	if bool(f.quiet) {
		b.WriteString(" 2>/dev/null")
	}
	if bool(f.ignoreErrors) {
		b.WriteString(" || true")
	}
	return b.String()
}

// envAssignment quotes the value half of a KEY=VALUE assignment for the shell.
func envAssignment(kv string) string {
	if i := strings.IndexByte(kv, '='); i >= 0 {
		return kv[:i+1] + shellQuote(kv[i+1:])
	}
	return shellQuote(kv)
}

// shellQuote wraps s in single quotes, safely escaping embedded single quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// errorCommand returns a command that ignores its input and fails with err.
func errorCommand(err error) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.Generate(ctx, func(_ context.Context, _ func([]byte) bool, sendErr func(error)) {
			sendErr(err)
		})
	})
}
