# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Manual build
go build -o raypaste

# Manual build with version info
go build -ldflags "-X github.com/raypaste/raypaste-cli/cmd.Version=v1.0.0 \
                   -X github.com/raypaste/raypaste-cli/cmd.GitCommit=$(git rev-parse --short HEAD) \
                   -X github.com/raypaste/raypaste-cli/cmd.BuildDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
         -o raypaste

# Check version
raypaste version
raypaste --version
raypaste -v

# Run all tests
go test ./...

# Run a specific test
go test ./internal/llm/ -run TestBuildRequest

# Lint
go tool golangci-lint run ./...

# Format
goimports -w ./cmd ./internal ./pkg
```

## Architecture

**Entry point:** `main.go` → `cmd.Execute()` (Cobra CLI)

**Two modes:**

- **Instant complete** — single-shot generation invoked directly as `raypaste "text"`; reads from args or stdin, copies to clipboard by default
- **Interactive** (`interactive`, aliases: `i`, `repl`) — REPL with streaming, slash commands (`/model`, `/prompt`, `/length`, `/copy`), readline line-editing with autocomplete and suggestion preview. **Note: context does not persist between requests — each message is a fresh, stateless request.**

**Request pipeline:**

```
Input → cmd/root.go (for raypaste "text")
      → cmd/interactive.go (for interactive mode)
      → internal/prompts/store.go       (load + render YAML prompt template)
      → internal/llm/router.go          (build OpenRouter request)
      → internal/llm/client.go          (HTTP to api.openrouter.ai)
      → internal/llm/streaming.go       (SSE stream parsing, if interactive)
      → internal/output/formatter.go    (Markdown colorization)
      → internal/clipboard/clipboard.go (copy to clipboard)
```

**Config loading hierarchy** (lowest → highest priority):

1. Defaults
2. `~/.raypaste/config.yaml`
3. `RAYPASTE_*` environment variables
4. CLI flags

Required: `RAYPASTE_API_KEY` (or `api_key` in config).

**Prompt system:** Built-in prompts live in `internal/prompts/defaults/`. User prompts are loaded from `~/.raypaste/prompts/` as YAML files with a `{{.LengthDirective}}` template variable.

**Model system:** Named aliases (`cerebras-llama-8b`, etc.) defined in `internal/config/models.go` map to OpenRouter model IDs. Custom models can be defined in config. Raw OpenRouter IDs also work directly.

**Autocomplete system:** Implemented in `internal/interactive/autocomplete.go`. Provides context-aware completions via the readline library:

- Completes slash command names (e.g., `/model`, `/prompt`)
- Completes model names after `/model` argument
- Completes prompt names after `/prompt` argument
- Uses case-insensitive prefix matching
- Renders dimmed suggestion preview after cursor in real-time (via `suggestionPainter`)

## Key packages

| Path                    | Purpose                                                  |
| ----------------------- | -------------------------------------------------------- |
| `cmd/`                  | Cobra command definitions and CLI wiring                 |
| `internal/config/`      | Config loading (Viper), model alias resolution           |
| `internal/interactive/` | REPL implementation, slash commands, autocomplete        |
| `internal/llm/`         | OpenRouter HTTP client, request building, SSE streaming  |
| `internal/prompts/`     | Prompt template loading and rendering                    |
| `internal/output/`      | Terminal Markdown colorization, respects `NO_COLOR`      |
| `internal/clipboard/`   | Cross-platform clipboard via `golang.design/x/clipboard` |
| `pkg/types/`            | Shared request/response types                            |

### Autocomplete Support

The interactive REPL provides two forms of autocomplete assistance:

#### Tab Completion

- Press **Tab** to trigger completion suggestions.
- If exactly one match exists, it is inserted automatically.
- If multiple matches exist, they are presented for selection.
- Completion context is determined by the current command:
  - Slash command name: filters available commands (e.g., `/m<Tab>` → `/model`, `/move`)
  - `/model <prefix>`: filters available model names
  - `/prompt <prefix>`: filters available prompt names

#### Suggestion Preview

- As you type, a dimmed suggestion preview appears after the cursor showing the first matching completion.
- The preview is non-intrusive and disappears when no matches exist.
- Press **Right Arrow** or **End** to accept the suggestion and advance the cursor, or continue typing to dismiss it.

#### Cursor Behavior

- Tab completion preserves cursor position at the end of inserted text.
- Arrow keys navigate the completion menu (if visible) without affecting suggestion preview state.
- Completing a command name (e.g., `/mo` → `/model`) leaves cursor after the command; spaces are not auto-inserted.
- Completing an argument (e.g., `/model cere` → `/model cerebras-llama-8b`) leaves cursor after the full name.

## Notes

- Streaming uses SSE (Server-Sent Events); context cancellation closes the TCP connection immediately to stop billing mid-stream.
- The `build` script is a bash file (no extension) — run with `./build`, not `make`.
- CI installs X11 dependencies for clipboard tests on Linux.
- Linter is pinned via Go's `tool` directive in `go.mod`; use `go tool golangci-lint` rather than a globally installed version.
