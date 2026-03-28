package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// VastAIClient is the Vast.ai REST API client.
type VastAIClient struct {
	httpClient *retryablehttp.Client
	baseURL    string
	apiKey     string
	userAgent  string

	// Service sub-objects for domain-specific API operations
	ApiKeys     *ApiKeyService
	EnvVars     *EnvVarService
	Subaccounts *SubaccountService
}

// NewVastAIClient creates a new Vast.ai API client with Bearer authentication,
// exponential backoff retry, and structured error handling.
func NewVastAIClient(apiKey, baseURL, version string) *VastAIClient {
	client := retryablehttp.NewClient()
	client.RetryMax = 5
	client.RetryWaitMin = 150 * time.Millisecond
	client.RetryWaitMax = 30 * time.Second
	client.CheckRetry = vastaiRetryPolicy
	client.Backoff = vastaiBackoff
	client.Logger = nil // silence default logger; use tflog instead (Pitfall 6)

	c := &VastAIClient{
		httpClient: client,
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		userAgent:  fmt.Sprintf("terraform-provider-vastai/%s", version),
	}

	// Initialize service sub-objects
	c.ApiKeys = &ApiKeyService{client: c}
	c.EnvVars = &EnvVarService{client: c}
	c.Subaccounts = &SubaccountService{client: c}

	return c
}

// vastaiRetryPolicy determines whether a request should be retried.
// Retries on 429 (rate limited), 5xx (except 501), and connection errors.
// Does not retry on 4xx client errors or context cancellation.
func vastaiRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// Check context first
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	// Connection errors should be retried
	if err != nil {
		return true, nil
	}

	// Rate limited
	if resp.StatusCode == 429 {
		return true, nil
	}

	// Server errors (except 501 Not Implemented)
	if resp.StatusCode >= 500 && resp.StatusCode != 501 {
		return true, nil
	}

	return false, nil
}

// vastaiBackoff calculates the backoff duration for retries using 150ms base
// with 1.5x exponential multiplier. Respects Retry-After header on 429 responses.
func vastaiBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// Respect Retry-After header on 429 responses
	if resp != nil && resp.StatusCode == 429 {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.ParseFloat(retryAfter, 64); err == nil {
				return time.Duration(seconds * float64(time.Second))
			}
		}
	}

	// Exponential backoff: min * 1.5^attemptNum
	wait := float64(min) * math.Pow(1.5, float64(attemptNum))

	// Guard against floating-point overflow (Inf or values exceeding max)
	if math.IsInf(wait, 0) || math.IsNaN(wait) || wait > float64(max) {
		return max
	}

	return time.Duration(wait)
}

// do executes an HTTP request and decodes the response.
func (c *VastAIClient) do(ctx context.Context, req *retryablehttp.Request, result interface{}) error {
	// Log request at DEBUG level
	tflog.Debug(ctx, "Vast.ai API request", map[string]interface{}{
		"method": req.Method,
		"url":    req.URL.String(),
	})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Log response at DEBUG level
	tflog.Debug(ctx, "Vast.ai API response", map[string]interface{}{
		"status": resp.StatusCode,
	})

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	// Log response body at TRACE level
	tflog.Trace(ctx, "Vast.ai API response body", map[string]interface{}{
		"body": string(body),
	})

	// Handle error responses
	if resp.StatusCode >= 400 {
		message := extractErrorMessage(body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    message,
			Method:     req.Method,
			Path:       req.URL.Path,
		}
	}

	// Decode successful response
	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("decoding response body: %w", err)
		}
	}

	return nil
}

// extractErrorMessage attempts to extract an error message from a JSON response body.
// Tries {"error": "..."} and {"msg": "..."} patterns.
func extractErrorMessage(body []byte) string {
	var errorResp map[string]interface{}
	if err := json.Unmarshal(body, &errorResp); err != nil {
		return string(body)
	}

	if msg, ok := errorResp["error"].(string); ok {
		return msg
	}
	if msg, ok := errorResp["msg"].(string); ok {
		return msg
	}

	return string(body)
}

// Get sends a GET request to the given path and decodes the response into result.
func (c *VastAIClient) Get(ctx context.Context, path string, result interface{}) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return c.do(ctx, req, result)
}

// Post sends a POST request with a JSON body to the given path and decodes the response into result.
func (c *VastAIClient) Post(ctx context.Context, path string, body, result interface{}) error {
	req, err := c.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	return c.do(ctx, req, result)
}

// Put sends a PUT request with a JSON body to the given path and decodes the response into result.
func (c *VastAIClient) Put(ctx context.Context, path string, body, result interface{}) error {
	req, err := c.newRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	return c.do(ctx, req, result)
}

// Delete sends a DELETE request to the given path and decodes the response into result.
func (c *VastAIClient) Delete(ctx context.Context, path string, result interface{}) error {
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return c.do(ctx, req, result)
}

// DeleteWithBody sends a DELETE request with a JSON body to the given path and decodes the response into result.
// This is needed for APIs like environment variable deletion where the identifier is sent in the request body.
func (c *VastAIClient) DeleteWithBody(ctx context.Context, path string, body, result interface{}) error {
	req, err := c.newRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return err
	}
	return c.do(ctx, req, result)
}
