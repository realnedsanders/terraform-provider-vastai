package overlay

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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

// Ensure OverlayResource satisfies the required interfaces.
var (
	_ resource.Resource                = &OverlayResource{}
	_ resource.ResourceWithImportState = &OverlayResource{}
)

// OverlayResource defines the resource implementation.
type OverlayResource struct {
	client *client.VastAIClient
}

// NewOverlayResource returns a new overlay resource instance.
func NewOverlayResource() resource.Resource {
	return &OverlayResource{}
}

// Metadata returns the resource type name.
func (r *OverlayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_overlay"
}

// Schema defines the schema for the resource.
func (r *OverlayResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai overlay network. Overlays provide virtual networking on top of clusters, " +
			"allowing instances to communicate over a private network.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique numeric identifier of the overlay.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the overlay network.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "ID of the cluster this overlay belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal_subnet": schema.StringAttribute{
				Description: "Subnet assigned to the overlay.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

// Configure adds the provider-configured client to the resource.
func (r *OverlayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new overlay network.
func (r *OverlayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model OverlayResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set create timeout
	createTimeout, diags := model.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	clusterID, err := strconv.Atoi(model.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Cluster ID",
			fmt.Sprintf("Could not parse cluster_id %q as integer: %s", model.ClusterID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Creating overlay", map[string]interface{}{
		"name":       model.Name.ValueString(),
		"cluster_id": clusterID,
	})

	overlay, err := r.client.Overlays.Create(ctx, clusterID, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Overlay",
			fmt.Sprintf("Could not create overlay: %s", err),
		)
		return
	}

	// Map response to model
	model.ID = types.StringValue(strconv.Itoa(overlay.OverlayID))
	model.InternalSubnet = types.StringValue(overlay.InternalSubnet)

	tflog.Debug(ctx, "Created overlay", map[string]interface{}{
		"id": overlay.OverlayID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the overlay state from the API.
func (r *OverlayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model OverlayResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set read timeout
	readTimeout, diags := model.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Overlay ID",
			fmt.Sprintf("Could not parse overlay ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading overlay", map[string]interface{}{
		"id": id,
	})

	// List all overlays and find the one matching our ID
	overlays, err := r.client.Overlays.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Overlay",
			fmt.Sprintf("Could not list overlays: %s", err),
		)
		return
	}

	var found *client.Overlay
	for i := range overlays {
		if overlays[i].OverlayID == id {
			found = &overlays[i]
			break
		}
	}

	if found == nil {
		tflog.Warn(ctx, "Overlay not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	model.Name = types.StringValue(found.Name)
	model.ClusterID = types.StringValue(strconv.Itoa(found.ClusterID))
	model.InternalSubnet = types.StringValue(found.InternalSubnet)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is not supported -- all fields are ForceNew.
func (r *OverlayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Overlay resources cannot be updated in-place. All attributes require replacement.",
	)
}

// Delete deletes an overlay network.
func (r *OverlayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model OverlayResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set delete timeout
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
			"Error Parsing Overlay ID",
			fmt.Sprintf("Could not parse overlay ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting overlay", map[string]interface{}{
		"id": id,
	})

	if err := r.client.Overlays.Delete(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Overlay",
			fmt.Sprintf("Could not delete overlay %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Deleted overlay", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing overlay by its numeric ID.
// Usage: terraform import vastai_overlay.example <id>
func (r *OverlayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
