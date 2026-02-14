/*
Copyright © 2026 Raypaste
*/
package output

import (
	"testing"
)

func TestIsMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "markdown with headers",
			input: `# Title
## Subtitle
Some text here`,
			expected: true,
		},
		{
			name: "markdown with lists",
			input: `Here are some items:
- Item 1
- Item 2
- Item 3`,
			expected: true,
		},
		{
			name: "markdown with ordered lists",
			input: `Steps to follow:
1. First step
2. Second step
3. Third step`,
			expected: true,
		},
		{
			name:     "markdown with code blocks",
			input:    "```go\nfunc main() {\n}\n```",
			expected: true,
		},
		{
			name:     "markdown with bold and italic",
			input:    `This is **bold** and this is *italic*`,
			expected: true,
		},
		{
			name:     "markdown with inline code",
			input:    "Use the `fmt.Println()` function",
			expected: true,
		},
		{
			name:     "markdown with links",
			input:    "Check out [this link](https://example.com)",
			expected: true,
		},
		{
			name:     "plain text",
			input:    "This is just plain text without any markdown syntax",
			expected: false,
		},
		{
			name: "mixed content with low markdown density",
			input: `This is mostly plain text.
It has many lines.
But only one markdown element:
- A single bullet point
And then lots more plain text.
More text here.
Even more text.
Still more text.
Text continues.
More plain text.`,
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("IsMarkdown() = %v, expected %v for input:\n%s", result, tt.expected, tt.input)
			}
		})
	}
}

func TestColorizeMarkdown(t *testing.T) {
	// Test that ColorizeMarkdown doesn't panic and returns a string
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "markdown text",
			input: `# Header
- List item`,
		},
		{
			name:  "plain text",
			input: "Just plain text",
		},
		{
			name:  "empty string",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorizeMarkdown(tt.input)
			if result == "" && tt.input != "" {
				t.Errorf("ColorizeMarkdown() returned empty string for non-empty input")
			}
		})
	}
}

func TestGeneratingMessage(t *testing.T) {
	msg := GeneratingMessage()
	if msg == "" {
		t.Error("GeneratingMessage() returned empty string")
	}
}

func TestCopiedMessage(t *testing.T) {
	msg := CopiedMessage()
	if msg == "" {
		t.Error("CopiedMessage() returned empty string")
	}
}

func TestAnsiColors(t *testing.T) {
	// Test that the new color functions don't panic and return content
	testCases := []struct {
		name string
		f    func(...interface{}) string
	}{
		{"Black", Black},
		{"BoldRed", BoldRed},
		{"HiGreen", HiGreen},
		{"BoldHiBlue", BoldHiBlue},
		{"Underline", Underline},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.f("test")
			if result == "" {
				t.Errorf("%s() returned empty string", tc.name)
			}
		})
	}
}
