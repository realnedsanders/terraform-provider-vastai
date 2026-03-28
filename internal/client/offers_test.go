package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestOfferService_Search
// ---------------------------------------------------------------------------

func TestOfferService_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/bundles/" {
			t.Errorf("expected path /api/v0/bundles/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// Flat body -- filters are top-level (no q wrapper)
		// Verify gpu_ram conversion: 24 GB * 1000 = 24000 MB
		gpuRam, ok := body["gpu_ram"].(map[string]interface{})
		if !ok {
			t.Fatal("expected gpu_ram filter in body")
		}
		gpuRamGte, ok := gpuRam["gte"].(float64)
		if !ok {
			t.Fatal("expected gpu_ram gte to be a number")
		}
		if gpuRamGte != 24000 {
			t.Errorf("expected gpu_ram gte 24000 (24 GB * 1000), got %f", gpuRamGte)
		}

		// Verify gpu_name filter
		gpuName, ok := body["gpu_name"].(map[string]interface{})
		if !ok {
			t.Fatal("expected gpu_name filter in body")
		}
		if gpuName["eq"] != "RTX 4090" {
			t.Errorf("expected gpu_name eq %q, got %v", "RTX 4090", gpuName["eq"])
		}

		// Verify num_gpus filter
		numGpus, ok := body["num_gpus"].(map[string]interface{})
		if !ok {
			t.Fatal("expected num_gpus filter in body")
		}
		if numGpus["eq"] != float64(2) {
			t.Errorf("expected num_gpus eq 2, got %v", numGpus["eq"])
		}

		// Verify limit is in the flat body
		if body["limit"] != float64(5) {
			t.Errorf("expected limit 5, got %v", body["limit"])
		}

		// Verify allocated_storage is present (default 5.0)
		if body["allocated_storage"] != float64(5) {
			t.Errorf("expected allocated_storage 5.0, got %v", body["allocated_storage"])
		}

		// Verify base filters at top level
		verified, ok := body["verified"].(map[string]interface{})
		if !ok {
			t.Fatal("expected verified filter in body")
		}
		if verified["eq"] != true {
			t.Error("expected verified eq true")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"offers": []map[string]interface{}{
				{
					"id":        1001,
					"gpu_name":  "RTX 4090",
					"num_gpus":  2,
					"gpu_ram":   24000,
					"dph_total": 0.50,
				},
				{
					"id":        1002,
					"gpu_name":  "RTX 4090",
					"num_gpus":  2,
					"gpu_ram":   24000,
					"dph_total": 0.65,
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")

	gpuRam := 24.0
	numGPUs := 2
	offers, err := c.Offers.Search(context.Background(), &OfferSearchParams{
		GPUName:  "RTX 4090",
		NumGPUs:  &numGPUs,
		GPURamGB: &gpuRam,
		Limit:    5,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(offers) != 2 {
		t.Fatalf("expected 2 offers, got %d", len(offers))
	}
	if offers[0].ID != 1001 {
		t.Errorf("expected first offer ID 1001, got %d", offers[0].ID)
	}
	if offers[0].DPHTotal != 0.50 {
		t.Errorf("expected first offer dph_total 0.50, got %f", offers[0].DPHTotal)
	}
}

// ---------------------------------------------------------------------------
// TestOfferService_Search_RawQuery
// ---------------------------------------------------------------------------

func TestOfferService_Search_RawQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/bundles/" {
			t.Errorf("expected path /api/v0/bundles/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// When raw_query is set, q should be the raw string, not a structured map
		q, ok := body["q"].(string)
		if !ok {
			t.Fatalf("expected q to be a string (raw query), got %T", body["q"])
		}
		if q != `{"gpu_name": {"eq": "A100"}}` {
			t.Errorf("expected raw query passthrough, got %q", q)
		}

		// Verify allocated_storage is present (default 5.0)
		if body["allocated_storage"] != float64(5) {
			t.Errorf("expected allocated_storage 5.0, got %v", body["allocated_storage"])
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

	c := NewVastAIClient("test-key", server.URL, "test")
	offers, err := c.Offers.Search(context.Background(), &OfferSearchParams{
		RawQuery: `{"gpu_name": {"eq": "A100"}}`,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(offers) != 0 {
		t.Errorf("expected 0 offers, got %d", len(offers))
	}
}

// ---------------------------------------------------------------------------
// TestOfferService_Search_Defaults
// ---------------------------------------------------------------------------

func TestOfferService_Search_Defaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/bundles/" {
			t.Errorf("expected path /api/v0/bundles/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// Verify defaults: limit=10 at top level (flat body)
		if body["limit"] != float64(10) {
			t.Errorf("expected default limit 10, got %v", body["limit"])
		}

		// Verify default order_by at top level (flat body)
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

		// Verify allocated_storage default
		if body["allocated_storage"] != float64(5) {
			t.Errorf("expected default allocated_storage 5.0, got %v", body["allocated_storage"])
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

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Offers.Search(context.Background(), &OfferSearchParams{})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestOfferService_Search_DatacenterOnly
// ---------------------------------------------------------------------------

func TestOfferService_Search_DatacenterOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/bundles/" {
			t.Errorf("expected path /api/v0/bundles/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// Flat body -- hosting_type is at top level
		hostingType, ok := body["hosting_type"].(map[string]interface{})
		if !ok {
			t.Fatal("expected hosting_type filter for datacenter-only")
		}
		if hostingType["eq"] != float64(1) {
			t.Errorf("expected hosting_type eq 1 for datacenter, got %v", hostingType["eq"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"offers": []interface{}{}}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	datacenterOnly := true
	_, err := c.Offers.Search(context.Background(), &OfferSearchParams{
		DatacenterOnly: &datacenterOnly,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
}
