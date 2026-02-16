/*
Copyright © 2026 Raypaste
*/
package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/prompts"
	"github.com/raypaste/raypaste-cli/pkg/types"
)

// initTestCfg sets the package-level cfg used by handleSlashCommand.
func initTestCfg(t *testing.T) {
	t.Helper()
	cfg = &config.Config{
		DefaultModel:  "cerebras-gpt-oss-120b",
		DefaultLength: types.OutputLengthMedium,
		Models:        make(map[string]config.Model),
		Temperature:   0.7,
	}
}

// newTestReplState returns a replState backed by a real prompts.Store.
func newTestReplState(t *testing.T) *replState {
	t.Helper()
	store, err := prompts.NewStore()
	if err != nil {
		t.Fatalf("prompts.NewStore() error = %v", err)
	}
	return &replState{
		model:      "cerebras-gpt-oss-120b",
		length:     types.OutputLengthMedium,
		promptName: "metaprompt",
		store:      store,
	}
}

// ── getInput ─────────────────────────────────────────────────────────────────

func TestGetInputFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"single arg", []string{"hello"}, "hello"},
		{"multiple args joined with space", []string{"hello", "world"}, "hello world"},
		{"single arg with internal space", []string{"hello world"}, "hello world"},
		{"three args", []string{"a", "b", "c"}, "a b c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInput(tt.args)
			if err != nil {
				t.Fatalf("getInput() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("getInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ── formatWelcomeLines ───────────────────────────────────────────────────────

func TestFormatWelcomeLines(t *testing.T) {
	state := &replState{
		model:      "test-model-xyz",
		length:     types.OutputLengthShort,
		promptName: "metaprompt",
	}

	lines := formatWelcomeLines(state)

	if len(lines) == 0 {
		t.Fatal("formatWelcomeLines() returned empty slice")
	}

	joined := strings.Join(lines, "")

	// Values are embedded inside ANSI escape codes but the raw string is still present.
	checks := []struct {
		desc    string
		contain string
	}{
		{"model name", "test-model-xyz"},
		{"length value", "short"},
		{"prompt name", "metaprompt"},
	}
	for _, c := range checks {
		if !strings.Contains(joined, c.contain) {
			t.Errorf("formatWelcomeLines() output missing %s (%q)", c.desc, c.contain)
		}
	}
}

// ── collectPastedInput ───────────────────────────────────────────────────────

func TestCollectPastedInput(t *testing.T) {
	t.Run("single line times out and returns first line", func(t *testing.T) {
		ch := make(chan readResult) // unbuffered — no additional lines arrive
		got := collectPastedInput(ch, "first line")
		if got != "first line" {
			t.Errorf("got %q, want %q", got, "first line")
		}
	})

	t.Run("collects buffered lines before timeout", func(t *testing.T) {
		ch := make(chan readResult, 3)
		ch <- readResult{line: "second line"}
		ch <- readResult{line: "  third line  "} // trimmed by collectPastedInput
		got := collectPastedInput(ch, "first line")
		want := "first line\nsecond line\nthird line"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("stops immediately on error in channel", func(t *testing.T) {
		ch := make(chan readResult, 2)
		ch <- readResult{err: io.EOF}
		got := collectPastedInput(ch, "first line")
		if got != "first line" {
			t.Errorf("got %q, want %q", got, "first line")
		}
	})

	t.Run("stops when channel is closed after buffered lines", func(t *testing.T) {
		ch := make(chan readResult, 1)
		ch <- readResult{line: "second line"}
		close(ch)
		got := collectPastedInput(ch, "first line")
		want := "first line\nsecond line"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("empty additional lines preserved for structure", func(t *testing.T) {
		ch := make(chan readResult, 2)
		ch <- readResult{line: ""}
		ch <- readResult{line: "third line"}
		got := collectPastedInput(ch, "first line")
		want := "first line\n\nthird line"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

// ── drainLines ───────────────────────────────────────────────────────────────

func TestDrainLines(t *testing.T) {
	t.Run("drains all buffered items", func(t *testing.T) {
		ch := make(chan readResult, 3)
		ch <- readResult{line: "a"}
		ch <- readResult{line: "b"}
		ch <- readResult{line: "c"}
		drainLines(ch)
		if len(ch) != 0 {
			t.Errorf("channel not fully drained, %d item(s) remain", len(ch))
		}
	})

	t.Run("does not block on empty channel", func(t *testing.T) {
		ch := make(chan readResult, 1)
		drainLines(ch) // must return immediately
	})

	t.Run("does not panic on closed channel", func(t *testing.T) {
		ch := make(chan readResult, 2)
		ch <- readResult{line: "a"}
		close(ch)
		drainLines(ch) // reads the item then returns on closed channel
	})
}

// ── handleSlashCommand ───────────────────────────────────────────────────────

func TestHandleSlashCommand_ExitCommands(t *testing.T) {
	initTestCfg(t)
	for _, cmd := range []string{"/quit", "/exit"} {
		t.Run(cmd+" returns true", func(t *testing.T) {
			if !handleSlashCommand(cmd, newTestReplState(t)) {
				t.Errorf("handleSlashCommand(%q) = false, want true", cmd)
			}
		})
	}
}

func TestHandleSlashCommand_NonExitCommands(t *testing.T) {
	initTestCfg(t)

	nonExit := []string{"/help", "/clear", "/unknown-command"}
	for _, cmd := range nonExit {
		t.Run(cmd+" returns false", func(t *testing.T) {
			if handleSlashCommand(cmd, newTestReplState(t)) {
				t.Errorf("handleSlashCommand(%q) = true, want false", cmd)
			}
		})
	}
}

func TestHandleSlashCommand_Length(t *testing.T) {
	initTestCfg(t)

	tests := []struct {
		name       string
		input      string
		wantLength types.OutputLength
		wantExit   bool
	}{
		{"no args shows current", "/length", types.OutputLengthMedium, false},
		{"set short", "/length short", types.OutputLengthShort, false},
		{"set long", "/length long", types.OutputLengthLong, false},
		{"set medium", "/length medium", types.OutputLengthMedium, false},
		{"alias /l short", "/l short", types.OutputLengthShort, false},
		// Invalid value: state should be unchanged.
		{"invalid value preserves state", "/length bad", types.OutputLengthMedium, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newTestReplState(t)
			state.length = types.OutputLengthMedium

			got := handleSlashCommand(tt.input, state)
			if got != tt.wantExit {
				t.Errorf("handleSlashCommand(%q) exit = %v, want %v", tt.input, got, tt.wantExit)
			}
			if state.length != tt.wantLength {
				t.Errorf("state.length = %q, want %q", state.length, tt.wantLength)
			}
		})
	}
}

func TestHandleSlashCommand_Model(t *testing.T) {
	initTestCfg(t)

	tests := []struct {
		name      string
		input     string
		wantModel string
	}{
		{"no args preserves model", "/model", "cerebras-gpt-oss-120b"},
		{"sets model alias", "/model gpt-4o", "gpt-4o"},
		{"sets raw openrouter id", "/model openai/gpt-4o", "openai/gpt-4o"},
		{"alias /m", "/m my-model", "my-model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newTestReplState(t)
			handleSlashCommand(tt.input, state)
			if state.model != tt.wantModel {
				t.Errorf("state.model = %q, want %q", state.model, tt.wantModel)
			}
		})
	}
}

func TestHandleSlashCommand_Prompt(t *testing.T) {
	initTestCfg(t)

	t.Run("no args preserves promptName", func(t *testing.T) {
		state := newTestReplState(t)
		handleSlashCommand("/prompt", state)
		if state.promptName != "metaprompt" {
			t.Errorf("state.promptName = %q, want %q", state.promptName, "metaprompt")
		}
	})

	t.Run("sets valid prompt", func(t *testing.T) {
		state := newTestReplState(t)
		handleSlashCommand("/prompt bulletlist", state)
		if state.promptName != "bulletlist" {
			t.Errorf("state.promptName = %q, want %q", state.promptName, "bulletlist")
		}
	})

	t.Run("invalid prompt preserves promptName", func(t *testing.T) {
		state := newTestReplState(t)
		handleSlashCommand("/prompt nonexistent-prompt", state)
		if state.promptName != "metaprompt" {
			t.Errorf("state.promptName = %q, want %q", state.promptName, "metaprompt")
		}
	})

	t.Run("alias /p sets prompt", func(t *testing.T) {
		state := newTestReplState(t)
		handleSlashCommand("/p bulletlist", state)
		if state.promptName != "bulletlist" {
			t.Errorf("state.promptName = %q, want %q", state.promptName, "bulletlist")
		}
	})
}

func TestHandleSlashCommand_Copy(t *testing.T) {
	initTestCfg(t)

	t.Run("no last response returns false without panicking", func(t *testing.T) {
		state := newTestReplState(t)
		state.lastResponse = ""
		if handleSlashCommand("/copy", state) {
			t.Error("handleSlashCommand(\"/copy\") = true, want false")
		}
	})

	t.Run("alias /c with no last response returns false", func(t *testing.T) {
		state := newTestReplState(t)
		state.lastResponse = ""
		if handleSlashCommand("/c", state) {
			t.Error("handleSlashCommand(\"/c\") = true, want false")
		}
	})

	t.Run("with last response returns false", func(t *testing.T) {
		state := newTestReplState(t)
		state.lastResponse = "some generated content"
		// clipboard.CopyWithWarning may fail in headless environments; that's OK —
		// the command still returns false either way.
		if handleSlashCommand("/copy", state) {
			t.Error("handleSlashCommand(\"/copy\") = true, want false")
		}
	})
}
