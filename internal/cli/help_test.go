package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn and returns whatever it printed to os.Stdout.
// Mutates the process-global os.Stdout, so callers must not run in parallel
// with anything else in this package that writes to stdout.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	// Restore stdout even if fn panics.
	defer func() {
		os.Stdout = orig
		_ = r.Close()
	}()

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	runErr := fn()

	_ = w.Close()
	<-done

	return buf.String(), runErr
}

func TestRunHelpAgents(t *testing.T) {
	t.Run("no args prints overview", func(t *testing.T) {
		out, err := captureStdout(t, func() error {
			return runHelpAgents(nil, nil)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := agentHelpOverview + "\n"
		if out != want {
			t.Errorf("output mismatch\n  got:  %q\n  want: %q", out, want)
		}
	})

	t.Run("workflow topic prints workflow", func(t *testing.T) {
		out, err := captureStdout(t, func() error {
			return runHelpAgents(nil, []string{"workflow"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := agentHelpWorkflow + "\n"
		if out != want {
			t.Errorf("output mismatch\n  got:  %q\n  want: %q", out, want)
		}
	})

	t.Run("all topic prints overview then separator then workflow", func(t *testing.T) {
		out, err := captureStdout(t, func() error {
			return runHelpAgents(nil, []string{"all"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := agentHelpOverview + "\n\n---\n" + agentHelpWorkflow + "\n"
		if out != want {
			t.Errorf("output mismatch\n  got:  %q\n  want: %q", out, want)
		}
	})

	t.Run("unknown topic returns error", func(t *testing.T) {
		_, err := captureStdout(t, func() error {
			return runHelpAgents(nil, []string{"nonexistent-topic-xyz"})
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unknown help topic: nonexistent-topic-xyz") {
			t.Errorf("error %q does not contain expected substring", err.Error())
		}
	})
}
