/*
Copyright © 2026 Raypaste
*/
package clipboard

import (
	"fmt"
	"sync"

	"golang.design/x/clipboard"
)

var (
	initialized bool
	initMutex   sync.Mutex
	initError   error
)

// Init initializes the clipboard system
// This must be called before any clipboard operations
func Init() error {
	initMutex.Lock()
	defer initMutex.Unlock()

	if initialized {
		return initError
	}

	initError = clipboard.Init()
	initialized = true
	return initError
}

// Copy copies text to the clipboard
// Automatically initializes the clipboard if not already done
func Copy(text string) error {
	if !initialized {
		if err := Init(); err != nil {
			return fmt.Errorf("failed to initialize clipboard: %w", err)
		}
	}

	if initError != nil {
		return fmt.Errorf("clipboard not available: %w", initError)
	}

	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}

// IsAvailable checks if the clipboard is available
func IsAvailable() bool {
	if !initialized {
		if err := Init(); err != nil {
			return false
		}
	}
	return initError == nil
}

// CopyWithWarning copies text to clipboard and returns a warning message if it fails
// This is a convenience function for CLI usage
func CopyWithWarning(text string) string {
	if err := Copy(text); err != nil {
		return fmt.Sprintf("Warning: Could not copy to clipboard: %v", err)
	}
	return ""
}
