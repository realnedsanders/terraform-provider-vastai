package workergroup

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WorkerGroupResourceModel describes the resource data model for vastai_worker_group.
// Worker groups bind to serverless endpoints and define the compute configuration
// (template, GPU requirements, search parameters) for inference workloads.
//
// Note: Autoscaling parameters (min_load, target_util, cold_mult) are intentionally
// omitted from this model. Per Pitfall 3 from research, these fields are "not currently
// used at the workergroup level" -- autoscaling is driven by the parent endpoint.
// Configure autoscaling via the vastai_endpoint resource instead.
type WorkerGroupResourceModel struct {
	// Primary identifier (worker group ID as string for Terraform compatibility)
	ID types.String `tfsdk:"id"`

	// Required -- binds to parent endpoint (ForceNew: cannot re-parent)
	EndpointID types.Int64 `tfsdk:"endpoint_id"`

	// Optional+Computed -- inferred from endpoint_id, useful for display
	EndpointName types.String `tfsdk:"endpoint_name"`

	// Template configuration -- at least one of template_hash or template_id required
	TemplateHash types.String `tfsdk:"template_hash"`
	TemplateID   types.Int64  `tfsdk:"template_id"`

	// Optional search and launch configuration
	SearchParams types.String  `tfsdk:"search_params"`
	LaunchArgs   types.String  `tfsdk:"launch_args"`
	GpuRAM       types.Float64 `tfsdk:"gpu_ram"`

	// Worker scaling configuration
	TestWorkers types.Int64 `tfsdk:"test_workers"`
	ColdWorkers types.Int64 `tfsdk:"cold_workers"`

	// Timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
