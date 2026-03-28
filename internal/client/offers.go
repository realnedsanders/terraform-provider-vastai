package client

import (
	"context"
	"fmt"
)

// OfferService handles GPU offer search API operations.
type OfferService struct {
	client *VastAIClient
}

// OfferSearchParams defines the structured search parameters for GPU offers.
// Pointer types indicate optional fields -- nil means "not set" (omitted from query).
type OfferSearchParams struct {
	GPUName        string   `json:"-"` // Mapped to query filter
	NumGPUs        *int     `json:"-"` // Mapped to query filter
	GPURamGB       *float64 `json:"-"` // Converted to MB (*1000) for API
	MaxPrice       *float64 `json:"-"` // Max $/hr (dph_total)
	DatacenterOnly *bool    `json:"-"` // Filter to datacenter hosting only
	Region         string   `json:"-"` // Geographic region filter
	OfferType      string   `json:"-"` // "on-demand", "bid", etc.
	OrderBy        string   `json:"-"` // Sort field (default: "dph_total")
	Limit          int      `json:"-"` // Max results (default: 10)
	RawQuery       string   `json:"-"` // Raw query JSON -- bypasses structured filters
	MinDisk        *float64 `json:"-"` // Minimum disk space in GB
}

// Offer represents a single GPU offer from the Vast.ai marketplace.
type Offer struct {
	ID                int     `json:"id"`
	MachineID         int     `json:"machine_id"`
	GPUName           string  `json:"gpu_name"`
	NumGPUs           int     `json:"num_gpus"`
	GPURAM            float64 `json:"gpu_ram"`      // In MB from API
	GPUTotalRAM       float64 `json:"gpu_totalram"` // In MB from API
	CPUCoresEffective float64 `json:"cpu_cores_effective"`
	CPURAM            float64 `json:"cpu_ram"`
	DiskSpace         float64 `json:"disk_space"`
	DPHTotal          float64 `json:"dph_total"` // $/hr total
	DLPerf            float64 `json:"dlperf"`
	DLPerfPerDPH      float64 `json:"dlperf_per_dphtotal"`
	InetUp            float64 `json:"inet_up"`
	InetDown          float64 `json:"inet_down"`
	Reliability       float64 `json:"reliability2"`
	Geolocation       string  `json:"geolocation"`
	HostingType       int     `json:"hosting_type"`
	Verification      string  `json:"verification"`
	StaticIP          bool    `json:"static_ip"`
	DirectPortCount   int     `json:"direct_port_count"`
	PCIGen            float64 `json:"pci_gen"`
	PCIeBW            float64 `json:"pcie_bw"`
	CUDAMaxGood       float64 `json:"cuda_max_good"`
	MinBid            float64 `json:"min_bid"`
	Rentable          bool    `json:"rentable"`
	Rented            bool    `json:"rented"`
	BundleID          int     `json:"bundle_id"`
	StorageCost       float64 `json:"storage_cost"`
	Duration          float64 `json:"duration"`
}

// offerSearchResponse wraps the offer search API response.
type offerSearchResponse struct {
	Offers []Offer `json:"offers"`
}

// Search searches for GPU offers matching the given parameters.
// Sends PUT /search/asks/ with a structured query or raw query passthrough.
func (s *OfferService) Search(ctx context.Context, params *OfferSearchParams) ([]Offer, error) {
	// Build the request body
	body := s.buildSearchBody(params)

	var resp offerSearchResponse
	if err := s.client.Put(ctx, "/search/asks/", body, &resp); err != nil {
		return nil, fmt.Errorf("searching offers: %w", err)
	}
	return resp.Offers, nil
}

// buildSearchBody constructs the search request body from OfferSearchParams.
func (s *OfferService) buildSearchBody(params *OfferSearchParams) map[string]interface{} {
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "dph_total"
	}

	// If raw query is provided, use it directly
	if params.RawQuery != "" {
		return map[string]interface{}{
			"select_cols": []string{"*"},
			"q":           params.RawQuery,
			"limit":       limit,
		}
	}

	// Build structured query
	query := map[string]interface{}{
		"verified": map[string]interface{}{"eq": true},
		"external": map[string]interface{}{"eq": false},
		"rentable": map[string]interface{}{"eq": true},
		"rented":   map[string]interface{}{"eq": false},
		"order":    []interface{}{[]interface{}{orderBy, "asc"}},
		"type":     "on-demand",
	}

	// GPU name filter
	if params.GPUName != "" {
		query["gpu_name"] = map[string]interface{}{"eq": params.GPUName}
	}

	// Number of GPUs filter
	if params.NumGPUs != nil {
		query["num_gpus"] = map[string]interface{}{"eq": *params.NumGPUs}
	}

	// GPU RAM filter (convert GB to MB by multiplying by 1000)
	if params.GPURamGB != nil {
		gpuRAMMB := *params.GPURamGB * 1000
		query["gpu_ram"] = map[string]interface{}{"gte": gpuRAMMB}
	}

	// Max price filter
	if params.MaxPrice != nil {
		query["dph_total"] = map[string]interface{}{"lte": *params.MaxPrice}
	}

	// Datacenter-only filter
	if params.DatacenterOnly != nil && *params.DatacenterOnly {
		query["hosting_type"] = map[string]interface{}{"eq": 0}
	}

	// Region filter
	if params.Region != "" {
		query["geolocation"] = map[string]interface{}{"in": params.Region}
	}

	// Minimum disk space filter
	if params.MinDisk != nil {
		query["disk_space"] = map[string]interface{}{"gte": *params.MinDisk}
	}

	// Offer type filter
	if params.OfferType != "" {
		query["type"] = params.OfferType
	}

	return map[string]interface{}{
		"select_cols": []string{"*"},
		"q":           query,
		"limit":       limit,
	}
}
