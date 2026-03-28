package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestTeamService_CreateTeam
// ---------------------------------------------------------------------------

func TestTeamService_CreateTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/team/" {
			t.Errorf("expected path /api/v0/team/, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["team_name"] != "my-team" {
			t.Errorf("expected team_name 'my-team', got %q", body["team_name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Team{ID: 42, TeamName: "my-team"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	team, err := c.Teams.CreateTeam(context.Background(), "my-team")
	if err != nil {
		t.Fatalf("CreateTeam returned error: %v", err)
	}
	if team.ID != 42 {
		t.Errorf("expected ID 42, got %d", team.ID)
	}
	if team.TeamName != "my-team" {
		t.Errorf("expected team_name 'my-team', got %q", team.TeamName)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_CreateTeam_Error
// ---------------------------------------------------------------------------

func TestTeamService_CreateTeam_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "team already exists"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Teams.CreateTeam(context.Background(), "existing-team")
	if err == nil {
		t.Fatal("expected error on 409 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_DestroyTeam
// ---------------------------------------------------------------------------

func TestTeamService_DestroyTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/team/" {
			t.Errorf("expected path /api/v0/team/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Teams.DestroyTeam(context.Background())
	if err != nil {
		t.Fatalf("DestroyTeam returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_CreateRole
// ---------------------------------------------------------------------------

func TestTeamService_CreateRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/team/roles/" {
			t.Errorf("expected path /api/v0/team/roles/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["name"] != "admin" {
			t.Errorf("expected name 'admin', got %v", body["name"])
		}
		if body["permissions"] == nil {
			t.Error("expected permissions field in body")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TeamRole{
			ID:          10,
			Name:        "admin",
			Permissions: json.RawMessage(`{"api":{"instance_read":{}}}`),
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	perms := json.RawMessage(`{"api":{"instance_read":{}}}`)
	role, err := c.Teams.CreateRole(context.Background(), "admin", perms)
	if err != nil {
		t.Fatalf("CreateRole returned error: %v", err)
	}
	if role.ID != 10 {
		t.Errorf("expected ID 10, got %d", role.ID)
	}
	if role.Name != "admin" {
		t.Errorf("expected name 'admin', got %q", role.Name)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_ListRoles
// ---------------------------------------------------------------------------

func TestTeamService_ListRoles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/team/roles-full/" {
			t.Errorf("expected path /api/v0/team/roles-full/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]TeamRole{
			{ID: 1, Name: "admin"},
			{ID: 2, Name: "viewer"},
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	roles, err := c.Teams.ListRoles(context.Background())
	if err != nil {
		t.Fatalf("ListRoles returned error: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if roles[0].Name != "admin" {
		t.Errorf("expected first role 'admin', got %q", roles[0].Name)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_GetRole
// ---------------------------------------------------------------------------

func TestTeamService_GetRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/team/roles/admin/" {
			t.Errorf("expected path /api/v0/team/roles/admin/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TeamRole{
			ID:          10,
			Name:        "admin",
			Permissions: json.RawMessage(`{"api":{"instance_read":{}}}`),
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	role, err := c.Teams.GetRole(context.Background(), "admin")
	if err != nil {
		t.Fatalf("GetRole returned error: %v", err)
	}
	if role.ID != 10 {
		t.Errorf("expected ID 10, got %d", role.ID)
	}
	if role.Name != "admin" {
		t.Errorf("expected name 'admin', got %q", role.Name)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_GetRole_Error
// ---------------------------------------------------------------------------

func TestTeamService_GetRole_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "role not found"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Teams.GetRole(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error on 404 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_UpdateRole
// ---------------------------------------------------------------------------

func TestTeamService_UpdateRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		// Update uses ID, not name (Pitfall 3)
		if r.URL.Path != "/api/v0/team/roles/10/" {
			t.Errorf("expected path /api/v0/team/roles/10/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["name"] != "super-admin" {
			t.Errorf("expected name 'super-admin', got %v", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TeamRole{
			ID:          10,
			Name:        "super-admin",
			Permissions: json.RawMessage(`{"api":{"instance_read":{},"instance_write":{}}}`),
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	perms := json.RawMessage(`{"api":{"instance_read":{},"instance_write":{}}}`)
	role, err := c.Teams.UpdateRole(context.Background(), 10, "super-admin", perms)
	if err != nil {
		t.Fatalf("UpdateRole returned error: %v", err)
	}
	if role.Name != "super-admin" {
		t.Errorf("expected name 'super-admin', got %q", role.Name)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_DeleteRole
// ---------------------------------------------------------------------------

func TestTeamService_DeleteRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		// Delete uses name, not ID (Pitfall 3)
		if r.URL.Path != "/api/v0/team/roles/admin/" {
			t.Errorf("expected path /api/v0/team/roles/admin/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Teams.DeleteRole(context.Background(), "admin")
	if err != nil {
		t.Fatalf("DeleteRole returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_InviteMember
// ---------------------------------------------------------------------------

func TestTeamService_InviteMember(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Verify query params are in the URL (Pitfall 5)
		if r.URL.Query().Get("email") != "user@example.com" {
			t.Errorf("expected email query param 'user@example.com', got %q", r.URL.Query().Get("email"))
		}
		if r.URL.Query().Get("role") != "admin" {
			t.Errorf("expected role query param 'admin', got %q", r.URL.Query().Get("role"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Teams.InviteMember(context.Background(), "user@example.com", "admin")
	if err != nil {
		t.Fatalf("InviteMember returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_InviteMember_Error
// ---------------------------------------------------------------------------

func TestTeamService_InviteMember_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid email"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Teams.InviteMember(context.Background(), "bad-email", "admin")
	if err == nil {
		t.Fatal("expected error on 400 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_ListMembers
// ---------------------------------------------------------------------------

func TestTeamService_ListMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/team/members/" {
			t.Errorf("expected path /api/v0/team/members/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]TeamMember{
			{ID: 1, Email: "alice@example.com", Role: "admin"},
			{ID: 2, Email: "bob@example.com", Role: "viewer"},
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	members, err := c.Teams.ListMembers(context.Background())
	if err != nil {
		t.Fatalf("ListMembers returned error: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0].Email != "alice@example.com" {
		t.Errorf("expected first member email 'alice@example.com', got %q", members[0].Email)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_RemoveMember
// ---------------------------------------------------------------------------

func TestTeamService_RemoveMember(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/team/members/42/" {
			t.Errorf("expected path /api/v0/team/members/42/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Teams.RemoveMember(context.Background(), 42)
	if err != nil {
		t.Fatalf("RemoveMember returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestTeamService_RemoveMember_Error
// ---------------------------------------------------------------------------

func TestTeamService_RemoveMember_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "member not found"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Teams.RemoveMember(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error on 404 response, got nil")
	}
}
