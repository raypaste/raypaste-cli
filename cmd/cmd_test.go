/*
Copyright © 2026 Raypaste
*/
package cmd

import (
	"testing"
)

func TestGetInputFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"single arg", []string{"hello"}, "hello"},
		{"multiple args joined with space", []string{"hello", "world"}, "hello world"},
		{"single arg with internal space", []string{"hello world"}, "hello world"},
		{"three args", []string{"a", "b", "c"}, "a b c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInput(tt.args)
			if err != nil {
				t.Fatalf("getInput() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("getInput() = %q, want %q", got, tt.want)
			}
		})
	}
}
