/*
Copyright © 2026 Raypaste
*/
package output

import (
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var (
	// Color definitions
	lightBlue   = color.New(color.FgCyan).SprintFunc()
	green       = color.New(color.FgGreen).SprintFunc()
	headerStyle = color.New(color.FgCyan, color.Bold).SprintFunc()
	listStyle   = color.New(color.FgYellow).SprintFunc()
	quoteStyle  = color.New(color.FgHiBlack).SprintFunc()
	tableStyle  = color.New(color.FgBlue).SprintFunc()
	codeStyle   = color.New(color.FgGreen).SprintFunc()
	boldStyle   = color.New(color.Bold).SprintFunc()
	italicStyle = color.New(color.FgMagenta).SprintFunc()
	linkStyle   = color.New(color.FgBlue, color.Underline).SprintFunc()
	ruleStyle   = color.New(color.Faint).SprintFunc()

	// ANSI Basic Colors
	Black   = color.New(color.FgBlack).SprintFunc()
	Red     = color.New(color.FgRed).SprintFunc()
	Green   = color.New(color.FgGreen).SprintFunc()
	Yellow  = color.New(color.FgYellow).SprintFunc()
	Blue    = color.New(color.FgBlue).SprintFunc()
	Magenta = color.New(color.FgMagenta).SprintFunc()
	Cyan    = color.New(color.FgCyan).SprintFunc()
	White   = color.New(color.FgWhite).SprintFunc()

	// ANSI Bold Colors
	Bold        = color.New(color.Bold).SprintFunc()
	BoldBlack   = color.New(color.FgBlack, color.Bold).SprintFunc()
	BoldRed     = color.New(color.FgRed, color.Bold).SprintFunc()
	BoldGreen   = color.New(color.FgGreen, color.Bold).SprintFunc()
	BoldYellow  = color.New(color.FgYellow, color.Bold).SprintFunc()
	BoldBlue    = color.New(color.FgBlue, color.Bold).SprintFunc()
	BoldMagenta = color.New(color.FgMagenta, color.Bold).SprintFunc()
	BoldCyan    = color.New(color.FgCyan, color.Bold).SprintFunc()
	BoldWhite   = color.New(color.FgWhite, color.Bold).SprintFunc()

	// ANSI High Intensity Colors
	HiBlack   = color.New(color.FgHiBlack).SprintFunc()
	HiRed     = color.New(color.FgHiRed).SprintFunc()
	HiGreen   = color.New(color.FgHiGreen).SprintFunc()
	HiYellow  = color.New(color.FgHiYellow).SprintFunc()
	HiBlue    = color.New(color.FgHiBlue).SprintFunc()
	HiMagenta = color.New(color.FgHiMagenta).SprintFunc()
	HiCyan    = color.New(color.FgHiCyan).SprintFunc()
	HiWhite   = color.New(color.FgHiWhite).SprintFunc()

	// ANSI Bold High Intensity Colors
	BoldHiBlack   = color.New(color.FgHiBlack, color.Bold).SprintFunc()
	BoldHiRed     = color.New(color.FgHiRed, color.Bold).SprintFunc()
	BoldHiGreen   = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	BoldHiYellow  = color.New(color.FgHiYellow, color.Bold).SprintFunc()
	BoldHiBlue    = color.New(color.FgHiBlue, color.Bold).SprintFunc()
	BoldHiMagenta = color.New(color.FgHiMagenta, color.Bold).SprintFunc()
	BoldHiCyan    = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	BoldHiWhite   = color.New(color.FgHiWhite, color.Bold).SprintFunc()

	// Underline
	Underline    = color.New(color.Underline).SprintFunc()
	headerRE     = regexp.MustCompile(`^#{1,6}\s`)
	unorderedRE  = regexp.MustCompile(`^\s*[-*+]\s`)
	orderedRE    = regexp.MustCompile(`^\s*\d+\.\s`)
	quoteRE      = regexp.MustCompile(`^\s*>\s`)
	tableRE      = regexp.MustCompile(`^\s*\|\s.*\s\|\s*$`)
	ruleRE       = regexp.MustCompile(`^\s*---\s*$`)
	unorderedCap = regexp.MustCompile(`^(\s*[-*+]\s)(.*)$`)
	orderedCap   = regexp.MustCompile(`^(\s*\d+\.\s)(.*)$`)
	quoteCap     = regexp.MustCompile(`^(\s*>\s)(.*)$`)
	inlineCodeRE = regexp.MustCompile("`[^`]+`")
	boldRE       = regexp.MustCompile(`\*\*[^*\n]+\*\*`)
	italicRE     = regexp.MustCompile(`_[^_\n]+_`)
	linkRE       = regexp.MustCompile(`\[[^\]]+\]\([^)]+\)`)

	// Markdown patterns
	markdownPatterns = []*regexp.Regexp{
		headerRE,                             // Headers
		unorderedRE,                          // Unordered lists
		orderedRE,                            // Ordered lists
		regexp.MustCompile(`\*\*.*?\*\*`),    // Bold
		regexp.MustCompile(`\*.*?\*`),        // Italic
		inlineCodeRE,                         // Inline code
		regexp.MustCompile("```"),            // Code blocks
		regexp.MustCompile(`\[.*?\]\(.*?\)`), // Links
		quoteRE,                              // Blockquotes
		ruleRE,                               // Horizontal rules
		tableRE,                              // Tables
	}
)

// IsMarkdown checks if the text appears to be markdown
func IsMarkdown(text string) bool {
	lines := strings.Split(text, "\n")
	markdownLineCount := 0

	// Check first 20 lines or all lines if fewer
	checkLines := len(lines)
	if checkLines > 20 {
		checkLines = 20
	}

	for i := 0; i < checkLines; i++ {
		line := lines[i]
		for _, pattern := range markdownPatterns {
			if pattern.MatchString(line) {
				markdownLineCount++
				break
			}
		}
	}

	// If more than 20% of checked lines have markdown syntax, consider it markdown
	threshold := float64(checkLines) * 0.2
	return float64(markdownLineCount) >= threshold
}

// ColorizeMarkdown applies syntax highlighting to markdown text
func ColorizeMarkdown(text string) string {
	if color.NoColor || !IsMarkdown(text) {
		return text
	}

	lines := strings.Split(text, "\n")
	var out strings.Builder
	inCodeBlock := false

	for idx, line := range lines {
		out.WriteString(colorizeMarkdownLine(line, &inCodeBlock))
		if idx < len(lines)-1 {
			out.WriteByte('\n')
		}
	}

	return out.String()
}

// StreamingColorizer handles token-by-token colorization for streaming output
type StreamingColorizer struct {
	buffer      strings.Builder
	inCodeBlock bool
}

// NewStreamingColorizer creates a new streaming colorizer
func NewStreamingColorizer() *StreamingColorizer {
	return &StreamingColorizer{}
}

// ProcessToken processes a token and returns the colorized version
// For streaming, we apply colors in real-time as tokens arrive
func (sc *StreamingColorizer) ProcessToken(token string) string {
	if color.NoColor {
		return token
	}

	var out strings.Builder
	for _, r := range token {
		if r != '\n' {
			sc.buffer.WriteRune(r)
			continue
		}

		out.WriteString(colorizeMarkdownLine(sc.buffer.String(), &sc.inCodeBlock))
		out.WriteRune('\n')
		sc.buffer.Reset()
	}

	return out.String()
}

// Finalize returns any remaining buffered content
func (sc *StreamingColorizer) Finalize() string {
	if color.NoColor || sc.buffer.Len() == 0 {
		return ""
	}

	remaining := colorizeMarkdownLine(sc.buffer.String(), &sc.inCodeBlock)
	sc.buffer.Reset()
	return remaining
}

// ReadingInputMessage returns a colored "Reading input..." message
func ReadingInputMessage() string {
	return lightBlue("Reading input...")
}

// GeneratingMessage returns a colored "Generating..." status line.
// length is always shown; contextFile is appended only when non-empty.
func GeneratingMessage(length string, contextFile string) string {
	msg := "Generating... | output length: " + length
	if contextFile != "" {
		msg += " | with project context from " + contextFile
	}
	return lightBlue(msg)
}

// CopiedMessage returns a colored "✓ Copied to clipboard" message
func CopiedMessage() string {
	return green("✓ Copied to clipboard")
}

func colorizeMarkdownLine(line string, inCodeBlock *bool) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "```") {
		*inCodeBlock = !*inCodeBlock
		return codeStyle(line)
	}

	if *inCodeBlock {
		return codeStyle(line)
	}

	switch {
	case headerRE.MatchString(line):
		return headerStyle(line)
	case unorderedRE.MatchString(line):
		return colorizeLineWithPrefix(line, unorderedCap, listStyle)
	case orderedRE.MatchString(line):
		return colorizeLineWithPrefix(line, orderedCap, listStyle)
	case quoteRE.MatchString(line):
		return colorizeLineWithPrefix(line, quoteCap, quoteStyle)
	case tableRE.MatchString(line):
		return tableStyle(line)
	case ruleRE.MatchString(line):
		return ruleStyle(line)
	default:
		return applyInlineStyles(line)
	}
}

func colorizeLineWithPrefix(line string, re *regexp.Regexp, style func(...interface{}) string) string {
	matches := re.FindStringSubmatch(line)
	if len(matches) != 3 {
		return applyInlineStyles(line)
	}

	return style(matches[1]) + applyInlineStyles(matches[2])
}

func applyInlineStyles(line string) string {
	line = inlineCodeRE.ReplaceAllStringFunc(line, func(match string) string {
		return codeStyle(match)
	})
	line = boldRE.ReplaceAllStringFunc(line, func(match string) string {
		return boldStyle(match)
	})
	line = italicRE.ReplaceAllStringFunc(line, func(match string) string {
		return italicStyle(match)
	})
	return linkRE.ReplaceAllStringFunc(line, func(match string) string {
		return linkStyle(match)
	})
}
