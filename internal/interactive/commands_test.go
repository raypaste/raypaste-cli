package interactive

import (
	"testing"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/pkg/types"
)

func TestHandleSlashCommandExitCommands(t *testing.T) {
	for _, cmd := range []string{"/quit", "/exit"} {
		t.Run(cmd+" returns true", func(t *testing.T) {
			if !handleSlashCommand(cmd, newTestState(t), map[string]config.Model{}) {
				t.Errorf("handleSlashCommand(%q) = false, want true", cmd)
			}
		})
	}
}

func TestHandleSlashCommandNonExitCommands(t *testing.T) {
	nonExit := []string{"/help", "/clear", "/unknown-command"}
	for _, cmd := range nonExit {
		t.Run(cmd+" returns false", func(t *testing.T) {
			if handleSlashCommand(cmd, newTestState(t), map[string]config.Model{}) {
				t.Errorf("handleSlashCommand(%q) = true, want false", cmd)
			}
		})
	}
}

func TestHandleSlashCommandLength(t *testing.T) {
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
			state := newTestState(t)
			state.Length = types.OutputLengthMedium

			got := handleSlashCommand(tt.input, state, map[string]config.Model{})
			if got != tt.wantExit {
				t.Errorf("handleSlashCommand(%q) exit = %v, want %v", tt.input, got, tt.wantExit)
			}
			if state.Length != tt.wantLength {
				t.Errorf("state.Length = %q, want %q", state.Length, tt.wantLength)
			}
		})
	}
}

func TestHandleSlashCommandModel(t *testing.T) {
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
			state := newTestState(t)
			handleSlashCommand(tt.input, state, map[string]config.Model{})
			if state.Model != tt.wantModel {
				t.Errorf("state.Model = %q, want %q", state.Model, tt.wantModel)
			}
		})
	}
}

func TestHandleSlashCommandPrompt(t *testing.T) {
	t.Run("no args preserves promptName", func(t *testing.T) {
		state := newTestState(t)
		handleSlashCommand("/prompt", state, map[string]config.Model{})
		if state.PromptName != "metaprompt" {
			t.Errorf("state.PromptName = %q, want %q", state.PromptName, "metaprompt")
		}
	})

	t.Run("sets valid prompt", func(t *testing.T) {
		state := newTestState(t)
		handleSlashCommand("/prompt bulletlist", state, map[string]config.Model{})
		if state.PromptName != "bulletlist" {
			t.Errorf("state.PromptName = %q, want %q", state.PromptName, "bulletlist")
		}
	})

	t.Run("invalid prompt preserves promptName", func(t *testing.T) {
		state := newTestState(t)
		handleSlashCommand("/prompt nonexistent-prompt", state, map[string]config.Model{})
		if state.PromptName != "metaprompt" {
			t.Errorf("state.PromptName = %q, want %q", state.PromptName, "metaprompt")
		}
	})

	t.Run("alias /p sets prompt", func(t *testing.T) {
		state := newTestState(t)
		handleSlashCommand("/p bulletlist", state, map[string]config.Model{})
		if state.PromptName != "bulletlist" {
			t.Errorf("state.PromptName = %q, want %q", state.PromptName, "bulletlist")
		}
	})
}

func TestHandleSlashCommandCopy(t *testing.T) {
	t.Run("no last response returns false without panicking", func(t *testing.T) {
		state := newTestState(t)
		state.LastResponse = ""
		if handleSlashCommand("/copy", state, map[string]config.Model{}) {
			t.Error("handleSlashCommand(\"/copy\") = true, want false")
		}
	})

	t.Run("alias /c with no last response returns false", func(t *testing.T) {
		state := newTestState(t)
		state.LastResponse = ""
		if handleSlashCommand("/c", state, map[string]config.Model{}) {
			t.Error("handleSlashCommand(\"/c\") = true, want false")
		}
	})

	t.Run("with last response returns false", func(t *testing.T) {
		state := newTestState(t)
		state.LastResponse = "some generated content"
		// clipboard.CopyWithWarning may fail in headless environments; that's OK —
		// the command still returns false either way.
		if handleSlashCommand("/copy", state, map[string]config.Model{}) {
			t.Error("handleSlashCommand(\"/copy\") = true, want false")
		}
	})
}
