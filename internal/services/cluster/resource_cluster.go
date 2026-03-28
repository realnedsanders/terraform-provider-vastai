package cluster

import (
	"context"
	"fmt"
	"strconv"
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

// Ensure ClusterResource satisfies the required interfaces.
var (
	_ resource.Resource                = &ClusterResource{}
	_ resource.ResourceWithImportState = &ClusterResource{}
)

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	client *client.VastAIClient
}

// NewClusterResource returns a new cluster resource instance.
func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// Metadata returns the resource type name.
func (r *ClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

// Schema defines the schema for the resource.
func (r *ClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai cluster. Clusters group machines into a private network with a designated manager node.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique numeric identifier of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet": schema.StringAttribute{
				Description: "Subnet for the cluster (e.g., 10.0.0.0/24).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"manager_id": schema.StringAttribute{
				Description: "Machine ID to be the cluster manager.",
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
func (r *ClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new cluster.
func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model ClusterResourceModel

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

	managerID, err := strconv.Atoi(model.ManagerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Manager ID",
			fmt.Sprintf("Could not parse manager_id %q as integer: %s", model.ManagerID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Creating cluster", map[string]interface{}{
		"subnet":     model.Subnet.ValueString(),
		"manager_id": managerID,
	})

	cluster, err := r.client.Clusters.Create(ctx, model.Subnet.ValueString(), managerID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Cluster",
			fmt.Sprintf("Could not create cluster: %s", err),
		)
		return
	}

	// Map response to model
	model.ID = types.StringValue(strconv.Itoa(cluster.ID))

	tflog.Debug(ctx, "Created cluster", map[string]interface{}{
		"id": cluster.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the cluster state from the API.
func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ClusterResourceModel

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

	id := model.ID.ValueString()

	tflog.Debug(ctx, "Reading cluster", map[string]interface{}{
		"id": id,
	})

	// List all clusters and find the one matching our ID
	clusters, err := r.client.Clusters.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Cluster",
			fmt.Sprintf("Could not list clusters: %s", err),
		)
		return
	}

	cluster, ok := clusters[id]
	if !ok {
		tflog.Warn(ctx, "Cluster not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	model.Subnet = types.StringValue(cluster.Subnet)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is not supported -- all fields are ForceNew.
func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Cluster resources cannot be updated in-place. All attributes require replacement.",
	)
}

// Delete deletes a cluster.
func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ClusterResourceModel

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
			"Error Parsing Cluster ID",
			fmt.Sprintf("Could not parse cluster ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting cluster", map[string]interface{}{
		"id": id,
	})

	if err := r.client.Clusters.Delete(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Cluster",
			fmt.Sprintf("Could not delete cluster %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Deleted cluster", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing cluster by its numeric ID.
// Usage: terraform import vastai_cluster.example <id>
func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
