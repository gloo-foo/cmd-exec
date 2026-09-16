package command

import (
	"strings"

	gloo "github.com/gloo-foo/framework"
)

// Plan resolves parameters into the concrete program and arguments Exec would
// run, without spawning anything. It is the dry-run / preview counterpart of
// Exec: a caller can log or inspect the exact command line first. It returns
// ErrNoCommand when no program is named.
//
// When no flag needs shell features the program and its arguments are returned
// directly; otherwise the program is wrapped in a POSIX shell invocation
// ("sh -c <command line>") so the flags take effect.
func Plan(parameters ...any) (name string, args []string, err error) {
	f, rest := foldOptions(parameters)
	params := gloo.NewParameters[gloo.File, struct{}](rest...)
	argv := argvOf(params.Positional)
	if len(argv) == 0 {
		return "", nil, ErrNoCommand
	}
	name, args = resolve(argv, f)
	return name, args, nil
}

// argvOf extracts the program and its arguments from the classified positional
// values. NewParameters converts every string positional to a gloo.File, so the
// program name and each argument arrive as File values; other positional kinds
// (e.g. an io.Reader) carry no argv token and are skipped.
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
	needShell := bool(f.shouldUseShell) ||
		string(f.shell) != "" ||
		string(f.workingDir) != "" ||
		len(f.envVars) > 0 ||
		bool(f.isQuiet) ||
		bool(f.shouldIgnoreErrors)

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
	parts := make([]string, len(argv))
	for i, a := range argv {
		parts[i] = shellQuote(shellWord(a))
	}

	line := envPrefix(f) + strings.Join(parts, " ")
	if bool(f.isQuiet) {
		line += " 2>/dev/null"
	}
	if bool(f.shouldIgnoreErrors) {
		line += " || true"
	}
	return line
}

// envPrefix builds the leading "cd <dir> && K=V " portion of the shell command
// line from the working-directory and environment-variable flags.
func envPrefix(f flags) string {
	prefix := ""
	if dir := string(f.workingDir); dir != "" {
		prefix = "cd " + shellQuote(shellWord(dir)) + " && "
	}
	for _, e := range f.envVars {
		prefix += envAssignment(e) + " "
	}
	return prefix
}

// envAssignment quotes the value half of a KEY=VALUE assignment for the shell.
func envAssignment(kv ExecEnvVar) string {
	if i := strings.IndexByte(string(kv), '='); i >= 0 {
		return string(kv)[:i+1] + shellQuote(shellWord(string(kv)[i+1:]))
	}
	return shellQuote(shellWord(kv))
}

// shellWord is a single token destined for a POSIX shell command line.
type shellWord string

// shellQuote wraps the word in single quotes, safely escaping embedded single
// quotes.
func shellQuote(w shellWord) string {
	return "'" + strings.ReplaceAll(string(w), "'", `'\''`) + "'"
}
