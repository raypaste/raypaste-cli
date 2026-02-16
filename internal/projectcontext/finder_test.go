/*
Copyright © 2026 Raypaste
*/
package projectcontext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoFiles(t *testing.T) {
	dir := t.TempDir()
	result := Load(dir)
	if result.Content != "" || result.Filename != "" {
		t.Errorf("expected empty Result, got Content=%q Filename=%q", result.Content, result.Filename)
	}
}

func TestLoad_ClaudeMd(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "claude context")

	result := Load(dir)
	if result.Content != "claude context" {
		t.Errorf("Content = %q, want %q", result.Content, "claude context")
	}
	if result.Filename != "CLAUDE.md" {
		t.Errorf("Filename = %q, want %q", result.Filename, "CLAUDE.md")
	}
}

func TestLoad_AgentsMd(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "agents context")

	result := Load(dir)
	if result.Content != "agents context" {
		t.Errorf("Content = %q, want %q", result.Content, "agents context")
	}
	if result.Filename != "AGENTS.md" {
		t.Errorf("Filename = %q, want %q", result.Filename, "AGENTS.md")
	}
}

func TestLoad_ClaudeMdPriorityOverAgentsMd(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "claude wins")
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "agents loses")

	result := Load(dir)
	if result.Content != "claude wins" {
		t.Errorf("Content = %q, want %q", result.Content, "claude wins")
	}
	if result.Filename != "CLAUDE.md" {
		t.Errorf("Filename = %q, want %q", result.Filename, "CLAUDE.md")
	}
}

func TestLoad_CursorRulesFallback(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(rulesDir, "my-rule.md"), "cursor rule content")

	result := Load(dir)
	if result.Content != "cursor rule content" {
		t.Errorf("Content = %q, want %q", result.Content, "cursor rule content")
	}
	if result.Filename != "my-rule.md" {
		t.Errorf("Filename = %q, want %q", result.Filename, "my-rule.md")
	}
}

func TestLoad_CursorRulesFirstAlphabetically(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(rulesDir, "zzz-last.md"), "last")
	writeFile(t, filepath.Join(rulesDir, "aaa-first.md"), "first content")
	writeFile(t, filepath.Join(rulesDir, "mmm-middle.md"), "middle")

	result := Load(dir)
	if result.Content != "first content" {
		t.Errorf("Content = %q, want %q", result.Content, "first content")
	}
	if result.Filename != "aaa-first.md" {
		t.Errorf("Filename = %q, want %q", result.Filename, "aaa-first.md")
	}
}

func TestLoad_CursorRulesNotUsedWhenClaudeMdExists(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "claude content")
	rulesDir := filepath.Join(dir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(rulesDir, "rule.md"), "cursor content")

	result := Load(dir)
	if result.Filename != "CLAUDE.md" {
		t.Errorf("Filename = %q, want CLAUDE.md", result.Filename)
	}
}

func TestLoad_SearchesUpward(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "CLAUDE.md"), "root claude")

	// Create a nested subdirectory and start Load from there
	sub := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	result := Load(sub)
	if result.Content != "root claude" {
		t.Errorf("Content = %q, want %q", result.Content, "root claude")
	}
	if result.Filename != "CLAUDE.md" {
		t.Errorf("Filename = %q, want %q", result.Filename, "CLAUDE.md")
	}
}

func TestLoad_NearestFileWins(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "CLAUDE.md"), "root claude")

	sub := filepath.Join(root, "child")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "CLAUDE.md"), "child claude")

	result := Load(sub)
	if result.Content != "child claude" {
		t.Errorf("Content = %q, want %q", result.Content, "child claude")
	}
}

func TestLoad_AgentsMdFoundUpward(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "AGENTS.md"), "agents upward")

	sub := filepath.Join(root, "deep", "dir")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	result := Load(sub)
	if result.Content != "agents upward" {
		t.Errorf("Content = %q, want %q", result.Content, "agents upward")
	}
	if result.Filename != "AGENTS.md" {
		t.Errorf("Filename = %q, want AGENTS.md, got %q", result.Filename, result.Filename)
	}
}

func TestLoad_EmptyCursorRulesDir(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	// No files in rules dir

	result := Load(dir)
	if result.Content != "" || result.Filename != "" {
		t.Errorf("expected empty Result for empty rules dir, got Content=%q Filename=%q", result.Content, result.Filename)
	}
}

// writeFile is a helper to create a file with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
