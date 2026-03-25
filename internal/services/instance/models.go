package instance

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// InstanceResourceModel describes the resource data model for vastai_instance.
// Attributes are classified per SCHM-03: Required (user must set), Optional (user may set),
// Computed (API-set, read-only), or Optional+Computed (user may set, otherwise server default).
type InstanceResourceModel struct {
	// Required, immutable (RequiresReplace) -- changing these forces a new resource
	OfferID types.Int64   `tfsdk:"offer_id"`
	DiskGB  types.Float64 `tfsdk:"disk_gb"`

	// Required/Optional, mutable -- can be changed in-place via API
	Image          types.String  `tfsdk:"image"`           // Optional+Computed (can come from template)
	Status         types.String  `tfsdk:"status"`          // Optional+Computed ("running" or "stopped"), per D-08
	Label          types.String  `tfsdk:"label"`           // Optional, mutable via PUT /instances/{id}/
	BidPrice       types.Float64 `tfsdk:"bid_price"`       // Optional, nil = on-demand, mutable per D-11
	Onstart        types.String  `tfsdk:"onstart"`         // Optional, mutable via update_template
	Env            types.Map     `tfsdk:"env"`             // Optional, map of string per D-19
	TemplateHashID types.String  `tfsdk:"template_hash_id"` // Optional, mutable
	SSHKeyIDs      types.Set     `tfsdk:"ssh_key_ids"`     // Optional, set of strings per D-12/D-19
	ImageLogin     types.String  `tfsdk:"image_login"`     // Optional, Sensitive per SCHM-02
	UseSSH         types.Bool    `tfsdk:"use_ssh"`         // Optional+Computed per D-15
	UseJupyterLab  types.Bool    `tfsdk:"use_jupyter_lab"` // Optional+Computed per D-15
	CancelUnavail  types.Bool    `tfsdk:"cancel_unavail"`  // Optional

	// Computed, stable (UseStateForUnknown) -- set once at creation, won't change
	ID        types.String `tfsdk:"id"`         // Contract ID as string
	MachineID types.Int64  `tfsdk:"machine_id"` // Physical machine ID
	SSHHost   types.String `tfsdk:"ssh_host"`   // SSH connection hostname
	SSHPort   types.Int64  `tfsdk:"ssh_port"`   // SSH connection port
	NumGPUs   types.Int64  `tfsdk:"num_gpus"`   // Number of GPUs
	GPUName   types.String `tfsdk:"gpu_name"`   // GPU model name
	CreatedAt types.String `tfsdk:"created_at"` // From start_date

	// Computed, dynamic (no UseStateForUnknown) -- changes on every read
	ActualStatus types.String  `tfsdk:"actual_status"`  // Raw API status
	DPHTotal     types.Float64 `tfsdk:"cost_per_hour"`  // Current price
	GPURamGB     types.Float64 `tfsdk:"gpu_ram_gb"`     // GPU RAM in GB
	CPURamGB     types.Float64 `tfsdk:"cpu_ram_gb"`     // CPU RAM in GB
	CPUCores     types.Float64 `tfsdk:"cpu_cores"`      // Effective CPU cores
	InetUp       types.Float64 `tfsdk:"inet_up_mbps"`   // Upload speed in Mbps
	InetDown     types.Float64 `tfsdk:"inet_down_mbps"` // Download speed in Mbps
	Reliability  types.Float64 `tfsdk:"reliability"`    // Host reliability score
	Geolocation  types.String  `tfsdk:"geolocation"`    // Geographic location
	IsBid        types.Bool    `tfsdk:"is_bid"`         // Whether instance uses bid pricing
	StatusMsg    types.String  `tfsdk:"status_msg"`     // Current status message

	// Timeouts per SCHM-06
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// InstanceDataSourceModel describes the data model for the vastai_instance data source (singular).
// Looks up a single instance by its ID.
type InstanceDataSourceModel struct {
	ID               types.String  `tfsdk:"id"`
	MachineID        types.Int64   `tfsdk:"machine_id"`
	GPUName          types.String  `tfsdk:"gpu_name"`
	NumGPUs          types.Int64   `tfsdk:"num_gpus"`
	GPURamGB         types.Float64 `tfsdk:"gpu_ram_gb"`
	CPUCores         types.Float64 `tfsdk:"cpu_cores"`
	CPURamGB         types.Float64 `tfsdk:"cpu_ram_gb"`
	DiskSpaceGB      types.Float64 `tfsdk:"disk_space_gb"`
	ActualStatus     types.String  `tfsdk:"actual_status"`
	IntendedStatus   types.String  `tfsdk:"intended_status"`
	SSHHost          types.String  `tfsdk:"ssh_host"`
	SSHPort          types.Int64   `tfsdk:"ssh_port"`
	CostPerHour      types.Float64 `tfsdk:"cost_per_hour"`
	Label            types.String  `tfsdk:"label"`
	Image            types.String  `tfsdk:"image"`
	Geolocation      types.String  `tfsdk:"geolocation"`
	IsBid            types.Bool    `tfsdk:"is_bid"`
	Reliability      types.Float64 `tfsdk:"reliability"`
	InetUpMbps       types.Float64 `tfsdk:"inet_up_mbps"`
	InetDownMbps     types.Float64 `tfsdk:"inet_down_mbps"`
	StatusMsg        types.String  `tfsdk:"status_msg"`
	TemplateHashID   types.String  `tfsdk:"template_hash_id"`
	Onstart          types.String  `tfsdk:"onstart"`
}

// InstancesDataSourceModel describes the data model for the vastai_instances data source (plural).
// Lists all instances with optional label filtering.
type InstancesDataSourceModel struct {
	Label     types.String `tfsdk:"label"`
	Instances types.List   `tfsdk:"instances"`
}

// InstanceDataModel describes a single instance in the instances list (read-only view).
type InstanceDataModel struct {
	ID             types.String  `tfsdk:"id"`
	MachineID      types.Int64   `tfsdk:"machine_id"`
	GPUName        types.String  `tfsdk:"gpu_name"`
	NumGPUs        types.Int64   `tfsdk:"num_gpus"`
	GPURamGB       types.Float64 `tfsdk:"gpu_ram_gb"`
	CPUCores       types.Float64 `tfsdk:"cpu_cores"`
	CPURamGB       types.Float64 `tfsdk:"cpu_ram_gb"`
	DiskSpaceGB    types.Float64 `tfsdk:"disk_space_gb"`
	ActualStatus   types.String  `tfsdk:"actual_status"`
	IntendedStatus types.String  `tfsdk:"intended_status"`
	SSHHost        types.String  `tfsdk:"ssh_host"`
	SSHPort        types.Int64   `tfsdk:"ssh_port"`
	CostPerHour    types.Float64 `tfsdk:"cost_per_hour"`
	Label          types.String  `tfsdk:"label"`
	Image          types.String  `tfsdk:"image"`
	Geolocation    types.String  `tfsdk:"geolocation"`
	IsBid          types.Bool    `tfsdk:"is_bid"`
	Reliability    types.Float64 `tfsdk:"reliability"`
	InetUpMbps     types.Float64 `tfsdk:"inet_up_mbps"`
	InetDownMbps   types.Float64 `tfsdk:"inet_down_mbps"`
	StatusMsg      types.String  `tfsdk:"status_msg"`
	TemplateHashID types.String  `tfsdk:"template_hash_id"`
	Onstart        types.String  `tfsdk:"onstart"`
}

// instanceDataModelAttrTypes returns the attribute types for InstanceDataModel,
// used for constructing types.List and types.Object values.
func instanceDataModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"machine_id":       types.Int64Type,
		"gpu_name":         types.StringType,
		"num_gpus":         types.Int64Type,
		"gpu_ram_gb":       types.Float64Type,
		"cpu_cores":        types.Float64Type,
		"cpu_ram_gb":       types.Float64Type,
		"disk_space_gb":    types.Float64Type,
		"actual_status":    types.StringType,
		"intended_status":  types.StringType,
		"ssh_host":         types.StringType,
		"ssh_port":         types.Int64Type,
		"cost_per_hour":    types.Float64Type,
		"label":            types.StringType,
		"image":            types.StringType,
		"geolocation":      types.StringType,
		"is_bid":           types.BoolType,
		"reliability":      types.Float64Type,
		"inet_up_mbps":     types.Float64Type,
		"inet_down_mbps":   types.Float64Type,
		"status_msg":       types.StringType,
		"template_hash_id": types.StringType,
		"onstart":          types.StringType,
	}
}
