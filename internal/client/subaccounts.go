package client

import (
	"context"
	"fmt"
)

// SubaccountService handles subaccount-related API operations.
type SubaccountService struct {
	client *VastAIClient
}

// Subaccount represents a subaccount from the Vast.ai API.
type Subaccount struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// SubaccountCreateRequest represents the request body for creating a subaccount.
type SubaccountCreateRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	HostOnly bool   `json:"host_only"`
	ParentID string `json:"parent_id"`
}

// SubaccountListResponse represents the response from the subaccounts list endpoint.
type SubaccountListResponse struct {
	Users []Subaccount `json:"users"`
}

// Create creates a new subaccount.
// Sends POST /users/ with email, username, password, host_only, and parent_id="me".
func (s *SubaccountService) Create(ctx context.Context, email, username, password string, hostOnly bool) (*Subaccount, error) {
	body := SubaccountCreateRequest{
		Email:    email,
		Username: username,
		Password: password,
		HostOnly: hostOnly,
		ParentID: "me",
	}
	var resp Subaccount
	if err := s.client.Post(ctx, "/users/", body, &resp); err != nil {
		return nil, fmt.Errorf("creating subaccount: %w", err)
	}
	return &resp, nil
}

// List retrieves all subaccounts owned by the authenticated user.
// Sends GET /subaccounts?owner=me. Unwraps .Users from the response.
func (s *SubaccountService) List(ctx context.Context) ([]Subaccount, error) {
	var resp SubaccountListResponse
	if err := s.client.Get(ctx, "/subaccounts?owner=me", &resp); err != nil {
		return nil, fmt.Errorf("listing subaccounts: %w", err)
	}
	return resp.Users, nil
}
