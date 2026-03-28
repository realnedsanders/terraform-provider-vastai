package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ApiKeyService handles API key-related API operations.
type ApiKeyService struct {
	client *VastAIClient
}

// ApiKey represents an API key object from the Vast.ai API.
type ApiKey struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Key         string          `json:"key,omitempty"`
	Permissions json.RawMessage `json:"permissions,omitempty"`
	KeyParams   string          `json:"key_params,omitempty"`
	CreatedAt   string          `json:"created_at,omitempty"`
}

// Create creates a new API key with the given name, permissions, and optional key_params.
// Sends POST /auth/apikeys/ with {"name": name, "permissions": perms, "key_params": keyParams}.
// Returns the full ApiKey including the key value (only available on create).
func (s *ApiKeyService) Create(ctx context.Context, name string, permissions json.RawMessage, keyParams string) (*ApiKey, error) {
	body := map[string]interface{}{
		"name":        name,
		"permissions": permissions,
		"key_params":  keyParams,
	}
	var resp ApiKey
	if err := s.client.Post(ctx, "/auth/apikeys/", body, &resp); err != nil {
		return nil, fmt.Errorf("creating API key: %w", err)
	}
	return &resp, nil
}

// List retrieves all API keys for the authenticated user.
// Sends GET /auth/apikeys/. Returns array of ApiKey (without key value).
func (s *ApiKeyService) List(ctx context.Context) ([]ApiKey, error) {
	var resp []ApiKey
	if err := s.client.Get(ctx, "/auth/apikeys/", &resp); err != nil {
		return nil, fmt.Errorf("listing API keys: %w", err)
	}
	return resp, nil
}

// Delete deletes an API key by ID.
// Sends DELETE /auth/apikeys/{id}/.
func (s *ApiKeyService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/auth/apikeys/%d/", id)
	if err := s.client.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting API key %d: %w", id, err)
	}
	return nil
}
