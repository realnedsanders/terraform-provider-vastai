package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestSubaccountService_Create
// ---------------------------------------------------------------------------

func TestSubaccountService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/users/" {
			t.Errorf("expected path /api/v0/users/, got %s", r.URL.Path)
		}

		var body SubaccountCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Email != "sub@example.com" {
			t.Errorf("expected email 'sub@example.com', got %q", body.Email)
		}
		if body.Username != "subuser" {
			t.Errorf("expected username 'subuser', got %q", body.Username)
		}
		if body.Password != "secret123" {
			t.Errorf("expected password 'secret123', got %q", body.Password)
		}
		if body.HostOnly != false {
			t.Errorf("expected host_only false, got %v", body.HostOnly)
		}
		if body.ParentID != "me" {
			t.Errorf("expected parent_id 'me', got %q", body.ParentID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(Subaccount{
			ID:       200,
			Email:    "sub@example.com",
			Username: "subuser",
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	sub, err := c.Subaccounts.Create(context.Background(), "sub@example.com", "subuser", "secret123", false)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if sub.ID != 200 {
		t.Errorf("expected ID 200, got %d", sub.ID)
	}
	if sub.Email != "sub@example.com" {
		t.Errorf("expected email 'sub@example.com', got %q", sub.Email)
	}
	if sub.Username != "subuser" {
		t.Errorf("expected username 'subuser', got %q", sub.Username)
	}
}

// ---------------------------------------------------------------------------
// TestSubaccountService_Create_Error
// ---------------------------------------------------------------------------

func TestSubaccountService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "email already in use"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Subaccounts.Create(context.Background(), "existing@example.com", "user", "pass", false)
	if err == nil {
		t.Fatal("expected error on 409 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestSubaccountService_Create_HostOnly
// ---------------------------------------------------------------------------

func TestSubaccountService_Create_HostOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body SubaccountCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.HostOnly != true {
			t.Errorf("expected host_only true, got %v", body.HostOnly)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(Subaccount{ID: 201, Email: "host@example.com", Username: "hostuser"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	sub, err := c.Subaccounts.Create(context.Background(), "host@example.com", "hostuser", "pass", true)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if sub.ID != 201 {
		t.Errorf("expected ID 201, got %d", sub.ID)
	}
}

// ---------------------------------------------------------------------------
// TestSubaccountService_List
// ---------------------------------------------------------------------------

func TestSubaccountService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// The path will include the query parameter
		if r.URL.Path != "/api/v0/subaccounts" {
			t.Errorf("expected path /api/v0/subaccounts, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("owner") != "me" {
			t.Errorf("expected owner=me query param, got %q", r.URL.Query().Get("owner"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(SubaccountListResponse{
			Users: []Subaccount{
				{ID: 100, Email: "sub1@example.com", Username: "sub1"},
				{ID: 101, Email: "sub2@example.com", Username: "sub2"},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	subs, err := c.Subaccounts.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 subaccounts, got %d", len(subs))
	}
	if subs[0].Email != "sub1@example.com" {
		t.Errorf("expected first subaccount email 'sub1@example.com', got %q", subs[0].Email)
	}
}

// ---------------------------------------------------------------------------
// TestSubaccountService_List_Error
// ---------------------------------------------------------------------------

func TestSubaccountService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Subaccounts.List(context.Background())
	if err == nil {
		t.Fatal("expected error on 403 response, got nil")
	}
}
