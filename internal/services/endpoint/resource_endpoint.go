package endpoint

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &EndpointResource{}
	_ resource.ResourceWithConfigure   = &EndpointResource{}
	_ resource.ResourceWithImportState = &EndpointResource{}
)

// EndpointResource defines the resource implementation.
type EndpointResource struct {
	client *client.VastAIClient
}

// NewEndpointResource creates a new endpoint resource instance.
func NewEndpointResource() resource.Resource {
	return &EndpointResource{}
}

// Metadata returns the resource type name.
func (r *EndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

// Schema defines the schema for the endpoint resource.
func (r *EndpointResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai serverless endpoint. Endpoints provide autoscaling inference " +
			"serving with configurable load balancing, cold start behavior, and worker scaling. " +
			"Autoscaling is configured directly on the endpoint resource (SRVL-03).",

		Attributes: map[string]schema.Attribute{
			// Primary identifier
			"id": schema.StringAttribute{
				Description: "Endpoint ID. Used as the primary identifier for reads, updates, deletes, and imports.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Required attributes
			"endpoint_name": schema.StringAttribute{
				Description: "Name of the serverless endpoint.",
				Required:    true,
			},

			// Optional+Computed autoscaling attributes (server defaults per D-04, D-05)
			"min_load": schema.Float64Attribute{
				Description: "Minimum floor load in perf units/s (default: 0).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"min_cold_load": schema.Float64Attribute{
				Description: "Minimum floor load allowing cold workers (default: 0).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"target_util": schema.Float64Attribute{
				Description: "Target capacity utilization fraction (default: 0.9).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Float64{
					float64validator.Between(0, 1),
				},
			},
			"cold_mult": schema.Float64Attribute{
				Description: "Cold capacity as multiple of hot capacity (default: 2.5).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(1.0),
				},
			},
			"cold_workers": schema.Int64Attribute{
				Description: "Minimum cold workers when no load (default: 5).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"max_workers": schema.Int64Attribute{
				Description: "Maximum workers the endpoint can have (default: 20).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},

			// Optional+Computed (update-only per Pitfall 6: new endpoints start as active)
			"endpoint_state": schema.StringAttribute{
				Description: "Endpoint runtime state (active, suspended, stopped). New endpoints start as active.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("active", "suspended", "stopped"),
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
func (r *EndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new serverless endpoint via the Vast.ai API.
func (r *EndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model EndpointResourceModel

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
	createReq := &client.CreateEndpointRequest{
		EndpointName: model.EndpointName.ValueString(),
	}

	if !model.MinLoad.IsNull() && !model.MinLoad.IsUnknown() {
		createReq.MinLoad = model.MinLoad.ValueFloat64()
	}
	if !model.MinColdLoad.IsNull() && !model.MinColdLoad.IsUnknown() {
		createReq.MinColdLoad = model.MinColdLoad.ValueFloat64()
	}
	if !model.TargetUtil.IsNull() && !model.TargetUtil.IsUnknown() {
		createReq.TargetUtil = model.TargetUtil.ValueFloat64()
	}
	if !model.ColdMult.IsNull() && !model.ColdMult.IsUnknown() {
		createReq.ColdMult = model.ColdMult.ValueFloat64()
	}
	if !model.ColdWorkers.IsNull() && !model.ColdWorkers.IsUnknown() {
		createReq.ColdWorkers = int(model.ColdWorkers.ValueInt64())
	}
	if !model.MaxWorkers.IsNull() && !model.MaxWorkers.IsUnknown() {
		createReq.MaxWorkers = int(model.MaxWorkers.ValueInt64())
	}

	tflog.Debug(ctx, "Creating endpoint", map[string]interface{}{
		"endpoint_name": model.EndpointName.ValueString(),
	})

	endpoint, err := r.client.Endpoints.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Endpoint",
			fmt.Sprintf("Could not create endpoint: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Endpoint created", map[string]interface{}{
		"id": endpoint.ID,
	})

	// Map API response to model
	readEndpointIntoModel(endpoint, &model)

	// If endpoint_state is set and not "active", issue an Update call after create (Pitfall 6)
	if !model.EndpointState.IsNull() && !model.EndpointState.IsUnknown() {
		desiredState := model.EndpointState.ValueString()
		if desiredState != "" && desiredState != "active" {
			updateReq := &client.UpdateEndpointRequest{
				EndpointState: desiredState,
			}
			id, parseErr := strconv.Atoi(model.ID.ValueString())
			if parseErr != nil {
				resp.Diagnostics.AddError(
					"Error Parsing Endpoint ID",
					fmt.Sprintf("Could not parse endpoint ID %q as integer: %s", model.ID.ValueString(), parseErr),
				)
				return
			}
			if err := r.client.Endpoints.Update(ctx, id, updateReq); err != nil {
				resp.Diagnostics.AddError(
					"Error Setting Endpoint State",
					fmt.Sprintf("Could not set endpoint state to %q after creation: %s", desiredState, err),
				)
				return
			}
			model.EndpointState = types.StringValue(desiredState)
		}
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the endpoint state from the Vast.ai API.
func (r *EndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model EndpointResourceModel

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
			"Error Parsing Endpoint ID",
			fmt.Sprintf("Could not parse endpoint ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading endpoint", map[string]interface{}{
		"id": id,
	})

	// Read via list (no single-GET endpoint per Pitfall 1)
	endpoints, err := r.client.Endpoints.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Endpoint",
			fmt.Sprintf("Could not list endpoints: %s", err),
		)
		return
	}

	// Find matching endpoint by ID
	var found *client.Endpoint
	for i := range endpoints {
		if endpoints[i].ID == id {
			found = &endpoints[i]
			break
		}
	}

	if found == nil {
		// Endpoint not found -- remove from state
		tflog.Warn(ctx, "Endpoint not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	readEndpointIntoModel(found, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update modifies an existing endpoint via the Vast.ai API.
func (r *EndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model EndpointResourceModel

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
	var state EndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Endpoint ID",
			fmt.Sprintf("Could not parse endpoint ID %q as integer: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Build update request with pointer fields for partial updates
	updateReq := &client.UpdateEndpointRequest{
		EndpointName: model.EndpointName.ValueString(),
	}

	if !model.MinLoad.IsNull() && !model.MinLoad.IsUnknown() {
		v := model.MinLoad.ValueFloat64()
		updateReq.MinLoad = &v
	}
	if !model.MinColdLoad.IsNull() && !model.MinColdLoad.IsUnknown() {
		v := model.MinColdLoad.ValueFloat64()
		updateReq.MinColdLoad = &v
	}
	if !model.TargetUtil.IsNull() && !model.TargetUtil.IsUnknown() {
		v := model.TargetUtil.ValueFloat64()
		updateReq.TargetUtil = &v
	}
	if !model.ColdMult.IsNull() && !model.ColdMult.IsUnknown() {
		v := model.ColdMult.ValueFloat64()
		updateReq.ColdMult = &v
	}
	if !model.ColdWorkers.IsNull() && !model.ColdWorkers.IsUnknown() {
		v := int(model.ColdWorkers.ValueInt64())
		updateReq.ColdWorkers = &v
	}
	if !model.MaxWorkers.IsNull() && !model.MaxWorkers.IsUnknown() {
		v := int(model.MaxWorkers.ValueInt64())
		updateReq.MaxWorkers = &v
	}
	if !model.EndpointState.IsNull() && !model.EndpointState.IsUnknown() {
		updateReq.EndpointState = model.EndpointState.ValueString()
	}

	tflog.Debug(ctx, "Updating endpoint", map[string]interface{}{
		"id":   id,
		"name": model.EndpointName.ValueString(),
	})

	if err := r.client.Endpoints.Update(ctx, id, updateReq); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Endpoint",
			fmt.Sprintf("Could not update endpoint %d: %s", id, err),
		)
		return
	}

	// Read-after-write: re-read via List to get updated state
	endpoints, err := r.client.Endpoints.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Endpoint After Update",
			fmt.Sprintf("Could not list endpoints after update: %s", err),
		)
		return
	}

	var found *client.Endpoint
	for i := range endpoints {
		if endpoints[i].ID == id {
			found = &endpoints[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Error Reading Endpoint After Update",
			fmt.Sprintf("Endpoint %d not found after update", id),
		)
		return
	}

	readEndpointIntoModel(found, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete removes an endpoint via the Vast.ai API.
func (r *EndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model EndpointResourceModel

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
			"Error Parsing Endpoint ID",
			fmt.Sprintf("Could not parse endpoint ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting endpoint", map[string]interface{}{
		"id": id,
	})

	if err := r.client.Endpoints.Delete(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Endpoint",
			fmt.Sprintf("Could not delete endpoint %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Endpoint deleted", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing endpoint by its ID.
// Usage: terraform import vastai_endpoint.example <endpoint_id>.
func (r *EndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readEndpointIntoModel maps a client.Endpoint to an EndpointResourceModel.
// Preserves non-API fields (Timeouts).
func readEndpointIntoModel(endpoint *client.Endpoint, model *EndpointResourceModel) {
	model.ID = types.StringValue(strconv.Itoa(endpoint.ID))
	model.EndpointName = types.StringValue(endpoint.EndpointName)
	model.MinLoad = types.Float64Value(endpoint.MinLoad)
	model.MinColdLoad = types.Float64Value(endpoint.MinColdLoad)
	model.TargetUtil = types.Float64Value(endpoint.TargetUtil)
	model.ColdMult = types.Float64Value(endpoint.ColdMult)
	model.ColdWorkers = types.Int64Value(int64(endpoint.ColdWorkers))
	model.MaxWorkers = types.Int64Value(int64(endpoint.MaxWorkers))
	model.EndpointState = types.StringValue(endpoint.EndpointState)
}
