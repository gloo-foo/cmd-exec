package command

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	gloo "github.com/gloo-foo/framework"
)

// runError executes cmd over empty input and returns the first error it emits.
// It never spawns a real process: the only command executed here is the
// error command from Exec's no-command path, which emits a sentinel and exits.
func runError(t *testing.T, cmd gloo.Command[[]byte, []byte]) error {
	t.Helper()
	ctx := context.Background()
	source := gloo.ByteReaderSource([]io.Reader{strings.NewReader("")})
	_, err := gloo.Collect(ctx, cmd.Execute(ctx, source.Stream(ctx)))
	return err
}

func TestExec_NoCommandEmitsSentinel(t *testing.T) {
	// With no program named, Exec returns a command that fails with ErrNoCommand
	// rather than spawning anything.
	err := runError(t, Exec())
	if !errors.Is(err, ErrNoCommand) {
		t.Fatalf("err=%v, want ErrNoCommand", err)
	}
}

func TestErrNoCommand_Message(t *testing.T) {
	// The sentinel renders its own message (the Error method's contract).
	if got := ErrNoCommand.Error(); got != "exec: no command specified" {
		t.Fatalf("got %q", got)
	}
}

func TestExec_BuildsCommandForNamedProgram(t *testing.T) {
	// Naming a program yields a runnable command. Constructing it must not spawn
	// the process, so this asserts only that a command was built.
	if Exec("tr", "a-z", "A-Z") == nil {
		t.Fatal("Exec returned nil command for a named program")
	}
}
