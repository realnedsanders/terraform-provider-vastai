package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestSSHKeyService_Create
// ---------------------------------------------------------------------------

func TestSSHKeyService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/ssh/" {
			t.Errorf("expected path /api/v0/ssh/, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["ssh_key"] != "ssh-ed25519 AAAA... user@host" {
			t.Errorf("expected ssh_key field, got %q", body["ssh_key"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(sshKeyCreateResponse{
			Key: SSHKey{
				ID:        10,
				PublicKey: "ssh-ed25519 AAAA... user@host",
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	key, err := c.SSHKeys.Create(context.Background(), "ssh-ed25519 AAAA... user@host")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if key.ID != 10 {
		t.Errorf("expected ID 10, got %d", key.ID)
	}
}

// ---------------------------------------------------------------------------
// TestSSHKeyService_List
// ---------------------------------------------------------------------------

func TestSSHKeyService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/ssh/" {
			t.Errorf("expected path /api/v0/ssh/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode([]SSHKey{
			{ID: 1, SSHKey: "key1"},
			{ID: 2, SSHKey: "key2"},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	keys, err := c.SSHKeys.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].ID != 1 {
		t.Errorf("expected first key ID 1, got %d", keys[0].ID)
	}
}

// ---------------------------------------------------------------------------
// TestSSHKeyService_Update
// ---------------------------------------------------------------------------

func TestSSHKeyService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/ssh/10/" {
			t.Errorf("expected path /api/v0/ssh/10/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["id"] != float64(10) {
			t.Errorf("expected id 10, got %v", body["id"])
		}
		if body["ssh_key"] != "ssh-rsa BBBB... user@host" {
			t.Errorf("expected ssh_key field, got %v", body["ssh_key"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(SSHKey{
			ID:     10,
			SSHKey: "ssh-rsa BBBB... user@host",
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	key, err := c.SSHKeys.Update(context.Background(), 10, "ssh-rsa BBBB... user@host")
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if key.SSHKey != "ssh-rsa BBBB... user@host" {
		t.Errorf("expected updated ssh_key, got %q", key.SSHKey)
	}
}

// ---------------------------------------------------------------------------
// TestSSHKeyService_Delete
// ---------------------------------------------------------------------------

func TestSSHKeyService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/ssh/10/" {
			t.Errorf("expected path /api/v0/ssh/10/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.SSHKeys.Delete(context.Background(), 10)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestSSHKeyService_AttachToInstance
// ---------------------------------------------------------------------------

func TestSSHKeyService_AttachToInstance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/ssh/" {
			t.Errorf("expected path /api/v0/instances/42/ssh/, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["ssh_key"] != "ssh-ed25519 AAAA... user@host" {
			t.Errorf("expected ssh_key in body, got %q", body["ssh_key"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.SSHKeys.AttachToInstance(context.Background(), 42, "ssh-ed25519 AAAA... user@host")
	if err != nil {
		t.Fatalf("AttachToInstance returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestSSHKeyService_DetachFromInstance
// ---------------------------------------------------------------------------

func TestSSHKeyService_DetachFromInstance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/ssh/10/" {
			t.Errorf("expected path /api/v0/instances/42/ssh/10/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.SSHKeys.DetachFromInstance(context.Background(), 42, 10)
	if err != nil {
		t.Fatalf("DetachFromInstance returned error: %v", err)
	}
}
