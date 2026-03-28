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
// TestOverlayService_Create
// ---------------------------------------------------------------------------

func TestOverlayService_Create(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// First call: POST /api/v0/overlay/ (create)
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/overlay/" {
				t.Errorf("expected path /api/v0/overlay/, got %s", r.URL.Path)
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body["cluster_id"] != float64(42) {
				t.Errorf("expected cluster_id=42, got %v", body["cluster_id"])
			}
			if body["name"] != "test-overlay" {
				t.Errorf("expected name=test-overlay, got %v", body["name"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"msg": "overlay created",
			})
			return
		}

		// Second call: GET /api/v0/overlay/ (list for create-then-read)
		if r.Method != http.MethodGet {
			t.Errorf("expected GET for list, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/overlay/" {
			t.Errorf("expected path /api/v0/overlay/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"overlay_id":      1,
				"name":            "existing-overlay",
				"internal_subnet": "10.0.0.0/24",
				"cluster_id":      42,
				"instances":       []int{},
			},
			{
				"overlay_id":      2,
				"name":            "test-overlay",
				"internal_subnet": "10.0.1.0/24",
				"cluster_id":      42,
				"instances":       []int{},
			},
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	overlay, err := c.Overlays.Create(context.Background(), 42, "test-overlay")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if overlay.OverlayID != 2 {
		t.Errorf("expected OverlayID 2, got %d", overlay.OverlayID)
	}
	if overlay.Name != "test-overlay" {
		t.Errorf("expected name test-overlay, got %q", overlay.Name)
	}
	if overlay.ClusterID != 42 {
		t.Errorf("expected ClusterID 42, got %d", overlay.ClusterID)
	}
}

// ---------------------------------------------------------------------------
// TestOverlayService_Create_Error
// ---------------------------------------------------------------------------

func TestOverlayService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid cluster"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Overlays.Create(context.Background(), 999, "test-overlay")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestOverlayService_List
// ---------------------------------------------------------------------------

func TestOverlayService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/overlay/" {
			t.Errorf("expected path /api/v0/overlay/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"overlay_id":      1,
				"name":            "overlay-a",
				"internal_subnet": "10.0.0.0/24",
				"cluster_id":      10,
				"instances":       []int{100, 200},
			},
			{
				"overlay_id":      2,
				"name":            "overlay-b",
				"internal_subnet": "10.0.1.0/24",
				"cluster_id":      20,
				"instances":       []int{},
			},
		})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	overlays, err := c.Overlays.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(overlays) != 2 {
		t.Fatalf("expected 2 overlays, got %d", len(overlays))
	}
	if overlays[0].OverlayID != 1 {
		t.Errorf("expected first overlay ID 1, got %d", overlays[0].OverlayID)
	}
	if overlays[0].Name != "overlay-a" {
		t.Errorf("expected first overlay name overlay-a, got %q", overlays[0].Name)
	}
	if len(overlays[0].Instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(overlays[0].Instances))
	}
}

// ---------------------------------------------------------------------------
// TestOverlayService_List_Error
// ---------------------------------------------------------------------------

func TestOverlayService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Overlays.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestOverlayService_Delete
// ---------------------------------------------------------------------------

func TestOverlayService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/overlay/" {
			t.Errorf("expected path /api/v0/overlay/, got %s", r.URL.Path)
		}

		bodyBytes, _ := io.ReadAll(r.Body)
		var body map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["overlay_id"] != float64(2) {
			t.Errorf("expected overlay_id=2, got %v", body["overlay_id"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Overlays.Delete(context.Background(), 2)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestOverlayService_Delete_Error
// ---------------------------------------------------------------------------

func TestOverlayService_Delete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "overlay not found"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Overlays.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestOverlayService_JoinInstance
// ---------------------------------------------------------------------------

func TestOverlayService_JoinInstance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/overlay/" {
			t.Errorf("expected path /api/v0/overlay/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["name"] != "test-overlay" {
			t.Errorf("expected name=test-overlay, got %v", body["name"])
		}
		if body["instance_id"] != float64(500) {
			t.Errorf("expected instance_id=500, got %v", body["instance_id"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Overlays.JoinInstance(context.Background(), "test-overlay", 500)
	if err != nil {
		t.Fatalf("JoinInstance returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestOverlayService_JoinInstance_Error
// ---------------------------------------------------------------------------

func TestOverlayService_JoinInstance_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid instance"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Overlays.JoinInstance(context.Background(), "test-overlay", 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
