/*
Copyright © 2026 Raypaste
*/
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/raypaste/raypaste-cli/pkg/types"
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1/chat/completions"
	defaultTimeout    = 30 * time.Second
)

// Client represents an OpenRouter API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new OpenRouter API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Complete sends a completion request to OpenRouter and returns the full response
func (c *Client) Complete(ctx context.Context, req types.CompletionRequest) (string, error) {
	// Ensure stream is false for non-streaming
	req.Stream = false

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", openRouterBaseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set GetBody for retry support
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	// Set headers
	c.setHeaders(httpReq)

	// Send request with retry
	resp, err := c.doWithRetry(httpReq)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", c.handleErrorResponse(resp)
	}

	// Parse response
	var completionResp types.CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract content
	if len(completionResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return completionResp.Choices[0].Message.Content, nil
}

// StreamComplete sends a streaming completion request and calls the callback for each token.
//
// The provided context controls the request lifetime — cancelling it aborts the
// connection immediately, which triggers OpenRouter's stream cancellation and stops
// billing for supported providers (including Cerebras).
//
// See: https://openrouter.ai/docs/api/reference/streaming#stream-cancellation
func (c *Client) StreamComplete(ctx context.Context, req types.CompletionRequest, callback func(string) error) error {
	// Ensure stream is true
	req.Stream = true

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", openRouterBaseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set GetBody for potential retry support
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	// Set headers
	c.setHeaders(httpReq)

	// Use a client without a blanket timeout for streaming.
	// The context controls cancellation; a client-level Timeout would kill
	// long streams that legitimately take longer than the default timeout.
	streamClient := &http.Client{}

	// Send request (no retry for streaming)
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	// Close the response body promptly on context cancellation.
	// This aborts the TCP connection, which is how OpenRouter detects stream
	// cancellation and stops model processing / billing.
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = resp.Body.Close()
		case <-done:
		}
	}()
	defer func() {
		close(done)
		_ = resp.Body.Close()
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	// Process streaming response
	return processStreamingResponse(resp.Body, callback)
}

// setHeaders sets the required headers for OpenRouter API
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/raypaste/raypaste-cli")
	req.Header.Set("X-Title", "raypaste-cli")
}

// doWithRetry sends the request with simple retry logic
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		// Retry once on network error or 5xx
		time.Sleep(1 * time.Second)

		// Need to recreate the request body if it was consumed
		if req.Body != nil {
			if req.GetBody != nil {
				newBody, err := req.GetBody()
				if err == nil {
					req.Body = newBody
				}
			}
		}

		resp, err = c.httpClient.Do(req)
	}
	return resp, err
}

// handleErrorResponse parses and returns an error from the API response
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("API error (status %d): failed to read error body: %w", resp.StatusCode, err)
	}

	var errResp types.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if errResp.Error.Message != "" {
		return fmt.Errorf("API error: %s", errResp.Error.Message)
	}

	return fmt.Errorf("API error (status %d)", resp.StatusCode)
}
