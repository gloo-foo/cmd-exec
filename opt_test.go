package command

import "testing"

func TestWith_FoldsEveryOptionKind(t *testing.T) {
	// Each option kind lands in its flag field; the receiver is a value, so the
	// original zero flags are untouched.
	var zero flags
	f, _ := zero.with(ExecWorkingDir("/tmp"))
	f, _ = f.with(ExecEnvVar("K=V"))
	f, _ = f.with(ExecShell("bash"))
	f, _ = f.with(ExecUseShell)
	f, _ = f.with(ExecIgnoreErrors)
	f, _ = f.with(ExecQuiet)

	if f.workingDir != "/tmp" || f.shell != "bash" {
		t.Fatalf("workingDir=%q shell=%q", f.workingDir, f.shell)
	}
	if len(f.envVars) != 1 || f.envVars[0] != "K=V" {
		t.Fatalf("envVars=%q", f.envVars)
	}
	if !bool(f.shouldUseShell) || !bool(f.shouldIgnoreErrors) || !bool(f.isQuiet) {
		t.Fatalf("bool flags=%v %v %v, want all true", f.shouldUseShell, f.shouldIgnoreErrors, f.isQuiet)
	}
	if zero.workingDir != "" || zero.envVars != nil {
		t.Fatalf("zero flags mutated: %+v", zero)
	}
}

func TestWith_ReportsNonOptions(t *testing.T) {
	// A plain string is the program or an argument, not an option.
	var f flags
	if _, isOption := f.with("ls"); isOption {
		t.Fatal("plain string classified as an option")
	}
	if _, isOption := f.with(ExecNoQuiet); !isOption {
		t.Fatal("disabled quiet flag not classified as an option")
	}
}

func TestFoldOptions_SeparatesOptionsFromArgv(t *testing.T) {
	// Options fold into flags; everything else passes through, in order.
	f, rest := foldOptions([]any{"echo", ExecQuiet, "hi", ExecWorkingDir("/tmp")})
	if !bool(f.isQuiet) || f.workingDir != "/tmp" {
		t.Fatalf("flags=%+v", f)
	}
	if len(rest) != 2 || rest[0] != "echo" || rest[1] != "hi" {
		t.Fatalf("rest=%v", rest)
	}
}
