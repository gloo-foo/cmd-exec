package command

type (
	ExecWorkingDir string
	ExecEnvVar     string
	ExecShell      string
)

type execShellFlag bool

const (
	ExecUseShell execShellFlag = true
	ExecNoShell  execShellFlag = false
)

type execIgnoreErrorsFlag bool

const (
	ExecIgnoreErrors   execIgnoreErrorsFlag = true
	ExecNoIgnoreErrors execIgnoreErrorsFlag = false
)

type execQuietFlag bool

const (
	ExecQuiet   execQuietFlag = true
	ExecNoQuiet execQuietFlag = false
)

type flags struct {
	workingDir   ExecWorkingDir
	shell        ExecShell
	envVars      []ExecEnvVar
	useShell     execShellFlag
	ignoreErrors execIgnoreErrorsFlag
	quiet        execQuietFlag
}

func (w ExecWorkingDir) Configure(flags *flags)       { flags.workingDir = w }
func (e ExecEnvVar) Configure(flags *flags)           { flags.envVars = append(flags.envVars, e) }
func (s ExecShell) Configure(flags *flags)            { flags.shell = s }
func (s execShellFlag) Configure(flags *flags)        { flags.useShell = s }
func (i execIgnoreErrorsFlag) Configure(flags *flags) { flags.ignoreErrors = i }
func (q execQuietFlag) Configure(flags *flags)        { flags.quiet = q }
