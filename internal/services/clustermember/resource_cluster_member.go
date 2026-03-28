package clustermember

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure ClusterMemberResource satisfies the required interfaces.
var (
	_ resource.Resource                = &ClusterMemberResource{}
	_ resource.ResourceWithConfigure   = &ClusterMemberResource{}
	_ resource.ResourceWithImportState = &ClusterMemberResource{}
)

// ClusterMemberResource defines the resource implementation.
type ClusterMemberResource struct {
	client *client.VastAIClient
}

// NewClusterMemberResource returns a new cluster member resource instance.
func NewClusterMemberResource() resource.Resource {
	return &ClusterMemberResource{}
}

// Metadata returns the resource type name.
func (r *ClusterMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_member"
}

// Schema defines the schema for the resource.
func (r *ClusterMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages membership of a machine in a Vast.ai cluster. Uses composite ID format: cluster_id/machine_id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID: cluster_id/machine_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "ID of the cluster to join.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"machine_id": schema.StringAttribute{
				Description: "ID of the machine to add to the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"new_manager_id": schema.StringAttribute{
				Description: "If removing the cluster manager, specify the new manager machine ID. Only used during destroy.",
				Optional:    true,
			},
			"is_cluster_manager": schema.BoolAttribute{
				Description: "Whether this machine is the cluster manager.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"local_ip": schema.StringAttribute{
				Description: "Local IP address of the machine within the cluster network.",
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
func (r *ClusterMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create adds a machine to a cluster.
func (r *ClusterMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model ClusterMemberResourceModel

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

	machineID, err := strconv.Atoi(model.MachineID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Machine ID",
			fmt.Sprintf("Could not parse machine_id %q as integer: %s", model.MachineID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Adding machine to cluster", map[string]interface{}{
		"cluster_id": clusterID,
		"machine_id": machineID,
	})

	if err := r.client.Clusters.JoinMachine(ctx, clusterID, []int{machineID}); err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Machine to Cluster",
			fmt.Sprintf("Could not add machine %d to cluster %d: %s", machineID, clusterID, err),
		)
		return
	}

	// Set composite ID
	model.ID = types.StringValue(fmt.Sprintf("%d/%d", clusterID, machineID))

	// Read back to get computed fields (is_cluster_manager, local_ip)
	clusters, err := r.client.Clusters.List(ctx)
	if err != nil {
		// Non-fatal: we successfully joined, just can't read computed fields
		tflog.Warn(ctx, "Could not read cluster after join to populate computed fields", map[string]interface{}{
			"error": err.Error(),
		})
		model.IsClusterManager = types.BoolValue(false)
		model.LocalIP = types.StringValue("")
	} else {
		clusterIDStr := strconv.Itoa(clusterID)
		if cluster, ok := clusters[clusterIDStr]; ok {
			found := false
			for _, node := range cluster.Nodes {
				if node.MachineID == machineID {
					model.IsClusterManager = types.BoolValue(node.IsClusterManager)
					model.LocalIP = types.StringValue(node.LocalIP)
					found = true
					break
				}
			}
			if !found {
				model.IsClusterManager = types.BoolValue(false)
				model.LocalIP = types.StringValue("")
			}
		} else {
			model.IsClusterManager = types.BoolValue(false)
			model.LocalIP = types.StringValue("")
		}
	}

	tflog.Debug(ctx, "Added machine to cluster", map[string]interface{}{
		"id": model.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the cluster member state from the API.
func (r *ClusterMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ClusterMemberResourceModel

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

	clusterID, err := strconv.Atoi(model.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Cluster ID",
			fmt.Sprintf("Could not parse cluster_id %q as integer: %s", model.ClusterID.ValueString(), err),
		)
		return
	}

	machineID, err := strconv.Atoi(model.MachineID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Machine ID",
			fmt.Sprintf("Could not parse machine_id %q as integer: %s", model.MachineID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading cluster member", map[string]interface{}{
		"cluster_id": clusterID,
		"machine_id": machineID,
	})

	// List all clusters and find the matching cluster
	clusters, err := r.client.Clusters.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Cluster Member",
			fmt.Sprintf("Could not list clusters: %s", err),
		)
		return
	}

	clusterIDStr := strconv.Itoa(clusterID)
	cluster, ok := clusters[clusterIDStr]
	if !ok {
		tflog.Warn(ctx, "Cluster not found, removing member from state", map[string]interface{}{
			"cluster_id": clusterID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Check if machine is in the nodes list
	found := false
	for _, node := range cluster.Nodes {
		if node.MachineID == machineID {
			model.IsClusterManager = types.BoolValue(node.IsClusterManager)
			model.LocalIP = types.StringValue(node.LocalIP)
			found = true
			break
		}
	}

	if !found {
		tflog.Warn(ctx, "Machine not found in cluster, removing from state", map[string]interface{}{
			"cluster_id": clusterID,
			"machine_id": machineID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is not supported -- key fields are ForceNew.
func (r *ClusterMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Cluster member resources cannot be updated in-place. Key attributes require replacement.",
	)
}

// Delete removes a machine from a cluster.
func (r *ClusterMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ClusterMemberResourceModel

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

	clusterID, err := strconv.Atoi(model.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Cluster ID",
			fmt.Sprintf("Could not parse cluster_id %q as integer: %s", model.ClusterID.ValueString(), err),
		)
		return
	}

	machineID, err := strconv.Atoi(model.MachineID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Machine ID",
			fmt.Sprintf("Could not parse machine_id %q as integer: %s", model.MachineID.ValueString(), err),
		)
		return
	}

	// Parse new_manager_id if set
	var newManagerIDPtr *int
	if !model.NewManagerID.IsNull() && !model.NewManagerID.IsUnknown() {
		newManagerIDVal, err := strconv.Atoi(model.NewManagerID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing New Manager ID",
				fmt.Sprintf("Could not parse new_manager_id %q as integer: %s", model.NewManagerID.ValueString(), err),
			)
			return
		}
		newManagerIDPtr = &newManagerIDVal
	}

	tflog.Debug(ctx, "Removing machine from cluster", map[string]interface{}{
		"cluster_id": clusterID,
		"machine_id": machineID,
	})

	if err := r.client.Clusters.RemoveMachine(ctx, clusterID, machineID, newManagerIDPtr); err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Machine from Cluster",
			fmt.Sprintf("Could not remove machine %d from cluster %d: %s", machineID, clusterID, err),
		)
		return
	}

	tflog.Debug(ctx, "Removed machine from cluster", map[string]interface{}{
		"cluster_id": clusterID,
		"machine_id": machineID,
	})
}

// ImportState imports an existing cluster member by "cluster_id/machine_id" composite string.
// Usage: terraform import vastai_cluster_member.example <cluster_id>/<machine_id>.
func (r *ClusterMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'cluster_id/machine_id', got: %q", req.ID),
		)
		return
	}

	clusterID := parts[0]
	machineID := parts[1]

	// Validate both parts are valid integers
	if _, err := strconv.Atoi(clusterID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Cluster ID",
			fmt.Sprintf("cluster_id %q is not a valid integer: %s", clusterID, err),
		)
		return
	}
	if _, err := strconv.Atoi(machineID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Machine ID",
			fmt.Sprintf("machine_id %q is not a valid integer: %s", machineID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("machine_id"), machineID)...)
}
