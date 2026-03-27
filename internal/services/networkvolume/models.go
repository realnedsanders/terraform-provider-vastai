package networkvolume

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NetworkVolumeResourceModel describes the resource data model for vastai_network_volume.
// The primary identifier is the volume contract ID stored as a string for
// Terraform compatibility (same pattern as volume resource).
type NetworkVolumeResourceModel struct {
	// Primary identifier (volume contract ID as string)
	ID types.String `tfsdk:"id"`

	// Creation-time attributes (all ForceNew)
	OfferID types.Int64  `tfsdk:"offer_id"`
	Size    types.Int64  `tfsdk:"size"`
	Name    types.String `tfsdk:"name"`

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

// NetworkVolumeOffersDataSourceModel describes the data source model for vastai_network_volume_offers.
// Filter attributes are Optional; result attributes are Computed.
type NetworkVolumeOffersDataSourceModel struct {
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

// NetworkVolumeOfferModel represents a single network volume offer in Terraform state.
// Includes network-volume-specific bandwidth fields (nw_disk_min_bw, nw_disk_max_bw, nw_disk_avg_bw)
// and cluster_id that are not present on local volume offers.
type NetworkVolumeOfferModel struct {
	ID           types.Int64   `tfsdk:"id"`
	DiskSpace    types.Float64 `tfsdk:"disk_space"`
	StorageCost  types.Float64 `tfsdk:"storage_cost"`
	InetUp       types.Float64 `tfsdk:"inet_up"`
	InetDown     types.Float64 `tfsdk:"inet_down"`
	Reliability  types.Float64 `tfsdk:"reliability"`
	Duration     types.Float64 `tfsdk:"duration"`
	Verification types.String  `tfsdk:"verification"`
	HostID       types.Int64   `tfsdk:"host_id"`
	ClusterID    types.Int64   `tfsdk:"cluster_id"`
	Geolocation  types.String  `tfsdk:"geolocation"`
	// Network-volume-specific bandwidth metrics
	NWDiskMinBW types.Float64 `tfsdk:"nw_disk_min_bw"`
	NWDiskMaxBW types.Float64 `tfsdk:"nw_disk_max_bw"`
	NWDiskAvgBW types.Float64 `tfsdk:"nw_disk_avg_bw"`
}
