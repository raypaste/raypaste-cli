# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.4] - 2026-02-16

### Changed

- Updated main README and Github workflows README

## [0.2.3] - 2026-02-16

### Changed

- Improved interactive mode UI and adjusted colors.

### Fixed

- **Inline code markdown formatting**: Fixed underscores in inline code blocks (e.g., `` `usage.prompt_tokens` ``) being incorrectly interpreted as italic markdown. Now properly protects inline code blocks from other markdown pattern processing using placeholder substitution.

## [0.2.2] - 2026-02-15

### Fixed

- **GitHub Actions tag release workflow**: Fixed trigger from `pull_request` event to `push` event on main branch to ensure tag-release runs reliably after PR merges

## [0.2.1] - 2026-02-15

### Added

- **Improved GitHub Actions release workflow**: 2 new actions, check for valid new version in changelog and tag release on merge to main
- **Auto-complete support**: Tab complete for /model and /prompt and for "/" slash command suggestions.

### Changed

- Moved and organized bulk of interactive mode logic to separate files in `/internal/interactive/`

## [0.2.0] - 2026-02-15

### Added

- **Project context awareness**: Automatically loads project context from `CLAUDE.md`, `AGENTS.md`, or `.cursor/rules/` files to inform prompt generation
- `/clear` command in interactive mode to clear the screen
- `version` command to display version information (`raypaste version`)
- `--version` and `-v` flags to display version information
- Version embedding via ldflags during build process (includes version, git commit, and build date)
- Comprehensive test suite for command handling (`cmd/cmd_test.go`)
- Tests for project context finder (`internal/projectcontext/finder_test.go`)
- Enhanced status messages showing output length and project context file when active

### Changed

- **Default model updated to `cerebras-gpt-oss-120b`** (prev `cerebras-llama-8b`)
- Prompt templates now support `{{.Context}}` variable for project-aware generation
- `GeneratingMessage()` now includes output length and context file information
- Updated GoReleaser configuration to embed version information in release binaries
- Build script now embeds version information via ldflags

## [0.1.6] - 2026-02-15

### Added

- Comprehensive streaming test coverage for multiple content payload types (string, array, object formats)

### Changed

- **Enhanced streaming output parsing** - Now correctly handles OpenAI GPT-5 models with array/object content payloads in addition to string content
- **Improved GitHub Actions release workflow** - Multi-platform builds now run on native platforms (macOS, Linux, Windows) for better compatibility
- **macOS clipboard support** - Binaries are now built on `macos-latest` runners with `CGO_ENABLED=1` to ensure clipboard functionality works correctly
- Updated interactive mode welcome message to highlight `Ctrl+D` as an alternative exit method alongside `/quit` command
- Improved clipboard troubleshooting documentation with detailed root cause explanation and comprehensive solutions

### Fixed

- Fixed OpenAI GPT-5 model ID from `openai/gpt5-nano` to `openai/gpt-5-nano` across all configurations
- Fixed streaming response handling for GPT-5 models by implementing `max_completion_tokens` and `reasoning_effort` parameters to prevent empty outputs
- Fixed streaming content extraction to support complex nested content structures (text objects, content arrays) from various API providers
- Fixed macOS clipboard failures caused by CGO-disabled builds in release binaries

## [0.1.5] - 2026-02-15

### Changed

- **Copy to clipboard now enabled by default** - Generated prompts are automatically copied to clipboard in both `generate` and `interactive` modes
- Use the `--no-copy` flag in generate mode to disable this behavior if needed

## [0.1.4] - 2026-02-15

### Fixed

- Fixed Homebrew cask installation URL - removed spaces from archive name template to create valid URIs

## [0.1.3] - 2026-02-15

### Added

- Automated GitHub Actions release workflow with multi-platform binary builds
- Homebrew tap integration for easy installation via `brew tap raypaste/tap && brew install --cask raypaste`
- GoReleaser configuration for streamlined release process
- Automatic changelog extraction for GitHub releases

### Changed

- Release process now fully automated when pushing version tags

## [0.1.2] - 2026-02-14

### Fixed

- Fixed interactive mode breaking when pasting long multi-line inputs by implementing proper input buffering and detection
- Stream cancellation now properly handled with context cancellation triggering TCP connection close for immediate API billing stop
- Improved timeout handling: moved context deadline from request level to generation level for better control of long-running streams
- Added mid-stream error detection from OpenRouter API, now properly catches error chunks and finish_reason indicators
- SSE comment lines (starting with `:`) now properly ignored per SSE specification

### Changed

- Updated metaprompt to avoid assumptions on technology specifics - no longer adds tech constraints unless explicitly mentioned in user input
- Enhanced interactive mode prompt handling with input validation and visual feedback improvements
- Improved stream handling with drain mechanism for buffered lines during cancellation
- Added context support to streaming generation with 120-second timeout

## [0.1.1] - 2026-02-14

### Added

- Colored output formatting for CLI messages and markdown content
- Syntax highlighting for code blocks in streaming output
- Color-coded status messages (success in green, errors in red, info in cyan/yellow)
- Enhanced visual feedback for interactive mode

### Fixed

- Corrected module path in go.mod from `raypaste-cli` to `github.com/raypaste/raypaste-cli` to enable proper `go install` functionality
- Updated all internal import paths to use the correct module path

## [0.1.0] - 2026-02-14

### Added

- Initial release of raypaste-cli
- Generate command for one-shot meta-prompt generation
- Interactive REPL mode with streaming output
- Support for multiple output lengths (short, medium, long)
- OpenRouter API integration with configurable models
- Built-in models: cerebras-llama-8b, cerebras-gpt-oss-120b, openai-gpt5-nano
- Clipboard integration for auto-copying results
- Custom prompt template system with YAML-based templates
- Built-in prompt templates: metaprompt, bulletlist
- Flexible configuration via YAML file, environment variables, and CLI flags
- Configuration hierarchy: defaults → config file → env vars → CLI flags
- Slash commands in interactive mode (/length, /model, /prompt, /copy, /help, /quit)
- Automatic retry logic for network errors
- Comprehensive documentation and examples
- GitHub Actions workflows for testing and linting
- Go 1.24 tool directive for consistent development tooling

### Features

- **Fast Generation**: Optimized for speed using Cerebras inference
- **Multiple Output Lengths**: Control response size with max_tokens and system directives
- **Interactive Mode**: REPL with streaming SSE output
- **Clipboard Integration**: Cross-platform clipboard support
- **Custom Prompts**: Create and manage prompt templates
- **Flexible Configuration**: Multiple configuration methods
- **Model Flexibility**: Use built-in aliases or any OpenRouter model ID

[0.2.2]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.2.2
[0.2.1]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.2.1
[0.2.0]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.2.0
[0.1.6]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.6
[0.1.5]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.5
[0.1.4]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.4
[0.1.3]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.3
[0.1.2]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.2
[0.1.1]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.1
[0.1.0]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.0
