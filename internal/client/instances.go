package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ExtraEnvMap is a map[string]string that can unmarshal from the API's
// [[key, val], ...] format (array of 2-element arrays) as well as from
// a standard JSON object {"key": "val", ...}.
type ExtraEnvMap map[string]string

// UnmarshalJSON implements json.Unmarshaler for ExtraEnvMap.
// Handles both [[key, val], ...] (API format) and {"key": "val"} (standard).
func (e *ExtraEnvMap) UnmarshalJSON(data []byte) error {
	// Try array-of-arrays format first: [[key, val], ...]
	var pairs [][]string
	if err := json.Unmarshal(data, &pairs); err == nil {
		m := make(map[string]string, len(pairs))
		for _, pair := range pairs {
			if len(pair) == 2 {
				m[pair[0]] = pair[1]
			}
		}
		*e = m
		return nil
	}

	// Fall back to standard object format: {"key": "val"}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("extra_env: cannot unmarshal %s", string(data))
	}
	*e = m
	return nil
}

// InstanceService handles instance-related API operations.
type InstanceService struct {
	client *VastAIClient
}

// CreateInstanceRequest is the JSON body for PUT /asks/{offer_id}/.
type CreateInstanceRequest struct {
	ClientID       string            `json:"client_id"`
	Image          string            `json:"image"`
	Env            map[string]string `json:"env,omitempty"`
	Price          *float64          `json:"price,omitempty"` // nil = on-demand pricing
	Disk           float64           `json:"disk"`
	Label          string            `json:"label,omitempty"`
	Onstart        string            `json:"onstart,omitempty"`
	Runtype        string            `json:"runtype,omitempty"` // ssh, jupyter, args
	TemplateHashID string            `json:"template_hash_id,omitempty"`
	ImageLogin     string            `json:"image_login,omitempty"`
	CancelUnavail  bool              `json:"cancel_unavail,omitempty"`
	PythonUTF8     bool              `json:"python_utf8,omitempty"`
	LangUTF8       bool              `json:"lang_utf8,omitempty"`
	UseJupyterLab  bool              `json:"use_jupyter_lab,omitempty"`
	JupyterDir     string            `json:"jupyter_dir,omitempty"`
	Force          bool              `json:"force,omitempty"`
}

// CreateInstanceResponse is the JSON response from instance creation.
type CreateInstanceResponse struct {
	Success     bool `json:"success"`
	NewContract int  `json:"new_contract"`
}

// UpdateTemplateRequest is the JSON body for PUT /instances/update_template/{id}/.
type UpdateTemplateRequest struct {
	ID             int               `json:"id,omitempty"`
	Image          string            `json:"image,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	Onstart        string            `json:"onstart,omitempty"`
	TemplateHashID string            `json:"template_hash_id,omitempty"`
}

// Instance represents the full instance object from GET /instances/{id}/.
type Instance struct {
	ID                int         `json:"id"`
	MachineID         int         `json:"machine_id"`
	ActualStatus      string      `json:"actual_status"`
	IntendedStatus    string      `json:"intended_status"`
	CurState          string      `json:"cur_state"`
	NextState         string      `json:"next_state"`
	NumGPUs           int         `json:"num_gpus"`
	GPUName           string      `json:"gpu_name"`
	GPUUtil           float64     `json:"gpu_util"`
	GPURAM            float64     `json:"gpu_ram"`
	GPUTotalRAM       float64     `json:"gpu_totalram"`
	CPUCoresEffective float64     `json:"cpu_cores_effective"`
	CPURAM            float64     `json:"cpu_ram"`
	DiskSpace         float64     `json:"disk_space"`
	SSHHost           string      `json:"ssh_host"`
	SSHPort           int         `json:"ssh_port"`
	DPHTotal          float64     `json:"dph_total"`
	ImageUUID         string      `json:"image_uuid"`
	Label             string      `json:"label"`
	InetUp            float64     `json:"inet_up"`
	InetDown          float64     `json:"inet_down"`
	Reliability2      float64     `json:"reliability2"`
	StartDate         float64     `json:"start_date"`
	IsBid             bool        `json:"is_bid"`
	MinBid            float64     `json:"min_bid"`
	Geolocation       string      `json:"geolocation"`
	HostingType       int         `json:"hosting_type"`
	TemplateID        int         `json:"template_id"`
	TemplateHashID    string      `json:"template_hash_id"`
	StatusMsg         string      `json:"status_msg"`
	ExtraEnv          ExtraEnvMap `json:"extra_env"`
	Onstart           string      `json:"onstart"`
	Verification      string      `json:"verification"`
	DirectPortCount   int         `json:"direct_port_count"`
}

// defaultPollInterval is the default interval between status polls.
const defaultPollInterval = 5 * time.Second

// instanceGetWrapper wraps the single-instance API response.
type instanceGetWrapper struct {
	Instances Instance `json:"instances"`
}

// instanceListWrapper wraps the instance list API response.
type instanceListWrapper struct {
	Instances []Instance `json:"instances"`
}

// Create creates a new instance from an offer.
// Sends PUT /asks/{offerID}/ and returns the contract response.
func (s *InstanceService) Create(ctx context.Context, offerID int, req *CreateInstanceRequest) (*CreateInstanceResponse, error) {
	path := fmt.Sprintf("/asks/%d/", offerID)
	var resp CreateInstanceResponse
	if err := s.client.Put(ctx, path, req, &resp); err != nil {
		return nil, fmt.Errorf("creating instance from offer %d: %w", offerID, err)
	}
	return &resp, nil
}

// Get retrieves a single instance by ID.
// Sends GET /instances/{id}/?owner=me and unwraps the response.
func (s *InstanceService) Get(ctx context.Context, id int) (*Instance, error) {
	path := fmt.Sprintf("/instances/%d/?owner=me", id)
	var wrapper instanceGetWrapper
	if err := s.client.Get(ctx, path, &wrapper); err != nil {
		return nil, fmt.Errorf("getting instance %d: %w", id, err)
	}
	return &wrapper.Instances, nil
}

// List retrieves all instances owned by the authenticated user.
// Sends GET /instances?owner=me and unwraps the response.
func (s *InstanceService) List(ctx context.Context) ([]Instance, error) {
	path := "/instances?owner=me"
	var wrapper instanceListWrapper
	if err := s.client.Get(ctx, path, &wrapper); err != nil {
		return nil, fmt.Errorf("listing instances: %w", err)
	}
	return wrapper.Instances, nil
}

// Start starts a stopped instance.
// Sends PUT /instances/{id}/ with {"state": "running"}.
func (s *InstanceService) Start(ctx context.Context, id int) error {
	path := fmt.Sprintf("/instances/%d/", id)
	body := map[string]string{"state": "running"}
	if err := s.client.Put(ctx, path, body, nil); err != nil {
		return fmt.Errorf("starting instance %d: %w", id, err)
	}
	return nil
}

// Stop stops a running instance.
// Sends PUT /instances/{id}/ with {"state": "stopped"}.
func (s *InstanceService) Stop(ctx context.Context, id int) error {
	path := fmt.Sprintf("/instances/%d/", id)
	body := map[string]string{"state": "stopped"}
	if err := s.client.Put(ctx, path, body, nil); err != nil {
		return fmt.Errorf("stopping instance %d: %w", id, err)
	}
	return nil
}

// Destroy destroys an instance.
// Sends DELETE /instances/{id}/.
func (s *InstanceService) Destroy(ctx context.Context, id int) error {
	path := fmt.Sprintf("/instances/%d/", id)
	if err := s.client.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("destroying instance %d: %w", id, err)
	}
	return nil
}

// SetLabel updates the label on an instance.
// Sends PUT /instances/{id}/ with {"label": label}.
func (s *InstanceService) SetLabel(ctx context.Context, id int, label string) error {
	path := fmt.Sprintf("/instances/%d/", id)
	body := map[string]string{"label": label}
	if err := s.client.Put(ctx, path, body, nil); err != nil {
		return fmt.Errorf("setting label on instance %d: %w", id, err)
	}
	return nil
}

// ChangeBid changes the bid price on a spot/interruptible instance.
// Sends PUT /instances/bid_price/{id}/ with {"client_id": "me", "price": price}.
func (s *InstanceService) ChangeBid(ctx context.Context, id int, price float64) error {
	path := fmt.Sprintf("/instances/bid_price/%d/", id)
	body := map[string]interface{}{
		"client_id": "me",
		"price":     price,
	}
	if err := s.client.Put(ctx, path, body, nil); err != nil {
		return fmt.Errorf("changing bid on instance %d: %w", id, err)
	}
	return nil
}

// UpdateTemplate updates the template configuration on a running instance.
// Sends PUT /instances/update_template/{id}/.
func (s *InstanceService) UpdateTemplate(ctx context.Context, id int, req *UpdateTemplateRequest) error {
	path := fmt.Sprintf("/instances/update_template/%d/", id)
	if err := s.client.Put(ctx, path, req, nil); err != nil {
		return fmt.Errorf("updating template on instance %d: %w", id, err)
	}
	return nil
}

// WaitForStatus polls the instance until actual_status matches the target status,
// the context is cancelled, or the timeout expires.
// For "destroyed" target, a 404 response is treated as success.
// Returns an error if the instance enters the terminal "exited" state unexpectedly.
func (s *InstanceService) WaitForStatus(ctx context.Context, id int, targetStatus string, timeout time.Duration) (*Instance, error) {
	// Only add a new timeout if the context doesn't already have a deadline,
	// avoiding nested timeouts when the caller already set one on ctx.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		instance, err := s.Get(ctx, id)
		if err != nil {
			// On 404 when waiting for destroy, treat as success
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == 404 && targetStatus == "destroyed" {
				tflog.Debug(ctx, "Instance not found (404), treating as destroyed", map[string]interface{}{
					"instance_id": id,
				})
				return nil, nil
			}
			return nil, fmt.Errorf("polling instance %d status: %w", id, err)
		}

		tflog.Debug(ctx, "Instance status poll", map[string]interface{}{
			"instance_id":   id,
			"actual_status": instance.ActualStatus,
			"target_status": targetStatus,
		})

		if instance.ActualStatus == targetStatus {
			return instance, nil
		}

		// Check for terminal "exited" state when not explicitly waiting for it
		if instance.ActualStatus == "exited" && targetStatus != "exited" {
			return instance, fmt.Errorf("instance %d entered terminal state 'exited' while waiting for '%s': %s",
				id, targetStatus, instance.StatusMsg)
		}

		select {
		case <-ctx.Done():
			return instance, fmt.Errorf("timed out waiting for instance %d to reach status '%s' (current: '%s')",
				id, targetStatus, instance.ActualStatus)
		case <-ticker.C:
			// continue polling
		}
	}
}
