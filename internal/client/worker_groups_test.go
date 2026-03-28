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
// TestWorkerGroupService_Create
// ---------------------------------------------------------------------------

func TestWorkerGroupService_Create(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// First call: POST /api/v0/autojobs/ (create)
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/autojobs/" {
				t.Errorf("expected path /api/v0/autojobs/, got %s", r.URL.Path)
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
			if body["autoscaler_instance"] != "prod" {
				t.Errorf("expected autoscaler_instance=prod, got %v", body["autoscaler_instance"])
			}
			if body["endpoint_id"] != float64(100) {
				t.Errorf("expected endpoint_id=100, got %v", body["endpoint_id"])
			}
			if body["template_hash"] != "abc123hash" {
				t.Errorf("expected template_hash=abc123hash, got %v", body["template_hash"])
			}
			if body["test_workers"] != float64(3) {
				t.Errorf("expected test_workers=3, got %v", body["test_workers"])
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
			}); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
			return
		}

		// Second call: GET /api/v0/autojobs/ (list for create-then-read)
		if r.Method != http.MethodGet {
			t.Errorf("expected GET for list, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/autojobs/" {
			t.Errorf("expected path /api/v0/autojobs/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"results": []map[string]interface{}{
				{
					"id":            50,
					"endpoint_name": "test-endpoint",
					"endpoint_id":   100,
					"template_hash": "oldhash",
					"template_id":   10,
					"search_params": "gpu_ram>=16",
					"gpu_ram":       16.0,
					"min_load":      0.0,
					"target_util":   0.9,
					"cold_mult":     2.0,
					"cold_workers":  3,
					"test_workers":  2,
				},
				{
					"id":            75,
					"endpoint_name": "test-endpoint",
					"endpoint_id":   100,
					"template_hash": "abc123hash",
					"template_id":   20,
					"search_params": "gpu_ram>=24 num_gpus=2",
					"gpu_ram":       24.0,
					"min_load":      0.0,
					"target_util":   0.9,
					"cold_mult":     2.5,
					"cold_workers":  5,
					"test_workers":  3,
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	wg, err := c.WorkerGroups.Create(context.Background(), &CreateWorkerGroupRequest{
		EndpointID:   100,
		TemplateHash: "abc123hash",
		SearchParams: "gpu_ram>=24 num_gpus=2",
		GpuRAM:       24.0,
		TargetUtil:   0.9,
		ColdMult:     2.5,
		ColdWorkers:  5,
		TestWorkers:  3,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	// Should return the highest ID (75, not 50) -- create-then-read finds newest
	if wg.ID != 75 {
		t.Errorf("expected worker group ID 75 (highest), got %d", wg.ID)
	}
	if wg.TemplateHash != "abc123hash" {
		t.Errorf("expected template_hash abc123hash, got %q", wg.TemplateHash)
	}
	if wg.EndpointID != 100 {
		t.Errorf("expected endpoint_id 100, got %d", wg.EndpointID)
	}
	if wg.TestWorkers != 3 {
		t.Errorf("expected test_workers 3, got %d", wg.TestWorkers)
	}
}

// ---------------------------------------------------------------------------
// TestWorkerGroupService_List
// ---------------------------------------------------------------------------

func TestWorkerGroupService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/autojobs/" {
			t.Errorf("expected path /api/v0/autojobs/, got %s", r.URL.Path)
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
					"id":            10,
					"endpoint_name": "llama-ep",
					"endpoint_id":   100,
					"template_hash": "hash1",
					"template_id":   1,
					"search_params": "gpu_ram>=24",
					"launch_args":   "--model llama",
					"gpu_ram":       24.0,
					"min_load":      0.0,
					"target_util":   0.9,
					"cold_mult":     2.0,
					"cold_workers":  5,
					"test_workers":  3,
				},
				{
					"id":            20,
					"endpoint_name": "whisper-ep",
					"endpoint_id":   200,
					"template_hash": "hash2",
					"template_id":   2,
					"search_params": "gpu_ram>=8",
					"gpu_ram":       8.0,
					"min_load":      1.0,
					"target_util":   0.85,
					"cold_mult":     3.0,
					"cold_workers":  2,
					"test_workers":  1,
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	groups, err := c.WorkerGroups.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 worker groups, got %d", len(groups))
	}
	if groups[0].ID != 10 {
		t.Errorf("expected first group ID 10, got %d", groups[0].ID)
	}
	if groups[0].EndpointName != "llama-ep" {
		t.Errorf("expected endpoint_name llama-ep, got %q", groups[0].EndpointName)
	}
	if groups[0].GpuRAM != 24.0 {
		t.Errorf("expected gpu_ram 24.0, got %f", groups[0].GpuRAM)
	}
	if groups[0].LaunchArgs != "--model llama" {
		t.Errorf("expected launch_args '--model llama', got %q", groups[0].LaunchArgs)
	}
	if groups[1].ID != 20 {
		t.Errorf("expected second group ID 20, got %d", groups[1].ID)
	}
	if groups[1].TargetUtil != 0.85 {
		t.Errorf("expected target_util 0.85, got %f", groups[1].TargetUtil)
	}
}

// ---------------------------------------------------------------------------
// TestWorkerGroupService_Update
// ---------------------------------------------------------------------------

func TestWorkerGroupService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/autojobs/55/" {
			t.Errorf("expected path /api/v0/autojobs/55/, got %s", r.URL.Path)
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
		if body["autojob_id"] != float64(55) {
			t.Errorf("expected autojob_id=55, got %v", body["autojob_id"])
		}
		if body["search_params"] != "gpu_ram>=48 num_gpus=4" {
			t.Errorf("expected search_params update, got %v", body["search_params"])
		}
		if body["test_workers"] != float64(5) {
			t.Errorf("expected test_workers=5, got %v", body["test_workers"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	testWorkers := 5
	err := c.WorkerGroups.Update(context.Background(), 55, &UpdateWorkerGroupRequest{
		SearchParams: "gpu_ram>=48 num_gpus=4",
		TestWorkers:  &testWorkers,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestWorkerGroupService_Delete
// ---------------------------------------------------------------------------

func TestWorkerGroupService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/autojobs/55/" {
			t.Errorf("expected path /api/v0/autojobs/55/, got %s", r.URL.Path)
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		// Verify body contains client_id and autojob_id (Pitfall 2: delete needs JSON body)
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
		if body["autojob_id"] != float64(55) {
			t.Errorf("expected autojob_id=55, got %v", body["autojob_id"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	err := c.WorkerGroups.Delete(context.Background(), 55)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}
