package instance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &InstanceDataSource{}
var _ datasource.DataSourceWithConfigure = &InstanceDataSource{}

// InstanceDataSource defines the data source implementation for looking up a single instance by ID.
type InstanceDataSource struct {
	client *client.VastAIClient
}

// NewInstanceDataSource creates a new instance data source instance.
func NewInstanceDataSource() datasource.DataSource {
	return &InstanceDataSource{}
}

// Metadata returns the data source type name.
func (d *InstanceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Schema defines the schema for the instance data source.
// All attributes except id are Computed. Every attribute has a Description per SCHM-04.
func (d *InstanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a single Vast.ai instance by its ID. Returns all instance attributes including " +
			"hardware specs, network info, status, and configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the instance to look up.",
				Required:    true,
			},
			"machine_id": schema.Int64Attribute{
				Description: "Physical machine ID hosting this instance.",
				Computed:    true,
			},
			"gpu_name": schema.StringAttribute{
				Description: "GPU model name (e.g., 'RTX 4090', 'A100 SXM4').",
				Computed:    true,
			},
			"num_gpus": schema.Int64Attribute{
				Description: "Number of GPUs allocated to this instance.",
				Computed:    true,
			},
			"gpu_ram_gb": schema.Float64Attribute{
				Description: "GPU memory per GPU in GB.",
				Computed:    true,
			},
			"cpu_cores": schema.Float64Attribute{
				Description: "Number of effective CPU cores available.",
				Computed:    true,
			},
			"cpu_ram_gb": schema.Float64Attribute{
				Description: "System RAM in GB.",
				Computed:    true,
			},
			"disk_space_gb": schema.Float64Attribute{
				Description: "Allocated disk space in GB.",
				Computed:    true,
			},
			"actual_status": schema.StringAttribute{
				Description: "Current actual status of the instance (e.g., 'running', 'stopped', 'loading', 'exited').",
				Computed:    true,
			},
			"intended_status": schema.StringAttribute{
				Description: "Intended status of the instance ('running' or 'stopped').",
				Computed:    true,
			},
			"ssh_host": schema.StringAttribute{
				Description: "SSH connection hostname for this instance.",
				Computed:    true,
			},
			"ssh_port": schema.Int64Attribute{
				Description: "SSH connection port for this instance.",
				Computed:    true,
			},
			"cost_per_hour": schema.Float64Attribute{
				Description: "Current cost per hour in USD.",
				Computed:    true,
			},
			"label": schema.StringAttribute{
				Description: "User-assigned label for this instance.",
				Computed:    true,
			},
			"image": schema.StringAttribute{
				Description: "Docker image running on this instance.",
				Computed:    true,
			},
			"geolocation": schema.StringAttribute{
				Description: "Geographic location of the machine hosting this instance.",
				Computed:    true,
			},
			"is_bid": schema.BoolAttribute{
				Description: "Whether this instance uses bid (interruptible) pricing.",
				Computed:    true,
			},
			"reliability": schema.Float64Attribute{
				Description: "Host reliability score (0.0 to 1.0).",
				Computed:    true,
			},
			"inet_up_mbps": schema.Float64Attribute{
				Description: "Internet upload speed in Mbps.",
				Computed:    true,
			},
			"inet_down_mbps": schema.Float64Attribute{
				Description: "Internet download speed in Mbps.",
				Computed:    true,
			},
			"status_msg": schema.StringAttribute{
				Description: "Current status message providing detail about the instance state.",
				Computed:    true,
			},
			"template_hash_id": schema.StringAttribute{
				Description: "Hash ID of the template associated with this instance.",
				Computed:    true,
			},
			"onstart": schema.StringAttribute{
				Description: "Startup script command executed when the instance starts.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *InstanceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.VastAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.VastAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

// Read fetches a single instance by ID from the Vast.ai API.
func (d *InstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model InstanceDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ID
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Instance ID",
			fmt.Sprintf("Could not parse instance ID %q as an integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading instance data source", map[string]interface{}{
		"instance_id": id,
	})

	// Call API
	instance, err := d.client.Instances.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Instance",
			fmt.Sprintf("An unexpected error occurred while reading instance %d: %s", id, err),
		)
		return
	}

	// Map API response to model (RAM values from API are in MB, convert to GB)
	model.ID = types.StringValue(strconv.Itoa(instance.ID))
	model.MachineID = types.Int64Value(int64(instance.MachineID))
	model.GPUName = types.StringValue(instance.GPUName)
	model.NumGPUs = types.Int64Value(int64(instance.NumGPUs))
	model.GPURamGB = types.Float64Value(instance.GPURAM / 1000.0)
	model.CPUCores = types.Float64Value(instance.CPUCoresEffective)
	model.CPURamGB = types.Float64Value(instance.CPURAM / 1000.0)
	model.DiskSpaceGB = types.Float64Value(instance.DiskSpace)
	model.ActualStatus = types.StringValue(instance.ActualStatus)
	model.IntendedStatus = types.StringValue(instance.IntendedStatus)
	model.SSHHost = types.StringValue(instance.SSHHost)
	model.SSHPort = types.Int64Value(int64(instance.SSHPort))
	model.CostPerHour = types.Float64Value(instance.DPHTotal)
	model.Label = types.StringValue(instance.Label)
	model.Image = types.StringValue(instance.ImageUUID)
	model.Geolocation = types.StringValue(instance.Geolocation)
	model.IsBid = types.BoolValue(instance.IsBid)
	model.Reliability = types.Float64Value(instance.Reliability2)
	model.InetUpMbps = types.Float64Value(instance.InetUp)
	model.InetDownMbps = types.Float64Value(instance.InetDown)
	model.StatusMsg = types.StringValue(instance.StatusMsg)
	model.TemplateHashID = types.StringValue(instance.TemplateHashID)
	model.Onstart = types.StringValue(instance.Onstart)

	tflog.Debug(ctx, "Read instance data source complete", map[string]interface{}{
		"instance_id": id,
		"status":      instance.ActualStatus,
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
