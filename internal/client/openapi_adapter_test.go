package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestOpenAPIClientPreservesRequestMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v0/users/current/" {
			t.Errorf("path = %q, want /api/v0/users/current/", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-key" {
			t.Errorf("Authorization = %q, want Bearer secret-key", got)
		}
		if got := r.URL.Query().Get("api_key"); got != "secret-key" {
			t.Errorf("api_key = %q, want secret-key", got)
		}
		if got := r.Header.Get("User-Agent"); got != "terraform-provider-vastai/1.2.3" {
			t.Errorf("User-Agent = %q, want terraform-provider-vastai/1.2.3", got)
		}
		if got := r.URL.Query().Get("owner"); got != "me" {
			t.Errorf("owner = %q, want me", got)
		}
		w.Header().Set("Content-Type", applicationJSON)
		if err := json.NewEncoder(w).Encode(User{ID: 7}); err != nil {
			t.Fatalf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := NewVastAIClient("secret-key", server.URL, "1.2.3")
	user, err := client.Users.GetCurrent(context.Background())
	if err != nil {
		t.Fatalf("GetCurrent returned error: %v", err)
	}
	if user.ID != 7 {
		t.Errorf("user ID = %d, want 7", user.ID)
	}
}

func TestOpenAPIClientPreservesRetryPolicy(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			http.Error(w, "temporary failure", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", applicationJSON)
		if err := json.NewEncoder(w).Encode(endpointListResponse{Success: true}); err != nil {
			t.Fatalf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := NewVastAIClient("secret-key", server.URL, "test")
	if _, err := client.Endpoints.List(context.Background()); err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("attempts = %d, want 2", got)
	}
}
