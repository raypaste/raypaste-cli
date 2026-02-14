# Custom Prompt Guide

This guide explains how to create and use custom prompts in raypaste-cli.

## Overview

raypaste supports two types of prompts:

1. **Built-in prompts** - Compiled into the binary (defined in Go code)
2. **User prompts** - YAML files in `~/.raypaste/prompts/`

## Built-in Prompts

### metaprompt (default)
- **Description**: Generate an optimized meta-prompt from a user's goal
- **Supported lengths**: short, medium, long
- **Use case**: General-purpose prompt optimization

### bulletlist
- **Description**: Organize text by relation and output as a short bulleted list
- **Supported lengths**: short, medium (long mode is intentionally disabled)
- **Use case**: Organizing information into structured, relational bullet points

## Creating Custom Prompts

### Quick Start: Your First Custom Prompt

Let's create an ASCII art prompt that only supports medium mode. Run this command:

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
# Generate ASCII art
raypaste gen "coffee cup" -p ascii-art

# Or try other inputs
raypaste gen "happy cat" -p ascii-art
raypaste gen "rocket ship" -p ascii-art

# This will fail because we only defined 'medium':
raypaste gen "tree" -p ascii-art -l long
# Error: prompt 'ascii-art' does not support output length 'long'
```

### Method 1: User Prompt (YAML) - Recommended

Create a YAML file in `~/.raypaste/prompts/`:

```yaml
name: my-prompt
description: "Brief description of what this prompt does"

# The system prompt template
# Use {{.LengthDirective}} to inject length-specific guidance
system: |
  You are an expert assistant. Your task is to...

  Output length guidance: {{.LengthDirective}}

  CRITICAL:
  - Follow these specific instructions
  - Output format requirements
  - Any other constraints

# Optional: Define custom length directives
# If omitted, default directives will be used
length_directives:
  short: "Your custom short directive here"
  medium: "Your custom medium directive here"
  long: "Your custom long directive here"
```

**Usage:**
```bash
raypaste gen "your input" -p my-prompt
```

### Method 2: Built-in Prompt (Go Code)

For prompts that should be bundled with the CLI:

1. Create a new file in `internal/prompts/defaults/`:

```go
// internal/prompts/defaults/mytemplate.go
package defaults

const MyTemplateTemplate = `Your template content here with {{.LengthDirective}}`
const MyTemplateName = "mytemplate"
const MyTemplateDescription = "What this template does"
```

2. Register it in `internal/prompts/store.go`:

```go
func (s *Store) loadBuiltInPrompts() error {
    // ... existing prompts ...
    
    myPrompt := &Prompt{
        Name:        defaults.MyTemplateName,
        Description: defaults.MyTemplateDescription,
        System:      defaults.MyTemplateTemplate,
        LengthDirectives: map[string]string{
            string(types.OutputLengthShort):  llm.LengthParams[types.OutputLengthShort].Directive,
            string(types.OutputLengthMedium): llm.LengthParams[types.OutputLengthMedium].Directive,
            string(types.OutputLengthLong):   llm.LengthParams[types.OutputLengthLong].Directive,
        },
    }
    
    s.prompts[defaults.MyTemplateName] = myPrompt
    return nil
}
```

3. Rebuild the CLI:
```bash
./build
```

## Restricting Output Lengths

To limit a prompt to specific output lengths, omit unwanted lengths from `length_directives`:

### Example: Short and Medium Only

```yaml
name: quick-summary
description: "Generate quick summaries (no long mode)"
system: |
  Summarize the following text concisely.
  
  {{.LengthDirective}}

length_directives:
  short: "1-2 sentence summary"
  medium: "1 paragraph summary"
  # long is intentionally omitted
```

When a user tries to use long mode with this prompt:
```bash
raypaste gen "text to summarize" -p quick-summary -l long
# Error: prompt 'quick-summary' does not support output length 'long'
```

### Example: Medium and Long Only

```yaml
length_directives:
  medium: "Detailed analysis with examples"
  long: "Comprehensive analysis with examples and edge cases"
  # short is intentionally omitted
```

## Template Variables

Currently supported template variables:

- `{{.LengthDirective}}` - Automatically replaced with the length-specific directive

## Default Length Directives

If you don't specify custom `length_directives`, these defaults are used:

| Length | Max Tokens | Default Directive |
|--------|------------|-------------------|
| short  | 300        | "Keep the generated prompt concise — under 100 words. Focus on the core instruction only." |
| medium | 800        | "Generate a moderately detailed prompt (~150-300 words) with context, constraints, and desired output format." |
| long   | 1500       | "Generate a comprehensive prompt (300-500+ words) including examples, edge cases, tone guidance, and detailed formatting instructions." |

## Best Practices

1. **Clear Instructions**: Be explicit about what the prompt should do
2. **Output Format**: Specify the exact format you want (bullets, paragraphs, code, etc.)
3. **Constraints**: Use CRITICAL sections to emphasize important rules
4. **Length Directives**: Customize them for your specific use case
5. **Testing**: Test with all supported lengths to ensure consistent behavior

## Examples

### Example 1: Code Review Prompt

```yaml
name: code-review
description: "Generate a code review prompt"
system: |
  You are a code review expert. Generate a detailed prompt for reviewing code.

  Output length guidance: {{.LengthDirective}}

  CRITICAL:
  - Focus on code quality, security, and best practices
  - Output ONLY the review prompt itself
  - No preamble or explanations

length_directives:
  short: "Focus on critical issues only"
  medium: "Cover functionality, style, and best practices"
  long: "Include security, performance, testing, and documentation"
```

### Example 2: Email Writer (Short/Medium Only)

```yaml
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

### Example 3: Using Built-in Bulletlist

```bash
# Organize meeting notes into bullets (short mode)
raypaste gen "Discussed Q1 goals, budget concerns, new hires, office relocation" -p bulletlist -l short

# Organize project requirements (medium mode)
raypaste gen "User auth, database schema, API endpoints, frontend UI" -p bulletlist -l medium

# Try long mode (will fail)
raypaste gen "some text" -p bulletlist -l long
# Error: prompt 'bulletlist' does not support output length 'long'
```

## Troubleshooting

### Prompt Not Found
```
Error: prompt not found: my-prompt
```
**Solution**: Check that the YAML file exists in `~/.raypaste/prompts/` and has the correct `name` field.

### Length Not Supported
```
Error: prompt 'my-prompt' does not support output length 'long'
```
**Solution**: This is intentional. Use a supported length (check the prompt's `length_directives`).

### Template Parse Error
```
Error: failed to parse template: ...
```
**Solution**: Check your template syntax. Make sure `{{.LengthDirective}}` is spelled correctly.

## Contributing Prompts

If you create a useful prompt, consider contributing it as a built-in prompt:

1. Create the prompt file in `internal/prompts/defaults/`
2. Register it in `internal/prompts/store.go`
3. Add documentation to README.md
4. Submit a pull request

## See Also

- `config.yaml.example` - Configuration examples
- `prompt.yaml.example` - Complete prompt template example
- `README.md` - Main documentation
