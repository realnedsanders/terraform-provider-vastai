package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestNetworkVolumeService_Create
// ---------------------------------------------------------------------------

func TestNetworkVolumeService_Create(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// First call: PUT /api/v0/network_volumes/ (create)
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/network_volumes/" {
				t.Errorf("expected path /api/v0/network_volumes/, got %s", r.URL.Path)
			}

			// Verify auth header
			if r.Header.Get("Authorization") != "Bearer test-api-key" {
				t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body["size"] != float64(100) {
				t.Errorf("expected size 100, got %v", body["size"])
			}
			if body["id"] != float64(888) {
				t.Errorf("expected offer id 888, got %v", body["id"])
			}
			if body["name"] != "test-nv" {
				t.Errorf("expected name test-nv, got %v", body["name"])
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      456,
				"success": true,
			}); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
			return
		}

		// Second call: GET /api/v0/volumes?owner=me&type=network_volume (list)
		if r.Method != http.MethodGet {
			t.Errorf("expected GET for list, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/volumes" {
			t.Errorf("expected path /api/v0/volumes, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "network_volume" {
			t.Errorf("expected type=network_volume, got %s", r.URL.Query().Get("type"))
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"volumes": []map[string]interface{}{
				{
					"id":         456,
					"label":      "test-nv",
					"disk_space": 100.0,
					"status":     "active",
					"machine_id": 789,
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	vol, err := c.NetworkVolumes.Create(context.Background(), &CreateNetworkVolumeRequest{
		Size:    100,
		OfferID: 888,
		Name:    "test-nv",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if vol.ID != 456 {
		t.Errorf("expected volume ID 456, got %d", vol.ID)
	}
	if vol.Label != "test-nv" {
		t.Errorf("expected label test-nv, got %q", vol.Label)
	}
}

// ---------------------------------------------------------------------------
// TestNetworkVolumeService_List
// ---------------------------------------------------------------------------

func TestNetworkVolumeService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Pitfall 6: Same /volumes endpoint, different type parameter
		if r.URL.Path != "/api/v0/volumes" {
			t.Errorf("expected path /api/v0/volumes, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("owner") != "me" {
			t.Errorf("expected owner=me, got %s", r.URL.Query().Get("owner"))
		}
		if r.URL.Query().Get("type") != "network_volume" {
			t.Errorf("expected type=network_volume, got %s", r.URL.Query().Get("type"))
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"volumes": []map[string]interface{}{
				{
					"id":           10,
					"cluster_id":   20,
					"label":        "nv-1",
					"disk_space":   500.0,
					"status":       "active",
					"machine_id":   100,
					"geolocation":  "EU",
					"reliability2": 0.95,
				},
				{
					"id":          11,
					"label":       "nv-2",
					"disk_space":  1000.0,
					"status":      "provisioning",
					"geolocation": "US",
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	volumes, err := c.NetworkVolumes.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(volumes) != 2 {
		t.Fatalf("expected 2 network volumes, got %d", len(volumes))
	}
	if volumes[0].ID != 10 {
		t.Errorf("expected first volume ID 10, got %d", volumes[0].ID)
	}
	if volumes[0].ClusterID != 20 {
		t.Errorf("expected ClusterID 20, got %d", volumes[0].ClusterID)
	}
	if volumes[0].DiskSpace != 500.0 {
		t.Errorf("expected DiskSpace 500.0, got %f", volumes[0].DiskSpace)
	}
	if volumes[1].ID != 11 {
		t.Errorf("expected second volume ID 11, got %d", volumes[1].ID)
	}
}

// ---------------------------------------------------------------------------
// TestNetworkVolumeService_Delete
// ---------------------------------------------------------------------------

func TestNetworkVolumeService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		// Same delete endpoint as local volumes
		if r.URL.Path != "/api/v0/volumes/" {
			t.Errorf("expected path /api/v0/volumes/, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "456" {
			t.Errorf("expected query id=456, got %s", r.URL.Query().Get("id"))
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	err := c.NetworkVolumes.Delete(context.Background(), 456)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestNetworkVolumeService_SearchOffers
// ---------------------------------------------------------------------------

func TestNetworkVolumeService_SearchOffers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/network_volumes/search/" {
			t.Errorf("expected path /api/v0/network_volumes/search/, got %s", r.URL.Path)
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// Flat body -- filters are top-level (no q wrapper)
		diskSpace, ok := body["disk_space"].(map[string]interface{})
		if !ok {
			t.Fatal("expected disk_space filter in body")
		}
		if diskSpace["gte"] != float64(250) {
			t.Errorf("expected disk_space gte 250, got %v", diskSpace["gte"])
		}

		// Verify default filters at top level
		verified, ok := body["verified"].(map[string]interface{})
		if !ok {
			t.Fatal("expected default verified filter in body")
		}
		if verified["eq"] != true {
			t.Error("expected verified eq true")
		}
		external, ok := body["external"].(map[string]interface{})
		if !ok {
			t.Fatal("expected default external filter in body")
		}
		if external["eq"] != false {
			t.Error("expected external eq false")
		}

		// Verify allocated_storage at top level
		if body["allocated_storage"] != float64(5) {
			t.Errorf("expected allocated_storage 5, got %v", body["allocated_storage"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"offers": []map[string]interface{}{
				{
					"id":             701,
					"disk_space":     500.0,
					"storage_cost":   0.10,
					"inet_up":        1000.0,
					"inet_down":      2000.0,
					"reliability":    0.99,
					"duration":       172800.0,
					"verification":   "verified",
					"host_id":        201,
					"cluster_id":     301,
					"geolocation":    "US",
					"nw_disk_min_bw": 100.0,
					"nw_disk_max_bw": 500.0,
					"nw_disk_avg_bw": 300.0,
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")

	diskSpace := 250.0
	offers, err := c.NetworkVolumes.SearchOffers(context.Background(), &NetworkVolumeOfferSearchParams{
		DiskSpace:        &diskSpace,
		Limit:            10,
		AllocatedStorage: 5,
	})
	if err != nil {
		t.Fatalf("SearchOffers returned error: %v", err)
	}
	if len(offers) != 1 {
		t.Fatalf("expected 1 offer, got %d", len(offers))
	}

	o := offers[0]
	if o.ID != 701 {
		t.Errorf("expected offer ID 701, got %d", o.ID)
	}
	if o.DiskSpace != 500.0 {
		t.Errorf("expected DiskSpace 500.0, got %f", o.DiskSpace)
	}
	if o.StorageCost != 0.10 {
		t.Errorf("expected StorageCost 0.10, got %f", o.StorageCost)
	}
	// Verify network-volume-specific fields
	if o.ClusterID != 301 {
		t.Errorf("expected ClusterID 301, got %d", o.ClusterID)
	}
	if o.NWDiskMinBW != 100.0 {
		t.Errorf("expected NWDiskMinBW 100.0, got %f", o.NWDiskMinBW)
	}
	if o.NWDiskMaxBW != 500.0 {
		t.Errorf("expected NWDiskMaxBW 500.0, got %f", o.NWDiskMaxBW)
	}
	if o.NWDiskAvgBW != 300.0 {
		t.Errorf("expected NWDiskAvgBW 300.0, got %f", o.NWDiskAvgBW)
	}
	if o.Geolocation != "US" {
		t.Errorf("expected Geolocation US, got %q", o.Geolocation)
	}
	if o.Reliability != 0.99 {
		t.Errorf("expected Reliability 0.99, got %f", o.Reliability)
	}
}
