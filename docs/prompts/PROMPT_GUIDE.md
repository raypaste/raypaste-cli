# Custom Prompt Guide

This guide explains how to create and use custom prompts in raypaste-cli.

## Overview

raypaste supports two types of prompts:

1. **Built-in prompts** — Compiled into the binary (defined in Go code)
2. **Custom prompts** — YAML files in `~/.raypaste/prompts/`, managed via CLI or created manually

## Built-in Prompts

### metaprompt (default)
- **Description**: Generate an optimized meta-prompt from a user's goal
- **Supported lengths**: short, medium, long
- **Use case**: General-purpose prompt optimization

### bulletlist
- **Description**: Organize text by relation and output as a short bulleted list
- **Supported lengths**: short, medium (long mode is intentionally disabled)
- **Use case**: Organizing information into structured, relational bullet points

## Quick Start: Create Your Own Custom Prompt

The easiest way to create a custom prompt is with the CLI:

```bash
raypaste config prompt add ascii-art
```

This launches an interactive wizard that walks you through:
1. **Description** — what the prompt does (not sent to the LLM)
2. **System prompt** — the template text (supports `{{.LengthDirective}}` and `{{.Context}}`)
3. **Length directives** — per-length guidance (text or integer token counts)

Once created, use it:

```bash
# Instant completion mode (use single quotes to protect backticks)
raypaste 'coffee cup' -p ascii-art

# Interactive mode
raypaste interactive
# then type: /prompt ascii-art
```

### Non-Interactive Creation

You can also create prompts entirely via flags:

```bash
raypaste config prompt add ascii-art \
  --description "Convert text into ASCII art" \
  --system "You are an ASCII art expert. Create creative ASCII art of the input.

Output length guidance: {{.LengthDirective}}

CRITICAL:
- Output ONLY the ASCII art itself, no explanations
- Keep it readable in a terminal" \
  --medium "Create a medium-sized ASCII art (5-15 lines)"
```

Or load the system prompt from a file:

```bash
raypaste config prompt add ascii-art \
  --description "Convert text into ASCII art" \
  --from-file ./my-prompt.txt \
  --short "Small ASCII art (3-5 lines)" \
  --medium "Medium ASCII art (5-15 lines)"
```

## Managing Prompts via CLI

### Add a prompt

```bash
raypaste config prompt add <name>                    # Interactive mode
raypaste config prompt add <name> --description "…"  # With flags
raypaste config prompt add <name> --from-file ./f    # System prompt from file
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--description` | `-d` | Description of the prompt |
| `--system` | `-s` | System prompt template text |
| `--from-file` | `-f` | Read system prompt from a file |
| `--short` | | Directive for short output length |
| `--medium` | | Directive for medium output length |
| `--long` | | Directive for long output length |

**Prompt names** must contain only letters, numbers, hyphens, and underscores.

### List all prompts

```bash
raypaste config prompt list
```

Shows all prompts with `[built-in]`/`[custom]` tags and supported lengths.

### Show prompt details

```bash
raypaste config prompt show metaprompt
raypaste config prompt show my-custom-prompt
```

Displays the full system prompt text and length directive configuration.

### Remove a prompt

```bash
raypaste config prompt remove my-custom-prompt       # With confirmation
raypaste config prompt remove my-custom-prompt -f    # Force (no confirmation)
```

Built-in prompts cannot be removed.

## Manual YAML Creation (Advanced)

For power users, you can create YAML files directly in `~/.raypaste/prompts/`. Both `.yaml` and `.yml` extensions are supported.

```yaml
name: my-prompt
description: "Brief description of what this prompt does"

# The system prompt template
# Use {{.LengthDirective}} for length-specific guidance
# Use {{.Context}} for project context (from CLAUDE.md, AGENTS.md, .cursor/rules/)
system: |
  You are an expert assistant. Your task is to...

  Project context: {{.Context}}
  Output length guidance: {{.LengthDirective}}

  CRITICAL:
  - Follow these specific instructions
  - Output format requirements

# Define which lengths this prompt supports
# Omit a length to disable it for this prompt
length_directives:
  short: "Your custom short directive here"
  medium: "Your custom medium directive here"
  long: "Your custom long directive here"
```

**Usage:**
```bash
raypaste 'your input' -p my-prompt
raypaste 'your input' -p my-prompt -l short
```

## Template Variables

Two template variables are available in system prompts:

### `{{.LengthDirective}}`

Replaced with the length-specific directive for the current output length. If the directive is an integer (used as a `max_tokens` override), this renders as an empty string.

### `{{.Context}}`

Replaced with project context automatically detected from the working directory. raypaste looks for these files (in priority order):
- `CLAUDE.md`
- `AGENTS.md`
- `.cursor/rules/`

This lets prompts be context-aware without the user manually pasting project info.

## Length Directives

Length directives control how much output the LLM generates. Each prompt defines directives for `short`, `medium`, and/or `long`.

### Two Types of Directives

**Text directives** — A string injected into `{{.LengthDirective}}` in the system prompt:

```yaml
length_directives:
  short: "Be concise, 2-3 sentences max"
  medium: "Moderate detail, 1-2 paragraphs"
```

**Integer directives** — A plain number that sets the `max_tokens` API parameter directly. The `{{.LengthDirective}}` template variable renders as an empty string:

```yaml
length_directives:
  short: "200"    # Sets max_tokens=200
  medium: "500"   # Sets max_tokens=500
  long: "1200"    # Sets max_tokens=1200
```

Integer directives are useful for prompts where you want precise token control (e.g., SQL generation) rather than injecting text guidance.

### Default Length Parameters

If you don't specify custom `length_directives`, or leave a directive value empty, these defaults are used:

| Length | Max Tokens | Default Directive |
|--------|------------|-------------------|
| short  | 550        | "Keep the generated prompt concise — under 150 words. Focus on the core instruction only." |
| medium | 850        | "Generate a moderately detailed prompt (~200-350 words) with context, constraints, and desired output format." |
| long   | 1600       | "Generate a comprehensive prompt (400-600+ words) including examples, edge cases, tone guidance, and detailed formatting instructions." |

### Restricting Output Lengths

To limit a prompt to specific output lengths, omit unwanted lengths from `length_directives`:

```yaml
# Short and medium only — long will error
length_directives:
  short: "1-2 sentence summary"
  medium: "1 paragraph summary"
```

```bash
raypaste 'text to summarize' -p quick-summary -l long
# Error: prompt 'quick-summary' does not support output length 'long'
```

## Examples

### Code Review Prompt

```bash
raypaste config prompt add code-review \
  --description "Generate a code review prompt" \
  --system "You are a code review expert. Generate a detailed prompt for reviewing code.

Output length guidance: {{.LengthDirective}}
Project context: {{.Context}}

CRITICAL:
- Focus on code quality, security, and best practices
- Output ONLY the review prompt itself
- No preamble or explanations" \
  --short "Focus on critical issues only" \
  --medium "Cover functionality, style, and best practices" \
  --long "Include security, performance, testing, and documentation"
```

### SQL Prompt with Integer Directives

```bash
raypaste config prompt add sql \
  --description "Generate SQL queries" \
  --system "Act as a SQL expert. Write the SQL query for the user's request.

Project context: {{.Context}}

CRITICAL:
- Output ONLY the SQL query, no explanations
- Use standard SQL unless the context indicates a specific database" \
  --short "200" \
  --medium "500" \
  --long "1200"
```

Here, `"200"` sets `max_tokens=200` instead of injecting text into the template.

### Email Writer (Short/Medium Only)

```yaml
# ~/.raypaste/prompts/email-writer.yaml
name: email-writer
description: "Generate professional email drafts"
system: |
  You are a professional email writer. Generate a well-structured email based on the user's request.

  Output length guidance: {{.LengthDirective}}

  CRITICAL:
  - Include subject line
  - Professional tone
  - Clear call-to-action
  - Output ONLY the email itself

length_directives:
  short: "Brief, direct email under 100 words"
  medium: "Detailed email with context, 150-250 words"
  # long mode not supported for emails
```

### Using Built-in Bulletlist

```bash
# Organize meeting notes (short mode)
raypaste 'Discussed Q1 goals, budget concerns, new hires, office relocation' -p bulletlist -l short

# Organize project requirements (medium mode)
raypaste 'User auth, database schema, API endpoints, frontend UI' -p bulletlist -l medium

# Long mode errors (intentionally unsupported)
raypaste 'some text' -p bulletlist -l long
# Error: prompt 'bulletlist' does not support output length 'long'
```

## Best Practices

1. **Clear instructions** — Be explicit about what the prompt should do
2. **Output format** — Specify the exact format you want (bullets, paragraphs, code, etc.)
3. **Constraints** — Use CRITICAL sections to emphasize important rules
4. **Length directives** — Customize them for your specific use case; use integers for precise token control
5. **Project context** — Include `{{.Context}}` when the prompt benefits from knowing about the user's project
6. **Testing** — Test with all supported lengths to ensure consistent behavior
7. **Single quotes** — In instant completion mode, use single quotes to protect backticks in your input

## Contributing Built-in Prompts

If you create a useful prompt, consider contributing it as a built-in:

1. Create the prompt file in `internal/prompts/defaults/`
2. Register it in `internal/prompts/store.go` (`loadBuiltInPrompts()`)
3. Add documentation to README.md
4. Submit a pull request

## Troubleshooting

### Prompt Not Found
```
Error: prompt not found: my-prompt
```
**Solution**: Run `raypaste config prompt list` to see available prompts. If you created the YAML manually, check that the file exists in `~/.raypaste/prompts/` and has a `name` field matching what you're requesting.

### Length Not Supported
```
Error: prompt 'my-prompt' does not support output length 'long'
```
**Solution**: This is intentional. The prompt only defines directives for certain lengths. Run `raypaste config prompt show my-prompt` to see which lengths are supported.

### Template Parse Error
```
Error: failed to parse template: ...
```
**Solution**: Check your template syntax. Ensure `{{.LengthDirective}}` and `{{.Context}}` are spelled correctly with the exact casing shown.

### Prompt Already Exists
```
Error: prompt 'my-prompt' already exists. Use 'remove' first to overwrite
```
**Solution**: Remove the existing prompt first, then re-add it:
```bash
raypaste config prompt remove my-prompt
raypaste config prompt add my-prompt
```

## See Also

- `prompt.yaml.example` — Complete prompt template example
- `README.md` — Main documentation
