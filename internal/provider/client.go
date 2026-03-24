package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenRouterClient struct {
	APIKey  string
	BaseURL string
	http    *http.Client
}

func (c *OpenRouterClient) httpClient() *http.Client {
	if c.http == nil {
		c.http = &http.Client{}
	}
	return c.http
}

func (c *OpenRouterClient) get(path string, out any) error {
	req, err := http.NewRequest("GET", c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openrouter API error %d: %s", resp.StatusCode, body)
	}

	return json.Unmarshal(body, out)
}
