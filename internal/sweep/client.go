// Package sweep provides shared utilities for test resource sweepers.
//
// This package is intentionally separate from acctest to avoid import cycles.
// The acctest package imports the provider package (which imports all service packages),
// so service-level sweep_test.go files cannot import acctest. This package only
// depends on the client package, breaking the cycle.
package sweep

import (
	"fmt"
	"os"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// SharedClient creates a VastAIClient for sweeper operations.
// Reads VASTAI_API_KEY and optional VASTAI_API_URL from environment.
func SharedClient() (*client.VastAIClient, error) {
	apiKey := os.Getenv("VASTAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("VASTAI_API_KEY must be set for sweepers")
	}
	apiURL := os.Getenv("VASTAI_API_URL")
	if apiURL == "" {
		apiURL = "https://console.vast.ai"
	}
	return client.NewVastAIClient(apiKey, apiURL, "test"), nil
}
