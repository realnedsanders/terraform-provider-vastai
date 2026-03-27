package volume

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &VolumeOffersDataSource{}
var _ datasource.DataSourceWithConfigure = &VolumeOffersDataSource{}

// VolumeOffersDataSource defines the data source implementation.
type VolumeOffersDataSource struct {
	client *client.VastAIClient
}

// NewVolumeOffersDataSource creates a new volume offers data source instance.
func NewVolumeOffersDataSource() datasource.DataSource {
	return &VolumeOffersDataSource{}
}

// Metadata returns the data source type name.
func (d *VolumeOffersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_offers"
}

// volumeOfferModelAttrTypes returns the attribute types for VolumeOfferModel,
// used for constructing types.List and types.Object values.
func volumeOfferModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.Int64Type,
		"cuda_max_good":  types.Float64Type,
		"cpu_ghz":        types.Float64Type,
		"disk_bw":        types.Float64Type,
		"disk_space":     types.Float64Type,
		"disk_name":      types.StringType,
		"storage_cost":   types.Float64Type,
		"driver_version": types.StringType,
		"inet_up":        types.Float64Type,
		"inet_down":      types.Float64Type,
		"reliability":    types.Float64Type,
		"duration":       types.Float64Type,
		"machine_id":     types.Int64Type,
		"verification":   types.StringType,
		"host_id":        types.Int64Type,
		"geolocation":    types.StringType,
	}
}

// volumeOfferNestedAttributes returns the schema attributes for a single volume offer,
// shared between the "offers" list and "most_affordable" single nested attribute.
func volumeOfferNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Description: "Unique volume offer ID in the Vast.ai marketplace.",
			Computed:    true,
		},
		"cuda_max_good": schema.Float64Attribute{
			Description: "Maximum supported CUDA version on this machine.",
			Computed:    true,
		},
		"cpu_ghz": schema.Float64Attribute{
			Description: "CPU clock speed in GHz.",
			Computed:    true,
		},
		"disk_bw": schema.Float64Attribute{
			Description: "Disk bandwidth in MB/s.",
			Computed:    true,
		},
		"disk_space": schema.Float64Attribute{
			Description: "Available disk space in GB.",
			Computed:    true,
		},
		"disk_name": schema.StringAttribute{
			Description: "Name/model of the physical disk.",
			Computed:    true,
		},
		"storage_cost": schema.Float64Attribute{
			Description: "Storage cost per GB per month in USD.",
			Computed:    true,
		},
		"driver_version": schema.StringAttribute{
			Description: "GPU driver version on the host machine.",
			Computed:    true,
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
		"duration": schema.Float64Attribute{
			Description: "Machine uptime duration in days.",
			Computed:    true,
		},
		"machine_id": schema.Int64Attribute{
			Description: "ID of the physical machine hosting this offer.",
			Computed:    true,
		},
		"verification": schema.StringAttribute{
			Description: "Host verification status.",
			Computed:    true,
		},
		"host_id": schema.Int64Attribute{
			Description: "ID of the host.",
			Computed:    true,
		},
		"geolocation": schema.StringAttribute{
			Description: "Geographic location of the machine (e.g., 'US', 'EU').",
			Computed:    true,
		},
	}
}

// Schema defines the schema for the volume offers data source.
func (d *VolumeOffersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Search for available volume offers on the Vast.ai marketplace. " +
			"Returns a filtered and sorted list of offers matching the specified criteria, " +
			"along with a convenience `most_affordable` attribute for the cheapest matching offer.",

		Attributes: map[string]schema.Attribute{
			// Filter attributes (all Optional)
			"disk_space": schema.Float64Attribute{
				Description: "Minimum disk space in GB. Only offers with at least this much disk space will be returned.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"max_storage_cost": schema.Float64Attribute{
				Description: "Maximum storage cost per GB per month in USD. Only offers at or below this cost will be returned.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"inet_up": schema.Float64Attribute{
				Description: "Minimum internet upload speed in Mbps.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"inet_down": schema.Float64Attribute{
				Description: "Minimum internet download speed in Mbps.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"reliability": schema.Float64Attribute{
				Description: "Minimum host reliability score. Valid range: 0.0 to 1.0.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.Between(0, 1),
				},
			},
			"geolocation": schema.StringAttribute{
				Description: "Filter by geographic location (e.g., 'US', 'EU'). Matches against the offer's geolocation field.",
				Optional:    true,
			},
			"verified": schema.BoolAttribute{
				Description: "If true, only return offers from verified machines.",
				Optional:    true,
			},
			"static_ip": schema.BoolAttribute{
				Description: "If true, only return offers from machines with static IP addresses.",
				Optional:    true,
			},
			"disk_bw": schema.Float64Attribute{
				Description: "Minimum disk bandwidth in MB/s.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"order_by": schema.StringAttribute{
				Description: "Field to sort results by. Valid values: 'storage_cost', 'disk_space', 'inet_up', 'inet_down', 'reliability', 'duration'. Default: 'storage_cost'.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("storage_cost", "disk_space", "inet_up", "inet_down", "reliability", "duration"),
				},
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum number of offers to return. Valid range: 1 to 1000. Default: 10.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1000),
				},
			},
			"allocated_storage": schema.Float64Attribute{
				Description: "Storage amount in GB for pricing calculations. Affects the computed storage_cost " +
					"in results. Defaults to 1.0 if not set. Set this to match your intended volume size " +
					"for accurate pricing.",
				Optional: true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"raw_query": schema.StringAttribute{
				Description: "Raw query JSON string to pass directly to the Vast.ai volume search API. " +
					"When set, all structured filter attributes are ignored. " +
					"Use this for advanced queries not supported by the structured filters.",
				Optional: true,
			},

			// Result attributes (Computed)
			"offers": schema.ListNestedAttribute{
				Description: "List of volume offers matching the search criteria, sorted by the `order_by` field.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: volumeOfferNestedAttributes(),
				},
			},
			"most_affordable": schema.SingleNestedAttribute{
				Description: "The most affordable offer from the results (first result when sorted by storage_cost). " +
					"Convenience attribute to avoid indexing into the offers list.",
				Computed:   true,
				Attributes: volumeOfferNestedAttributes(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *VolumeOffersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read queries the Vast.ai API for volume offers matching the configured filters.
func (d *VolumeOffersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model VolumeOffersDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build search params from model
	params := &client.VolumeOfferSearchParams{}

	if !model.DiskSpace.IsNull() && !model.DiskSpace.IsUnknown() {
		v := model.DiskSpace.ValueFloat64()
		params.DiskSpace = &v
	}

	if !model.MaxStorageCost.IsNull() && !model.MaxStorageCost.IsUnknown() {
		v := model.MaxStorageCost.ValueFloat64()
		params.StorageCost = &v
	}

	if !model.InetUp.IsNull() && !model.InetUp.IsUnknown() {
		v := model.InetUp.ValueFloat64()
		params.InetUp = &v
	}

	if !model.InetDown.IsNull() && !model.InetDown.IsUnknown() {
		v := model.InetDown.ValueFloat64()
		params.InetDown = &v
	}

	if !model.Reliability.IsNull() && !model.Reliability.IsUnknown() {
		v := model.Reliability.ValueFloat64()
		params.Reliability = &v
	}

	if !model.Geolocation.IsNull() && !model.Geolocation.IsUnknown() {
		params.Geolocation = model.Geolocation.ValueString()
	}

	if !model.Verified.IsNull() && !model.Verified.IsUnknown() {
		v := model.Verified.ValueBool()
		params.Verified = &v
	}

	if !model.StaticIP.IsNull() && !model.StaticIP.IsUnknown() {
		v := model.StaticIP.ValueBool()
		params.StaticIP = &v
	}

	if !model.DiskBW.IsNull() && !model.DiskBW.IsUnknown() {
		v := model.DiskBW.ValueFloat64()
		params.DiskBW = &v
	}

	if !model.OrderBy.IsNull() && !model.OrderBy.IsUnknown() {
		params.OrderBy = model.OrderBy.ValueString()
	}

	if !model.Limit.IsNull() && !model.Limit.IsUnknown() {
		params.Limit = int(model.Limit.ValueInt64())
	}

	if !model.AllocatedStorage.IsNull() && !model.AllocatedStorage.IsUnknown() {
		params.AllocatedStorage = model.AllocatedStorage.ValueFloat64()
	}

	if !model.RawQuery.IsNull() && !model.RawQuery.IsUnknown() {
		params.RawQuery = model.RawQuery.ValueString()
	}

	// Call API
	tflog.Debug(ctx, "Searching volume offers", map[string]interface{}{
		"limit":    params.Limit,
		"order_by": params.OrderBy,
	})

	offers, err := d.client.Volumes.SearchOffers(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Search Volume Offers",
			fmt.Sprintf("An unexpected error occurred while searching volume offers: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Volume offers search complete", map[string]interface{}{
		"count": len(offers),
	})

	// Convert API offers to Terraform model
	offerModels := make([]VolumeOfferModel, len(offers))
	for i, o := range offers {
		offerModels[i] = apiVolumeOfferToModel(o)
	}

	// Set offers list
	offersList, diags := volumeOfferModelsToList(offerModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Offers = offersList

	// Set most_affordable (first result, already sorted by order_by)
	if len(offerModels) > 0 {
		mostAffordable, diags := volumeOfferModelToObject(offerModels[0])
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		model.MostAffordable = mostAffordable
	} else {
		model.MostAffordable = types.ObjectNull(volumeOfferModelAttrTypes())
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// apiVolumeOfferToModel converts a client.VolumeOffer to a VolumeOfferModel.
// No unit conversions needed -- all fields map directly.
func apiVolumeOfferToModel(o client.VolumeOffer) VolumeOfferModel {
	return VolumeOfferModel{
		ID:            types.Int64Value(int64(o.ID)),
		CUDAMaxGood:   types.Float64Value(o.CUDAMaxGood),
		CPUGhz:        types.Float64Value(o.CPUGhz),
		DiskBW:        types.Float64Value(o.DiskBW),
		DiskSpace:     types.Float64Value(o.DiskSpace),
		DiskName:      types.StringValue(o.DiskName),
		StorageCost:   types.Float64Value(o.StorageCost),
		DriverVersion: types.StringValue(o.DriverVersion),
		InetUp:        types.Float64Value(o.InetUp),
		InetDown:      types.Float64Value(o.InetDown),
		Reliability:   types.Float64Value(o.Reliability),
		Duration:      types.Float64Value(o.Duration),
		MachineID:     types.Int64Value(int64(o.MachineID)),
		Verification:  types.StringValue(o.Verification),
		HostID:        types.Int64Value(int64(o.HostID)),
		Geolocation:   types.StringValue(o.Geolocation),
	}
}

// volumeOfferModelToAttrValues converts a VolumeOfferModel to a map of attr.Value.
func volumeOfferModelToAttrValues(m VolumeOfferModel) map[string]attr.Value {
	return map[string]attr.Value{
		"id":             m.ID,
		"cuda_max_good":  m.CUDAMaxGood,
		"cpu_ghz":        m.CPUGhz,
		"disk_bw":        m.DiskBW,
		"disk_space":     m.DiskSpace,
		"disk_name":      m.DiskName,
		"storage_cost":   m.StorageCost,
		"driver_version": m.DriverVersion,
		"inet_up":        m.InetUp,
		"inet_down":      m.InetDown,
		"reliability":    m.Reliability,
		"duration":       m.Duration,
		"machine_id":     m.MachineID,
		"verification":   m.Verification,
		"host_id":        m.HostID,
		"geolocation":    m.Geolocation,
	}
}

// volumeOfferModelsToList converts a slice of VolumeOfferModel to a types.List.
func volumeOfferModelsToList(models []VolumeOfferModel) (types.List, diag.Diagnostics) {
	attrTypes := volumeOfferModelAttrTypes()
	if len(models) == 0 {
		return types.ListValueMust(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{}), nil
	}

	elems := make([]attr.Value, len(models))
	for i, m := range models {
		obj, diags := types.ObjectValue(attrTypes, volumeOfferModelToAttrValues(m))
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}
		elems[i] = obj
	}

	return types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
}

// volumeOfferModelToObject converts a single VolumeOfferModel to a types.Object.
func volumeOfferModelToObject(m VolumeOfferModel) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(volumeOfferModelAttrTypes(), volumeOfferModelToAttrValues(m))
}
