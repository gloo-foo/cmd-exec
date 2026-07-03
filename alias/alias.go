// Package alias provides unprefixed type aliases for exec command flags.
// This allows users to import and use shorter names:
//
//	import "github.com/gloo-foo/cmd-exec/alias"
//	exec.Exec("ls", "-la", alias.WorkingDir("/tmp"))
package alias

import command "github.com/gloo-foo/cmd-exec"

// working directory
type WorkingDir = command.ExecWorkingDir

// environment variable
type EnvVar = command.ExecEnvVar

// shell to use
type Shell = command.ExecShell

// use shell
const UseShell = command.ExecUseShell

// default: direct execution
const NoShell = command.ExecNoShell

// ignore errors
const IgnoreErrors = command.ExecIgnoreErrors

// default: fail on error
const NoIgnoreErrors = command.ExecNoIgnoreErrors

// -q flag: quiet mode
const Quiet = command.ExecQuiet

// default: not quiet
const NoQuiet = command.ExecNoQuiet
