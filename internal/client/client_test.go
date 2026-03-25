package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// TestNewVastAIClient
// ---------------------------------------------------------------------------

func TestNewVastAIClient(t *testing.T) {
	c := NewVastAIClient("test-key", "https://console.vast.ai/", "1.0.0")

	// baseURL trailing slash should be stripped
	if c.baseURL != "https://console.vast.ai" {
		t.Errorf("expected baseURL without trailing slash, got %q", c.baseURL)
	}

	// apiKey should be stored
	if c.apiKey != "test-key" {
		t.Errorf("expected apiKey %q, got %q", "test-key", c.apiKey)
	}

	// userAgent should contain provider version
	expected := "terraform-provider-vastai/1.0.0"
	if c.userAgent != expected {
		t.Errorf("expected userAgent %q, got %q", expected, c.userAgent)
	}
}

// ---------------------------------------------------------------------------
// TestNewRequest_AuthHeader
// ---------------------------------------------------------------------------

func TestNewRequest_AuthHeader(t *testing.T) {
	c := NewVastAIClient("secret-api-key", "https://console.vast.ai", "test")
	req, err := c.newRequest(context.Background(), http.MethodGet, "/instances", nil)
	if err != nil {
		t.Fatalf("newRequest returned error: %v", err)
	}
	if req == nil {
		t.Fatal("newRequest returned nil request")
	}

	auth := req.Header.Get("Authorization")
	if auth != "Bearer secret-api-key" {
		t.Errorf("expected Authorization header %q, got %q", "Bearer secret-api-key", auth)
	}

	// API key must never appear in URL
	if strings.Contains(req.URL.String(), "secret-api-key") {
		t.Error("API key found in URL -- must only be in Authorization header")
	}
	if strings.Contains(req.URL.String(), "api_key") {
		t.Error("api_key query parameter found in URL -- must only be in Authorization header")
	}
}

// ---------------------------------------------------------------------------
// TestNewRequest_UserAgent
// ---------------------------------------------------------------------------

func TestNewRequest_UserAgent(t *testing.T) {
	c := NewVastAIClient("key", "https://console.vast.ai", "2.5.0")
	req, err := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("newRequest returned error: %v", err)
	}
	if req == nil {
		t.Fatal("newRequest returned nil request")
	}

	ua := req.Header.Get("User-Agent")
	expected := "terraform-provider-vastai/2.5.0"
	if ua != expected {
		t.Errorf("expected User-Agent %q, got %q", expected, ua)
	}
}

// ---------------------------------------------------------------------------
// TestNewRequest_Headers
// ---------------------------------------------------------------------------

func TestNewRequest_Headers(t *testing.T) {
	c := NewVastAIClient("key", "https://console.vast.ai", "test")
	req, err := c.newRequest(context.Background(), http.MethodPost, "/test", nil)
	if err != nil {
		t.Fatalf("newRequest returned error: %v", err)
	}
	if req == nil {
		t.Fatal("newRequest returned nil request")
	}

	ct := req.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}

	accept := req.Header.Get("Accept")
	if accept != "application/json" {
		t.Errorf("expected Accept %q, got %q", "application/json", accept)
	}
}

// ---------------------------------------------------------------------------
// TestNewRequest_URLConstruction
// ---------------------------------------------------------------------------

func TestNewRequest_URLConstruction(t *testing.T) {
	c := NewVastAIClient("key", "https://console.vast.ai", "test")
	req, err := c.newRequest(context.Background(), http.MethodGet, "/instances", nil)
	if err != nil {
		t.Fatalf("newRequest returned error: %v", err)
	}
	if req == nil {
		t.Fatal("newRequest returned nil request")
	}

	expected := "https://console.vast.ai/api/v0/instances"
	if req.URL.String() != expected {
		t.Errorf("expected URL %q, got %q", expected, req.URL.String())
	}
}

// ---------------------------------------------------------------------------
// TestNewRequest_WithBody
// ---------------------------------------------------------------------------

func TestNewRequest_WithBody(t *testing.T) {
	c := NewVastAIClient("key", "https://console.vast.ai", "test")

	body := map[string]interface{}{
		"name": "test-instance",
		"gpu":  2,
	}
	req, err := c.newRequest(context.Background(), http.MethodPost, "/instances", body)
	if err != nil {
		t.Fatalf("newRequest returned error: %v", err)
	}
	if req == nil {
		t.Fatal("newRequest returned nil request")
	}

	// Read the body from the request
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}

	if parsed["name"] != "test-instance" {
		t.Errorf("expected body name %q, got %v", "test-instance", parsed["name"])
	}
	if parsed["gpu"] != float64(2) {
		t.Errorf("expected body gpu %v, got %v", 2, parsed["gpu"])
	}
}

// ---------------------------------------------------------------------------
// TestRetryPolicy_RateLimited
// ---------------------------------------------------------------------------

func TestRetryPolicy_RateLimited(t *testing.T) {
	resp := &http.Response{StatusCode: 429}
	retry, err := vastaiRetryPolicy(context.Background(), resp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !retry {
		t.Error("expected retry on 429, got no retry")
	}
}

// ---------------------------------------------------------------------------
// TestRetryPolicy_ServerErrors
// ---------------------------------------------------------------------------

func TestRetryPolicy_ServerErrors(t *testing.T) {
	for _, code := range []int{500, 502, 503, 504} {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			resp := &http.Response{StatusCode: code}
			retry, err := vastaiRetryPolicy(context.Background(), resp, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !retry {
				t.Errorf("expected retry on %d, got no retry", code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestRetryPolicy_NotImplemented
// ---------------------------------------------------------------------------

func TestRetryPolicy_NotImplemented(t *testing.T) {
	resp := &http.Response{StatusCode: 501}
	retry, err := vastaiRetryPolicy(context.Background(), resp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retry {
		t.Error("expected no retry on 501, got retry")
	}
}

// ---------------------------------------------------------------------------
// TestRetryPolicy_ClientErrors
// ---------------------------------------------------------------------------

func TestRetryPolicy_ClientErrors(t *testing.T) {
	for _, code := range []int{400, 401, 403, 404} {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			resp := &http.Response{StatusCode: code}
			retry, err := vastaiRetryPolicy(context.Background(), resp, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if retry {
				t.Errorf("expected no retry on %d, got retry", code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestRetryPolicy_ContextCanceled
// ---------------------------------------------------------------------------

func TestRetryPolicy_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	retry, err := vastaiRetryPolicy(ctx, nil, nil)
	if retry {
		t.Error("expected no retry on canceled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestRetryPolicy_ConnectionError
// ---------------------------------------------------------------------------

func TestRetryPolicy_ConnectionError(t *testing.T) {
	connErr := fmt.Errorf("connection refused")
	retry, err := vastaiRetryPolicy(context.Background(), nil, connErr)
	if err != nil {
		t.Fatalf("unexpected error from retry policy: %v", err)
	}
	if !retry {
		t.Error("expected retry on connection error, got no retry")
	}
}

// ---------------------------------------------------------------------------
// TestBackoff_ExponentialGrowth
// ---------------------------------------------------------------------------

func TestBackoff_ExponentialGrowth(t *testing.T) {
	minWait := 150 * time.Millisecond
	maxWait := 30 * time.Second

	for attempt := 0; attempt < 5; attempt++ {
		got := vastaiBackoff(minWait, maxWait, attempt, nil)
		expected := time.Duration(float64(minWait) * math.Pow(1.5, float64(attempt)))

		// Allow 1ms tolerance for floating point
		diff := got - expected
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Millisecond {
			t.Errorf("attempt %d: expected %v, got %v (diff: %v)", attempt, expected, got, diff)
		}
	}
}

// ---------------------------------------------------------------------------
// TestBackoff_RetryAfterHeader
// ---------------------------------------------------------------------------

func TestBackoff_RetryAfterHeader(t *testing.T) {
	resp := &http.Response{
		StatusCode: 429,
		Header:     http.Header{},
	}
	resp.Header.Set("Retry-After", "2.5")

	got := vastaiBackoff(150*time.Millisecond, 30*time.Second, 0, resp)
	expected := 2500 * time.Millisecond

	// Allow 10ms tolerance
	diff := got - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > 10*time.Millisecond {
		t.Errorf("expected ~%v from Retry-After, got %v", expected, got)
	}
}

// ---------------------------------------------------------------------------
// TestBackoff_MaxCap
// ---------------------------------------------------------------------------

func TestBackoff_MaxCap(t *testing.T) {
	minWait := 150 * time.Millisecond
	maxWait := 30 * time.Second

	// At very high attempt numbers, backoff should not exceed max
	got := vastaiBackoff(minWait, maxWait, 100, nil)
	if got > maxWait {
		t.Errorf("expected backoff capped at %v, got %v", maxWait, got)
	}
	if got != maxWait {
		t.Errorf("expected backoff to equal max %v at high attempt, got %v", maxWait, got)
	}
}

// ---------------------------------------------------------------------------
// TestAPIError_ErrorString
// ---------------------------------------------------------------------------

func TestAPIError_ErrorString(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Message:    "instance not found",
		Method:     "GET",
		Path:       "/api/v0/instances/12345",
	}

	expected := "Vast.ai API error: GET /api/v0/instances/12345 returned 404: instance not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

// ---------------------------------------------------------------------------
// TestGet_Success (integration with httptest)
// ---------------------------------------------------------------------------

func TestGet_Success(t *testing.T) {
	type testResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}
		// Verify User-Agent
		if r.Header.Get("User-Agent") != "terraform-provider-vastai/test" {
			t.Errorf("expected User-Agent %q, got %q", "terraform-provider-vastai/test", r.Header.Get("User-Agent"))
		}
		// Verify URL path
		if r.URL.Path != "/api/v0/instances/1" {
			t.Errorf("expected path /api/v0/instances/1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testResponse{ID: 1, Name: "gpu-box"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")

	var result testResponse
	err := c.Get(context.Background(), "/instances/1", &result)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Name != "gpu-box" {
		t.Errorf("expected Name %q, got %q", "gpu-box", result.Name)
	}
}

// ---------------------------------------------------------------------------
// TestPost_Success (integration with httptest)
// ---------------------------------------------------------------------------

func TestPost_Success(t *testing.T) {
	type postBody struct {
		Name string `json:"name"`
	}
	type postResponse struct {
		ID int `json:"id"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer auth, got %q", r.Header.Get("Authorization"))
		}

		// Verify request body
		var body postBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Name != "new-instance" {
			t.Errorf("expected body name %q, got %q", "new-instance", body.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(postResponse{ID: 42})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")

	var result postResponse
	err := c.Post(context.Background(), "/instances", postBody{Name: "new-instance"}, &result)
	if err != nil {
		t.Fatalf("Post returned error: %v", err)
	}

	if result.ID != 42 {
		t.Errorf("expected ID 42, got %d", result.ID)
	}
}

// ---------------------------------------------------------------------------
// TestGet_APIError
// ---------------------------------------------------------------------------

func TestGet_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "instance not found"})
	}))
	defer server.Close()

	c := NewVastAIClient("test-key", server.URL, "test")

	var result map[string]interface{}
	err := c.Get(context.Background(), "/instances/999", &result)
	if err == nil {
		t.Fatal("expected error on 404 response, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if apiErr.Method != "GET" {
		t.Errorf("expected method GET, got %s", apiErr.Method)
	}
	if !strings.Contains(apiErr.Path, "/instances/999") {
		t.Errorf("expected path containing /instances/999, got %s", apiErr.Path)
	}
	if apiErr.Message != "instance not found" {
		t.Errorf("expected message %q, got %q", "instance not found", apiErr.Message)
	}
}

// ---------------------------------------------------------------------------
// TestAPIKeyNotInURL
// ---------------------------------------------------------------------------

func TestAPIKeyNotInURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the API key never appears in the URL
		if strings.Contains(r.URL.String(), "my-secret-key") {
			t.Error("API key found in request URL")
		}
		if strings.Contains(r.URL.RawQuery, "api_key") {
			t.Error("api_key query parameter found in request URL")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	c := NewVastAIClient("my-secret-key", server.URL, "test")

	var result map[string]interface{}
	err := c.Get(context.Background(), "/instances", &result)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
}
