package command

import (
	"errors"
	"strings"
	"testing"

	gloo "github.com/gloo-foo/framework"
)

func TestArgvOf_KeepsFilePositionalsInOrder(t *testing.T) {
	// NewParameters converts string positionals to gloo.File; argvOf turns each
	// back into a token, preserving order.
	got := argvOf([]any{gloo.File("tr"), gloo.File("a-z"), gloo.File("A-Z")})
	want := []string{"tr", "a-z", "A-Z"}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestArgvOf_SkipsNonFilePositionals(t *testing.T) {
	// A positional that is not a File (e.g. an io.Reader handed to a File-typed
	// command) carries no argv token and is skipped.
	got := argvOf([]any{gloo.File("cat"), strings.NewReader("ignored")})
	if len(got) != 1 || got[0] != "cat" {
		t.Fatalf("got %q, want [cat]", got)
	}
}

func TestPlan_NoProgramReturnsErrNoCommand(t *testing.T) {
	// With nothing but flags, there is no program to run.
	name, args, err := Plan(ExecQuiet)
	if !errors.Is(err, ErrNoCommand) {
		t.Fatalf("err=%v, want ErrNoCommand", err)
	}
	if name != "" || args != nil {
		t.Fatalf("name=%q args=%q, want empty", name, args)
	}
}

func TestResolve_DirectWhenNoFlags(t *testing.T) {
	name, args := resolve([]string{"tr", "a-z", "A-Z"}, flags{})
	if name != "tr" {
		t.Fatalf("name=%q, want tr", name)
	}
	if len(args) != 2 || args[0] != "a-z" || args[1] != "A-Z" {
		t.Fatalf("args=%q", args)
	}
}

func TestResolve_ShellWhenWorkingDir(t *testing.T) {
	name, args := resolve([]string{"ls"}, flags{workingDir: ExecWorkingDir("/tmp")})
	if name != "sh" || len(args) != 2 || args[0] != "-c" {
		t.Fatalf("name=%q args=%q", name, args)
	}
	if args[1] != "cd '/tmp' && 'ls'" {
		t.Fatalf("cmdline=%q", args[1])
	}
}

func TestResolve_HonoursExplicitShell(t *testing.T) {
	// An explicit shell both forces shell routing and names the interpreter.
	name, args := resolve([]string{"echo", "hi"}, flags{shell: ExecShell("bash")})
	if name != "bash" || len(args) != 2 || args[0] != "-c" {
		t.Fatalf("name=%q args=%q", name, args)
	}
	if args[1] != "'echo' 'hi'" {
		t.Fatalf("cmdline=%q", args[1])
	}
}

func TestCommandLine_EnvAndQuoting(t *testing.T) {
	got := commandLine([]string{"echo", "a b"}, flags{envVars: []ExecEnvVar{"FOO=bar baz"}})
	want := "FOO='bar baz' 'echo' 'a b'"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCommandLine_QuietAndIgnoreErrors(t *testing.T) {
	got := commandLine([]string{"false"}, flags{isQuiet: ExecQuiet, shouldIgnoreErrors: ExecIgnoreErrors})
	want := "'false' 2>/dev/null || true"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("it's"); got != `'it'\''s'` {
		t.Fatalf("got %q", got)
	}
}

func TestEnvAssignment(t *testing.T) {
	if got := envAssignment("KEY=a b"); got != "KEY='a b'" {
		t.Fatalf("got %q", got)
	}
	if got := envAssignment("BARE"); got != "'BARE'" {
		t.Fatalf("got %q", got)
	}
}
