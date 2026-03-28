package networkvolume

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
	_ resource.Resource                = &NetworkVolumeResource{}
	_ resource.ResourceWithConfigure   = &NetworkVolumeResource{}
	_ resource.ResourceWithImportState = &NetworkVolumeResource{}
)

// NetworkVolumeResource defines the resource implementation.
type NetworkVolumeResource struct {
	client *client.VastAIClient
}

// NewNetworkVolumeResource creates a new network volume resource instance.
func NewNetworkVolumeResource() resource.Resource {
	return &NetworkVolumeResource{}
}

// Metadata returns the resource type name.
func (r *NetworkVolumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_volume"
}

// Schema defines the schema for the network volume resource.
func (r *NetworkVolumeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai network volume. Network volumes provide network-attached persistent storage " +
			"that can be mounted across instances. Network volumes are created from network volume offers " +
			"(use the vastai_network_volume_offers data source to search for available offers).",

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
				Description: "ID of the network volume offer to create from (from vastai_network_volume_offers data source). " +
					"Changes force replacement. " +
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
				Description: "Network volume size in GB. Changes force replacement. " +
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
				Description: "Optional name/label for the network volume. Changes force replacement.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Computed fields from API (read-only)
			"status": schema.StringAttribute{
				Description: "Current status of the network volume (e.g., 'active', 'inactive').",
				Computed:    true,
			},
			"disk_space": schema.Float64Attribute{
				Description: "Actual allocated disk space in GB.",
				Computed:    true,
			},
			"machine_id": schema.Int64Attribute{
				Description: "ID of the machine hosting this network volume.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"geolocation": schema.StringAttribute{
				Description: "Geographic location of the machine hosting this network volume.",
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
				Description: "Name of the physical disk hosting this network volume.",
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
func (r *NetworkVolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new network volume via the Vast.ai API.
func (r *NetworkVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model NetworkVolumeResourceModel

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

	if model.OfferID.IsNull() || model.OfferID.IsUnknown() {
		resp.Diagnostics.AddError("Missing offer_id", "offer_id is required when creating a network volume from an offer")
		return
	}
	if model.Size.IsNull() || model.Size.IsUnknown() {
		resp.Diagnostics.AddError("Missing size", "size is required when creating a network volume from an offer")
		return
	}

	createReq := &client.CreateNetworkVolumeRequest{
		Size:    int(model.Size.ValueInt64()),
		OfferID: int(model.OfferID.ValueInt64()),
	}

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		createReq.Name = model.Name.ValueString()
	}

	tflog.Debug(ctx, "Creating network volume from offer", map[string]interface{}{
		"offer_id": model.OfferID.ValueInt64(),
		"size":     model.Size.ValueInt64(),
	})

	vol, err := r.client.NetworkVolumes.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Network Volume",
			fmt.Sprintf("Could not create network volume: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Network volume created", map[string]interface{}{
		"id": vol.ID,
	})

	// Set ID and populate computed fields
	model.ID = types.StringValue(strconv.FormatInt(int64(vol.ID), 10))
	readNetworkVolumeIntoModel(vol, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the network volume state from the Vast.ai API.
func (r *NetworkVolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model NetworkVolumeResourceModel

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
			"Error Parsing Network Volume ID",
			fmt.Sprintf("Could not parse network volume ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading network volume", map[string]interface{}{
		"id": id,
	})

	// Read via list (no single-GET endpoint per Pitfall 6)
	// Uses GET /volumes?owner=me&type=network_volume
	volumes, err := r.client.NetworkVolumes.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Network Volume",
			fmt.Sprintf("Could not list network volumes: %s", err),
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
		// Network volume not found -- remove from state
		tflog.Warn(ctx, "Network volume not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	readNetworkVolumeIntoModel(found, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update returns an error -- network volumes are immutable.
// All mutable attributes are ForceNew so this should never be called.
func (r *NetworkVolumeResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Network Volume Update Not Supported",
		"Network volumes are immutable; changes require replacement. "+
			"All creation-time attributes are marked as ForceNew.",
	)
}

// Delete removes a network volume via the Vast.ai API.
func (r *NetworkVolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model NetworkVolumeResourceModel

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
			"Error Parsing Network Volume ID",
			fmt.Sprintf("Could not parse network volume ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting network volume", map[string]interface{}{
		"id": id,
	})

	if err := r.client.NetworkVolumes.Delete(ctx, int(id)); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Network Volume",
			fmt.Sprintf("Could not delete network volume %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Network volume deleted", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing network volume by its contract ID.
// Usage: terraform import vastai_network_volume.example <network_volume_id>.
func (r *NetworkVolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readNetworkVolumeIntoModel maps a client.Volume to a NetworkVolumeResourceModel.
// Preserves non-API fields (Timeouts, creation-time attributes).
func readNetworkVolumeIntoModel(vol *client.Volume, model *NetworkVolumeResourceModel) {
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
