/*
Copyright © 2026 Raypaste
*/
package projectcontext

import (
	"os"
	"path/filepath"
	"sort"
)

// Result holds the loaded project context and the source filename.
// Filename is empty when no context file was found.
type Result struct {
	Content  string
	Filename string // base filename, e.g. "CLAUDE.md"
}

// candidateFiles lists specific project context files to try and resolve in priority order.
var candidateFiles = []string{"CLAUDE.md", "AGENTS.md"}

// Load searches upward from startDir for a project context file using the
// following priority order:
//  1. CLAUDE.md
//  2. AGENTS.md
//  3. First (alphabetically) file inside a .cursor/rules/ directory
//
// Returns a Result with the file contents and base filename. If no file is
// found, Result is zero-valued (empty Content and Filename).
func Load(startDir string) Result {
	// Try named candidates in priority order, each searched upward.
	for _, name := range candidateFiles {
		path, err := findUpward(startDir, name)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return Result{Content: string(data), Filename: name}
	}

	// Fall back to the first file inside a .cursor/rules/ directory.
	if path, err := findCursorRule(startDir); err == nil {
		if data, err := os.ReadFile(path); err == nil {
			return Result{Content: string(data), Filename: filepath.Base(path)}
		}
	}

	return Result{}
}

// findUpward walks up the directory tree from startDir looking for a file
// with the given name. Returns the full path of the first match found, or
// an error if no match exists before reaching the filesystem root.
func findUpward(startDir, filename string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding the file.
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// findCursorRule searches upward from startDir for a .cursor/rules/ directory
// and returns the path to the first file found inside it (sorted alphabetically).
func findCursorRule(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		rulesDir := filepath.Join(dir, ".cursor", "rules")
		if info, err := os.Stat(rulesDir); err == nil && info.IsDir() {
			entries, err := os.ReadDir(rulesDir)
			if err == nil {
				// Collect regular files only, sorted alphabetically.
				var files []string
				for _, e := range entries {
					if !e.IsDir() {
						files = append(files, filepath.Join(rulesDir, e.Name()))
					}
				}
				sort.Strings(files)
				if len(files) > 0 {
					return files[0], nil
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
