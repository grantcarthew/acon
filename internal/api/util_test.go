package api

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestTruncateStringUTF8Safe(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxRunes int
		want     string
	}{
		{
			name:     "shorter than max returns unchanged",
			input:    "hello",
			maxRunes: 10,
			want:     "hello",
		},
		{
			name:     "exactly max length returns unchanged",
			input:    "hello",
			maxRunes: 5,
			want:     "hello",
		},
		{
			name:     "longer than max gets truncated with ellipsis",
			input:    "hello world",
			maxRunes: 5,
			want:     "hello...",
		},
		{
			name:     "empty input returns empty",
			input:    "",
			maxRunes: 10,
			want:     "",
		},
		{
			name:     "zero maxRunes truncates everything",
			input:    "hello",
			maxRunes: 0,
			want:     "...",
		},
		{
			name:     "japanese characters truncated by rune not byte",
			input:    "日本語テキスト",
			maxRunes: 3,
			want:     "日本語...",
		},
		{
			name:     "emoji counted as runes",
			input:    "🚀🎉🔥💥",
			maxRunes: 2,
			want:     "🚀🎉...",
		},
		{
			name:     "mixed ascii and multibyte truncates safely",
			input:    "hello 日本",
			maxRunes: 7,
			want:     "hello 日...",
		},
		{
			name:     "japanese exactly at limit unchanged",
			input:    "日本語",
			maxRunes: 3,
			want:     "日本語",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateStringUTF8Safe(tt.input, tt.maxRunes)
			if got != tt.want {
				t.Errorf("truncateStringUTF8Safe(%q, %d)\n  got:  %q\n  want: %q", tt.input, tt.maxRunes, got, tt.want)
			}
		})
	}
}

func TestClientLogVerbose(t *testing.T) {
	t.Run("writes when VerboseLog is set", func(t *testing.T) {
		var buf bytes.Buffer
		c := &Client{VerboseLog: &buf}
		c.logVerbose("hello %s\n", "world")
		if got := buf.String(); got != "hello world\n" {
			t.Errorf("got %q, want %q", got, "hello world\n")
		}
	})

	t.Run("writes nothing anywhere when VerboseLog is nil", func(t *testing.T) {
		// Redirect stderr and stdout to detect any rogue fallback writes.
		// Mutates process-global FDs; do not call t.Parallel here.
		origStderr, origStdout := os.Stderr, os.Stdout
		rErr, wErr, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe stderr: %v", err)
		}
		rOut, wOut, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe stdout: %v", err)
		}
		os.Stderr, os.Stdout = wErr, wOut
		// Always restore the FDs, even if the call under test panics.
		defer func() {
			os.Stderr, os.Stdout = origStderr, origStdout
			_ = rErr.Close()
			_ = rOut.Close()
		}()

		errCh := make(chan []byte, 1)
		outCh := make(chan []byte, 1)
		go func() { b, _ := io.ReadAll(rErr); errCh <- b }()
		go func() { b, _ := io.ReadAll(rOut); outCh <- b }()

		c := &Client{VerboseLog: nil}
		c.logVerbose("ignored %s\n", "value")

		_ = wErr.Close()
		_ = wOut.Close()

		if got := <-errCh; len(got) != 0 {
			t.Errorf("stderr received %q, want nothing", got)
		}
		if got := <-outCh; len(got) != 0 {
			t.Errorf("stdout received %q, want nothing", got)
		}
	})

	t.Run("formats arguments correctly", func(t *testing.T) {
		var buf bytes.Buffer
		c := &Client{VerboseLog: &buf}
		c.logVerbose("status=%d body=%s", 200, "ok")
		if got := buf.String(); !strings.Contains(got, "status=200 body=ok") {
			t.Errorf("got %q, missing expected formatted output", got)
		}
	})
}
