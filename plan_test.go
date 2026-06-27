package command

import "testing"

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
	got := commandLine([]string{"false"}, flags{quiet: ExecQuiet, ignoreErrors: ExecIgnoreErrors})
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
