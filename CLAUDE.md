# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Review this plan thoroughly before making any code changes. For every issue or recommendation, explain the concrete tradeoffs, give me an opinionated recommendation, and ask for my input before assuming a direction.

My engineering preferences (use these to guide your recommendations):

DRY is important—flag repetition aggressively.

Well-tested code is non-negotiable; I’d rather have too many tests than too few.

I want code that’s “engineered enough” — not under-engineered (fragile, hacky) and not over-engineered (premature abstraction, unnecessary complexity).

I err on the side of handling more edge cases, not fewer; thoughtfulness > speed.

Bias toward explicit over clever.

1. Architecture review

Evaluate:

Overall system design and component boundaries.

Dependency graph and coupling concerns.

Data flow patterns and potential bottlenecks.

Scaling characteristics and single points of failure.

Security architecture (auth, data access, API boundaries).

2. Code quality review

Evaluate:

Code organization and module structure.

DRY violations—be aggressive here.

Error handling patterns and missing edge cases (call these out explicitly).

Technical debt hotspots.

Areas that are over-engineered or under-engineered relative to my preferences.

3. Test review

Evaluate:

Test coverage gaps (unit, integration, e2e).

Test quality and assertion strength.

Missing edge case coverage—be thorough.

Untested failure modes and error paths.

4. Performance review

Evaluate:

N+1 queries and database access patterns.

Memory-usage concerns.

Caching opportunities.

Slow or high-complexity code paths.

For each issue you find (for every specific issue (bug, smell, design concern, or risk)):

Describe the problem concretely, with file and line references.

Present 2–3 options, including “do nothing” where that’s reasonable.

For each option, specify: implementation effort, risk, impact on other code, and maintenance burden.

Give me your recommended option and why, mapped to my preferences above.

Then explicitly ask whether I agree or want to choose a different direction before proceeding.

Workflow and interaction

Do not assume my priorities on timeline or scale.

After each section, pause and ask for my feedback before moving on.

BEFORE YOU START:

Ask if I want one of two options:

1/ BIG CHANGE: Work through this interactively, one section at a time (Architecture → Code Quality → Tests → Performance) with at most 4 top issues in each section.

2/ SMALL CHANGE: Work through interactively ONE question per review section.

FOR EACH STAGE OF REVIEW: output the explanation and pros and cons of each stage’s questions AND your opinionated recommendation and why, and then use AskUserQuestion. Also NUMBER issues and then give LETTERS for options and when using AskUserQuestion make sure each option clearly labels the issue NUMBER and option LETTER so the user doesn’t get confused. Make the recommended option always the 1st option.

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

# Config commands
raypaste config set api-key sk-or-v1-...
raypaste config set default-model cerebras-llama-8b
raypaste config set default-length short
raypaste config get api-key
raypaste config get default-model

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

**Three modes:**

- **Instant complete** — single-shot generation invoked directly as `raypaste "text"`; reads from args or stdin, copies to clipboard by default
- **Interactive** (`interactive`, aliases: `i`, `repl`) — REPL with streaming, slash commands (`/model`, `/prompt`, `/length`, `/copy`), readline line-editing with autocomplete and suggestion preview. **Note: context does not persist between requests — each message is a fresh, stateless request.**
- **Config** (`config`) — Manage configuration settings via CLI with `set` and `get` subcommands. Does not require API key to be set (allows setting the key itself).

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
| `internal/config/`      | Config loading (Viper), model alias resolution, config save |
| `internal/interactive/` | REPL implementation, slash commands, autocomplete        |
| `internal/llm/`         | OpenRouter HTTP client, request building, SSE streaming  |
| `internal/prompts/`     | Prompt template loading and rendering                    |
| `internal/output/`      | Terminal Markdown colorization, respects `NO_COLOR`      |
| `internal/clipboard/`   | Cross-platform clipboard via `golang.design/x/clipboard` |
| `pkg/types/`            | Shared request/response types                            |

### Config Command

The `config` command provides CLI-based configuration management:

```bash
raypaste config set [key] [value]   # Set a config value
raypaste config get [key]           # Get a config value
```

**Supported keys:** `api-key`, `default-model`, `default-length`, `disable-copy`, `temperature`

**Implementation:** `cmd/config.go` uses `config.SaveTo()` method which writes to `~/.raypaste/config.yaml` via Viper. The `initConfig()` in `cmd/root.go` skips API key validation when running `config` commands, allowing users to set their API key without having one already configured.

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
