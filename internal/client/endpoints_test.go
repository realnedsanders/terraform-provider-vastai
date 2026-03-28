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
// TestEndpointService_Create
// ---------------------------------------------------------------------------

func TestEndpointService_Create(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// First call: POST /api/v0/endptjobs/ (create)
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/endptjobs/" {
				t.Errorf("expected path /api/v0/endptjobs/, got %s", r.URL.Path)
			}

			// Verify auth header
			if r.Header.Get("Authorization") != "Bearer test-api-key" {
				t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body["client_id"] != "me" {
				t.Errorf("expected client_id=me, got %v", body["client_id"])
			}
			if body["endpoint_name"] != "test-endpoint" {
				t.Errorf("expected endpoint_name=test-endpoint, got %v", body["endpoint_name"])
			}
			if body["autoscaler_instance"] != "prod" {
				t.Errorf("expected autoscaler_instance=prod, got %v", body["autoscaler_instance"])
			}
			if body["target_util"] != 0.9 {
				t.Errorf("expected target_util=0.9, got %v", body["target_util"])
			}
			if body["max_workers"] != float64(20) {
				t.Errorf("expected max_workers=20, got %v", body["max_workers"])
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
			}); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
			return
		}

		// Second call: GET /api/v0/endptjobs/ (list for create-then-read)
		if r.Method != http.MethodGet {
			t.Errorf("expected GET for list, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/endptjobs/" {
			t.Errorf("expected path /api/v0/endptjobs/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"results": []map[string]interface{}{
				{
					"id":             100,
					"endpoint_name":  "old-endpoint",
					"target_util":    0.8,
					"cold_mult":      2.0,
					"max_workers":    10,
					"endpoint_state": "active",
				},
				{
					"id":             200,
					"endpoint_name":  "test-endpoint",
					"min_load":       0.0,
					"min_cold_load":  0.0,
					"target_util":    0.9,
					"cold_mult":      2.5,
					"cold_workers":   5,
					"max_workers":    20,
					"endpoint_state": "active",
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	ep, err := c.Endpoints.Create(context.Background(), &CreateEndpointRequest{
		EndpointName: "test-endpoint",
		TargetUtil:   0.9,
		ColdMult:     2.5,
		ColdWorkers:  5,
		MaxWorkers:   20,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if ep.ID != 200 {
		t.Errorf("expected endpoint ID 200, got %d", ep.ID)
	}
	if ep.EndpointName != "test-endpoint" {
		t.Errorf("expected endpoint name test-endpoint, got %q", ep.EndpointName)
	}
	if ep.TargetUtil != 0.9 {
		t.Errorf("expected target_util 0.9, got %f", ep.TargetUtil)
	}
	if ep.MaxWorkers != 20 {
		t.Errorf("expected max_workers 20, got %d", ep.MaxWorkers)
	}
}

// ---------------------------------------------------------------------------
// TestEndpointService_List
// ---------------------------------------------------------------------------

func TestEndpointService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/endptjobs/" {
			t.Errorf("expected path /api/v0/endptjobs/, got %s", r.URL.Path)
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"results": []map[string]interface{}{
				{
					"id":             10,
					"endpoint_name":  "llama-endpoint",
					"min_load":       0.0,
					"min_cold_load":  0.0,
					"target_util":    0.9,
					"cold_mult":      2.5,
					"cold_workers":   5,
					"max_workers":    20,
					"endpoint_state": "active",
				},
				{
					"id":             11,
					"endpoint_name":  "whisper-endpoint",
					"min_load":       1.0,
					"min_cold_load":  0.5,
					"target_util":    0.85,
					"cold_mult":      3.0,
					"cold_workers":   3,
					"max_workers":    10,
					"endpoint_state": "suspended",
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	endpoints, err := c.Endpoints.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(endpoints))
	}
	if endpoints[0].ID != 10 {
		t.Errorf("expected first endpoint ID 10, got %d", endpoints[0].ID)
	}
	if endpoints[0].EndpointName != "llama-endpoint" {
		t.Errorf("expected name llama-endpoint, got %q", endpoints[0].EndpointName)
	}
	if endpoints[0].TargetUtil != 0.9 {
		t.Errorf("expected target_util 0.9, got %f", endpoints[0].TargetUtil)
	}
	if endpoints[1].ID != 11 {
		t.Errorf("expected second endpoint ID 11, got %d", endpoints[1].ID)
	}
	if endpoints[1].EndpointState != "suspended" {
		t.Errorf("expected endpoint_state suspended, got %q", endpoints[1].EndpointState)
	}
}

// ---------------------------------------------------------------------------
// TestEndpointService_Update
// ---------------------------------------------------------------------------

func TestEndpointService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/endptjobs/42/" {
			t.Errorf("expected path /api/v0/endptjobs/42/, got %s", r.URL.Path)
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["client_id"] != "me" {
			t.Errorf("expected client_id=me, got %v", body["client_id"])
		}
		if body["endptjob_id"] != float64(42) {
			t.Errorf("expected endptjob_id=42, got %v", body["endptjob_id"])
		}
		if body["autoscaler_instance"] != "prod" {
			t.Errorf("expected autoscaler_instance=prod, got %v", body["autoscaler_instance"])
		}
		if body["max_workers"] != float64(30) {
			t.Errorf("expected max_workers=30, got %v", body["max_workers"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	maxWorkers := 30
	err := c.Endpoints.Update(context.Background(), 42, &UpdateEndpointRequest{
		MaxWorkers: &maxWorkers,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestEndpointService_Delete
// ---------------------------------------------------------------------------

func TestEndpointService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/endptjobs/42/" {
			t.Errorf("expected path /api/v0/endptjobs/42/, got %s", r.URL.Path)
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		// Verify body contains client_id and endptjob_id (Pitfall 2: delete needs JSON body)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var body map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["client_id"] != "me" {
			t.Errorf("expected client_id=me, got %v", body["client_id"])
		}
		if body["endptjob_id"] != float64(42) {
			t.Errorf("expected endptjob_id=42, got %v", body["endptjob_id"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	err := c.Endpoints.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}
