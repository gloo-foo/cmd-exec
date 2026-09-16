package alias_test

import (
	"slices"
	"testing"

	command "github.com/gloo-foo/cmd-exec"
	"github.com/gloo-foo/cmd-exec/alias"
)

// The alias package re-exports exec's flag types and constants under unprefixed
// names. A mis-wired re-export (say, IgnoreErrors bound to the disabled constant,
// or WorkingDir aliased to the wrong type) compiles cleanly, so only behavior can
// prove the wiring. Each test feeds an alias symbol to command.Plan and asserts
// the resolved program/arguments it must produce — without spawning anything.

func assertPlan(t *testing.T, gotName string, gotArgs []string, gotErr error, wantName string, wantArgs []string) {
	t.Helper()
	if gotErr != nil {
		t.Fatalf("unexpected error: %v", gotErr)
	}
	if gotName != wantName || !slices.Equal(gotArgs, wantArgs) {
		t.Fatalf("got %q %q, want %q %q", gotName, gotArgs, wantName, wantArgs)
	}
}

func TestAlias_DefaultRunsDirectly(t *testing.T) {
	// No flags: the program runs directly, no shell wrapper.
	name, args, err := command.Plan("tr", "a-z", "A-Z")
	assertPlan(t, name, args, err, "tr", []string{"a-z", "A-Z"})
}

func TestAlias_WorkingDirRoutesThroughShell(t *testing.T) {
	// WorkingDir must alias ExecWorkingDir: a cd prefix appears in the shell line.
	name, args, err := command.Plan("ls", alias.WorkingDir("/tmp"))
	assertPlan(t, name, args, err, "sh", []string{"-c", "cd '/tmp' && 'ls'"})
}

func TestAlias_EnvVarRoutesThroughShell(t *testing.T) {
	// EnvVar must alias ExecEnvVar: a KEY=VALUE assignment precedes the program.
	name, args, err := command.Plan("printenv", "GREETING", alias.EnvVar("GREETING=hi"))
	assertPlan(t, name, args, err, "sh", []string{"-c", "GREETING='hi' 'printenv' 'GREETING'"})
}

func TestAlias_ShellNamesTheInterpreter(t *testing.T) {
	// Shell must alias ExecShell: the named interpreter becomes the program.
	name, args, err := command.Plan("echo", "hi", alias.Shell("bash"))
	assertPlan(t, name, args, err, "bash", []string{"-c", "'echo' 'hi'"})
}

func TestAlias_UseShellForcesShellRouting(t *testing.T) {
	// UseShell must be the enabled constant: it forces the sh wrapper on its own.
	name, args, err := command.Plan("echo", "hi", alias.UseShell)
	assertPlan(t, name, args, err, "sh", []string{"-c", "'echo' 'hi'"})
}

func TestAlias_NoShellLeavesDirectExecution(t *testing.T) {
	// NoShell must be the disabled constant: it behaves like passing no flag.
	name, args, err := command.Plan("echo", "hi", alias.NoShell)
	assertPlan(t, name, args, err, "echo", []string{"hi"})
}

func TestAlias_QuietRedirectsStderr(t *testing.T) {
	// Quiet must be the enabled constant: it appends a stderr redirect.
	name, args, err := command.Plan("noisy", alias.Quiet)
	assertPlan(t, name, args, err, "sh", []string{"-c", "'noisy' 2>/dev/null"})
}

func TestAlias_NoQuietLeavesDirectExecution(t *testing.T) {
	// NoQuiet must be the disabled constant: it behaves like passing no flag.
	name, args, err := command.Plan("noisy", alias.NoQuiet)
	assertPlan(t, name, args, err, "noisy", []string{})
}

func TestAlias_IgnoreErrorsAppendsTrueGuard(t *testing.T) {
	// IgnoreErrors must be the enabled constant: it appends `|| true`.
	name, args, err := command.Plan("false", alias.IgnoreErrors)
	assertPlan(t, name, args, err, "sh", []string{"-c", "'false' || true"})
}

func TestAlias_NoIgnoreErrorsLeavesDirectExecution(t *testing.T) {
	// NoIgnoreErrors must be the disabled constant: it behaves like no flag.
	name, args, err := command.Plan("false", alias.NoIgnoreErrors)
	assertPlan(t, name, args, err, "false", []string{})
}
