package client

import (
	"context"
	"fmt"
)

// WorkerGroupService handles worker group-related API operations.
type WorkerGroupService struct {
	client *VastAIClient
}

// CreateWorkerGroupRequest is the JSON body for POST /autojobs/ (create worker group).
type CreateWorkerGroupRequest struct {
	ClientID           string  `json:"client_id"`
	EndpointName       string  `json:"endpoint_name,omitempty"`
	EndpointID         int     `json:"endpoint_id,omitempty"`
	TemplateHash       string  `json:"template_hash,omitempty"`
	TemplateID         int     `json:"template_id,omitempty"`
	SearchParams       string  `json:"search_params,omitempty"`
	LaunchArgs         string  `json:"launch_args,omitempty"`
	GpuRAM             float64 `json:"gpu_ram,omitempty"`
	MinLoad            float64 `json:"min_load"`
	TargetUtil         float64 `json:"target_util"`
	ColdMult           float64 `json:"cold_mult"`
	ColdWorkers        int     `json:"cold_workers"`
	TestWorkers        int     `json:"test_workers"`
	AutoscalerInstance string  `json:"autoscaler_instance"`
}

// UpdateWorkerGroupRequest is the JSON body for PUT /autojobs/{id}/ (update worker group).
// Pointer types allow omitting zero-value fields (omitempty) for partial updates.
type UpdateWorkerGroupRequest struct {
	ClientID     string   `json:"client_id"`
	AutoJobID    int      `json:"autojob_id"`
	EndpointName string   `json:"endpoint_name,omitempty"`
	EndpointID   int      `json:"endpoint_id,omitempty"`
	TemplateHash string   `json:"template_hash,omitempty"`
	TemplateID   int      `json:"template_id,omitempty"`
	SearchParams string   `json:"search_params,omitempty"`
	LaunchArgs   string   `json:"launch_args,omitempty"`
	GpuRAM       *float64 `json:"gpu_ram,omitempty"`
	MinLoad      *float64 `json:"min_load,omitempty"`
	TargetUtil   *float64 `json:"target_util,omitempty"`
	ColdMult     *float64 `json:"cold_mult,omitempty"`
	ColdWorkers  *int     `json:"cold_workers,omitempty"`
	TestWorkers  *int     `json:"test_workers,omitempty"`
}

// WorkerGroup represents a worker group from the API response.
type WorkerGroup struct {
	ID           int     `json:"id"`
	EndpointName string  `json:"endpoint_name"`
	EndpointID   int     `json:"endpoint_id"`
	TemplateHash string  `json:"template_hash"`
	TemplateID   int     `json:"template_id"`
	SearchParams string  `json:"search_params"`
	LaunchArgs   string  `json:"launch_args"`
	GpuRAM       float64 `json:"gpu_ram"`
	MinLoad      float64 `json:"min_load"`
	TargetUtil   float64 `json:"target_util"`
	ColdMult     float64 `json:"cold_mult"`
	ColdWorkers  int     `json:"cold_workers"`
	TestWorkers  int     `json:"test_workers"`
}

// workerGroupListResponse wraps the worker group list API response.
type workerGroupListResponse struct {
	Success bool          `json:"success"`
	Results []WorkerGroup `json:"results"`
	Message string        `json:"msg,omitempty"`
}

// Create creates a new worker group bound to an endpoint.
// Sends POST /autojobs/ with worker group configuration, then reads back via List
// to return the full worker group object (create-then-read pattern).
// Always sets client_id="me" and autoscaler_instance="prod" (Pitfall 5).
func (s *WorkerGroupService) Create(ctx context.Context, req *CreateWorkerGroupRequest) (*WorkerGroup, error) {
	req.ClientID = "me"
	req.AutoscalerInstance = "prod"

	var resp map[string]interface{}
	if err := s.client.Post(ctx, "/autojobs/", req, &resp); err != nil {
		return nil, fmt.Errorf("creating worker group: %w", err)
	}

	// Create-then-read: list all worker groups and find the newest one by highest ID
	groups, err := s.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading worker group after create: %w", err)
	}

	var newest *WorkerGroup
	for i := range groups {
		if newest == nil || groups[i].ID > newest.ID {
			newest = &groups[i]
		}
	}

	if newest == nil {
		return nil, fmt.Errorf("worker group not found after creation")
	}
	return newest, nil
}

// List retrieves all worker groups owned by the user.
// Sends GET /autojobs/ and checks the success field in the response.
// Pitfall 1: No single-GET endpoint exists; always use list and filter in Go.
func (s *WorkerGroupService) List(ctx context.Context) ([]WorkerGroup, error) {
	var resp workerGroupListResponse
	if err := s.client.Get(ctx, "/autojobs/", &resp); err != nil {
		return nil, fmt.Errorf("listing worker groups: %w", err)
	}
	return resp.Results, nil
}

// Update updates an existing worker group by ID.
// Sends PUT /autojobs/{id}/ with the updated fields.
// Always sets client_id="me" and autojob_id=id.
func (s *WorkerGroupService) Update(ctx context.Context, id int, req *UpdateWorkerGroupRequest) error {
	req.ClientID = "me"
	req.AutoJobID = id

	path := fmt.Sprintf("/autojobs/%d/", id)
	if err := s.client.Put(ctx, path, req, nil); err != nil {
		return fmt.Errorf("updating worker group %d: %w", id, err)
	}
	return nil
}

// Delete deletes a worker group by ID.
// Sends DELETE /autojobs/{id}/ with JSON body containing client_id and autojob_id.
// Pitfall 2: Delete requires a JSON body (uses DeleteWithBody, not Delete).
// Note: Deleting a worker group does NOT automatically destroy its associated instances.
func (s *WorkerGroupService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/autojobs/%d/", id)
	body := map[string]interface{}{
		"client_id":  "me",
		"autojob_id": id,
	}
	if err := s.client.DeleteWithBody(ctx, path, body, nil); err != nil {
		return fmt.Errorf("deleting worker group %d: %w", id, err)
	}
	return nil
}
