package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// InvoiceService handles invoice-related API operations.
// Uses the v1 API endpoint via GetFullPath (not the standard /api/v0 prefix).
type InvoiceService struct {
	client *VastAIClient
}

// Invoice represents a billing invoice from the Vast.ai API.
type Invoice struct {
	ID          int     `json:"id"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Timestamp   string  `json:"timestamp"`
}

// InvoiceListResponse wraps the invoice list API response.
type InvoiceListResponse struct {
	Results   []Invoice `json:"results"`
	Count     int       `json:"count"`
	Total     int       `json:"total"`
	NextToken string    `json:"next_token"`
}

// InvoiceListParams contains optional filtering parameters for listing invoices.
// Per D-06: support date filtering and type.
type InvoiceListParams struct {
	StartDate string
	EndDate   string
	Limit     int
	Type      string
}

// List retrieves invoices using the v1 API endpoint.
// Sends GET /api/v1/invoices/ with optional query parameters.
// Uses GetFullPath to bypass the default /api/v0 prefix.
func (s *InvoiceService) List(ctx context.Context, params InvoiceListParams) (*InvoiceListResponse, error) {
	queryParams := url.Values{}
	if params.StartDate != "" {
		queryParams.Set("start_date", params.StartDate)
	}
	if params.EndDate != "" {
		queryParams.Set("end_date", params.EndDate)
	}
	if params.Limit > 0 {
		queryParams.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Type != "" {
		queryParams.Set("type", params.Type)
	}

	path := "/api/v1/invoices/"
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	var resp InvoiceListResponse
	if err := s.client.GetFullPath(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("listing invoices: %w", err)
	}
	return &resp, nil
}
