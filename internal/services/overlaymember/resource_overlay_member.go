package overlaymember

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure OverlayMemberResource satisfies the required interfaces.
var (
	_ resource.Resource                = &OverlayMemberResource{}
	_ resource.ResourceWithImportState = &OverlayMemberResource{}
)

// OverlayMemberResource defines the resource implementation.
type OverlayMemberResource struct {
	client *client.VastAIClient
}

// NewOverlayMemberResource returns a new overlay member resource instance.
func NewOverlayMemberResource() resource.Resource {
	return &OverlayMemberResource{}
}

// Metadata returns the resource type name.
func (r *OverlayMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_overlay_member"
}

// Schema defines the schema for the resource.
func (r *OverlayMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Joins an instance to a Vast.ai overlay network. NOTE: Individual instance removal from an " +
			"overlay is not supported by the API. Destroying this resource removes it from Terraform state only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID: overlay_id/instance_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"overlay_name": schema.StringAttribute{
				Description: "Name of the overlay to join.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"overlay_id": schema.StringAttribute{
				Description: "ID of the overlay.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"instance_id": schema.StringAttribute{
				Description: "ID of the instance to join to the overlay.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
func (r *OverlayMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create joins an instance to an overlay network.
func (r *OverlayMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model OverlayMemberResourceModel

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

	instanceID, err := strconv.Atoi(model.InstanceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Instance ID",
			fmt.Sprintf("Could not parse instance_id %q as integer: %s", model.InstanceID.ValueString(), err),
		)
		return
	}

	overlayName := model.OverlayName.ValueString()

	tflog.Debug(ctx, "Joining instance to overlay", map[string]interface{}{
		"overlay_name": overlayName,
		"instance_id":  instanceID,
	})

	if err := r.client.Overlays.JoinInstance(ctx, overlayName, instanceID); err != nil {
		resp.Diagnostics.AddError(
			"Error Joining Instance to Overlay",
			fmt.Sprintf("Could not join instance %d to overlay %q: %s", instanceID, overlayName, err),
		)
		return
	}

	// Read back to get overlay_id
	overlays, err := r.client.Overlays.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Overlays After Join",
			fmt.Sprintf("Could not list overlays after joining instance: %s", err),
		)
		return
	}

	var overlayID int
	found := false
	for _, o := range overlays {
		if o.Name == overlayName {
			overlayID = o.OverlayID
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Overlay Not Found After Join",
			fmt.Sprintf("Could not find overlay %q after joining instance", overlayName),
		)
		return
	}

	model.OverlayID = types.StringValue(strconv.Itoa(overlayID))
	model.ID = types.StringValue(fmt.Sprintf("%d/%d", overlayID, instanceID))

	tflog.Debug(ctx, "Joined instance to overlay", map[string]interface{}{
		"id": model.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the overlay member state from the API.
func (r *OverlayMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model OverlayMemberResourceModel

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

	overlayID, err := strconv.Atoi(model.OverlayID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Overlay ID",
			fmt.Sprintf("Could not parse overlay_id %q as integer: %s", model.OverlayID.ValueString(), err),
		)
		return
	}

	instanceID, err := strconv.Atoi(model.InstanceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Instance ID",
			fmt.Sprintf("Could not parse instance_id %q as integer: %s", model.InstanceID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading overlay member", map[string]interface{}{
		"overlay_id":  overlayID,
		"instance_id": instanceID,
	})

	// List all overlays and find the matching overlay
	overlays, err := r.client.Overlays.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Overlay Member",
			fmt.Sprintf("Could not list overlays: %s", err),
		)
		return
	}

	var foundOverlay *client.Overlay
	for i := range overlays {
		if overlays[i].OverlayID == overlayID {
			foundOverlay = &overlays[i]
			break
		}
	}

	if foundOverlay == nil {
		tflog.Warn(ctx, "Overlay not found, removing member from state", map[string]interface{}{
			"overlay_id": overlayID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Check if instance is in the overlay's instances list
	instanceFound := false
	for _, instID := range foundOverlay.Instances {
		if instID == instanceID {
			instanceFound = true
			break
		}
	}

	if !instanceFound {
		tflog.Warn(ctx, "Instance not found in overlay, removing from state", map[string]interface{}{
			"overlay_id":  overlayID,
			"instance_id": instanceID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	model.OverlayName = types.StringValue(foundOverlay.Name)
	model.OverlayID = types.StringValue(strconv.Itoa(foundOverlay.OverlayID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is not supported -- all fields are ForceNew.
func (r *OverlayMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Overlay member resources cannot be updated in-place. All attributes require replacement.",
	)
}

// Delete is a no-op -- the Vast.ai API does not support removing individual instances from overlays.
// The overlay membership is removed from Terraform state only.
func (r *OverlayMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model OverlayMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Destroying overlay member (no-op: removing from state only)", map[string]interface{}{
		"id": model.ID.ValueString(),
	})

	// NO API CALL -- Pitfall 9: no remove-instance-from-overlay endpoint
	resp.Diagnostics.AddWarning(
		"Overlay Member Not Removed",
		"The overlay membership was removed from Terraform state but the instance was NOT removed from the overlay. "+
			"Individual instance removal is not supported by the Vast.ai API. Delete the overlay to remove all instances.",
	)
}

// ImportState imports an existing overlay member by "overlay_id/instance_id" composite string.
// Usage: terraform import vastai_overlay_member.example <overlay_id>/<instance_id>
func (r *OverlayMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'overlay_id/instance_id', got: %q", req.ID),
		)
		return
	}

	overlayID := parts[0]
	instanceID := parts[1]

	// Validate both parts are valid integers
	if _, err := strconv.Atoi(overlayID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Overlay ID",
			fmt.Sprintf("overlay_id %q is not a valid integer: %s", overlayID, err),
		)
		return
	}
	if _, err := strconv.Atoi(instanceID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Instance ID",
			fmt.Sprintf("instance_id %q is not a valid integer: %s", instanceID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("overlay_id"), overlayID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), instanceID)...)
}
