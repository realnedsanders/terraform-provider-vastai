package client

import (
	"context"
	"fmt"
)

// VolumeService handles volume-related API operations.
type VolumeService struct {
	client *VastAIClient
}

// CreateVolumeRequest is the JSON body for PUT /volumes/.
type CreateVolumeRequest struct {
	Size    int    `json:"size"`
	OfferID int    `json:"id"`
	Name    string `json:"name,omitempty"`
}

// CloneVolumeRequest is the JSON body for POST /volumes/copy/.
type CloneVolumeRequest struct {
	SourceID           int     `json:"src_id"`
	DestOfferID        int     `json:"dst_id"`
	Size               float64 `json:"size,omitempty"`
	DisableCompression bool    `json:"disable_compression,omitempty"`
}

// Volume represents a volume object from the Vast.ai API.
// Used for both local volumes and network volumes (same response shape from show__volumes).
type Volume struct {
	ID            int     `json:"id"`
	ClusterID     int     `json:"cluster_id"`
	Label         string  `json:"label"`
	DiskSpace     float64 `json:"disk_space"`
	Status        string  `json:"status"`
	DiskName      string  `json:"disk_name"`
	DriverVersion string  `json:"driver_version"`
	InetUp        float64 `json:"inet_up"`
	InetDown      float64 `json:"inet_down"`
	Reliability   float64 `json:"reliability2"`
	StartDate     float64 `json:"start_date"`
	MachineID     int     `json:"machine_id"`
	Verification  string  `json:"verification"`
	HostID        int     `json:"host_id"`
	Geolocation   string  `json:"geolocation"`
	Instances     []int   `json:"instances"`
}

// volumeListResponse wraps the volume list API response.
type volumeListResponse struct {
	Volumes []Volume `json:"volumes"`
}

// createVolumeResponse wraps the volume create API response.
type createVolumeResponse struct {
	ID      int  `json:"id"`
	Success bool `json:"success"`
}

// VolumeOfferSearchParams defines the structured search parameters for volume offers.
// Pointer types indicate optional fields -- nil means "not set" (omitted from query).
type VolumeOfferSearchParams struct {
	DiskSpace        *float64 `json:"-"` // Minimum disk space in GB
	StorageCost      *float64 `json:"-"` // Max $/GB/month
	InetUp           *float64 `json:"-"` // Min upload Mb/s
	InetDown         *float64 `json:"-"` // Min download Mb/s
	Reliability      *float64 `json:"-"` // Min reliability (0-1)
	Geolocation      string   `json:"-"` // Country code
	Verified         *bool    `json:"-"` // Machine verified
	StaticIP         *bool    `json:"-"` // Static IP
	DiskBW           *float64 `json:"-"` // Min disk bandwidth MB/s
	OrderBy          string   `json:"-"` // Sort field (default: "storage_cost")
	Limit            int      `json:"-"` // Max results (default: 10)
	AllocatedStorage float64  `json:"-"` // Storage amount for pricing (default: 1.0)
	RawQuery         string   `json:"-"` // Bypass structured filters
}

// VolumeOffer represents a single volume offer from the Vast.ai marketplace.
type VolumeOffer struct {
	ID            int     `json:"id"`
	CUDAMaxGood   float64 `json:"cuda_max_good"`
	CPUGhz        float64 `json:"cpu_ghz"`
	DiskBW        float64 `json:"disk_bw"`
	DiskSpace     float64 `json:"disk_space"`
	DiskName      string  `json:"disk_name"`
	StorageCost   float64 `json:"storage_cost"`
	DriverVersion string  `json:"driver_version"`
	InetUp        float64 `json:"inet_up"`
	InetDown      float64 `json:"inet_down"`
	Reliability   float64 `json:"reliability"`
	Duration      float64 `json:"duration"`
	MachineID     int     `json:"machine_id"`
	Verification  string  `json:"verification"`
	HostID        int     `json:"host_id"`
	Geolocation   string  `json:"geolocation"`
}

// volumeOfferSearchResponse wraps the volume offer search API response.
type volumeOfferSearchResponse struct {
	Offers []VolumeOffer `json:"offers"`
}

// Create creates a volume from an offer.
// Sends PUT /volumes/ with {size, id, name?}, then reads back via List
// since create response may only return {id, success}.
func (s *VolumeService) Create(ctx context.Context, req *CreateVolumeRequest) (*Volume, error) {
	var createResp createVolumeResponse
	if err := s.client.Put(ctx, "/volumes/", req, &createResp); err != nil {
		return nil, fmt.Errorf("creating volume: %w", err)
	}

	// Read back via List since create response is minimal
	volumes, err := s.List(ctx, "local_volume")
	if err != nil {
		return nil, fmt.Errorf("reading volume after create: %w", err)
	}

	for i := range volumes {
		if volumes[i].ID == createResp.ID {
			return &volumes[i], nil
		}
	}

	// If we can't find it by ID, return what we know
	return &Volume{ID: createResp.ID}, nil
}

// Clone creates a volume by cloning an existing one.
// Sends POST /volumes/copy/ with {src_id, dst_id, size?, disable_compression?}.
// Clone may not return the new volume ID directly, so returns error only.
func (s *VolumeService) Clone(ctx context.Context, req *CloneVolumeRequest) error {
	if err := s.client.Post(ctx, "/volumes/copy/", req, nil); err != nil {
		return fmt.Errorf("cloning volume: %w", err)
	}
	return nil
}

// List retrieves all volumes owned by the user, filtered by type.
// Sends GET /volumes?owner=me&type={volumeType}.
// Valid types: "local_volume", "network_volume", "all_volume".
// Pitfall 3: No single-volume GET endpoint exists -- must use list and filter.
func (s *VolumeService) List(ctx context.Context, volumeType string) ([]Volume, error) {
	path := fmt.Sprintf("/volumes?owner=me&type=%s", volumeType)
	var resp volumeListResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("listing volumes (type=%s): %w", volumeType, err)
	}
	return resp.Volumes, nil
}

// Delete deletes a volume by ID.
// Sends DELETE /volumes/?id={id}.
// Pitfall 2: Uses query parameter, NOT path parameter.
func (s *VolumeService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/volumes/?id=%d", id)
	if err := s.client.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting volume %d: %w", id, err)
	}
	return nil
}

// SearchOffers searches for volume offers matching the given parameters.
// Sends POST /volumes/search/ with structured query body.
// Pitfall 5: Always includes "allocated_storage" field (default 1.0).
func (s *VolumeService) SearchOffers(ctx context.Context, params *VolumeOfferSearchParams) ([]VolumeOffer, error) {
	body := s.buildSearchBody(params)

	var resp volumeOfferSearchResponse
	if err := s.client.Post(ctx, "/volumes/search/", body, &resp); err != nil {
		return nil, fmt.Errorf("searching volume offers: %w", err)
	}
	return resp.Offers, nil
}

// buildSearchBody constructs the search request body from VolumeOfferSearchParams.
func (s *VolumeService) buildSearchBody(params *VolumeOfferSearchParams) map[string]interface{} {
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "storage_cost"
	}

	allocatedStorage := params.AllocatedStorage
	if allocatedStorage <= 0 {
		allocatedStorage = 1.0
	}

	// If raw query is provided, use it directly
	if params.RawQuery != "" {
		return map[string]interface{}{
			"q":                 params.RawQuery,
			"order":             []interface{}{[]interface{}{orderBy, "asc"}},
			"limit":             limit,
			"allocated_storage": allocatedStorage,
		}
	}

	// Build structured query
	query := map[string]interface{}{}

	if params.DiskSpace != nil {
		query["disk_space"] = map[string]interface{}{"gte": *params.DiskSpace}
	}
	if params.StorageCost != nil {
		query["storage_cost"] = map[string]interface{}{"lte": *params.StorageCost}
	}
	if params.InetUp != nil {
		query["inet_up"] = map[string]interface{}{"gte": *params.InetUp}
	}
	if params.InetDown != nil {
		query["inet_down"] = map[string]interface{}{"gte": *params.InetDown}
	}
	if params.Reliability != nil {
		query["reliability"] = map[string]interface{}{"gte": *params.Reliability}
	}
	if params.Geolocation != "" {
		query["geolocation"] = map[string]interface{}{"eq": params.Geolocation}
	}
	if params.Verified != nil {
		query["verified"] = map[string]interface{}{"eq": *params.Verified}
	}
	if params.StaticIP != nil {
		query["static_ip"] = map[string]interface{}{"eq": *params.StaticIP}
	}
	if params.DiskBW != nil {
		query["disk_bw"] = map[string]interface{}{"gte": *params.DiskBW}
	}

	return map[string]interface{}{
		"q":                 query,
		"order":             []interface{}{[]interface{}{orderBy, "asc"}},
		"limit":             limit,
		"allocated_storage": allocatedStorage,
	}
}
