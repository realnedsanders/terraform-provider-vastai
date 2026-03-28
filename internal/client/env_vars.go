package client

import (
	"context"
	"fmt"
)

// EnvVarService handles environment variable-related API operations.
type EnvVarService struct {
	client *VastAIClient
}

// EnvVarMap represents the response from the secrets endpoint.
type EnvVarMap struct {
	Secrets map[string]string `json:"secrets"`
}

// Create creates a new environment variable.
// Sends POST /secrets/ with {"key": key, "value": value}.
func (s *EnvVarService) Create(ctx context.Context, key, value string) error {
	body := map[string]string{
		"key":   key,
		"value": value,
	}
	if err := s.client.Post(ctx, "/secrets/", body, nil); err != nil {
		return fmt.Errorf("creating environment variable %q: %w", key, err)
	}
	return nil
}

// List retrieves all environment variables for the authenticated user.
// Sends GET /secrets/. Returns a map of key->value pairs.
func (s *EnvVarService) List(ctx context.Context) (map[string]string, error) {
	var resp EnvVarMap
	if err := s.client.Get(ctx, "/secrets/", &resp); err != nil {
		return nil, fmt.Errorf("listing environment variables: %w", err)
	}
	return resp.Secrets, nil
}

// Update updates an existing environment variable.
// Sends PUT /secrets/ with {"key": key, "value": value}.
func (s *EnvVarService) Update(ctx context.Context, key, value string) error {
	body := map[string]string{
		"key":   key,
		"value": value,
	}
	if err := s.client.Put(ctx, "/secrets/", body, nil); err != nil {
		return fmt.Errorf("updating environment variable %q: %w", key, err)
	}
	return nil
}

// Delete deletes an environment variable by key name.
// Sends DELETE /secrets/ with {"key": key} in the request body.
// Uses DeleteWithBody because the API requires the key in the body, not the URL path.
func (s *EnvVarService) Delete(ctx context.Context, key string) error {
	body := map[string]string{
		"key": key,
	}
	if err := s.client.DeleteWithBody(ctx, "/secrets/", body, nil); err != nil {
		return fmt.Errorf("deleting environment variable %q: %w", key, err)
	}
	return nil
}
