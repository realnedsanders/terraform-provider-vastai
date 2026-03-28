package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestAuditLogService_List
// ---------------------------------------------------------------------------

func TestAuditLogService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/audit_logs/" {
			t.Errorf("expected path /api/v0/audit_logs/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"ip_address": "192.168.1.1",
				"api_key_id": 10,
				"created_at": "2025-01-15T10:30:00Z",
				"api_route":  "POST /api/v0/instances/create/",
				"args":       "{\"image\":\"pytorch/pytorch\"}",
			},
			{
				"ip_address": "10.0.0.1",
				"api_key_id": 20,
				"created_at": "2025-01-15T09:00:00Z",
				"api_route":  "GET /api/v0/instances/",
				"args":       "",
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	logs, err := c.AuditLogs.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(logs))
	}
	if logs[0].IPAddress != "192.168.1.1" {
		t.Errorf("expected ip_address 192.168.1.1, got %q", logs[0].IPAddress)
	}
	if logs[0].ApiKeyID != 10 {
		t.Errorf("expected api_key_id 10, got %d", logs[0].ApiKeyID)
	}
	if logs[0].CreatedAt != "2025-01-15T10:30:00Z" {
		t.Errorf("expected created_at, got %q", logs[0].CreatedAt)
	}
	if logs[0].ApiRoute != "POST /api/v0/instances/create/" {
		t.Errorf("expected api_route, got %q", logs[0].ApiRoute)
	}
	if logs[0].Args != "{\"image\":\"pytorch/pytorch\"}" {
		t.Errorf("expected args, got %q", logs[0].Args)
	}
	if logs[1].IPAddress != "10.0.0.1" {
		t.Errorf("expected second ip_address 10.0.0.1, got %q", logs[1].IPAddress)
	}
}

// ---------------------------------------------------------------------------
// TestAuditLogService_List_Empty
// ---------------------------------------------------------------------------

func TestAuditLogService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode([]map[string]interface{}{}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	logs, err := c.AuditLogs.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 entries, got %d", len(logs))
	}
}

// ---------------------------------------------------------------------------
// TestAuditLogService_List_Error
// ---------------------------------------------------------------------------

func TestAuditLogService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "server error"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.AuditLogs.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
