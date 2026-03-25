package offer

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GpuOffersDataSourceModel describes the data source data model.
// Filter attributes are Optional; result attributes are Computed.
type GpuOffersDataSourceModel struct {
	// Filter attributes (all Optional)
	GPUName         types.String  `tfsdk:"gpu_name"`
	NumGPUs         types.Int64   `tfsdk:"num_gpus"`
	GPURamGB        types.Float64 `tfsdk:"gpu_ram_gb"`
	MaxPricePerHour types.Float64 `tfsdk:"max_price_per_hour"`
	DatacenterOnly  types.Bool    `tfsdk:"datacenter_only"`
	Region          types.String  `tfsdk:"region"`
	OfferType       types.String  `tfsdk:"offer_type"`
	OrderBy         types.String  `tfsdk:"order_by"`
	Limit           types.Int64   `tfsdk:"limit"`
	RawQuery        types.String  `tfsdk:"raw_query"`

	// Result attributes (Computed)
	Offers          types.List   `tfsdk:"offers"`
	MostAffordable  types.Object `tfsdk:"most_affordable"`
}

// OfferModel represents a single GPU offer in Terraform state.
// All numeric RAM/storage values are in user-friendly GB units.
type OfferModel struct {
	ID               types.Int64   `tfsdk:"id"`
	MachineID        types.Int64   `tfsdk:"machine_id"`
	GPUName          types.String  `tfsdk:"gpu_name"`
	NumGPUs          types.Int64   `tfsdk:"num_gpus"`
	GPURamGB         types.Float64 `tfsdk:"gpu_ram_gb"`
	GPUTotalRAMGB    types.Float64 `tfsdk:"gpu_total_ram_gb"`
	CPUCores         types.Float64 `tfsdk:"cpu_cores"`
	CPURamGB         types.Float64 `tfsdk:"cpu_ram_gb"`
	DiskSpaceGB      types.Float64 `tfsdk:"disk_space_gb"`
	PricePerHour     types.Float64 `tfsdk:"price_per_hour"`
	DLPerf           types.Float64 `tfsdk:"dl_perf"`
	InetUp           types.Float64 `tfsdk:"inet_up"`
	InetDown         types.Float64 `tfsdk:"inet_down"`
	Reliability      types.Float64 `tfsdk:"reliability"`
	Geolocation      types.String  `tfsdk:"geolocation"`
	DatacenterHosted types.Bool    `tfsdk:"datacenter_hosted"`
	Verification     types.String  `tfsdk:"verification"`
	StaticIP         types.Bool    `tfsdk:"static_ip"`
	DirectPortCount  types.Int64   `tfsdk:"direct_port_count"`
	CUDAVersion      types.Float64 `tfsdk:"cuda_version"`
	MinBid           types.Float64 `tfsdk:"min_bid"`
	StorageCostPerGB types.Float64 `tfsdk:"storage_cost_per_gb"`
}
