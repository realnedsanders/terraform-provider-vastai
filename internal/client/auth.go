package client

import (
	"context"

	"github.com/hashicorp/go-retryablehttp"
)

// newRequest creates a new retryablehttp.Request with proper authentication headers.
// STUB: returns nil for TDD RED phase.
func (c *VastAIClient) newRequest(ctx context.Context, method, path string, body interface{}) (*retryablehttp.Request, error) {
	return nil, nil
}
