package command

type (
	// ExecWorkingDir runs the command in the named directory.
	ExecWorkingDir string
	// ExecEnvVar adds one KEY=VALUE environment variable (repeatable).
	ExecEnvVar string
	// ExecShell names the shell interpreter to route the command through.
	ExecShell string
)

// execShellFlag forces (or declines) shell routing without naming a shell.
type execShellFlag bool

const (
	ExecUseShell execShellFlag = true
	ExecNoShell  execShellFlag = false
)

// execIgnoreErrorsFlag makes the command succeed even on a non-zero exit.
type execIgnoreErrorsFlag bool

const (
	ExecIgnoreErrors   execIgnoreErrorsFlag = true
	ExecNoIgnoreErrors execIgnoreErrorsFlag = false
)

// execQuietFlag discards the program's stderr.
type execQuietFlag bool

const (
	ExecQuiet   execQuietFlag = true
	ExecNoQuiet execQuietFlag = false
)

// flags is the folded option set for an exec run. The zero value runs the
// program directly with the inherited environment.
type flags struct {
	workingDir         ExecWorkingDir
	shell              ExecShell
	envVars            []ExecEnvVar
	shouldUseShell     execShellFlag
	shouldIgnoreErrors execIgnoreErrorsFlag
	isQuiet            execQuietFlag
}

// with folds one option value into the flag set, reporting whether the value
// was an exec option. Anything else (the program and its arguments) is left
// for positional classification.
func (f flags) with(o any) (flags, bool) {
	switch v := o.(type) {
	case ExecWorkingDir:
		f.workingDir = v
	case ExecEnvVar:
		f.envVars = append(f.envVars, v)
	case ExecShell:
		f.shell = v
	case execShellFlag:
		f.shouldUseShell = v
	case execIgnoreErrorsFlag:
		f.shouldIgnoreErrors = v
	case execQuietFlag:
		f.isQuiet = v
	default:
		return f, false
	}
	return f, true
}

// foldOptions applies every recognized exec option to a zero flags value and
// returns the leftover arguments for the framework's positional classification.
func foldOptions(opts []any) (flags, []any) {
	var f flags
	var rest []any
	for _, o := range opts {
		next, isOption := f.with(o)
		if !isOption {
			rest = append(rest, o)
			continue
		}
		f = next
	}
	return f, rest
}
