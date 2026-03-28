package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestApiKeyService_Create
// ---------------------------------------------------------------------------

func TestApiKeyService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/auth/apikeys/" {
			t.Errorf("expected path /api/v0/auth/apikeys/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["name"] != "my-key" {
			t.Errorf("expected name 'my-key', got %v", body["name"])
		}
		if body["permissions"] == nil {
			t.Error("expected permissions field in body")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ApiKey{
			ID:   100,
			Name: "my-key",
			Key:  "vast-api-key-secret-value",
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	perms := json.RawMessage(`{"api":{"instance_read":{}}}`)
	key, err := c.ApiKeys.Create(context.Background(), "my-key", perms)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if key.ID != 100 {
		t.Errorf("expected ID 100, got %d", key.ID)
	}
	if key.Name != "my-key" {
		t.Errorf("expected name 'my-key', got %q", key.Name)
	}
	if key.Key != "vast-api-key-secret-value" {
		t.Errorf("expected key value, got %q", key.Key)
	}
}

// ---------------------------------------------------------------------------
// TestApiKeyService_Create_Error
// ---------------------------------------------------------------------------

func TestApiKeyService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid permissions"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	perms := json.RawMessage(`{}`)
	_, err := c.ApiKeys.Create(context.Background(), "bad-key", perms)
	if err == nil {
		t.Fatal("expected error on 400 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestApiKeyService_List
// ---------------------------------------------------------------------------

func TestApiKeyService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/auth/apikeys/" {
			t.Errorf("expected path /api/v0/auth/apikeys/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode([]ApiKey{
			{ID: 1, Name: "key-one"},
			{ID: 2, Name: "key-two"},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	keys, err := c.ApiKeys.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].Name != "key-one" {
		t.Errorf("expected first key name 'key-one', got %q", keys[0].Name)
	}
}

// ---------------------------------------------------------------------------
// TestApiKeyService_List_Error
// ---------------------------------------------------------------------------

func TestApiKeyService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("bad-key", server.URL, "test")
	_, err := c.ApiKeys.List(context.Background())
	if err == nil {
		t.Fatal("expected error on 401 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestApiKeyService_Delete
// ---------------------------------------------------------------------------

func TestApiKeyService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/auth/apikeys/100/" {
			t.Errorf("expected path /api/v0/auth/apikeys/100/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.ApiKeys.Delete(context.Background(), 100)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestApiKeyService_Delete_Error
// ---------------------------------------------------------------------------

func TestApiKeyService_Delete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "API key not found"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.ApiKeys.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error on 404 response, got nil")
	}
}
