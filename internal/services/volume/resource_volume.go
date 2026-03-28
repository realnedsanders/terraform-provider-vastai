package volume

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
	_ resource.Resource                = &VolumeResource{}
	_ resource.ResourceWithConfigure   = &VolumeResource{}
	_ resource.ResourceWithImportState = &VolumeResource{}
)

// VolumeResource defines the resource implementation.
type VolumeResource struct {
	client *client.VastAIClient
}

// NewVolumeResource creates a new volume resource instance.
func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

// Metadata returns the resource type name.
func (r *VolumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

// Schema defines the schema for the volume resource.
func (r *VolumeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai local volume. Volumes provide persistent storage that can be " +
			"attached to instances. Volumes are created from volume offers (use the vastai_volume_offers " +
			"data source to search for available offers) or cloned from existing volumes.",

		Attributes: map[string]schema.Attribute{
			// Primary identifier
			"id": schema.StringAttribute{
				Description: "Volume contract ID. Used as the primary identifier for reads, deletes, and imports.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Creation-time attributes (all RequiresReplace)
			"offer_id": schema.Int64Attribute{
				Description: "ID of the volume offer to create from (from vastai_volume_offers data source). " +
					"When cloning, this is the destination offer ID. Changes force replacement. " +
					"This is a creation-time attribute not returned by the API; preserved in state via UseStateForUnknown.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"size": schema.Int64Attribute{
				Description: "Volume size in GB. Changes force replacement. " +
					"This is a creation-time attribute not returned by the API; preserved in state via UseStateForUnknown.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Description: "Optional name/label for the volume. Changes force replacement.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"clone_from_id": schema.Int64Attribute{
				Description: "Source volume ID to clone from. When set, creates by cloning " +
					"(POST /volumes/copy/) instead of creating from offer. Changes force replacement.",
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"disable_compression": schema.BoolAttribute{
				Description: "Disable compression during clone. Only used when clone_from_id is set. Changes force replacement.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},

			// Computed fields from API (read-only)
			"status": schema.StringAttribute{
				Description: "Current status of the volume (e.g., 'active', 'inactive').",
				Computed:    true,
			},
			"disk_space": schema.Float64Attribute{
				Description: "Actual allocated disk space in GB.",
				Computed:    true,
			},
			"machine_id": schema.Int64Attribute{
				Description: "ID of the machine hosting this volume.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"geolocation": schema.StringAttribute{
				Description: "Geographic location of the machine hosting this volume.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"inet_up": schema.Float64Attribute{
				Description: "Internet upload speed in Mbps.",
				Computed:    true,
			},
			"inet_down": schema.Float64Attribute{
				Description: "Internet download speed in Mbps.",
				Computed:    true,
			},
			"reliability": schema.Float64Attribute{
				Description: "Host reliability score (0.0 to 1.0).",
				Computed:    true,
			},
			"disk_name": schema.StringAttribute{
				Description: "Name of the physical disk hosting this volume.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"driver_version": schema.StringAttribute{
				Description: "Driver version on the host machine.",
				Computed:    true,
			},
			"host_id": schema.Int64Attribute{
				Description: "ID of the host machine.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"verification": schema.StringAttribute{
				Description: "Host verification status.",
				Computed:    true,
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

// Configure adds the provider configured client to the resource.
func (r *VolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new volume via the Vast.ai API.
func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model VolumeResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout (5min default for create, 30min for clone via user override)
	createTimeout, diags := model.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	if !model.CloneFromID.IsNull() && !model.CloneFromID.IsUnknown() {
		// Clone path: POST /volumes/copy/
		r.createViaClone(ctx, &model, resp)
	} else {
		// Standard path: PUT /volumes/
		r.createFromOffer(ctx, &model, resp)
	}
}

// createFromOffer creates a volume from an offer (standard creation path).
func (r *VolumeResource) createFromOffer(ctx context.Context, model *VolumeResourceModel, resp *resource.CreateResponse) {
	if model.OfferID.IsNull() || model.OfferID.IsUnknown() {
		resp.Diagnostics.AddError("Missing offer_id", "offer_id is required when creating a volume from an offer")
		return
	}
	if model.Size.IsNull() || model.Size.IsUnknown() {
		resp.Diagnostics.AddError("Missing size", "size is required when creating a volume from an offer")
		return
	}

	createReq := &client.CreateVolumeRequest{
		Size:    int(model.Size.ValueInt64()),
		OfferID: int(model.OfferID.ValueInt64()),
	}

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		createReq.Name = model.Name.ValueString()
	}

	tflog.Debug(ctx, "Creating volume from offer", map[string]interface{}{
		"offer_id": model.OfferID.ValueInt64(),
		"size":     model.Size.ValueInt64(),
	})

	vol, err := r.client.Volumes.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Volume",
			fmt.Sprintf("Could not create volume: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Volume created", map[string]interface{}{
		"id": vol.ID,
	})

	// Set ID and populate computed fields
	model.ID = types.StringValue(strconv.FormatInt(int64(vol.ID), 10))
	readVolumeIntoModel(vol, model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// createViaClone creates a volume by cloning an existing one.
// Per D-01: clone needs both source ID (clone_from_id) AND destination offer ID (offer_id).
func (r *VolumeResource) createViaClone(ctx context.Context, model *VolumeResourceModel, resp *resource.CreateResponse) {
	cloneReq := &client.CloneVolumeRequest{
		SourceID:    int(model.CloneFromID.ValueInt64()),
		DestOfferID: int(model.OfferID.ValueInt64()),
		Size:        float64(model.Size.ValueInt64()),
	}

	if !model.DisableCompression.IsNull() && !model.DisableCompression.IsUnknown() {
		cloneReq.DisableCompression = model.DisableCompression.ValueBool()
	}

	tflog.Debug(ctx, "Cloning volume", map[string]interface{}{
		"source_id":     model.CloneFromID.ValueInt64(),
		"dest_offer_id": model.OfferID.ValueInt64(),
		"size":          model.Size.ValueInt64(),
	})

	err := r.client.Volumes.Clone(ctx, cloneReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Cloning Volume",
			fmt.Sprintf("Could not clone volume: %s", err),
		)
		return
	}

	// Clone doesn't return the new volume ID directly.
	// Read back via List to find the newly created volume.
	volumes, err := r.client.Volumes.List(ctx, "local_volume")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Volume After Clone",
			fmt.Sprintf("Volume was cloned but could not read back: %s", err),
		)
		return
	}

	// Find the most recently created volume (highest ID) as the clone result.
	// This is a best-effort approach since clone doesn't return the new ID.
	var newest *client.Volume
	for i := range volumes {
		if newest == nil || volumes[i].ID > newest.ID {
			newest = &volumes[i]
		}
	}

	if newest == nil {
		resp.Diagnostics.AddError(
			"Error Finding Cloned Volume",
			"Volume was cloned but could not be found in the volume list.",
		)
		return
	}

	tflog.Debug(ctx, "Volume cloned", map[string]interface{}{
		"id": newest.ID,
	})

	// Set ID and populate computed fields
	model.ID = types.StringValue(strconv.FormatInt(int64(newest.ID), 10))
	readVolumeIntoModel(newest, model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Read refreshes the volume state from the Vast.ai API.
func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model VolumeResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	readTimeout, diags := model.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	id, err := strconv.ParseInt(model.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Volume ID",
			fmt.Sprintf("Could not parse volume ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading volume", map[string]interface{}{
		"id": id,
	})

	// Read via list (no single-GET endpoint per Pitfall 3)
	volumes, err := r.client.Volumes.List(ctx, "local_volume")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Volume",
			fmt.Sprintf("Could not list volumes: %s", err),
		)
		return
	}

	// Find matching volume by ID
	var found *client.Volume
	for i := range volumes {
		if int64(volumes[i].ID) == id {
			found = &volumes[i]
			break
		}
	}

	if found == nil {
		// Volume not found -- remove from state
		tflog.Warn(ctx, "Volume not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	readVolumeIntoModel(found, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update returns an error -- volumes are immutable.
// All mutable attributes are ForceNew so this should never be called.
func (r *VolumeResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Volume Update Not Supported",
		"Volumes are immutable; changes require replacement. "+
			"All creation-time attributes are marked as ForceNew.",
	)
}

// Delete removes a volume via the Vast.ai API.
func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model VolumeResourceModel

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

	id, err := strconv.ParseInt(model.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Volume ID",
			fmt.Sprintf("Could not parse volume ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting volume", map[string]interface{}{
		"id": id,
	})

	if err := r.client.Volumes.Delete(ctx, int(id)); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Volume",
			fmt.Sprintf("Could not delete volume %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Volume deleted", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing volume by its contract ID.
// Usage: terraform import vastai_volume.example <volume_id>.
func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readVolumeIntoModel maps a client.Volume to a VolumeResourceModel.
// Preserves non-API fields (Timeouts, creation-time attributes).
func readVolumeIntoModel(vol *client.Volume, model *VolumeResourceModel) {
	model.Status = types.StringValue(vol.Status)
	model.DiskSpace = types.Float64Value(vol.DiskSpace)
	model.MachineID = types.Int64Value(int64(vol.MachineID))
	model.Geolocation = types.StringValue(vol.Geolocation)
	model.InetUp = types.Float64Value(vol.InetUp)
	model.InetDown = types.Float64Value(vol.InetDown)
	model.Reliability = types.Float64Value(vol.Reliability)
	model.DiskName = types.StringValue(vol.DiskName)
	model.DriverVersion = types.StringValue(vol.DriverVersion)
	model.HostID = types.Int64Value(int64(vol.HostID))
	model.Verification = types.StringValue(vol.Verification)
}
