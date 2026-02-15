# raypaste-cli

Ultra-fast AI-optimized meta prompts from your input text in your terminal.

A powerful [Cobra](https://cobra.dev/)-based CLI that generates meta-prompts and AI completions via OpenRouter, with configurable output lengths, fast model routing, and interactive REPL mode with streaming.

## Features

- **Fast Generation**: Generate optimized prompts in milliseconds using fast inference from small models running on
  hardware from Cerebras.
- **Multiple Output Lengths**: Choose between short, medium, or long responses
- **Interactive Mode**: REPL with streaming output and slash commands
- **Clipboard Integration**: Auto-copy results to clipboard
- **Custom Prompts**: Create and manage your own prompt templates
- **Flexible Configuration**: Configure via YAML, environment variables, or CLI flags
- **OpenRouter Integration**: Access to multiple LLM providers through a single API
- **Colored Output**: Automatic markdown detection with colored status messages for better readability

## Installation

### From Source

```bash
git clone https://github.com/raypaste/raypaste-cli.git
cd raypaste-cli
./build
```

Or manually:

```bash
go build -o raypaste
sudo mv raypaste /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/raypaste/raypaste-cli@latest
```

### Using Homebrew (macOS)

```bash
brew tap raypaste/tap
brew install --cask raypaste
```

## Quick Start

1. **Get an OpenRouter API key** from [https://openrouter.ai/keys](https://openrouter.ai/keys)

2. **Set your API key** (choose one method):

   **Option A: Environment Variable (Recommended for quick start)**

   ```bash
   export RAYPASTE_API_KEY=your_api_key_here
   ```

   To make it permanent, add to your shell config:

   ```bash
   # For zsh (macOS default)
   echo 'export RAYPASTE_API_KEY=your_api_key_here' >> ~/.zshrc
   source ~/.zshrc

   # For bash
   echo 'export RAYPASTE_API_KEY=your_api_key_here' >> ~/.bashrc
   source ~/.bashrc
   ```

   **Option B: Config File**

   ```bash
   mkdir -p ~/.raypaste
   cp config.yaml.example ~/.raypaste/config.yaml
   # Edit ~/.raypaste/config.yaml and add your API ke
   nano ~/.raypaste/config.yaml
   ```

   **Note**: The `.env` file in the project directory is for reference only. Go programs don't automatically load `.env` files. You must either export the environment variable or use the config file at `~/.raypaste/config.yaml`.

3. **Generate your first prompt**:

```bash
raypaste generate "help me write a blog post about Go"
```

## Usage

### Generate Command

Generate an optimized prompt from your input (one-shot mode):

```bash
# Basic usage
raypaste generate "create a REST API for a todo app"

# Short alias
raypaste gen "analyze CSV data"

# With flags
raypaste gen "write a blog post" --length long --copy

# From stdin
echo "my goal" | raypaste gen

# Specify model
raypaste gen "optimize this code" -m cerebras-gpt-oss-120b
```

**Flags:**

- `-l, --length`: Output length (short, medium, long) - default: medium
- `-m, --model`: Model alias or OpenRouter ID - default: cerebras-llama-8b
- `-p, --prompt`: Prompt template name - default: metaprompt
- `-c, --copy`: Auto-copy result to clipboard
- `--config`: Custom config file path

### Interactive Mode

Start an interactive REPL session with streaming output:

```bash
raypaste interactive
# or
raypaste i
```

**Slash Commands:**

- `/length <short|medium|long>` - Change output length
- `/model <alias>` - Switch model
- `/prompt <name>` - Switch prompt template
- `/copy` - Copy last response to clipboard
- `/help` - Show help
- `/quit` or `/exit` - Exit REPL

**Keyboard Shortcuts:**

- `Ctrl+C` - Cancel current generation
- `Ctrl+D` - Exit REPL

**Example Session:**

```
raypaste interactive mode
Model: cerebras-llama-8b | Length: medium | Prompt: metaprompt
Type /help for commands, /quit to exit

> create a function to parse JSON
[streaming output appears here...]

> /length short
Length set to: short

> optimize the above for performance
[shorter streaming output...]

> /copy
✓ Copied to clipboard

> /quit
Goodbye!
```

## Configuration

### Configuration Hierarchy

Configuration is loaded in the following order (later sources override earlier ones):

1. Default values
2. Config file (`~/.raypaste/config.yaml`)
3. Environment variables (`RAYPASTE_*`)
4. CLI flags

### Config File

Create `~/.raypaste/config.yaml`:

```yaml
# OpenRouter API key
api_key: "your_api_key_here"

# Default model
default_model: cerebras-llama-8b

# Default output length: short, medium, or long
default_length: medium

# Auto-copy results to clipboard
auto_copy: false

# Temperature for generation (0.0 to 1.0)
temperature: 0.7

# Custom model definitions
models:
  my-custom-model:
    id: "provider/model-name"
    provider: "provider-name"
    tier: "fast"
```

See `config.yaml.example` for a complete example.

### Environment Variables

```bash
export RAYPASTE_API_KEY=your_api_key_here
export RAYPASTE_DEFAULT_MODEL=cerebras-llama-8b
export RAYPASTE_DEFAULT_LENGTH=medium
```

See `.env.example` for a complete example.

## Models

### Built-in Models

| Alias                   | Model ID                           | Provider | Tier    |
| ----------------------- | ---------------------------------- | -------- | ------- |
| `cerebras-llama-8b`     | `meta-llama/llama-3.1-8b-instruct` | Cerebras | Fastest |
| `cerebras-gpt-oss-120b` | `openai/gpt-oss-120b`              | Cerebras | Fast    |
| `openai-gpt5-nano`      | `openai/gpt5-nano`                 | OpenAI   | Fast    |

### Using Custom Models

You can use any OpenRouter model by:

1. **Direct model ID**: Use the full OpenRouter model ID as the model flag

   ```bash
   raypaste gen "hello" -m "anthropic/claude-3-opus"
   ```

2. **Custom alias**: Define in `config.yaml`
   ```yaml
   models:
     my-claude:
       id: "anthropic/claude-3-opus"
       provider: "anthropic"
       tier: "powerful"
   ```
   Then use: `raypaste gen "hello" -m my-claude`

## Output Lengths

Output length controls both the desired response length and the `max_tokens` parameter:

| Length   | Max Tokens | Use Case                               |
| -------- | ---------- | -------------------------------------- |
| `short`  | 300        | Quick, concise prompts (~100 words)    |
| `medium` | 800        | Balanced detail (~150-300 words)       |
| `long`   | 1500       | Comprehensive prompts (300-500+ words) |

The system prompt includes guidance for each length to ensure appropriate output.

## Colored Output

raypaste automatically formats prompts with colored output:

- **Markdown Detection**: The CLI automatically detects when output appears to be markdown (based on syntax patterns like headers, lists, code blocks, etc.) and prepares it for potential syntax highlighting in future versions.

- **Disabling Colors**: Colors respect the `NO_COLOR` environment variable and terminal capabilities. To disable colors:
  ```bash
  NO_COLOR=1 raypaste gen "hello world"
  ```

## Custom Prompt Templates

### Built-in Prompts

raypaste includes the following built-in prompts:

| Name         | Description                                                   | Supported Lengths   |
| ------------ | ------------------------------------------------------------- | ------------------- |
| `metaprompt` | Generate an optimized meta-prompt from a user's goal          | short, medium, long |
| `bulletlist` | Organize text by relation and output as a short bulleted list | short, medium       |

### Creating Your First Custom Prompt

Let's create an ASCII art prompt to get you started. This prompt will only support medium mode:

Running the below adds `ascii-art.yaml` with the below prompt info/content to your `.raypaste/prompts` folder.

```bash
mkdir -p ~/.raypaste/prompts && cat > ~/.raypaste/prompts/ascii-art.yaml << 'EOF'
name: ascii-art
description: "Convert text into ASCII art/emoji representation"
system: |
  You are an ASCII art expert. Create creative ASCII art or emoji-based representations of the input text.

  Output length guidance: {{.LengthDirective}}

  CRITICAL:
  - Output ONLY the ASCII art itself, no explanations or preamble
  - Use creative arrangements of ASCII characters or emojis
  - Make it visually appealing and recognizable
  - Keep it readable in a terminal

length_directives:
  medium: "Create a medium-sized ASCII art (5-15 lines) with good detail and creativity"
EOF
```

**Try it out:**

```bash
raypaste gen "coffee cup" -p ascii-art
raypaste gen "happy cat" -p ascii-art
raypaste gen "rocket ship" -p ascii-art
```

### Creating Custom Prompts

Create YAML files in `~/.raypaste/prompts/`:

```yaml
# ~/.raypaste/prompts/code-review.yaml
name: code-review
description: "Generate a code review prompt"
system: |
  You are a code review expert. Generate a detailed prompt for reviewing code.

  Output length guidance: {{.LengthDirective}}

  Return only the generated prompt.

length_directives:
  short: "Keep the review prompt concise, focusing on critical issues only."
  medium: "Generate a balanced review prompt covering functionality, style, and best practices."
  long: "Generate a comprehensive review prompt including security, performance, testing, and documentation."
```

**Restricting Output Lengths**: To limit a prompt to specific output lengths, simply omit the unwanted lengths from `length_directives`. For example, to support only short and medium modes:

```yaml
length_directives:
  short: "Your short directive here"
  medium: "Your medium directive here"
  # long is intentionally omitted
```

See `prompt.yaml.example` for a complete example, or read the full [Custom Prompt Guide](PROMPT_GUIDE.md).

### Using Custom Prompts

```bash
raypaste gen "review my API code" -p code-review
```

### Template Variables

- `{{.LengthDirective}}` - Automatically replaced with length-specific guidance

## Examples

### Generate a Blog Post Outline

```bash
raypaste gen "write a blog post about microservices architecture" -l long -c
```

### Quick Code Snippet

```bash
raypaste gen "function to validate email addresses" -l short
```

### Interactive Brainstorming

```bash
raypaste i
> brainstorm features for a todo app
> /length long
> expand on the collaboration features
> /copy
> /quit
```

### Pipeline with Other Tools

```bash
# Generate and pipe to file
raypaste gen "API documentation structure" > api-docs-prompt.txt

# Use with other CLI tools
cat requirements.txt | raypaste gen "analyze these dependencies" -l medium
```

## Troubleshooting

### API Key Not Found

```
Error: API key not found. Set RAYPASTE_API_KEY environment variable or add to config.yaml
```

**Solution**: The CLI looks for your API key in two places:

1. **Environment Variable**: `RAYPASTE_API_KEY` must be exported in your current shell session

   ```bash
   export RAYPASTE_API_KEY=your_api_key_here
   ```

2. **Config File**: `~/.raypaste/config.yaml` (note: this is in your home directory, not the project directory)
   ```bash
   mkdir -p ~/.raypaste
   cp config.yaml.example ~/.raypaste/config.yaml
   # Edit the file and add your API key
   ```

**Common Issues**:

- Having a `.env` file in the project directory won't work - Go doesn't automatically load `.env` files
- Having `config.yaml` in the project directory won't work - the CLI looks in `~/.raypaste/config.yaml`
- Setting the variable without `export` won't work - it must be exported to be visible to the program
- You don't need to rebuild the executable after setting the API key

### Clipboard Not Working

```
Warning: Could not copy to clipboard: ...
```

**Solution**: The clipboard feature requires a display server. On headless systems, the warning is informational only and doesn't affect functionality.

### Model Not Found

```
Error: failed to resolve model: model not found
```

**Solution**: Check that the model alias exists in the built-in models or your custom models. Use a direct OpenRouter model ID if needed.

### Connection Timeout

```
Error: generation failed: context deadline exceeded
```

**Solution**: Check your internet connection. The CLI retries once automatically. For longer generations, the timeout is 30 seconds for generate and 60 seconds for interactive mode.

## Development

### Running Tests

```bash
go test ./...
```

### Building from Source

```bash
# Using the build script (recommended)
./build

# Or manually
go build -o raypaste
```

### Developer Workflow

This project uses Go 1.24's new `tool` directive for consistent development tooling.

**Prerequisites:**

- Go 1.24 or later
- `golangci-lint` (automatically installed via `go tool`)

**Running Linter Locally:**

```bash
# Using go tool (recommended - uses pinned version)
go tool golangci-lint run ./...
```

**Running Tests:**

```bash
# Run all tests
go test ./...

# Run tests with race detector and coverage
go test ./... -race -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

**Code Formatting:**

```bash
# Format and fix imports
goimports -w ./cmd ./internal ./pkg

# Or use the built-in formatter
go fmt ./...
```

**Before Committing:**

1. Run linter: `go tool golangci-lint run ./...`
2. Run tests: `go test ./... -race`
3. Verify no lint errors appear

The linter is configured in `.golangci.yml` and will be run automatically by GitHub Actions on pull requests. Running it locally before pushing ensures CI passes on first try.

**GitHub Actions Workflows:**

- `.github/workflows/go-test.yml` - Runs tests and coverage reporting
- `.github/workflows/go-lint.yml` - Runs linter on all changes

### Project Structure

```
.
├── cmd/                 # Cobra commands
│   ├── root.go          # Root command & config
│   ├── generate.go      # Generate command
│   └── interactive.go   # Interactive REPL
├── internal/
│   ├── config/          # Configuration management
│   ├── llm/             # OpenRouter client
│   ├── prompts/         # Prompt store & templates
│   └── clipboard/       # Clipboard operations
├── pkg/
│   └── types/           # Shared types
└── main.go              # Entry point
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [OpenRouter](https://openrouter.ai/) for LLM access
- Powered by [Viper](https://github.com/spf13/viper) for configuration
- Clipboard support via [golang.design/x/clipboard](https://golang.design/x/clipboard)
- Interactive mode with [readline](https://github.com/chzyer/readline)
