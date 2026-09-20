package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// TestInstanceService_Create
// ---------------------------------------------------------------------------

func TestInstanceService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/asks/12345/" {
			t.Errorf("expected path /api/v0/asks/12345/, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		var body CreateInstanceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Image != "pytorch/pytorch:latest" {
			t.Errorf("expected image %q, got %q", "pytorch/pytorch:latest", body.Image)
		}
		if body.Disk == nil || *body.Disk != 20.0 {
			t.Errorf("expected disk 20.0, got %v", body.Disk)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]int{"new_contract": 7835610}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	disk := 20.0
	resp, err := c.Instances.Create(context.Background(), 12345, &CreateInstanceRequest{
		ClientID: "me",
		Image:    "pytorch/pytorch:latest",
		Disk:     &disk,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if resp.NewContract != 7835610 {
		t.Errorf("expected NewContract 7835610, got %d", resp.NewContract)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_Create_BadRequestUsesHumanMessage
// ---------------------------------------------------------------------------

func TestInstanceService_Create_BadRequestUsesHumanMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"success":false,"error":"invalid_args","msg":"invalid env arguments, total length > 32KB"}`)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	_, err := c.Instances.Create(context.Background(), 28705908, &CreateInstanceRequest{})
	if err == nil {
		t.Fatal("expected Create to return an error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected wrapped *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d; want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
	if apiErr.Message != "invalid env arguments, total length > 32KB" {
		t.Errorf("Message = %q; want human validation message", apiErr.Message)
	}
	if apiErr.Code != "invalid_args" {
		t.Errorf("APIError.Code = %q; want %q", apiErr.Code, "invalid_args")
	}
	if !strings.Contains(err.Error(), "invalid env arguments, total length > 32KB") {
		t.Errorf("wrapped error %q does not contain human validation message", err)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_Get
// ---------------------------------------------------------------------------

func TestInstanceService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/" {
			t.Errorf("expected path /api/v0/instances/42/, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("owner") != "me" {
			t.Errorf("expected owner=me query param, got %q", r.URL.Query().Get("owner"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// API returns {"instances": {single object}} for single instance get
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"instances": map[string]interface{}{
				"id":            42,
				"machine_id":    100,
				"actual_status": "running",
				"gpu_name":      "RTX 4090",
				"num_gpus":      2,
				"label":         "my-gpu",
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	instance, err := c.Instances.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if instance.ID != 42 {
		t.Errorf("expected ID 42, got %d", instance.ID)
	}
	if instance.MachineID != 100 {
		t.Errorf("expected MachineID 100, got %d", instance.MachineID)
	}
	if instance.ActualStatus != "running" {
		t.Errorf("expected ActualStatus %q, got %q", "running", instance.ActualStatus)
	}
	if instance.GPUName != "RTX 4090" {
		t.Errorf("expected GPUName %q, got %q", "RTX 4090", instance.GPUName)
	}
	if instance.NumGPUs != 2 {
		t.Errorf("expected NumGPUs 2, got %d", instance.NumGPUs)
	}
	if instance.Label != "my-gpu" {
		t.Errorf("expected Label %q, got %q", "my-gpu", instance.Label)
	}
}

func TestInstanceService_Get_ResponseShapes(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantNotFound bool
		wantErr      bool
		wantID       int
	}{
		{
			name:         "null instance is absent",
			body:         `{"instances":null}`,
			wantNotFound: true,
			wantErr:      true,
		},
		{
			name:   "populated instance with null actual status is present",
			body:   `{"instances":{"id":42,"actual_status":null}}`,
			wantID: 42,
		},
		{
			name:   "populated instance with empty actual status is present",
			body:   `{"instances":{"id":42,"actual_status":""}}`,
			wantID: 42,
		},
		{
			name:    "missing instances field is malformed",
			body:    `{}`,
			wantErr: true,
		},
		{
			name:    "empty response is malformed",
			body:    ``,
			wantErr: true,
		},
		{
			name:    "top-level null is malformed",
			body:    `null`,
			wantErr: true,
		},
		{
			name:    "instance object without id is malformed",
			body:    `{"instances":{}}`,
			wantErr: true,
		},
		{
			name:    "instance array is malformed",
			body:    `{"instances":[]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c := NewVastAIClient("test-key", server.URL, "test")
			instance, err := c.Instances.Get(context.Background(), 42)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := errors.Is(err, ErrInstanceNotFound); got != tt.wantNotFound {
				t.Fatalf("errors.Is(err, ErrInstanceNotFound) = %v, want %v (err: %v)", got, tt.wantNotFound, err)
			}
			if tt.wantErr {
				if instance != nil {
					t.Fatalf("Get() instance = %#v, want nil on error", instance)
				}
				return
			}
			if instance == nil {
				t.Fatal("Get() returned nil instance")
			}
			if instance.ID != tt.wantID {
				t.Errorf("Get() instance ID = %d, want %d", instance.ID, tt.wantID)
			}
			if instance.ActualStatus != "" {
				t.Errorf("Get() actual_status = %q, want empty", instance.ActualStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_List
// ---------------------------------------------------------------------------

func TestInstanceService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"instances": []map[string]interface{}{
				{"id": 1, "actual_status": "running"},
				{"id": 2, "actual_status": "stopped"},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	instances, err := c.Instances.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].ID != 1 {
		t.Errorf("expected first instance ID 1, got %d", instances[0].ID)
	}
	if instances[1].ActualStatus != "stopped" {
		t.Errorf("expected second instance status %q, got %q", "stopped", instances[1].ActualStatus)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_Start
// ---------------------------------------------------------------------------

func TestInstanceService_Start(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/" {
			t.Errorf("expected path /api/v0/instances/42/, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["state"] != "running" {
			t.Errorf("expected state %q, got %q", "running", body["state"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Instances.Start(context.Background(), 42)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_Stop
// ---------------------------------------------------------------------------

func TestInstanceService_Stop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/" {
			t.Errorf("expected path /api/v0/instances/42/, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["state"] != "stopped" {
			t.Errorf("expected state %q, got %q", "stopped", body["state"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Instances.Stop(context.Background(), 42)
	if err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_Destroy
// ---------------------------------------------------------------------------

func TestInstanceService_Destroy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/" {
			t.Errorf("expected path /api/v0/instances/42/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Instances.Destroy(context.Background(), 42)
	if err != nil {
		t.Fatalf("Destroy returned error: %v", err)
	}
}

func TestInstanceService_Destroy_NotFoundIsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not_found","msg":"Instance not found"}`))
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	if err := c.Instances.Destroy(context.Background(), 42); err != nil {
		t.Fatalf("Destroy returned error for absent instance: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_SetLabel
// ---------------------------------------------------------------------------

func TestInstanceService_SetLabel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/" {
			t.Errorf("expected path /api/v0/instances/42/, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["label"] != "test-label" {
			t.Errorf("expected label %q, got %q", "test-label", body["label"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Instances.SetLabel(context.Background(), 42, "test-label")
	if err != nil {
		t.Fatalf("SetLabel returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_ChangeBid
// ---------------------------------------------------------------------------

func TestInstanceService_ChangeBid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/bid_price/42/" {
			t.Errorf("expected path /api/v0/instances/bid_price/42/, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["client_id"] != "me" {
			t.Errorf("expected client_id %q, got %v", "me", body["client_id"])
		}
		if body["price"] != 0.35 {
			t.Errorf("expected price 0.35, got %v", body["price"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Instances.ChangeBid(context.Background(), 42, 0.35)
	if err != nil {
		t.Fatalf("ChangeBid returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_UpdateTemplate
// ---------------------------------------------------------------------------

func TestInstanceService_UpdateTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/update_template/42/" {
			t.Errorf("expected path /api/v0/instances/update_template/42/, got %s", r.URL.Path)
		}

		var body UpdateTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Image != "new-image:v2" {
			t.Errorf("expected image %q, got %q", "new-image:v2", body.Image)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	err := c.Instances.UpdateTemplate(context.Background(), 42, &UpdateTemplateRequest{
		Image: "new-image:v2",
	})
	if err != nil {
		t.Fatalf("UpdateTemplate returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_WaitForStatus
// ---------------------------------------------------------------------------

func TestInstanceService_WaitForStatus(t *testing.T) {
	var callCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		status := "loading"
		if count >= 2 {
			status = "running"
		}

		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"instances": map[string]interface{}{
				"id":            42,
				"actual_status": status,
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")

	// Use a short poll interval via a patched client to speed up the test
	// We override the defaultPollInterval indirectly through a short timeout
	instance, err := c.Instances.WaitForStatus(context.Background(), 42, "running", 30*time.Second)
	if err != nil {
		t.Fatalf("WaitForStatus returned error: %v", err)
	}
	if instance == nil {
		t.Fatal("expected non-nil instance")
	}
	if instance.ActualStatus != "running" {
		t.Errorf("expected actual_status %q, got %q", "running", instance.ActualStatus)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_WaitForStatus_Timeout
// ---------------------------------------------------------------------------

func TestInstanceService_WaitForStatus_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"instances": map[string]interface{}{
				"id":            42,
				"actual_status": "loading",
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")

	// Use a very short timeout to trigger timeout quickly
	_, err := c.Instances.WaitForStatus(context.Background(), 42, "running", 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// Error message should mention timeout
	if !containsString(err.Error(), "timed out") {
		t.Errorf("expected error to mention 'timed out', got: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_WaitForStatus_404OnDestroy
// ---------------------------------------------------------------------------

func TestInstanceService_WaitForStatus_404OnDestroy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "instance not found"}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")

	// When waiting for "destroyed" status, 404 should be treated as success
	instance, err := c.Instances.WaitForStatus(context.Background(), 42, "destroyed", 10*time.Second)
	if err != nil {
		t.Fatalf("WaitForStatus returned error on 404 for destroy: %v", err)
	}
	if instance != nil {
		t.Error("expected nil instance on 404 destroy, got non-nil")
	}
}

func TestInstanceService_WaitForStatus_NullInstanceDuringStartIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"instances":null}`))
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	instance, err := c.Instances.WaitForStatus(context.Background(), 42, "running", 10*time.Second)
	if err == nil {
		t.Fatal("WaitForStatus returned nil error for absent instance")
	}
	if !errors.Is(err, ErrInstanceNotFound) {
		t.Fatalf("WaitForStatus error = %v, want ErrInstanceNotFound", err)
	}
	if instance != nil {
		t.Fatalf("WaitForStatus instance = %#v, want nil", instance)
	}
}

// ---------------------------------------------------------------------------
// TestInstanceService_WaitForStatus_ExitedTerminal
// ---------------------------------------------------------------------------

func TestInstanceService_WaitForStatus_ExitedTerminal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"instances": map[string]interface{}{
				"id":            42,
				"actual_status": "exited",
				"status_msg":    "process terminated",
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")

	// When instance enters "exited" state unexpectedly while waiting for "running"
	_, err := c.Instances.WaitForStatus(context.Background(), 42, "running", 10*time.Second)
	if err == nil {
		t.Fatal("expected error for terminal exited state, got nil")
	}
	if !containsString(err.Error(), "terminal state 'exited'") {
		t.Errorf("expected error to mention terminal exited state, got: %s", err.Error())
	}
}

// containsString checks if s contains substr (helper for error message checks).
func containsString(s, substr string) bool {
	return s != "" && len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestInstanceService_GetCreationStateFields verifies that the follow-up instance
// read exposes the fields needed to resolve omitted Optional+Computed values.
func TestInstanceService_GetCreationStateFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/" {
			t.Errorf("expected path /api/v0/instances/42/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"instances": map[string]interface{}{
				"id":            42,
				"image_runtype": "ssh ssh_direct ssh_proxy",
				"extra_env": [][]string{
					{"PROJECT", "vastai"},
					{"MODE", "test"},
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	instance, err := c.Instances.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if instance.ImageRuntype != "ssh ssh_direct ssh_proxy" {
		t.Errorf("ImageRuntype = %q, want %q", instance.ImageRuntype, "ssh ssh_direct ssh_proxy")
	}
	if got := instance.ExtraEnv["PROJECT"]; got != "vastai" {
		t.Errorf("ExtraEnv[PROJECT] = %q, want %q", got, "vastai")
	}
	if got := instance.ExtraEnv["MODE"]; got != "test" {
		t.Errorf("ExtraEnv[MODE] = %q, want %q", got, "test")
	}
}

// TestInstanceService_GetSSHKeys verifies the SSH endpoint's nested JSON
// encoding is decoded into the current instance attachments.
func TestInstanceService_GetSSHKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v0/instances/42/ssh/" {
			t.Errorf("expected path /api/v0/instances/42/ssh/, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"ssh_keys": `[{"id":101,"name":"first","public_key":"ssh-ed25519 AAAA"},{"id":202,"name":"second","public_key":"ssh-rsa BBBB"}]`,
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")
	keys, err := c.Instances.GetSSHKeys(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetSSHKeys returned error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].ID != 101 || keys[1].ID != 202 {
		t.Errorf("key IDs = [%d %d], want [101 202]", keys[0].ID, keys[1].ID)
	}
	if keys[0].PublicKey != "ssh-ed25519 AAAA" {
		t.Errorf("first PublicKey = %q, want %q", keys[0].PublicKey, "ssh-ed25519 AAAA")
	}
}

// TestCreateInstanceRequest_OptionalBooleans verifies omitted values do not
// override API/template defaults while an explicitly configured false is sent.
func TestCreateInstanceRequest_OptionalBooleans(t *testing.T) {
	data, err := json.Marshal(CreateInstanceRequest{})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	var omitted map[string]interface{}
	if err := json.Unmarshal(data, &omitted); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	for _, field := range []string{"cancel_unavail", "use_jupyter_lab"} {
		if _, ok := omitted[field]; ok {
			t.Errorf("omitted request unexpectedly contains %q: %s", field, data)
		}
	}

	configuredFalse := false
	data, err = json.Marshal(CreateInstanceRequest{
		CancelUnavail: &configuredFalse,
		UseJupyterLab: &configuredFalse,
	})
	if err != nil {
		t.Fatalf("Marshal explicit false returned error: %v", err)
	}
	var explicit map[string]interface{}
	if err := json.Unmarshal(data, &explicit); err != nil {
		t.Fatalf("Unmarshal explicit false returned error: %v", err)
	}
	for _, field := range []string{"cancel_unavail", "use_jupyter_lab"} {
		value, ok := explicit[field]
		if !ok {
			t.Errorf("explicit false request omitted %q: %s", field, data)
			continue
		}
		if value != false {
			t.Errorf("%s = %#v, want false", field, value)
		}
	}
}

func TestCreateInstanceRequest_OptionalDisk(t *testing.T) {
	data, err := json.Marshal(CreateInstanceRequest{})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	var omitted map[string]interface{}
	if err := json.Unmarshal(data, &omitted); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if _, ok := omitted["disk"]; ok {
		t.Errorf("omitted request unexpectedly contains disk: %s", data)
	}

	disk := 20.0
	data, err = json.Marshal(CreateInstanceRequest{Disk: &disk})
	if err != nil {
		t.Fatalf("Marshal explicit disk returned error: %v", err)
	}
	var explicit map[string]interface{}
	if err := json.Unmarshal(data, &explicit); err != nil {
		t.Fatalf("Unmarshal explicit disk returned error: %v", err)
	}
	if got, ok := explicit["disk"]; !ok || got != 20.0 {
		t.Errorf("explicit disk = %#v (present %v), want 20", got, ok)
	}
}
