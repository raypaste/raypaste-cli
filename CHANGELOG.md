# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.5]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.5
[0.1.4]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.4
[0.1.3]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.3
[0.1.2]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.2
[0.1.1]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.1
[0.1.0]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.0
