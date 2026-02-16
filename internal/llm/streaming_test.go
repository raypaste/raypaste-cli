/*
Copyright © 2026 Raypaste
*/
package llm

import (
	"strings"
	"testing"
)

func TestProcessStreamingResponse_StringDeltaContent(t *testing.T) {
	stream := strings.NewReader(strings.Join([]string{
		`data: {"choices":[{"delta":{"content":"Hello"}}]}`,
		`data: {"choices":[{"delta":{"content":" world"}}]}`,
		`data: [DONE]`,
	}, "\n"))

	var got strings.Builder
	err := processStreamingResponse(stream, func(token string) error {
		got.WriteString(token)
		return nil
	})
	if err != nil {
		t.Fatalf("processStreamingResponse() error = %v", err)
	}

	if got.String() != "Hello world" {
		t.Fatalf("processStreamingResponse() got %q, want %q", got.String(), "Hello world")
	}
}

func TestProcessStreamingResponse_ArrayDeltaContent(t *testing.T) {
	stream := strings.NewReader(strings.Join([]string{
		`data: {"choices":[{"delta":{"content":[{"type":"output_text","text":"Hello"},{"type":"output_text","text":" world"}]}}]}`,
		`data: [DONE]`,
	}, "\n"))

	var got strings.Builder
	err := processStreamingResponse(stream, func(token string) error {
		got.WriteString(token)
		return nil
	})
	if err != nil {
		t.Fatalf("processStreamingResponse() error = %v", err)
	}

	if got.String() != "Hello world" {
		t.Fatalf("processStreamingResponse() got %q, want %q", got.String(), "Hello world")
	}
}

func TestProcessStreamingResponse_MessageFallbackContent(t *testing.T) {
	stream := strings.NewReader(strings.Join([]string{
		`data: {"choices":[{"delta":{},"message":{"content":"fallback text"}}]}`,
		`data: [DONE]`,
	}, "\n"))

	var got strings.Builder
	err := processStreamingResponse(stream, func(token string) error {
		got.WriteString(token)
		return nil
	})
	if err != nil {
		t.Fatalf("processStreamingResponse() error = %v", err)
	}

	if got.String() != "fallback text" {
		t.Fatalf("processStreamingResponse() got %q, want %q", got.String(), "fallback text")
	}
}

func TestProcessStreamingResponse_StreamError(t *testing.T) {
	stream := strings.NewReader(strings.Join([]string{
		`data: {"error":{"message":"provider failed"}}`,
		`data: [DONE]`,
	}, "\n"))

	err := processStreamingResponse(stream, func(token string) error {
		return nil
	})
	if err == nil {
		t.Fatal("processStreamingResponse() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "provider failed") {
		t.Fatalf("processStreamingResponse() error = %q, want contains %q", err.Error(), "provider failed")
	}
}
