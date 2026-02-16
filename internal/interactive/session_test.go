package interactive

import (
	"io"
	"strings"
	"testing"

	"github.com/raypaste/raypaste-cli/internal/prompts"
	"github.com/raypaste/raypaste-cli/pkg/types"
)

func newTestState(t *testing.T) *State {
	t.Helper()
	store, err := prompts.NewStore()
	if err != nil {
		t.Fatalf("prompts.NewStore() error = %v", err)
	}
	return &State{
		Model:      "cerebras-gpt-oss-120b",
		Length:     types.OutputLengthMedium,
		PromptName: "metaprompt",
		Store:      store,
	}
}

func TestFormatWelcomeLines(t *testing.T) {
	state := &State{
		Model:      "test-model-xyz",
		Length:     types.OutputLengthShort,
		PromptName: "metaprompt",
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
