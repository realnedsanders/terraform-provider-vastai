package client

import (
	"context"
	"fmt"
)

// NetworkVolumeService handles network volume-related API operations.
type NetworkVolumeService struct {
	client *VastAIClient
}

// CreateNetworkVolumeRequest is the JSON body for PUT /network_volumes/.
type CreateNetworkVolumeRequest struct {
	Size    int    `json:"size"`
	OfferID int    `json:"id"`
	Name    string `json:"name,omitempty"`
}

// NetworkVolumeOfferSearchParams defines the structured search parameters for network volume offers.
// Uses the same filter fields as VolumeOfferSearchParams (vol_offers_fields is shared).
type NetworkVolumeOfferSearchParams struct {
	DiskSpace        *float64 `json:"-"` // Minimum disk space in GB
	StorageCost      *float64 `json:"-"` // Max $/GB/month
	InetUp           *float64 `json:"-"` // Min upload Mb/s
	InetDown         *float64 `json:"-"` // Min download Mb/s
	Reliability      *float64 `json:"-"` // Min reliability (0-1)
	Geolocation      string   `json:"-"` // Country code
	Verified         *bool    `json:"-"` // Machine verified
	StaticIP         *bool    `json:"-"` // Static IP
	DiskBW           *float64 `json:"-"` // Min disk bandwidth MB/s
	OrderBy          string   `json:"-"` // Sort field (default: "score")
	Limit            int      `json:"-"` // Max results (default: 10)
	AllocatedStorage float64  `json:"-"` // Storage amount for pricing (default: 1.0)
	RawQuery         string   `json:"-"` // Bypass structured filters
}

// NetworkVolumeOffer represents a single network volume offer from the Vast.ai marketplace.
// Includes network-volume-specific bandwidth metrics not present on local volume offers.
type NetworkVolumeOffer struct {
	ID           int     `json:"id"`
	DiskSpace    float64 `json:"disk_space"`
	StorageCost  float64 `json:"storage_cost"`
	InetUp       float64 `json:"inet_up"`
	InetDown     float64 `json:"inet_down"`
	Reliability  float64 `json:"reliability"`
	Duration     float64 `json:"duration"`
	Verification string  `json:"verification"`
	HostID       int     `json:"host_id"`
	ClusterID    int     `json:"cluster_id"`
	Geolocation  string  `json:"geolocation"`
	// Network-volume-specific bandwidth metrics
	NWDiskMinBW float64 `json:"nw_disk_min_bw"`
	NWDiskMaxBW float64 `json:"nw_disk_max_bw"`
	NWDiskAvgBW float64 `json:"nw_disk_avg_bw"`
}

// networkVolumeOfferSearchResponse wraps the network volume offer search API response.
type networkVolumeOfferSearchResponse struct {
	Offers []NetworkVolumeOffer `json:"offers"`
}

// Create creates a network volume from an offer.
// Sends PUT /network_volumes/ with {size, id, name?}, then reads back via List
// since create response may only return {id, success}.
func (s *NetworkVolumeService) Create(ctx context.Context, req *CreateNetworkVolumeRequest) (*Volume, error) {
	var createResp createVolumeResponse
	if err := s.client.Put(ctx, "/network_volumes/", req, &createResp); err != nil {
		return nil, fmt.Errorf("creating network volume: %w", err)
	}

	// Read back via List since create response is minimal
	volumes, err := s.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading network volume after create: %w", err)
	}

	for i := range volumes {
		if volumes[i].ID == createResp.ID {
			return &volumes[i], nil
		}
	}

	// If we can't find it by ID, return what we know
	return &Volume{ID: createResp.ID}, nil
}

// List retrieves all network volumes owned by the user.
// Sends GET /volumes?owner=me&type=network_volume.
// Pitfall 6: Uses same /volumes endpoint as local volumes, different type parameter.
func (s *NetworkVolumeService) List(ctx context.Context) ([]Volume, error) {
	path := "/volumes?owner=me&type=network_volume"
	var resp volumeListResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("listing network volumes: %w", err)
	}
	return resp.Volumes, nil
}

// Delete deletes a network volume by ID.
// Sends DELETE /volumes/?id={id} (same endpoint as local volumes).
func (s *NetworkVolumeService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/volumes/?id=%d", id)
	if err := s.client.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting network volume %d: %w", id, err)
	}
	return nil
}

// SearchOffers searches for network volume offers matching the given parameters.
// Sends POST /network_volumes/search/ with structured query body.
func (s *NetworkVolumeService) SearchOffers(ctx context.Context, params *NetworkVolumeOfferSearchParams) ([]NetworkVolumeOffer, error) {
	body := s.buildSearchBody(params)

	var resp networkVolumeOfferSearchResponse
	if err := s.client.Post(ctx, "/network_volumes/search/", body, &resp); err != nil {
		return nil, fmt.Errorf("searching network volume offers: %w", err)
	}
	return resp.Offers, nil
}

// buildSearchBody constructs the search request body from NetworkVolumeOfferSearchParams.
// Python SDK sends the query dict directly as a flat body (no "q" wrapping).
func (s *NetworkVolumeService) buildSearchBody(params *NetworkVolumeOfferSearchParams) map[string]interface{} {
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "score"
	}

	allocatedStorage := params.AllocatedStorage
	if allocatedStorage <= 0 {
		allocatedStorage = 1.0
	}

	// If raw query is provided, use it directly as a flat body
	if params.RawQuery != "" {
		return map[string]interface{}{
			"q":                 params.RawQuery,
			"order":             []interface{}{[]interface{}{orderBy, "desc"}},
			"limit":             limit,
			"allocated_storage": allocatedStorage,
		}
	}

	// Build flat query body with default filters (matches Python SDK defaults)
	// Python: {"verified": {"eq": True}, "external": {"eq": False}, "disk_space": {"gte": 1}}
	query := map[string]interface{}{
		"verified":   map[string]interface{}{"eq": true},
		"external":   map[string]interface{}{"eq": false},
		"disk_space": map[string]interface{}{"gte": float64(1)},
	}

	// Override defaults with explicit params
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

	// Add order, limit, allocated_storage as top-level keys in the flat body
	query["order"] = []interface{}{[]interface{}{orderBy, "desc"}}
	query["limit"] = limit
	query["allocated_storage"] = allocatedStorage

	return query
}
