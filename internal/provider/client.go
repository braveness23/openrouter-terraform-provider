package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// OpenRouterClient holds credentials and HTTP configuration for the OpenRouter API.
// APIKey is for inference endpoints (/models, /chat/completions, etc.).
// ManagementAPIKey is required for management endpoints (/keys, /guardrails, /credits).
type OpenRouterClient struct {
	APIKey           string
	ManagementAPIKey string
	BaseURL          string
	http             *http.Client
}

// APIError represents a non-2xx response from the OpenRouter API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("openrouter API error %d: %s", e.StatusCode, e.Body)
}

// IsNotFound returns true if the error is a 404 from the OpenRouter API.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

// envelope is the standard OpenRouter response wrapper: {"data": ...}
type envelope[T any] struct {
	Data T `json:"data"`
}

func (c *OpenRouterClient) httpClient() *http.Client {
	if c.http == nil {
		c.http = &http.Client{}
	}
	return c.http
}

// do is the core HTTP method. All other methods are wrappers around this.
func (c *OpenRouterClient) do(ctx context.Context, method, path, key string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}
	return nil
}

// Management API wrappers — require ManagementAPIKey.

func (c *OpenRouterClient) mgmtGet(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, c.ManagementAPIKey, nil, out)
}

func (c *OpenRouterClient) mgmtPost(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, c.ManagementAPIKey, body, out)
}

func (c *OpenRouterClient) mgmtPatch(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPatch, path, c.ManagementAPIKey, body, out)
}

func (c *OpenRouterClient) mgmtDelete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, c.ManagementAPIKey, nil, nil)
}

// Inference API wrappers — use standard APIKey.

func (c *OpenRouterClient) inferenceGet(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, c.APIKey, nil, out)
}

// get is a backward-compatible wrapper for the original scaffold.
func (c *OpenRouterClient) get(path string, out any) error {
	return c.do(context.Background(), http.MethodGet, path, c.APIKey, nil, out)
}
