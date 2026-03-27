package workergroup

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &WorkerGroupResource{}
	_ resource.ResourceWithConfigure   = &WorkerGroupResource{}
	_ resource.ResourceWithImportState = &WorkerGroupResource{}
)

// WorkerGroupResource defines the resource implementation.
type WorkerGroupResource struct {
	client *client.VastAIClient
}

// NewWorkerGroupResource creates a new worker group resource instance.
func NewWorkerGroupResource() resource.Resource {
	return &WorkerGroupResource{}
}

// Metadata returns the resource type name.
func (r *WorkerGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_worker_group"
}

// Schema defines the schema for the worker group resource.
func (r *WorkerGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai worker group. Worker groups bind to serverless endpoints and define " +
			"the compute configuration (template, GPU requirements, search parameters). " +
			"Autoscaling behavior is controlled at the endpoint level via the vastai_endpoint resource.",

		Attributes: map[string]schema.Attribute{
			// Primary identifier
			"id": schema.StringAttribute{
				Description: "Worker group ID. Used as the primary identifier for reads, updates, deletes, and imports.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Required -- binds to parent endpoint
			"endpoint_id": schema.Int64Attribute{
				Description: "ID of the parent endpoint this worker group belongs to. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},

			// Optional+Computed -- inferred from endpoint_id
			"endpoint_name": schema.StringAttribute{
				Description: "Name of the parent endpoint (computed from endpoint_id if not set).",
				Optional:    true,
				Computed:    true,
			},

			// Template configuration -- at least one required
			"template_hash": schema.StringAttribute{
				Description: "Template hash for worker instances. Either template_hash or template_id must be provided.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("template_id")),
				},
			},
			"template_id": schema.Int64Attribute{
				Description: "Numeric template ID for worker instances. Either template_hash or template_id must be provided.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeastOneOf(path.MatchRoot("template_hash")),
				},
			},

			// Search and launch configuration
			"search_params": schema.StringAttribute{
				Description: "Offer search filter string (e.g. 'gpu_ram>=23 num_gpus=2 gpu_name=RTX_4090'). " +
					"The API may apply default filters (verified=True rentable=True rented=False).",
				Optional: true,
			},
			"launch_args": schema.StringAttribute{
				Description: "Instance launch arguments string.",
				Optional:    true,
			},
			"gpu_ram": schema.Float64Attribute{
				Description: "Estimated GPU RAM requirement in GB.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},

			// Worker scaling
			"test_workers": schema.Int64Attribute{
				Description: "Number of workers for initial performance estimate (default: 3).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"cold_workers": schema.Int64Attribute{
				Description: "Minimum cold workers for this worker group.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *WorkerGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.VastAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.VastAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// Create creates a new worker group via the Vast.ai API.
func (r *WorkerGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model WorkerGroupResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout (5min default)
	createTimeout, diags := model.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Build API request from model
	createReq := &client.CreateWorkerGroupRequest{
		EndpointID: int(model.EndpointID.ValueInt64()),
		// Sensible defaults for API-required autoscaling fields (Pitfall 3:
		// these are not used at the worker group level, but the API requires them)
		MinLoad:    0,
		TargetUtil: 0.9,
		ColdMult:   2.0,
	}

	if !model.EndpointName.IsNull() && !model.EndpointName.IsUnknown() {
		createReq.EndpointName = model.EndpointName.ValueString()
	}
	if !model.TemplateHash.IsNull() && !model.TemplateHash.IsUnknown() {
		createReq.TemplateHash = model.TemplateHash.ValueString()
	}
	if !model.TemplateID.IsNull() && !model.TemplateID.IsUnknown() {
		createReq.TemplateID = int(model.TemplateID.ValueInt64())
	}
	if !model.SearchParams.IsNull() && !model.SearchParams.IsUnknown() {
		createReq.SearchParams = model.SearchParams.ValueString()
	}
	if !model.LaunchArgs.IsNull() && !model.LaunchArgs.IsUnknown() {
		createReq.LaunchArgs = model.LaunchArgs.ValueString()
	}
	if !model.GpuRAM.IsNull() && !model.GpuRAM.IsUnknown() {
		createReq.GpuRAM = model.GpuRAM.ValueFloat64()
	}
	if !model.TestWorkers.IsNull() && !model.TestWorkers.IsUnknown() {
		createReq.TestWorkers = int(model.TestWorkers.ValueInt64())
	} else {
		createReq.TestWorkers = 3 // default
	}
	if !model.ColdWorkers.IsNull() && !model.ColdWorkers.IsUnknown() {
		createReq.ColdWorkers = int(model.ColdWorkers.ValueInt64())
	}

	tflog.Debug(ctx, "Creating worker group", map[string]interface{}{
		"endpoint_id": model.EndpointID.ValueInt64(),
	})

	workerGroup, err := r.client.WorkerGroups.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Worker Group",
			fmt.Sprintf("Could not create worker group: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Worker group created", map[string]interface{}{
		"id": workerGroup.ID,
	})

	// Map API response to model
	readWorkerGroupIntoModel(workerGroup, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the worker group state from the Vast.ai API.
func (r *WorkerGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model WorkerGroupResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	readTimeout, diags := model.Timeouts.Read(ctx, 2*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Worker Group ID",
			fmt.Sprintf("Could not parse worker group ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading worker group", map[string]interface{}{
		"id": id,
	})

	// Read via list (no single-GET endpoint per Pitfall 1)
	groups, err := r.client.WorkerGroups.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Worker Group",
			fmt.Sprintf("Could not list worker groups: %s", err),
		)
		return
	}

	// Find matching worker group by ID
	var found *client.WorkerGroup
	for i := range groups {
		if groups[i].ID == id {
			found = &groups[i]
			break
		}
	}

	if found == nil {
		// Worker group not found -- remove from state
		tflog.Warn(ctx, "Worker group not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	readWorkerGroupIntoModel(found, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update modifies an existing worker group via the Vast.ai API.
func (r *WorkerGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model WorkerGroupResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	updateTimeout, diags := model.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Get the current ID from state
	var state WorkerGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Worker Group ID",
			fmt.Sprintf("Could not parse worker group ID %q as integer: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Build update request
	updateReq := &client.UpdateWorkerGroupRequest{
		EndpointID: int(model.EndpointID.ValueInt64()),
	}

	if !model.EndpointName.IsNull() && !model.EndpointName.IsUnknown() {
		updateReq.EndpointName = model.EndpointName.ValueString()
	}
	if !model.TemplateHash.IsNull() && !model.TemplateHash.IsUnknown() {
		updateReq.TemplateHash = model.TemplateHash.ValueString()
	}
	if !model.TemplateID.IsNull() && !model.TemplateID.IsUnknown() {
		updateReq.TemplateID = int(model.TemplateID.ValueInt64())
	}
	if !model.SearchParams.IsNull() && !model.SearchParams.IsUnknown() {
		updateReq.SearchParams = model.SearchParams.ValueString()
	}
	if !model.LaunchArgs.IsNull() && !model.LaunchArgs.IsUnknown() {
		updateReq.LaunchArgs = model.LaunchArgs.ValueString()
	}
	if !model.GpuRAM.IsNull() && !model.GpuRAM.IsUnknown() {
		v := model.GpuRAM.ValueFloat64()
		updateReq.GpuRAM = &v
	}
	if !model.TestWorkers.IsNull() && !model.TestWorkers.IsUnknown() {
		v := int(model.TestWorkers.ValueInt64())
		updateReq.TestWorkers = &v
	}
	if !model.ColdWorkers.IsNull() && !model.ColdWorkers.IsUnknown() {
		v := int(model.ColdWorkers.ValueInt64())
		updateReq.ColdWorkers = &v
	}

	tflog.Debug(ctx, "Updating worker group", map[string]interface{}{
		"id": id,
	})

	if err := r.client.WorkerGroups.Update(ctx, id, updateReq); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Worker Group",
			fmt.Sprintf("Could not update worker group %d: %s", id, err),
		)
		return
	}

	// Read-after-write: re-read via List to get updated state
	groups, err := r.client.WorkerGroups.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Worker Group After Update",
			fmt.Sprintf("Could not list worker groups after update: %s", err),
		)
		return
	}

	var found *client.WorkerGroup
	for i := range groups {
		if groups[i].ID == id {
			found = &groups[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Error Reading Worker Group After Update",
			fmt.Sprintf("Worker group %d not found after update", id),
		)
		return
	}

	readWorkerGroupIntoModel(found, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete removes a worker group via the Vast.ai API.
func (r *WorkerGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model WorkerGroupResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	deleteTimeout, diags := model.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Worker Group ID",
			fmt.Sprintf("Could not parse worker group ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting worker group", map[string]interface{}{
		"id": id,
	})

	if err := r.client.WorkerGroups.Delete(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Worker Group",
			fmt.Sprintf("Could not delete worker group %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Worker group deleted", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing worker group by its ID.
// Usage: terraform import vastai_worker_group.example <worker_group_id>
func (r *WorkerGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readWorkerGroupIntoModel maps a client.WorkerGroup to a WorkerGroupResourceModel.
// Preserves non-API fields (Timeouts). Sets null for empty/zero API values.
func readWorkerGroupIntoModel(wg *client.WorkerGroup, model *WorkerGroupResourceModel) {
	model.ID = types.StringValue(strconv.Itoa(wg.ID))
	model.EndpointID = types.Int64Value(int64(wg.EndpointID))
	model.EndpointName = types.StringValue(wg.EndpointName)

	// Template fields: set null for empty/zero values
	if wg.TemplateHash != "" {
		model.TemplateHash = types.StringValue(wg.TemplateHash)
	} else {
		model.TemplateHash = types.StringNull()
	}

	if wg.TemplateID != 0 {
		model.TemplateID = types.Int64Value(int64(wg.TemplateID))
	} else {
		model.TemplateID = types.Int64Null()
	}

	// Optional string fields: set null for empty values
	if wg.SearchParams != "" {
		model.SearchParams = types.StringValue(wg.SearchParams)
	} else {
		model.SearchParams = types.StringNull()
	}

	if wg.LaunchArgs != "" {
		model.LaunchArgs = types.StringValue(wg.LaunchArgs)
	} else {
		model.LaunchArgs = types.StringNull()
	}

	// Optional numeric fields: set null for zero values
	if wg.GpuRAM != 0 {
		model.GpuRAM = types.Float64Value(wg.GpuRAM)
	} else {
		model.GpuRAM = types.Float64Null()
	}

	// TestWorkers is always set (Computed)
	model.TestWorkers = types.Int64Value(int64(wg.TestWorkers))

	// ColdWorkers: set null for zero
	if wg.ColdWorkers != 0 {
		model.ColdWorkers = types.Int64Value(int64(wg.ColdWorkers))
	} else {
		model.ColdWorkers = types.Int64Null()
	}
}
