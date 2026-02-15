# Contributing to raypaste-cli

Thank you for your interest in contributing to raypaste-cli! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Common Tasks](#common-tasks)

## Code of Conduct

Be respectful, inclusive, and constructive in all interactions. We're committed to providing a welcoming environment for contributors of all backgrounds and experience levels.

## Getting Started

### Prerequisites

- **Go 1.24 or later** (check `.golang-ci.yml` and `go.mod` for minimum version)
- **Git**
- **Make** (optional, for convenience)
- An **OpenRouter API key** for testing (free tier available at https://openrouter.ai/keys)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/raypaste-cli.git
   cd raypaste-cli
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/raypaste/raypaste-cli.git
   ```

## Development Setup

### Install Dependencies

```bash
go mod download
```

### Build the Project

Using the provided build script:

```bash
./build
```

This compiles the binary and installs it to `/usr/local/bin/raypaste`.

Or manually:

```bash
go build -o raypaste
sudo mv raypaste /usr/local/bin/
```

### Configure for Development

Set up a local config file for testing:

```bash
mkdir -p ~/.raypaste
cp config.yaml.example ~/.raypaste/config.yaml
```

Then edit `~/.raypaste/config.yaml` and add your OpenRouter API key.

Or use environment variables:

```bash
export RAYPASTE_API_KEY=your_api_key_here
```

### Verify Installation

```bash
raypaste generate "test message"
```

## Making Changes

### Create a Branch

Use descriptive branch names:

```bash
git checkout -b feature/add-streaming-support
git checkout -b fix/handle-network-errors
git checkout -b docs/improve-readme
```

### Understand the Architecture

The project is organized into key packages:

- **`cmd/`** - Cobra commands (root, generate, interactive)
- **`internal/config/`** - Configuration management with Viper
- **`internal/llm/`** - OpenRouter API client and streaming
- **`internal/output/`** - Terminal formatting and colors
- **`internal/prompts/`** - Prompt template system
- **`internal/clipboard/`** - Cross-platform clipboard operations
- **`pkg/types/`** - Shared types

See the [`.cursor/rules/raypaste-project.mdc`](.cursor/rules/raypaste-project.mdc) file for detailed architecture and coding patterns.

### Code Organization Principles

- **Avoid nested anonymous structs** - Use explicit named types for better maintainability
- **Error handling** - Provide clear, actionable error messages
- **Configuration access** - Use the config package's methods with proper fallbacks
- **Prompt rendering** - Use the prompts package for template rendering
- **Colors and output** - Use `internal/output` package for all terminal output

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Run Tests for a Specific Package

```bash
go test ./internal/config -v
go test ./cmd -v
```

### Test Interactive Mode

To test the interactive REPL:

```bash
raypaste interactive
# or
raypaste i
```

Try various slash commands:

- `/length short` - Test output length switching
- `/model cerebras-llama-8b` - Test model switching
- `/prompt metaprompt` - Test prompt switching
- `/copy` - Test clipboard copying
- `/quit` - Test exit

### Linting

The project uses golangci-lint. Configuration is in `.golangci.yml`.

```bash
golangci-lint run
```

### Manual Integration Testing

Test the generate command:

```bash
# Basic usage
raypaste gen "create a REST API"

# With flags
raypaste gen "write code" --length long --copy

# From stdin
echo "my goal" | raypaste gen

# Specify a model
raypaste gen "test" -m cerebras-gpt-oss-120b
```

## Submitting Changes

### Commit Best Practices

Write clear, descriptive commit messages:

```bash
git commit -m "feat: add streaming support for real-time output"
git commit -m "fix: handle network timeouts gracefully"
git commit -m "docs: improve configuration examples"
git commit -m "refactor: simplify prompt rendering logic"
```

Use conventional commit prefixes:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation updates
- `refactor:` - Code refactoring without behavior change
- `test:` - Adding or updating tests
- `chore:` - Build, dependency, or tooling changes

### Keep Commits Atomic

Each commit should represent a single logical change. Avoid mixing unrelated changes in one commit.

### Push Your Branch

```bash
git push origin feature/your-feature-name
```

### Create a Pull Request

1. Go to the repository on GitHub
2. Click "New Pull Request"
3. Select your branch
4. Provide a clear title and description:
   - **What** - What does this change do?
   - **Why** - Why is this change needed?
   - **How** - How does it solve the problem?
5. Reference any related issues with `#issue-number`

### Pull Request Checklist

Before submitting, ensure:

- [ ] Your branch is up to date with `main`
- [ ] All tests pass: `go test ./...`
- [ ] Linting passes: `golangci-lint run`
- [ ] Code follows the project's conventions (see [Coding Standards](#coding-standards))
- [ ] You've added tests for new functionality
- [ ] You've updated documentation if needed
- [ ] Commit messages are clear and descriptive
- [ ] No debug code or commented-out code is left

### Code Review

A maintainer will review your PR. They may:

- Ask for clarifications or changes
- Suggest improvements
- Request additional tests

Please respond promptly to feedback and make requested changes in new commits (don't force-push).

## Project Structure

```
raypaste-cli/
├── cmd/                          # Cobra commands
│   ├── root.go                  # Root command setup
│   ├── generate.go              # One-shot generation command
│   └── interactive.go           # Interactive REPL command
├── internal/
│   ├── config/                  # Configuration management
│   │   ├── config.go           # Config loading and access
│   │   └── models.go           # Model definitions
│   ├── llm/                     # LLM integration
│   │   ├── client.go           # OpenRouter API client
│   │   ├── streaming.go        # SSE streaming parser
│   │   └── router.go           # Model routing and token mapping
│   ├── output/                  # Terminal output formatting
│   │   └── formatter.go        # Color formatting and markdown detection
│   ├── prompts/                 # Prompt template system
│   │   ├── store.go            # Template loading and rendering
│   │   └── defaults/           # Built-in templates
│   └── clipboard/               # Clipboard operations
│       └── clipboard.go        # Cross-platform clipboard
├── pkg/
│   └── types/                   # Shared types
│       └── types.go            # OutputLength, Message, etc.
├── main.go                      # Entry point
├── go.mod & go.sum              # Dependencies
├── config.yaml.example          # Config template
├── README.md                    # User documentation
├── CHANGELOG.md                 # Release notes
├── CONTRIBUTING.md              # This file
└── build                        # Build script
```

## Coding Standards

### Style Guidelines

Follow Go best practices and conventions:

- Use `gofmt` for formatting
- Use meaningful variable and function names
- Keep functions small and focused
- Document exported types and functions with comments
- Use `internal/` packages for code not meant to be imported externally

### Error Handling

Always handle errors explicitly:

```go
// Good
if err != nil {
    return fmt.Errorf("failed to fetch config: %w", err)
}

// For network errors, retry once
if err != nil || resp.StatusCode >= 500 {
    time.Sleep(1 * time.Second)
    resp, err = client.Do(req)
}
```

### Type Naming

Use explicit named types instead of anonymous structs:

```go
// Good
type APIResponse struct {
    Status string
    Data   string
}

// Avoid
type SomeFunc func() (struct {
    Status string
    Data   string
}, error)
```

### Configuration Access

Use the config package with proper fallbacks:

```go
cfg := config.GetConfig()
apiKey := cfg.GetAPIKey()           // Falls back to env var
model := cfg.GetDefaultModel()      // Falls back to default
length := cfg.GetDefaultLength()    // Falls back to medium
```

### Output and Colors

Always use the `internal/output` package:

```go
// Use these color functions
output.Red("Error message")
output.Green("Success message")
output.Yellow("Warning message")
output.Cyan("Info message")
output.Blue("Header text")
output.Bold("Important text")

// Respect NO_COLOR environment variable (automatic)
```

## Common Tasks

### Adding a New Model

1. Edit `internal/config/models.go`
2. Add to `DefaultModels` map:
   ```go
   "my-model": {
       ID:       "provider/model-name",
       Provider: "provider",
       Tier:     "fast",
   }
   ```
3. Test with: `raypaste gen "test" -m my-model`

### Adding a New Prompt Template

1. Create a new YAML file in `~/.raypaste/prompts/` or `internal/prompts/defaults/`
2. Template structure:
   ```yaml
   name: my-prompt
   description: "What this prompt does"
   system: |
     System prompt text.
     Use {{.LengthDirective}} for length-specific guidance.
   length_directives:
     short: "..."
     medium: "..."
     long: "..."
   ```
3. Test with: `raypaste gen "input" -p my-prompt`

### Modifying Output Lengths

1. Edit `internal/llm/router.go`
2. Update the `LengthParams` map
3. Run tests: `go test ./internal/llm`

### Adding CLI Flags

1. Edit the relevant command file in `cmd/`
2. Use Cobra's `Cmd.Flags()`:
   ```go
   generateCmd.Flags().StringVarP(&flagVar, "flag-name", "f", "default", "help text")
   ```
3. Test the flag: `raypaste gen "test" --flag-name value`

## Getting Help

- Check the [README.md](README.md) for usage information
- Review [PROMPT_GUIDE.md](PROMPT_GUIDE.md) for prompt template details
- Look at existing code for patterns
- Ask in GitHub issues or discussions

## Additional Resources

- [Cobra Documentation](https://cobra.dev/)
- [Viper Configuration](https://github.com/spf13/viper)
- [Go Error Handling](https://pkg.go.dev/errors)
- [OpenRouter API Docs](https://openrouter.ai/docs)

Thank you for contributing to raypaste-cli!
