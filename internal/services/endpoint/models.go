package endpoint

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EndpointResourceModel describes the resource data model for vastai_endpoint.
// The primary identifier is the endpoint ID stored as a string for Terraform
// compatibility (same pattern as other resources in this provider).
type EndpointResourceModel struct {
	// Primary identifier (endpoint ID as string)
	ID types.String `tfsdk:"id"`

	// Required attributes
	EndpointName types.String `tfsdk:"endpoint_name"`

	// Optional+Computed autoscaling attributes (server defaults per D-04, D-05)
	MinLoad     types.Float64 `tfsdk:"min_load"`
	MinColdLoad types.Float64 `tfsdk:"min_cold_load"`
	TargetUtil  types.Float64 `tfsdk:"target_util"`
	ColdMult    types.Float64 `tfsdk:"cold_mult"`
	ColdWorkers types.Int64   `tfsdk:"cold_workers"`
	MaxWorkers  types.Int64   `tfsdk:"max_workers"`

	// Optional+Computed (update-only per Pitfall 6)
	EndpointState types.String `tfsdk:"endpoint_state"`

	// Timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// EndpointsDataSourceModel describes the data source model for vastai_endpoints.
type EndpointsDataSourceModel struct {
	Endpoints types.List `tfsdk:"endpoints"`
}

// EndpointModel represents a single endpoint in the data source list.
type EndpointModel struct {
	ID            types.Int64   `tfsdk:"id"`
	EndpointName  types.String  `tfsdk:"endpoint_name"`
	MinLoad       types.Float64 `tfsdk:"min_load"`
	MinColdLoad   types.Float64 `tfsdk:"min_cold_load"`
	TargetUtil    types.Float64 `tfsdk:"target_util"`
	ColdMult      types.Float64 `tfsdk:"cold_mult"`
	ColdWorkers   types.Int64   `tfsdk:"cold_workers"`
	MaxWorkers    types.Int64   `tfsdk:"max_workers"`
	EndpointState types.String  `tfsdk:"endpoint_state"`
}
