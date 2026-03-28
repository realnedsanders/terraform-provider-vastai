package client

import (
	"context"
	"encoding/json"
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
	ID              int     `json:"id"`
	Amount          float64 `json:"amount"`
	Type            string  `json:"type"`
	Description     string  `json:"description"`
	Timestamp       string  `json:"timestamp"`
	UserID          int     `json:"user_id"`
	PaidOn          string  `json:"paid_on"`
	PaymentExpected string  `json:"payment_expected"`
	AmountCents     int     `json:"amount_cents"`
	IsCredit        bool    `json:"is_credit"`
	Service         string  `json:"service"`
	BalanceBefore   float64 `json:"balance_before"`
	BalanceAfter    float64 `json:"balance_after"`
}

// InvoiceListResponse wraps the invoice list API response.
type InvoiceListResponse struct {
	Results   []Invoice `json:"results"`
	Count     int       `json:"count"`
	Total     int       `json:"total"`
	NextToken string    `json:"next_token"`
}

// InvoiceListParams contains optional filtering parameters for listing invoices.
// Matches Python SDK show__invoices_v1: uses select_filters with nested date range,
// plus latest_first and after_token for pagination.
type InvoiceListParams struct {
	StartDate   float64 // Unix timestamp for start of date range
	EndDate     float64 // Unix timestamp for end of date range
	Limit       int
	Type        string
	LatestFirst bool
	AfterToken  string
}

// List retrieves invoices using the v1 API endpoint.
// Sends GET /api/v1/invoices/ with select_filters, latest_first, limit, and after_token
// as query parameters (matching Python SDK show__invoices_v1).
// Uses GetFullPath to bypass the default /api/v0 prefix.
func (s *InvoiceService) List(ctx context.Context, params InvoiceListParams) (*InvoiceListResponse, error) {
	queryParams := url.Values{}

	// Build select_filters as a nested JSON object: {"when": {"gte": start, "lte": end}}
	// Matches Python SDK: params['select_filters'] = {date_col: {'gte': start_timestamp, 'lte': end_timestamp}}
	selectFilters := map[string]interface{}{}
	if params.StartDate > 0 || params.EndDate > 0 {
		whenFilter := map[string]interface{}{}
		if params.StartDate > 0 {
			whenFilter["gte"] = params.StartDate
		}
		if params.EndDate > 0 {
			whenFilter["lte"] = params.EndDate
		}
		selectFilters["when"] = whenFilter
	}

	if len(selectFilters) > 0 {
		filtersJSON, err := json.Marshal(selectFilters)
		if err != nil {
			return nil, fmt.Errorf("marshaling select_filters: %w", err)
		}
		queryParams.Set("select_filters", string(filtersJSON))
	}

	if params.Limit > 0 {
		limit := params.Limit
		if limit > 100 {
			limit = 100
		}
		queryParams.Set("limit", strconv.Itoa(limit))
	}

	if params.LatestFirst {
		queryParams.Set("latest_first", "true")
	}

	if params.AfterToken != "" {
		queryParams.Set("after_token", params.AfterToken)
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
