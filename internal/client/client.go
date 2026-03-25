package client

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// VastAIClient is the Vast.ai REST API client.
type VastAIClient struct {
	httpClient *retryablehttp.Client
	baseURL    string
	apiKey     string
	userAgent  string
}

// NewVastAIClient creates a new Vast.ai API client.
// STUB: returns empty client for TDD RED phase.
func NewVastAIClient(apiKey, baseURL, version string) *VastAIClient {
	return &VastAIClient{}
}

// vastaiRetryPolicy determines whether a request should be retried.
// STUB: returns false for TDD RED phase.
func vastaiRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	return false, nil
}

// vastaiBackoff calculates the backoff duration for retries.
// STUB: returns 0 for TDD RED phase.
func vastaiBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	return 0
}

// do executes an HTTP request and decodes the response.
// STUB: returns nil for TDD RED phase.
func (c *VastAIClient) do(ctx context.Context, req *retryablehttp.Request, result interface{}) error {
	return nil
}

// Get sends a GET request to the given path.
func (c *VastAIClient) Get(ctx context.Context, path string, result interface{}) error {
	return nil
}

// Post sends a POST request with a JSON body to the given path.
func (c *VastAIClient) Post(ctx context.Context, path string, body, result interface{}) error {
	return nil
}

// Put sends a PUT request with a JSON body to the given path.
func (c *VastAIClient) Put(ctx context.Context, path string, body, result interface{}) error {
	return nil
}

// Delete sends a DELETE request to the given path.
func (c *VastAIClient) Delete(ctx context.Context, path string, result interface{}) error {
	return nil
}
