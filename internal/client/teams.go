package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// TeamService handles team, role, and member-related API operations.
type TeamService struct {
	client *VastAIClient
}

// Team represents a team object from the Vast.ai API.
type Team struct {
	ID       int    `json:"id"`
	TeamName string `json:"team_name"`
}

// TeamRole represents a team role from the Vast.ai API.
type TeamRole struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Permissions json.RawMessage `json:"permissions,omitempty"`
}

// TeamMember represents a team member from the Vast.ai API.
type TeamMember struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Username string `json:"username,omitempty"`
}

// ---------------------------------------------------------------------------
// Team CRUD
// ---------------------------------------------------------------------------

// CreateTeam creates a new team.
// Sends POST /team/ with {"team_name": teamName}.
func (s *TeamService) CreateTeam(ctx context.Context, teamName string) (*Team, error) {
	body := map[string]string{"team_name": teamName}
	var resp Team
	if err := s.client.Post(ctx, "/team/", body, &resp); err != nil {
		return nil, fmt.Errorf("creating team: %w", err)
	}
	return &resp, nil
}

// DestroyTeam destroys the team associated with the current API key context.
// Sends DELETE /team/ (no parameters -- parameterless delete per research).
func (s *TeamService) DestroyTeam(ctx context.Context) error {
	if err := s.client.Delete(ctx, "/team/", nil); err != nil {
		return fmt.Errorf("destroying team: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Role CRUD
// ---------------------------------------------------------------------------

// CreateRole creates a new team role.
// Sends POST /team/roles/ with {"name": name, "permissions": perms}.
func (s *TeamService) CreateRole(ctx context.Context, name string, permissions json.RawMessage) (*TeamRole, error) {
	body := map[string]interface{}{
		"name":        name,
		"permissions": permissions,
	}
	var resp TeamRole
	if err := s.client.Post(ctx, "/team/roles/", body, &resp); err != nil {
		return nil, fmt.Errorf("creating team role: %w", err)
	}
	return &resp, nil
}

// ListRoles retrieves all team roles.
// Sends GET /team/roles-full/.
func (s *TeamService) ListRoles(ctx context.Context) ([]TeamRole, error) {
	var resp []TeamRole
	if err := s.client.Get(ctx, "/team/roles-full/", &resp); err != nil {
		return nil, fmt.Errorf("listing team roles: %w", err)
	}
	return resp, nil
}

// GetRole retrieves a single team role by name.
// Sends GET /team/roles/{name}/.
func (s *TeamService) GetRole(ctx context.Context, name string) (*TeamRole, error) {
	path := fmt.Sprintf("/team/roles/%s/", url.PathEscape(name))
	var resp TeamRole
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("getting team role %q: %w", name, err)
	}
	return &resp, nil
}

// UpdateRole updates a team role by ID.
// Sends PUT /team/roles/{id}/ with {"name": name, "permissions": perms}.
// Note: Update uses ID (not name) in the path, per Pitfall 3.
func (s *TeamService) UpdateRole(ctx context.Context, id int, name string, permissions json.RawMessage) (*TeamRole, error) {
	path := fmt.Sprintf("/team/roles/%d/", id)
	body := map[string]interface{}{
		"name":        name,
		"permissions": permissions,
	}
	var resp TeamRole
	if err := s.client.Put(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("updating team role %d: %w", id, err)
	}
	return &resp, nil
}

// DeleteRole deletes a team role by name.
// Sends DELETE /team/roles/{name}/.
// Note: Delete uses name (not ID) in the path, per Pitfall 3.
func (s *TeamService) DeleteRole(ctx context.Context, name string) error {
	path := fmt.Sprintf("/team/roles/%s/", url.PathEscape(name))
	if err := s.client.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting team role %q: %w", name, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Member operations
// ---------------------------------------------------------------------------

// InviteMember invites a user to the team with a given role.
// Sends POST /team/invite/?email={email}&role={role}.
// Note: Uses query parameters, NOT JSON body (Pitfall 5).
func (s *TeamService) InviteMember(ctx context.Context, email, role string) error {
	path := fmt.Sprintf("/team/invite/?email=%s&role=%s", url.QueryEscape(email), url.QueryEscape(role))
	if err := s.client.Post(ctx, path, nil, nil); err != nil {
		return fmt.Errorf("inviting team member %q: %w", email, err)
	}
	return nil
}

// ListMembers retrieves all team members.
// Sends GET /team/members/.
func (s *TeamService) ListMembers(ctx context.Context) ([]TeamMember, error) {
	var resp []TeamMember
	if err := s.client.Get(ctx, "/team/members/", &resp); err != nil {
		return nil, fmt.Errorf("listing team members: %w", err)
	}
	return resp, nil
}

// RemoveMember removes a team member by ID.
// Sends DELETE /team/members/{id}/.
func (s *TeamService) RemoveMember(ctx context.Context, id int) error {
	path := fmt.Sprintf("/team/members/%d/", id)
	if err := s.client.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("removing team member %d: %w", id, err)
	}
	return nil
}
