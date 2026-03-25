package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
)

// newRequest creates a new retryablehttp.Request with proper authentication headers.
// API key is sent ONLY via Authorization: Bearer header, NEVER in URL query parameters.
// Sets User-Agent, Content-Type, and Accept headers on every request.
func (c *VastAIClient) newRequest(ctx context.Context, method, path string, body interface{}) (*retryablehttp.Request, error) {
	url := c.baseURL + "/api/v0" + path

	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	var req *retryablehttp.Request
	var err error

	if bodyReader != nil {
		req, err = retryablehttp.NewRequestWithContext(ctx, method, url, bodyReader)
	} else {
		req, err = retryablehttp.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// CRITICAL: API key goes ONLY in Authorization header, NEVER in URL (per D-09, FOUND-05)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}
