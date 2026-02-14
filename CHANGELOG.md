# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/raypaste/raypaste-cli/releases/tag/v0.1.0
