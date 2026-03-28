package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// TestInvoiceService_List
// ---------------------------------------------------------------------------

func TestInvoiceService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify the path is /api/v1/invoices/ NOT /api/v0/api/v1/invoices/
		// This confirms GetFullPath works correctly (no double prefix)
		if r.URL.Path != "/api/v1/invoices/" {
			t.Errorf("expected path /api/v1/invoices/, got %s", r.URL.Path)
		}

		// Verify auth header is still set
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"id":          1,
					"amount":      -5.25,
					"type":        "charge",
					"description": "GPU rental - RTX 4090",
					"timestamp":   "2025-01-15T10:30:00Z",
				},
				{
					"id":          2,
					"amount":      100.00,
					"type":        "credit",
					"description": "Account top-up",
					"timestamp":   "2025-01-14T08:00:00Z",
				},
			},
			"count":      2,
			"total":      50,
			"next_token": "abc123",
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	resp, err := c.Invoices.List(context.Background(), InvoiceListParams{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("expected count 2, got %d", resp.Count)
	}
	if resp.Total != 50 {
		t.Errorf("expected total 50, got %d", resp.Total)
	}
	if resp.NextToken != "abc123" {
		t.Errorf("expected next_token abc123, got %q", resp.NextToken)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Results))
	}
	if resp.Results[0].ID != 1 {
		t.Errorf("expected first invoice ID 1, got %d", resp.Results[0].ID)
	}
	if resp.Results[0].Amount != -5.25 {
		t.Errorf("expected amount -5.25, got %f", resp.Results[0].Amount)
	}
	if resp.Results[0].Type != "charge" {
		t.Errorf("expected type charge, got %q", resp.Results[0].Type)
	}
}

// ---------------------------------------------------------------------------
// TestInvoiceService_List_WithParams
// ---------------------------------------------------------------------------

func TestInvoiceService_List_WithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/invoices/" {
			t.Errorf("expected path /api/v1/invoices/, got %s", r.URL.Path)
		}

		// Verify query parameters use select_filters structure
		q := r.URL.Query()
		selectFilters := q.Get("select_filters")
		if selectFilters == "" {
			t.Fatal("expected select_filters query parameter")
		}

		// Parse the select_filters JSON
		var filters map[string]interface{}
		if err := json.Unmarshal([]byte(selectFilters), &filters); err != nil {
			t.Fatalf("failed to parse select_filters JSON: %v", err)
		}

		// Verify the when filter has gte and lte
		when, ok := filters["when"].(map[string]interface{})
		if !ok {
			t.Fatal("expected 'when' field in select_filters")
		}
		if when["gte"] != 1735689600.0 {
			t.Errorf("expected when.gte=1735689600, got %v", when["gte"])
		}
		if when["lte"] != 1738368000.0 {
			t.Errorf("expected when.lte=1738368000, got %v", when["lte"])
		}

		// Verify limit
		if q.Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", q.Get("limit"))
		}

		// Verify latest_first
		if q.Get("latest_first") != "true" {
			t.Errorf("expected latest_first=true, got %q", q.Get("latest_first"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{},
			"count":   0,
			"total":   0,
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	resp, err := c.Invoices.List(context.Background(), InvoiceListParams{
		StartDate:   1735689600, // 2025-01-01 00:00:00 UTC
		EndDate:     1738368000, // 2025-01-31 00:00:00 UTC
		Limit:       10,
		LatestFirst: true,
	})
	if err != nil {
		t.Fatalf("List with params returned error: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
}

// ---------------------------------------------------------------------------
// TestInvoiceService_List_WithAfterToken
// ---------------------------------------------------------------------------

func TestInvoiceService_List_WithAfterToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/invoices/" {
			t.Errorf("expected path /api/v1/invoices/, got %s", r.URL.Path)
		}

		q := r.URL.Query()
		if q.Get("after_token") != "page2token" {
			t.Errorf("expected after_token=page2token, got %q", q.Get("after_token"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{},
			"count":   0,
			"total":   0,
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Invoices.List(context.Background(), InvoiceListParams{
		AfterToken: "page2token",
	})
	if err != nil {
		t.Fatalf("List with after_token returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestInvoiceService_List_Error
// ---------------------------------------------------------------------------

func TestInvoiceService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("bad-key", server.URL, "test")
	_, err := c.Invoices.List(context.Background(), InvoiceListParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
