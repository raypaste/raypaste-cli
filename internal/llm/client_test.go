/*
Copyright © 2026 Raypaste
*/
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"raypaste-cli/pkg/types"
)

// TestGetBodySetForRetry verifies that GetBody is set to allow request body recreation during retries.
// Note: Go 1.24+ automatically sets GetBody for bytes.Reader, but we set it explicitly for:
// - Backward compatibility with older Go versions
// - Code clarity and documentation
// - Future-proofing
func TestGetBodySetForRetry(t *testing.T) {
	requestCount := int32(0)

	// Create a test server that fails on first request, succeeds on second
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)

		// Read the body to verify it's not empty
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}

		// Verify body is not empty on both requests
		if len(body) == 0 {
			t.Errorf("Request %d has empty body", count)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify body is valid JSON
		var req types.CompletionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Request %d has invalid JSON body: %v", count, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// First request fails with 500, second succeeds
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Second request succeeds
		w.Header().Set("Content-Type", "application/json")
		resp := types.CompletionResponse{
			Choices: []types.Choice{
				{
					Message: types.Message{
						Role:    "assistant",
						Content: "test response",
					},
				},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("test-key")

	// Create a request
	req := types.CompletionRequest{
		Model: "test-model",
		Messages: []types.Message{
			{Role: "user", Content: "test"},
		},
		MaxTokens:   100,
		Temperature: 0.7,
		Stream:      false,
	}

	// Marshal request to get body
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", server.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set GetBody (this is what the fix does)
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	client.setHeaders(httpReq)

	// Send request with retry
	resp, err := client.doWithRetry(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify we got a successful response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify we made exactly 2 requests (first failed, second succeeded)
	if requestCount != 2 {
		t.Errorf("Expected 2 requests, got %d", requestCount)
	}
}

// TestRetryWithBodyRecreation verifies that the retry logic correctly recreates the request body
func TestRetryWithBodyRecreation(t *testing.T) {
	requestCount := int32(0)
	var firstBody, secondBody []byte

	// Create a test server that captures request bodies
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)

		// Read and store the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}

		if count == 1 {
			firstBody = body
			// First request fails with 500
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		secondBody = body
		// Second request succeeds
		w.Header().Set("Content-Type", "application/json")
		resp := types.CompletionResponse{
			Choices: []types.Choice{
				{
					Message: types.Message{
						Role:    "assistant",
						Content: "test response",
					},
				},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("test-key")

	// Create a request
	req := types.CompletionRequest{
		Model: "test-model",
		Messages: []types.Message{
			{Role: "user", Content: "test message"},
		},
		MaxTokens:   100,
		Temperature: 0.7,
		Stream:      false,
	}

	// Marshal request to get body
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request with GetBody set (as our fix does)
	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", server.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set GetBody explicitly (this is what our fix does)
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	client.setHeaders(httpReq)

	// Send request with retry
	resp, err := client.doWithRetry(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify we got a successful response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify we made exactly 2 requests
	if requestCount != 2 {
		t.Errorf("Expected 2 requests, got %d", requestCount)
	}

	// Verify both requests had the same non-empty body
	if len(firstBody) == 0 {
		t.Error("First request had empty body")
	}
	if len(secondBody) == 0 {
		t.Error("Second request had empty body")
	}
	if !bytes.Equal(firstBody, secondBody) {
		t.Error("Request bodies differ between first and second request")
	}
}
