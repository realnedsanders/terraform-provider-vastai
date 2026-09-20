package client

import (
	"context"
	"encoding/json"
	"fmt"
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
	body, err := openAPIJSONBody(map[string]string{"team_name": teamName})
	if err != nil {
		return nil, fmt.Errorf("creating team: %w", err)
	}
	var result Team
	resp, err := s.client.openAPIClient.CreateTeamWithBody(ctx, applicationJSON, body)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("creating team: %w", err)
	}
	return &result, nil
}

// DestroyTeam destroys the team associated with the current API key context.
// Sends DELETE /team/ (no parameters -- parameterless delete per research).
func (s *TeamService) DestroyTeam(ctx context.Context) error {
	resp, err := s.client.openAPIClient.DestroyTeam(ctx)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, nil); err != nil {
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
	body, err := openAPIJSONBody(map[string]interface{}{
		"name":        name,
		"permissions": permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("creating team role: %w", err)
	}
	var result TeamRole
	resp, err := s.client.openAPIClient.CreateTeamRoleWithBody(ctx, applicationJSON, body)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("creating team role: %w", err)
	}
	return &result, nil
}

// ListRoles retrieves all team roles.
// Sends GET /team/roles-full/.
func (s *TeamService) ListRoles(ctx context.Context) ([]TeamRole, error) {
	var result []TeamRole
	resp, err := s.client.openAPIClient.ShowTeamRoles(ctx)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("listing team roles: %w", err)
	}
	return result, nil
}

// GetRole retrieves a single team role by name.
// Sends GET /team/roles/{name}/.
func (s *TeamService) GetRole(ctx context.Context, name string) (*TeamRole, error) {
	var result TeamRole
	resp, err := s.client.openAPIClient.ShowTeamRole(ctx, name)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("getting team role %q: %w", name, err)
	}
	return &result, nil
}

// UpdateRole updates a team role by ID.
// Sends PUT /team/roles/{id}/ with {"name": name, "permissions": perms}.
// Note: Update uses ID (not name) in the path, per Pitfall 3.
func (s *TeamService) UpdateRole(ctx context.Context, id int, name string, permissions json.RawMessage) (*TeamRole, error) {
	body, err := openAPIJSONBody(map[string]interface{}{
		"name":        name,
		"permissions": permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("updating team role %d: %w", id, err)
	}
	var result TeamRole
	resp, err := s.client.openAPIClient.UpdateTeamRoleWithBody(ctx, id, applicationJSON, body)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("updating team role %d: %w", id, err)
	}
	return &result, nil
}

// DeleteRole deletes a team role by name.
// Sends DELETE /team/roles/{name}/.
// Note: Delete uses name (not ID) in the path, per Pitfall 3.
func (s *TeamService) DeleteRole(ctx context.Context, name string) error {
	resp, err := s.client.openAPIClient.RemoveTeamRole(ctx, name)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, nil); err != nil {
		return fmt.Errorf("deleting team role %q: %w", name, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Member operations
// ---------------------------------------------------------------------------

// InviteMember invites a user to the team with a given role.
// The live API reads email and role from the query string even though the
// official OpenAPI document currently describes a JSON body.
func (s *TeamService) InviteMember(ctx context.Context, email, role string) error {
	resp, err := s.client.openAPIClient.InviteTeamMemberWithBody(ctx, applicationJSON, nil,
		withOpenAPIQuery(map[string]string{"email": email, "role": role}))
	if err := s.client.doOpenAPIResponse(ctx, resp, err, nil); err != nil {
		return fmt.Errorf("inviting team member %q: %w", email, err)
	}
	return nil
}

// ListMembers retrieves all team members.
// Sends GET /team/members/.
func (s *TeamService) ListMembers(ctx context.Context) ([]TeamMember, error) {
	var result []TeamMember
	resp, err := s.client.openAPIClient.ShowTeamMembers(ctx)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("listing team members: %w", err)
	}
	return result, nil
}

// RemoveMember removes a team member by ID.
// Sends DELETE /team/members/{id}/.
func (s *TeamService) RemoveMember(ctx context.Context, id int) error {
	resp, err := s.client.openAPIClient.RemoveTeamMember(ctx, id)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, nil); err != nil {
		return fmt.Errorf("removing team member %d: %w", id, err)
	}
	return nil
}
