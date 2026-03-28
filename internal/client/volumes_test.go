package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestVolumeService_Create
// ---------------------------------------------------------------------------

func TestVolumeService_Create(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// First call: PUT /api/v0/volumes/ (create)
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/volumes/" {
				t.Errorf("expected path /api/v0/volumes/, got %s", r.URL.Path)
			}

			// Verify auth header
			if r.Header.Get("Authorization") != "Bearer test-api-key" {
				t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body["size"] != float64(50) {
				t.Errorf("expected size 50, got %v", body["size"])
			}
			if body["id"] != float64(999) {
				t.Errorf("expected offer id 999, got %v", body["id"])
			}
			if body["name"] != "test-vol" {
				t.Errorf("expected name test-vol, got %v", body["name"])
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      123,
				"success": true,
			}); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
			return
		}

		// Second call: GET /api/v0/volumes?owner=me&type=local_volume (list)
		if r.Method != http.MethodGet {
			t.Errorf("expected GET for list, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/volumes" {
			t.Errorf("expected path /api/v0/volumes, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "local_volume" {
			t.Errorf("expected type=local_volume, got %s", r.URL.Query().Get("type"))
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"volumes": []map[string]interface{}{
				{
					"id":         123,
					"label":      "test-vol",
					"disk_space": 50.0,
					"status":     "active",
					"machine_id": 456,
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	vol, err := c.Volumes.Create(context.Background(), &CreateVolumeRequest{
		Size:    50,
		OfferID: 999,
		Name:    "test-vol",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if vol.ID != 123 {
		t.Errorf("expected volume ID 123, got %d", vol.ID)
	}
	if vol.Label != "test-vol" {
		t.Errorf("expected label test-vol, got %q", vol.Label)
	}
	if vol.DiskSpace != 50.0 {
		t.Errorf("expected disk_space 50.0, got %f", vol.DiskSpace)
	}
}

// ---------------------------------------------------------------------------
// TestVolumeService_Clone
// ---------------------------------------------------------------------------

func TestVolumeService_Clone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/volumes/copy/" {
			t.Errorf("expected path /api/v0/volumes/copy/, got %s", r.URL.Path)
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["src_id"] != float64(100) {
			t.Errorf("expected src_id 100, got %v", body["src_id"])
		}
		if body["dst_id"] != float64(200) {
			t.Errorf("expected dst_id 200, got %v", body["dst_id"])
		}
		if body["size"] != float64(75.5) {
			t.Errorf("expected size 75.5, got %v", body["size"])
		}
		if body["disable_compression"] != true {
			t.Errorf("expected disable_compression true, got %v", body["disable_compression"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	err := c.Volumes.Clone(context.Background(), &CloneVolumeRequest{
		SourceID:           100,
		DestOfferID:        200,
		Size:               75.5,
		DisableCompression: true,
	})
	if err != nil {
		t.Fatalf("Clone returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestVolumeService_List
// ---------------------------------------------------------------------------

func TestVolumeService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/volumes" {
			t.Errorf("expected path /api/v0/volumes, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("owner") != "me" {
			t.Errorf("expected owner=me, got %s", r.URL.Query().Get("owner"))
		}
		if r.URL.Query().Get("type") != "local_volume" {
			t.Errorf("expected type=local_volume, got %s", r.URL.Query().Get("type"))
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
					"id":             1,
					"cluster_id":     10,
					"label":          "vol-1",
					"disk_space":     100.0,
					"status":         "active",
					"disk_name":      "nvme0n1",
					"driver_version": "535.86.05",
					"inet_up":        1000.0,
					"inet_down":      2000.0,
					"reliability2":   0.99,
					"start_date":     1711900000.0,
					"machine_id":     789,
					"verification":   "verified",
					"host_id":        101,
					"geolocation":    "US",
					"instances":      []int{1001, 1002},
				},
				{
					"id":          2,
					"label":       "vol-2",
					"disk_space":  200.0,
					"status":      "provisioning",
					"machine_id":  790,
					"geolocation": "DE",
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	volumes, err := c.Volumes.List(context.Background(), "local_volume")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(volumes) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(volumes))
	}

	// Verify first volume fields
	v := volumes[0]
	if v.ID != 1 {
		t.Errorf("expected ID 1, got %d", v.ID)
	}
	if v.ClusterID != 10 {
		t.Errorf("expected ClusterID 10, got %d", v.ClusterID)
	}
	if v.Label != "vol-1" {
		t.Errorf("expected Label vol-1, got %q", v.Label)
	}
	if v.DiskSpace != 100.0 {
		t.Errorf("expected DiskSpace 100.0, got %f", v.DiskSpace)
	}
	if v.Status != "active" {
		t.Errorf("expected Status active, got %q", v.Status)
	}
	if v.DiskName != "nvme0n1" {
		t.Errorf("expected DiskName nvme0n1, got %q", v.DiskName)
	}
	if v.InetUp != 1000.0 {
		t.Errorf("expected InetUp 1000.0, got %f", v.InetUp)
	}
	if v.InetDown != 2000.0 {
		t.Errorf("expected InetDown 2000.0, got %f", v.InetDown)
	}
	if v.Reliability != 0.99 {
		t.Errorf("expected Reliability 0.99, got %f", v.Reliability)
	}
	if v.MachineID != 789 {
		t.Errorf("expected MachineID 789, got %d", v.MachineID)
	}
	if v.Verification != "verified" {
		t.Errorf("expected Verification verified, got %q", v.Verification)
	}
	if v.HostID != 101 {
		t.Errorf("expected HostID 101, got %d", v.HostID)
	}
	if v.Geolocation != "US" {
		t.Errorf("expected Geolocation US, got %q", v.Geolocation)
	}

	// Verify second volume
	if volumes[1].ID != 2 {
		t.Errorf("expected second volume ID 2, got %d", volumes[1].ID)
	}
	if volumes[1].Geolocation != "DE" {
		t.Errorf("expected second volume Geolocation DE, got %q", volumes[1].Geolocation)
	}
}

// ---------------------------------------------------------------------------
// TestVolumeService_Delete
// ---------------------------------------------------------------------------

func TestVolumeService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		// Pitfall 2: Uses query parameter, NOT path parameter
		if r.URL.Path != "/api/v0/volumes/" {
			t.Errorf("expected path /api/v0/volumes/, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "123" {
			t.Errorf("expected query id=123, got %s", r.URL.Query().Get("id"))
		}

		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	err := c.Volumes.Delete(context.Background(), 123)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestVolumeService_SearchOffers
// ---------------------------------------------------------------------------

func TestVolumeService_SearchOffers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/volumes/search/" {
			t.Errorf("expected path /api/v0/volumes/search/, got %s", r.URL.Path)
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
		// Verify disk_space gte filter (overridden from default 1 to 100)
		diskSpace, ok := body["disk_space"].(map[string]interface{})
		if !ok {
			t.Fatal("expected disk_space filter in body")
		}
		if diskSpace["gte"] != float64(100) {
			t.Errorf("expected disk_space gte 100, got %v", diskSpace["gte"])
		}

		// Verify storage_cost lte filter
		storageCost, ok := body["storage_cost"].(map[string]interface{})
		if !ok {
			t.Fatal("expected storage_cost filter in body")
		}
		if storageCost["lte"] != 0.5 {
			t.Errorf("expected storage_cost lte 0.5, got %v", storageCost["lte"])
		}

		// Verify inet_up gte filter
		inetUp, ok := body["inet_up"].(map[string]interface{})
		if !ok {
			t.Fatal("expected inet_up filter in body")
		}
		if inetUp["gte"] != float64(500) {
			t.Errorf("expected inet_up gte 500, got %v", inetUp["gte"])
		}

		// Verify geolocation eq filter
		geo, ok := body["geolocation"].(map[string]interface{})
		if !ok {
			t.Fatal("expected geolocation filter in body")
		}
		if geo["eq"] != "US" {
			t.Errorf("expected geolocation eq US, got %v", geo["eq"])
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

		// Verify limit at top level
		if body["limit"] != float64(5) {
			t.Errorf("expected limit 5, got %v", body["limit"])
		}

		// Verify allocated_storage at top level
		if body["allocated_storage"] != float64(2.5) {
			t.Errorf("expected allocated_storage 2.5, got %v", body["allocated_storage"])
		}

		// Verify order at top level
		order, ok := body["order"].([]interface{})
		if !ok {
			t.Fatal("expected order field to be an array")
		}
		orderPair, ok := order[0].([]interface{})
		if !ok {
			t.Fatal("expected order pair to be an array")
		}
		if orderPair[0] != "score" {
			t.Errorf("expected order by score, got %v", orderPair[0])
		}
		if orderPair[1] != "desc" {
			t.Errorf("expected order direction desc, got %v", orderPair[1])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"offers": []map[string]interface{}{
				{
					"id":             501,
					"cuda_max_good":  12.1,
					"cpu_ghz":        3.5,
					"disk_bw":        500.0,
					"disk_space":     200.0,
					"disk_name":      "nvme0n1",
					"storage_cost":   0.30,
					"driver_version": "535.86",
					"inet_up":        1000.0,
					"inet_down":      2000.0,
					"reliability":    0.98,
					"duration":       86400.0,
					"machine_id":     789,
					"verification":   "verified",
					"host_id":        101,
					"geolocation":    "US",
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")

	diskSpace := 100.0
	storageCost := 0.5
	inetUp := 500.0
	offers, err := c.Volumes.SearchOffers(context.Background(), &VolumeOfferSearchParams{
		DiskSpace:        &diskSpace,
		StorageCost:      &storageCost,
		InetUp:           &inetUp,
		Geolocation:      "US",
		Limit:            5,
		AllocatedStorage: 2.5,
	})
	if err != nil {
		t.Fatalf("SearchOffers returned error: %v", err)
	}
	if len(offers) != 1 {
		t.Fatalf("expected 1 offer, got %d", len(offers))
	}
	o := offers[0]
	if o.ID != 501 {
		t.Errorf("expected offer ID 501, got %d", o.ID)
	}
	if o.DiskSpace != 200.0 {
		t.Errorf("expected DiskSpace 200.0, got %f", o.DiskSpace)
	}
	if o.StorageCost != 0.30 {
		t.Errorf("expected StorageCost 0.30, got %f", o.StorageCost)
	}
	if o.InetUp != 1000.0 {
		t.Errorf("expected InetUp 1000.0, got %f", o.InetUp)
	}
	if o.Reliability != 0.98 {
		t.Errorf("expected Reliability 0.98, got %f", o.Reliability)
	}
	if o.Geolocation != "US" {
		t.Errorf("expected Geolocation US, got %q", o.Geolocation)
	}
	if o.MachineID != 789 {
		t.Errorf("expected MachineID 789, got %d", o.MachineID)
	}
}

// ---------------------------------------------------------------------------
// TestVolumeService_SearchOffers_RawQuery
// ---------------------------------------------------------------------------

func TestVolumeService_SearchOffers_RawQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// When raw_query is set, q should be the raw string, not a structured map
		q, ok := body["q"].(string)
		if !ok {
			t.Fatalf("expected q to be a string (raw query), got %T", body["q"])
		}
		if q != `{"disk_space": {"gte": 500}}` {
			t.Errorf("expected raw query passthrough, got %q", q)
		}

		// Verify allocated_storage is still present with default
		if body["allocated_storage"] != float64(1) {
			t.Errorf("expected allocated_storage 1.0 (default), got %v", body["allocated_storage"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"offers": []map[string]interface{}{},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	offers, err := c.Volumes.SearchOffers(context.Background(), &VolumeOfferSearchParams{
		RawQuery: `{"disk_space": {"gte": 500}}`,
	})
	if err != nil {
		t.Fatalf("SearchOffers returned error: %v", err)
	}
	if len(offers) != 0 {
		t.Errorf("expected 0 offers, got %d", len(offers))
	}
}

// ---------------------------------------------------------------------------
// TestVolumeService_SearchOffers_Defaults
// ---------------------------------------------------------------------------

func TestVolumeService_SearchOffers_Defaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// Verify defaults: limit=10 at top level
		if body["limit"] != float64(10) {
			t.Errorf("expected default limit 10, got %v", body["limit"])
		}

		// Verify default orderBy="storage_cost" at top level
		order, ok := body["order"].([]interface{})
		if !ok {
			t.Fatal("expected order field to be an array")
		}
		orderPair, ok := order[0].([]interface{})
		if !ok {
			t.Fatal("expected order pair to be an array")
		}
		if orderPair[0] != "score" {
			t.Errorf("expected default order by score, got %v", orderPair[0])
		}
		if orderPair[1] != "desc" {
			t.Errorf("expected default order direction desc, got %v", orderPair[1])
		}

		// Verify default allocated_storage=1.0 at top level
		if body["allocated_storage"] != float64(1) {
			t.Errorf("expected default allocated_storage 1.0, got %v", body["allocated_storage"])
		}

		// Verify default filters are applied (W-7)
		verified, ok := body["verified"].(map[string]interface{})
		if !ok {
			t.Fatal("expected default verified filter")
		}
		if verified["eq"] != true {
			t.Error("expected verified eq true")
		}
		external, ok := body["external"].(map[string]interface{})
		if !ok {
			t.Fatal("expected default external filter")
		}
		if external["eq"] != false {
			t.Error("expected external eq false")
		}
		diskSpace, ok := body["disk_space"].(map[string]interface{})
		if !ok {
			t.Fatal("expected default disk_space filter")
		}
		if diskSpace["gte"] != float64(1) {
			t.Errorf("expected disk_space gte 1, got %v", diskSpace["gte"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"offers": []map[string]interface{}{},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-api-key", server.URL, "test")
	_, err := c.Volumes.SearchOffers(context.Background(), &VolumeOfferSearchParams{})
	if err != nil {
		t.Fatalf("SearchOffers returned error: %v", err)
	}
}
