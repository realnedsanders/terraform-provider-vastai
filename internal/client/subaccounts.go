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
	body, err := openAPIJSONBody(map[string]interface{}{
		"email":     email,
		"username":  username,
		"password":  password,
		"host_only": hostOnly,
		"parent_id": "me",
	})
	if err != nil {
		return nil, fmt.Errorf("creating subaccount: %w", err)
	}
	var result Subaccount
	resp, err := s.client.openAPIClient.CreateSubaccountWithBody(ctx, applicationJSON, body)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("creating subaccount: %w", err)
	}
	return &result, nil
}

// List retrieves all subaccounts owned by the authenticated user.
// Sends GET /subaccounts?owner=me. Unwraps .Users from the response.
func (s *SubaccountService) List(ctx context.Context) ([]Subaccount, error) {
	var result SubaccountListResponse
	resp, err := s.client.openAPIClient.ShowSubaccounts(ctx, withOpenAPIQuery(map[string]string{"owner": "me"}))
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("listing subaccounts: %w", err)
	}
	return result.Users, nil
}
