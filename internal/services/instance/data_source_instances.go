package instance

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &InstancesDataSource{}
var _ datasource.DataSourceWithConfigure = &InstancesDataSource{}

// InstancesDataSource defines the data source implementation for listing all instances.
type InstancesDataSource struct {
	client *client.VastAIClient
}

// NewInstancesDataSource creates a new instances data source instance.
func NewInstancesDataSource() datasource.DataSource {
	return &InstancesDataSource{}
}

// Metadata returns the data source type name.
func (d *InstancesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instances"
}

// instanceNestedAttributes returns the schema attributes for a single instance
// in the instances list (read-only view). Every attribute has a Description per SCHM-04.
func instanceNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique contract ID of the instance.",
			Computed:    true,
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
			Description: "Total GPU VRAM across all GPUs in GB.",
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
	}
}

// Schema defines the schema for the instances data source.
func (d *InstancesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all Vast.ai instances owned by the authenticated user. " +
			"Optionally filter by label substring match.",

		Attributes: map[string]schema.Attribute{
			"label": schema.StringAttribute{
				Description: "Filter instances by label (substring match). Only instances whose label contains this string will be returned.",
				Optional:    true,
			},
			"instances": schema.ListNestedAttribute{
				Description: "List of instances matching the filter criteria.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: instanceNestedAttributes(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *InstancesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches all instances from the Vast.ai API and optionally filters by label.
func (d *InstancesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model InstancesDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading instances data source")

	// Call API
	instances, err := d.client.Instances.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to List Instances",
			fmt.Sprintf("An unexpected error occurred while listing instances: %s", err),
		)
		return
	}

	// Apply optional label filter (substring match)
	labelFilter := ""
	if !model.Label.IsNull() && !model.Label.IsUnknown() {
		labelFilter = model.Label.ValueString()
	}

	var filtered []client.Instance
	for _, inst := range instances {
		if labelFilter != "" && !strings.Contains(inst.Label, labelFilter) {
			continue
		}
		filtered = append(filtered, inst)
	}

	tflog.Debug(ctx, "Instances data source read complete", map[string]interface{}{
		"total":    len(instances),
		"filtered": len(filtered),
		"label":    labelFilter,
	})

	// Convert to Terraform list
	instancesList, diags := apiInstancesToList(ctx, filtered)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Instances = instancesList

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// apiInstanceToAttrValues converts a client.Instance to a map of attr.Value.
// RAM values from API are in MB, converted to GB.
func apiInstanceToAttrValues(inst client.Instance) map[string]attr.Value {
	return map[string]attr.Value{
		"id":               types.StringValue(strconv.Itoa(inst.ID)),
		"machine_id":       types.Int64Value(int64(inst.MachineID)),
		"gpu_name":         types.StringValue(inst.GPUName),
		"num_gpus":         types.Int64Value(int64(inst.NumGPUs)),
		"gpu_ram_gb":       types.Float64Value(inst.GPUTotalRAM / 1000.0),
		"cpu_cores":        types.Float64Value(inst.CPUCoresEffective),
		"cpu_ram_gb":       types.Float64Value(inst.CPURAM / 1000.0),
		"disk_space_gb":    types.Float64Value(inst.DiskSpace),
		"actual_status":    types.StringValue(inst.ActualStatus),
		"intended_status":  types.StringValue(inst.IntendedStatus),
		"ssh_host":         types.StringValue(inst.SSHHost),
		"ssh_port":         types.Int64Value(int64(inst.SSHPort)),
		"cost_per_hour":    types.Float64Value(inst.DPHTotal),
		"label":            types.StringValue(inst.Label),
		"image":            types.StringValue(inst.ImageUUID),
		"geolocation":      types.StringValue(inst.Geolocation),
		"is_bid":           types.BoolValue(inst.IsBid),
		"reliability":      types.Float64Value(inst.Reliability2),
		"inet_up_mbps":     types.Float64Value(inst.InetUp),
		"inet_down_mbps":   types.Float64Value(inst.InetDown),
		"status_msg":       types.StringValue(inst.StatusMsg),
		"template_hash_id": types.StringValue(inst.TemplateHashID),
		"onstart":          types.StringValue(inst.Onstart),
	}
}

// apiInstancesToList converts a slice of client.Instance to a types.List.
func apiInstancesToList(_ context.Context, instances []client.Instance) (types.List, diag.Diagnostics) {
	attrTypes := instanceDataModelAttrTypes()

	if len(instances) == 0 {
		return types.ListValueMust(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{}), nil
	}

	elems := make([]attr.Value, len(instances))
	for i, inst := range instances {
		obj, diags := types.ObjectValue(attrTypes, apiInstanceToAttrValues(inst))
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}
		elems[i] = obj
	}

	return types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
}
