package client

import "fmt"

// APIError represents a structured error from the Vast.ai API.
type APIError struct {
	StatusCode int
	Message    string
	Method     string
	Path       string
}

// Error returns a human-readable error string.
func (e *APIError) Error() string {
	return fmt.Sprintf("Vast.ai API error: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Message)
}
