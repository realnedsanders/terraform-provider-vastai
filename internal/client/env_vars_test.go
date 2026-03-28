package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestEnvVarService_Create
// ---------------------------------------------------------------------------

func TestEnvVarService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/secrets/" {
			t.Errorf("expected path /api/v0/secrets/, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["key"] != "MY_VAR" {
			t.Errorf("expected key 'MY_VAR', got %q", body["key"])
		}
		if body["value"] != "secret-value" {
			t.Errorf("expected value 'secret-value', got %q", body["value"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.EnvVars.Create(context.Background(), "MY_VAR", "secret-value")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestEnvVarService_Create_Error
// ---------------------------------------------------------------------------

func TestEnvVarService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid key"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.EnvVars.Create(context.Background(), "", "value")
	if err == nil {
		t.Fatal("expected error on 400 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestEnvVarService_List
// ---------------------------------------------------------------------------

func TestEnvVarService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/secrets/" {
			t.Errorf("expected path /api/v0/secrets/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(EnvVarMap{
			Secrets: map[string]string{
				"MY_VAR":  "val1",
				"DB_HOST": "localhost",
			},
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	vars, err := c.EnvVars.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(vars))
	}
	if vars["MY_VAR"] != "val1" {
		t.Errorf("expected MY_VAR='val1', got %q", vars["MY_VAR"])
	}
	if vars["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST='localhost', got %q", vars["DB_HOST"])
	}
}

// ---------------------------------------------------------------------------
// TestEnvVarService_List_Error
// ---------------------------------------------------------------------------

func TestEnvVarService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.EnvVars.List(context.Background())
	if err == nil {
		t.Fatal("expected error on 500 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestEnvVarService_Update
// ---------------------------------------------------------------------------

func TestEnvVarService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/secrets/" {
			t.Errorf("expected path /api/v0/secrets/, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["key"] != "MY_VAR" {
			t.Errorf("expected key 'MY_VAR', got %q", body["key"])
		}
		if body["value"] != "new-value" {
			t.Errorf("expected value 'new-value', got %q", body["value"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.EnvVars.Update(context.Background(), "MY_VAR", "new-value")
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestEnvVarService_Delete
// ---------------------------------------------------------------------------

func TestEnvVarService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/secrets/" {
			t.Errorf("expected path /api/v0/secrets/, got %s", r.URL.Path)
		}

		// Verify body is present on DELETE (DeleteWithBody pattern)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if len(bodyBytes) == 0 {
			t.Fatal("expected non-empty body on DELETE request (DeleteWithBody)")
		}

		var body map[string]string
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["key"] != "MY_VAR" {
			t.Errorf("expected key 'MY_VAR', got %q", body["key"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.EnvVars.Delete(context.Background(), "MY_VAR")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestEnvVarService_Delete_Error
// ---------------------------------------------------------------------------

func TestEnvVarService_Delete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "key not found"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.EnvVars.Delete(context.Background(), "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error on 404 response, got nil")
	}
}
