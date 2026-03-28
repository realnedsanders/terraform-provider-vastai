package volume

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VolumeResourceModel describes the resource data model for vastai_volume.
// The primary identifier is the volume contract ID stored as a string for
// Terraform compatibility (same pattern as SSH key resource).
type VolumeResourceModel struct {
	// Primary identifier (volume contract ID as string)
	ID types.String `tfsdk:"id"`

	// Creation-time attributes (all ForceNew)
	OfferID            types.Int64  `tfsdk:"offer_id"`
	Size               types.Int64  `tfsdk:"size"`
	Name               types.String `tfsdk:"name"`
	CloneFromID        types.Int64  `tfsdk:"clone_from_id"`
	DisableCompression types.Bool   `tfsdk:"disable_compression"`

	// Computed fields from API (read-only)
	Status        types.String  `tfsdk:"status"`
	DiskSpace     types.Float64 `tfsdk:"disk_space"`
	MachineID     types.Int64   `tfsdk:"machine_id"`
	Geolocation   types.String  `tfsdk:"geolocation"`
	InetUp        types.Float64 `tfsdk:"inet_up"`
	InetDown      types.Float64 `tfsdk:"inet_down"`
	Reliability   types.Float64 `tfsdk:"reliability"`
	DiskName      types.String  `tfsdk:"disk_name"`
	DriverVersion types.String  `tfsdk:"driver_version"`
	HostID        types.Int64   `tfsdk:"host_id"`
	Verification  types.String  `tfsdk:"verification"`

	// Timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// VolumeOffersDataSourceModel describes the data source model for vastai_volume_offers.
// Filter attributes are Optional; result attributes are Computed.
type VolumeOffersDataSourceModel struct {
	// Filter attributes (Optional)
	DiskSpace        types.Float64 `tfsdk:"disk_space"`
	MaxStorageCost   types.Float64 `tfsdk:"max_storage_cost"`
	InetUp           types.Float64 `tfsdk:"inet_up"`
	InetDown         types.Float64 `tfsdk:"inet_down"`
	Reliability      types.Float64 `tfsdk:"reliability"`
	Geolocation      types.String  `tfsdk:"geolocation"`
	Verified         types.Bool    `tfsdk:"verified"`
	StaticIP         types.Bool    `tfsdk:"static_ip"`
	DiskBW           types.Float64 `tfsdk:"disk_bw"`
	OrderBy          types.String  `tfsdk:"order_by"`
	Limit            types.Int64   `tfsdk:"limit"`
	AllocatedStorage types.Float64 `tfsdk:"allocated_storage"`
	RawQuery         types.String  `tfsdk:"raw_query"`

	// Result attributes (Computed)
	Offers         types.List   `tfsdk:"offers"`
	MostAffordable types.Object `tfsdk:"most_affordable"`
}

// VolumeOfferModel represents a single volume offer in Terraform state.
type VolumeOfferModel struct {
	ID            types.Int64   `tfsdk:"id"`
	CUDAMaxGood   types.Float64 `tfsdk:"cuda_max_good"`
	CPUGhz        types.Float64 `tfsdk:"cpu_ghz"`
	DiskBW        types.Float64 `tfsdk:"disk_bw"`
	DiskSpace     types.Float64 `tfsdk:"disk_space"`
	DiskName      types.String  `tfsdk:"disk_name"`
	StorageCost   types.Float64 `tfsdk:"storage_cost"`
	DriverVersion types.String  `tfsdk:"driver_version"`
	InetUp        types.Float64 `tfsdk:"inet_up"`
	InetDown      types.Float64 `tfsdk:"inet_down"`
	Reliability   types.Float64 `tfsdk:"reliability"`
	Duration      types.Float64 `tfsdk:"duration"`
	MachineID     types.Int64   `tfsdk:"machine_id"`
	Verification  types.String  `tfsdk:"verification"`
	HostID        types.Int64   `tfsdk:"host_id"`
	Geolocation   types.String  `tfsdk:"geolocation"`
}
