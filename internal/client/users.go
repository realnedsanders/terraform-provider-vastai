package client

import (
	"context"
	"fmt"
)

// UserService handles user profile-related API operations.
type UserService struct {
	client *VastAIClient
}

// User represents a user profile from the Vast.ai API.
// Contains a focused subset of fields from the SDK's user_fields tuple (per D-05).
type User struct {
	ID                      int     `json:"id"`
	Username                string  `json:"username"`
	Email                   string  `json:"email"`
	EmailVerified           bool    `json:"email_verified"`
	Fullname                string  `json:"fullname"`
	Balance                 float64 `json:"balance"`
	Credit                  float64 `json:"credit"`
	HasBilling              bool    `json:"has_billing"`
	SSHKey                  string  `json:"ssh_key"`
	BalanceThreshold        float64 `json:"balance_threshold"`
	BalanceThresholdEnabled bool    `json:"balance_threshold_enabled"`
}

// GetCurrent retrieves the current authenticated user's profile.
// Sends GET /users/current?owner=me.
func (s *UserService) GetCurrent(ctx context.Context) (*User, error) {
	var resp User
	if err := s.client.Get(ctx, "/users/current?owner=me", &resp); err != nil {
		return nil, fmt.Errorf("getting current user: %w", err)
	}
	return &resp, nil
}
