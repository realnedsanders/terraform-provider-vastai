package client

import (
	"context"
	"fmt"
)

// EndpointService handles serverless endpoint-related API operations.
type EndpointService struct {
	client *VastAIClient
}

// CreateEndpointRequest is the JSON body for POST /endptjobs/ (create endpoint).
// Pointer types for optional autoscaling fields ensure zero-values are omitted (W-4).
type CreateEndpointRequest struct {
	ClientID           string   `json:"client_id"`
	EndpointName       string   `json:"endpoint_name"`
	MinLoad            *float64 `json:"min_load,omitempty"`
	MinColdLoad        *float64 `json:"min_cold_load,omitempty"`
	TargetUtil         *float64 `json:"target_util,omitempty"`
	ColdMult           *float64 `json:"cold_mult,omitempty"`
	ColdWorkers        *int     `json:"cold_workers,omitempty"`
	MaxWorkers         *int     `json:"max_workers,omitempty"`
	AutoscalerInstance string   `json:"autoscaler_instance"`
}

// UpdateEndpointRequest is the JSON body for PUT /endptjobs/{id}/ (update endpoint).
// Pointer types allow omitting zero-value fields (omitempty) for partial updates.
type UpdateEndpointRequest struct {
	ClientID           string   `json:"client_id"`
	EndptJobID         int      `json:"endptjob_id"`
	EndpointName       string   `json:"endpoint_name,omitempty"`
	MinLoad            *float64 `json:"min_load,omitempty"`
	MinColdLoad        *float64 `json:"min_cold_load,omitempty"`
	TargetUtil         *float64 `json:"target_util,omitempty"`
	ColdMult           *float64 `json:"cold_mult,omitempty"`
	ColdWorkers        *int     `json:"cold_workers,omitempty"`
	MaxWorkers         *int     `json:"max_workers,omitempty"`
	EndpointState      string   `json:"endpoint_state,omitempty"`
	AutoscalerInstance string   `json:"autoscaler_instance"`
}

// Endpoint represents a serverless endpoint from the API response.
type Endpoint struct {
	ID            int     `json:"id"`
	EndpointName  string  `json:"endpoint_name"`
	MinLoad       float64 `json:"min_load"`
	MinColdLoad   float64 `json:"min_cold_load"`
	TargetUtil    float64 `json:"target_util"`
	ColdMult      float64 `json:"cold_mult"`
	ColdWorkers   int     `json:"cold_workers"`
	MaxWorkers    int     `json:"max_workers"`
	EndpointState string  `json:"endpoint_state"`
}

// endpointListResponse wraps the endpoint list API response.
type endpointListResponse struct {
	Success bool       `json:"success"`
	Results []Endpoint `json:"results"`
	Message string     `json:"msg,omitempty"`
}

// Create creates a new serverless endpoint.
// Sends POST /endptjobs/ with endpoint configuration, then reads back via List
// to return the full endpoint object (create-then-read pattern).
// Always sets client_id="me" and autoscaler_instance="prod" (Pitfall 5).
func (s *EndpointService) Create(ctx context.Context, req *CreateEndpointRequest) (*Endpoint, error) {
	req.ClientID = "me"
	req.AutoscalerInstance = "prod"

	var resp map[string]interface{}
	if err := s.client.Post(ctx, "/endptjobs/", req, &resp); err != nil {
		return nil, fmt.Errorf("creating endpoint: %w", err)
	}

	// Create-then-read: list all endpoints and find the newly created one by name
	endpoints, err := s.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading endpoint after create: %w", err)
	}

	// Find by name, iterating backwards for the most recently created match
	for i := len(endpoints) - 1; i >= 0; i-- {
		if endpoints[i].EndpointName == req.EndpointName {
			return &endpoints[i], nil
		}
	}

	return nil, fmt.Errorf("endpoint %q not found after creation", req.EndpointName)
}

// List retrieves all serverless endpoints owned by the user.
// Sends GET /endptjobs/ and checks the success field in the response.
// Pitfall 1: No single-GET endpoint exists; always use list and filter in Go.
func (s *EndpointService) List(ctx context.Context) ([]Endpoint, error) {
	var resp endpointListResponse
	if err := s.client.Get(ctx, "/endptjobs/", &resp); err != nil {
		return nil, fmt.Errorf("listing endpoints: %w", err)
	}
	return resp.Results, nil
}

// Update updates an existing serverless endpoint by ID.
// Sends PUT /endptjobs/{id}/ with the updated fields.
// Always sets client_id="me", endptjob_id=id, and autoscaler_instance="prod".
func (s *EndpointService) Update(ctx context.Context, id int, req *UpdateEndpointRequest) error {
	req.ClientID = "me"
	req.EndptJobID = id
	req.AutoscalerInstance = "prod"

	path := fmt.Sprintf("/endptjobs/%d/", id)
	if err := s.client.Put(ctx, path, req, nil); err != nil {
		return fmt.Errorf("updating endpoint %d: %w", id, err)
	}
	return nil
}

// Delete deletes a serverless endpoint by ID.
// Sends DELETE /endptjobs/{id}/ with JSON body containing client_id and endptjob_id.
// Pitfall 2: Delete requires a JSON body (uses DeleteWithBody, not Delete).
func (s *EndpointService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/endptjobs/%d/", id)
	body := map[string]interface{}{
		"client_id":   "me",
		"endptjob_id": id,
	}
	if err := s.client.DeleteWithBody(ctx, path, body, nil); err != nil {
		return fmt.Errorf("deleting endpoint %d: %w", id, err)
	}
	return nil
}
