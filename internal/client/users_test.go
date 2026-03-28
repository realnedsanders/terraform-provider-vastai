package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestUserService_GetCurrent
// ---------------------------------------------------------------------------

func TestUserService_GetCurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/users/current" {
			t.Errorf("expected path /api/v0/users/current, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("owner") != "me" {
			t.Errorf("expected owner=me query param, got %q", r.URL.Query().Get("owner"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                        42,
			"username":                  "testuser",
			"email":                     "test@example.com",
			"email_verified":            true,
			"fullname":                  "Test User",
			"balance":                   100.50,
			"credit":                    25.00,
			"has_billing":               true,
			"ssh_key":                   "ssh-ed25519 AAAA...",
			"balance_threshold":         10.0,
			"balance_threshold_enabled": true,
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	user, err := c.Users.GetCurrent(context.Background())
	if err != nil {
		t.Fatalf("GetCurrent returned error: %v", err)
	}
	if user.ID != 42 {
		t.Errorf("expected ID 42, got %d", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username testuser, got %q", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %q", user.Email)
	}
	if !user.EmailVerified {
		t.Error("expected email_verified=true")
	}
	if user.Balance != 100.50 {
		t.Errorf("expected balance 100.50, got %f", user.Balance)
	}
	if user.Credit != 25.00 {
		t.Errorf("expected credit 25.00, got %f", user.Credit)
	}
	if !user.HasBilling {
		t.Error("expected has_billing=true")
	}
	if user.SSHKey != "ssh-ed25519 AAAA..." {
		t.Errorf("expected ssh_key, got %q", user.SSHKey)
	}
	if user.BalanceThreshold != 10.0 {
		t.Errorf("expected balance_threshold 10.0, got %f", user.BalanceThreshold)
	}
	if !user.BalanceThresholdEnabled {
		t.Error("expected balance_threshold_enabled=true")
	}
}

// ---------------------------------------------------------------------------
// TestUserService_GetCurrent_Error
// ---------------------------------------------------------------------------

func TestUserService_GetCurrent_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
	}))
	defer server.Close()

	c := NewVastAIClient("bad-key", server.URL, "test")
	_, err := c.Users.GetCurrent(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
