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
// TestTemplateService_Create
// ---------------------------------------------------------------------------

func TestTemplateService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/template/" {
			t.Errorf("expected path /api/v0/template/, got %s", r.URL.Path)
		}

		var body CreateTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Name != "my-template" {
			t.Errorf("expected name %q, got %q", "my-template", body.Name)
		}
		if body.Image != "pytorch/pytorch:2.0" {
			t.Errorf("expected image %q, got %q", "pytorch/pytorch:2.0", body.Image)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(Template{
			ID:     100,
			HashID: "abc123",
			Name:   "my-template",
			Image:  "pytorch/pytorch:2.0",
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	tmpl, err := c.Templates.Create(context.Background(), &CreateTemplateRequest{
		Name:  "my-template",
		Image: "pytorch/pytorch:2.0",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if tmpl.ID != 100 {
		t.Errorf("expected ID 100, got %d", tmpl.ID)
	}
	if tmpl.HashID != "abc123" {
		t.Errorf("expected HashID %q, got %q", "abc123", tmpl.HashID)
	}
}

// ---------------------------------------------------------------------------
// TestTemplateService_Update
// ---------------------------------------------------------------------------

func TestTemplateService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/template/" {
			t.Errorf("expected path /api/v0/template/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["hash_id"] != "abc123" {
			t.Errorf("expected hash_id %q, got %v", "abc123", body["hash_id"])
		}
		if body["image"] != "pytorch/pytorch:2.1" {
			t.Errorf("expected image %q, got %v", "pytorch/pytorch:2.1", body["image"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(Template{
			ID:     100,
			HashID: "abc123",
			Image:  "pytorch/pytorch:2.1",
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	tmpl, err := c.Templates.Update(context.Background(), "abc123", &CreateTemplateRequest{
		Image: "pytorch/pytorch:2.1",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if tmpl.Image != "pytorch/pytorch:2.1" {
		t.Errorf("expected Image %q, got %q", "pytorch/pytorch:2.1", tmpl.Image)
	}
}

// ---------------------------------------------------------------------------
// TestTemplateService_Delete
// ---------------------------------------------------------------------------

func TestTemplateService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/template/" {
			t.Errorf("expected path /api/v0/template/, got %s", r.URL.Path)
		}

		// Verify body contains hash_id (Pitfall 5: delete sends hash_id in body)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var body map[string]string
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["hash_id"] != "abc123" {
			t.Errorf("expected hash_id %q in body, got %q", "abc123", body["hash_id"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Templates.Delete(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestTemplateService_Search
// ---------------------------------------------------------------------------

func TestTemplateService_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify query params
		if r.URL.Query().Get("select_cols") != "[*]" {
			t.Errorf("expected select_cols=[*], got %q", r.URL.Query().Get("select_cols"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"templates": []map[string]interface{}{
				{"id": 1, "hash_id": "hash1", "name": "template-1"},
				{"id": 2, "hash_id": "hash2", "name": "template-2"},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	templates, err := c.Templates.Search(context.Background(), "pytorch")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}
	if templates[0].Name != "template-1" {
		t.Errorf("expected first template name %q, got %q", "template-1", templates[0].Name)
	}
}
