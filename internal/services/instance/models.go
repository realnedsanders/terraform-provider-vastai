package instance

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
