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
// TestClusterService_Create
// ---------------------------------------------------------------------------

func TestClusterService_Create(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// First call: POST /api/v0/cluster/ (create)
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/cluster/" {
				t.Errorf("expected path /api/v0/cluster/, got %s", r.URL.Path)
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body["subnet"] != "10.0.0.0/24" {
				t.Errorf("expected subnet=10.0.0.0/24, got %v", body["subnet"])
			}
			if body["manager_id"] != float64(100) {
				t.Errorf("expected manager_id=100, got %v", body["manager_id"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"msg": "cluster created",
			})
			return
		}

		// Second call: GET /api/v0/clusters/ (list for create-then-read)
		if r.Method != http.MethodGet {
			t.Errorf("expected GET for list, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/clusters/" {
			t.Errorf("expected path /api/v0/clusters/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clusters": map[string]interface{}{
				"42": map[string]interface{}{
					"subnet": "10.0.0.0/24",
					"nodes": []map[string]interface{}{
						{
							"machine_id":         100,
							"is_cluster_manager": true,
							"local_ip":           "10.0.0.1",
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	cluster, err := c.Clusters.Create(context.Background(), "10.0.0.0/24", 100)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if cluster.ID != 42 {
		t.Errorf("expected ID 42, got %d", cluster.ID)
	}
	if cluster.Subnet != "10.0.0.0/24" {
		t.Errorf("expected subnet 10.0.0.0/24, got %q", cluster.Subnet)
	}
	if len(cluster.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(cluster.Nodes))
	}
	if cluster.Nodes[0].MachineID != 100 {
		t.Errorf("expected node machine_id 100, got %d", cluster.Nodes[0].MachineID)
	}
}

// ---------------------------------------------------------------------------
// TestClusterService_Create_Error
// ---------------------------------------------------------------------------

func TestClusterService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid subnet"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Clusters.Create(context.Background(), "bad-subnet", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestClusterService_List
// ---------------------------------------------------------------------------

func TestClusterService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/clusters/" {
			t.Errorf("expected path /api/v0/clusters/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clusters": map[string]interface{}{
				"10": map[string]interface{}{
					"subnet": "10.0.0.0/24",
					"nodes":  []interface{}{},
				},
				"20": map[string]interface{}{
					"subnet": "192.168.1.0/24",
					"nodes": []map[string]interface{}{
						{
							"machine_id":         200,
							"is_cluster_manager": true,
							"local_ip":           "192.168.1.1",
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	clusters, err := c.Clusters.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}

	cluster10, ok := clusters["10"]
	if !ok {
		t.Fatal("expected cluster with key '10'")
	}
	if cluster10.ID != 10 {
		t.Errorf("expected ID 10, got %d", cluster10.ID)
	}
	if cluster10.Subnet != "10.0.0.0/24" {
		t.Errorf("expected subnet 10.0.0.0/24, got %q", cluster10.Subnet)
	}

	cluster20, ok := clusters["20"]
	if !ok {
		t.Fatal("expected cluster with key '20'")
	}
	if cluster20.ID != 20 {
		t.Errorf("expected ID 20, got %d", cluster20.ID)
	}
	if len(cluster20.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(cluster20.Nodes))
	}
}

// ---------------------------------------------------------------------------
// TestClusterService_List_Error
// ---------------------------------------------------------------------------

func TestClusterService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Clusters.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestClusterService_Delete
// ---------------------------------------------------------------------------

func TestClusterService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/cluster/" {
			t.Errorf("expected path /api/v0/cluster/, got %s", r.URL.Path)
		}

		bodyBytes, _ := io.ReadAll(r.Body)
		var body map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["cluster_id"] != float64(42) {
			t.Errorf("expected cluster_id=42, got %v", body["cluster_id"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Clusters.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestClusterService_Delete_Error
// ---------------------------------------------------------------------------

func TestClusterService_Delete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "cluster not found"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Clusters.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestClusterService_JoinMachine
// ---------------------------------------------------------------------------

func TestClusterService_JoinMachine(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/cluster/" {
			t.Errorf("expected path /api/v0/cluster/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["cluster_id"] != float64(42) {
			t.Errorf("expected cluster_id=42, got %v", body["cluster_id"])
		}
		machineIDs, ok := body["machine_ids"].([]interface{})
		if !ok {
			t.Fatalf("expected machine_ids to be array, got %T", body["machine_ids"])
		}
		if len(machineIDs) != 2 {
			t.Errorf("expected 2 machine_ids, got %d", len(machineIDs))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Clusters.JoinMachine(context.Background(), 42, []int{200, 300})
	if err != nil {
		t.Fatalf("JoinMachine returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestClusterService_RemoveMachine
// ---------------------------------------------------------------------------

func TestClusterService_RemoveMachine(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/cluster/remove_machine/" {
			t.Errorf("expected path /api/v0/cluster/remove_machine/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["cluster_id"] != float64(42) {
			t.Errorf("expected cluster_id=42, got %v", body["cluster_id"])
		}
		if body["machine_id"] != float64(200) {
			t.Errorf("expected machine_id=200, got %v", body["machine_id"])
		}
		// Should not have new_manager_id when nil
		if _, ok := body["new_manager_id"]; ok {
			t.Error("expected no new_manager_id when nil")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Clusters.RemoveMachine(context.Background(), 42, 200, nil)
	if err != nil {
		t.Fatalf("RemoveMachine returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestClusterService_RemoveMachine_WithNewManager
// ---------------------------------------------------------------------------

func TestClusterService_RemoveMachine_WithNewManager(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/cluster/remove_machine/" {
			t.Errorf("expected path /api/v0/cluster/remove_machine/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["cluster_id"] != float64(42) {
			t.Errorf("expected cluster_id=42, got %v", body["cluster_id"])
		}
		if body["machine_id"] != float64(200) {
			t.Errorf("expected machine_id=200, got %v", body["machine_id"])
		}
		if body["new_manager_id"] != float64(300) {
			t.Errorf("expected new_manager_id=300, got %v", body["new_manager_id"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	newManager := 300
	err := c.Clusters.RemoveMachine(context.Background(), 42, 200, &newManager)
	if err != nil {
		t.Fatalf("RemoveMachine returned error: %v", err)
	}
}
