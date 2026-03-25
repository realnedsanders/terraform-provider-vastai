package offer

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
var _ datasource.DataSource = &GpuOffersDataSource{}
var _ datasource.DataSourceWithConfigure = &GpuOffersDataSource{}

// GpuOffersDataSource defines the data source implementation.
type GpuOffersDataSource struct {
	client *client.VastAIClient
}

// NewGpuOffersDataSource creates a new GPU offers data source instance.
func NewGpuOffersDataSource() datasource.DataSource {
	return &GpuOffersDataSource{}
}

// Metadata returns the data source type name.
func (d *GpuOffersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gpu_offers"
}

// offerModelAttrTypes returns the attribute types for OfferModel, used for
// constructing types.List and types.Object values.
func offerModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.Int64Type,
		"machine_id":        types.Int64Type,
		"gpu_name":          types.StringType,
		"num_gpus":          types.Int64Type,
		"gpu_ram_gb":        types.Float64Type,
		"gpu_total_ram_gb":  types.Float64Type,
		"cpu_cores":         types.Float64Type,
		"cpu_ram_gb":        types.Float64Type,
		"disk_space_gb":     types.Float64Type,
		"price_per_hour":    types.Float64Type,
		"dl_perf":           types.Float64Type,
		"inet_up":           types.Float64Type,
		"inet_down":         types.Float64Type,
		"reliability":       types.Float64Type,
		"geolocation":       types.StringType,
		"datacenter_hosted": types.BoolType,
		"verification":      types.StringType,
		"static_ip":         types.BoolType,
		"direct_port_count": types.Int64Type,
		"cuda_version":      types.Float64Type,
		"min_bid":           types.Float64Type,
		"storage_cost_per_gb": types.Float64Type,
	}
}

// offerNestedAttributes returns the schema attributes for a single offer,
// shared between the "offers" list and "most_affordable" single nested attribute.
func offerNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Description: "Unique offer ID in the Vast.ai marketplace.",
			Computed:    true,
		},
		"machine_id": schema.Int64Attribute{
			Description: "ID of the physical machine hosting this offer.",
			Computed:    true,
		},
		"gpu_name": schema.StringAttribute{
			Description: "GPU model name (e.g., 'RTX 4090', 'A100').",
			Computed:    true,
		},
		"num_gpus": schema.Int64Attribute{
			Description: "Number of GPUs available in this offer.",
			Computed:    true,
		},
		"gpu_ram_gb": schema.Float64Attribute{
			Description: "GPU memory per GPU in GB.",
			Computed:    true,
		},
		"gpu_total_ram_gb": schema.Float64Attribute{
			Description: "Total GPU memory across all GPUs in GB.",
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
			Description: "Available disk space in GB.",
			Computed:    true,
		},
		"price_per_hour": schema.Float64Attribute{
			Description: "Total price per hour in USD (dph_total).",
			Computed:    true,
		},
		"dl_perf": schema.Float64Attribute{
			Description: "Deep learning performance score.",
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
		"geolocation": schema.StringAttribute{
			Description: "Geographic location of the machine (e.g., 'US', 'EU').",
			Computed:    true,
		},
		"datacenter_hosted": schema.BoolAttribute{
			Description: "Whether the machine is hosted in a datacenter (true) or is a consumer machine (false).",
			Computed:    true,
		},
		"verification": schema.StringAttribute{
			Description: "Host verification status.",
			Computed:    true,
		},
		"static_ip": schema.BoolAttribute{
			Description: "Whether the machine has a static IP address.",
			Computed:    true,
		},
		"direct_port_count": schema.Int64Attribute{
			Description: "Number of directly accessible ports.",
			Computed:    true,
		},
		"cuda_version": schema.Float64Attribute{
			Description: "Maximum supported CUDA version.",
			Computed:    true,
		},
		"min_bid": schema.Float64Attribute{
			Description: "Minimum bid price in USD/hr for interruptible instances.",
			Computed:    true,
		},
		"storage_cost_per_gb": schema.Float64Attribute{
			Description: "Storage cost per GB per month in USD.",
			Computed:    true,
		},
	}
}

// Schema defines the schema for the GPU offers data source.
func (d *GpuOffersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Search for available GPU offers on the Vast.ai marketplace. " +
			"Returns a filtered and sorted list of offers matching the specified criteria, " +
			"along with a convenience `most_affordable` attribute for the cheapest matching offer.",

		Attributes: map[string]schema.Attribute{
			// Filter attributes (all Optional)
			"gpu_name": schema.StringAttribute{
				Description: "Filter by GPU model name (e.g., 'RTX 4090', 'A100'). Must match exactly.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"num_gpus": schema.Int64Attribute{
				Description: "Filter by exact number of GPUs. Valid range: 1 to 16.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 16),
				},
			},
			"gpu_ram_gb": schema.Float64Attribute{
				Description: "Minimum GPU memory per GPU in GB. Offers with at least this much VRAM will be returned.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(1.0),
				},
			},
			"max_price_per_hour": schema.Float64Attribute{
				Description: "Maximum price per hour in USD. Only offers at or below this price will be returned.",
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0.001),
				},
			},
			"datacenter_only": schema.BoolAttribute{
				Description: "If true, only return offers from datacenter-hosted machines. Default: false (includes all hosting types).",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "Filter by geographic region (e.g., 'US', 'EU'). Matches against the offer's geolocation field.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"offer_type": schema.StringAttribute{
				Description: "Type of offer pricing. Valid values: 'on-demand', 'bid', 'reserved'. Default: 'on-demand'.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("on-demand", "bid", "reserved"),
				},
			},
			"order_by": schema.StringAttribute{
				Description: "Field to sort results by. Common values: 'dph_total' (price), 'dlperf_per_dphtotal' (performance/price), 'gpu_ram' (VRAM). Default: 'dph_total'.",
				Optional:    true,
				Computed:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum number of offers to return. Valid range: 1 to 1000. Default: 10.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1000),
				},
			},
			"raw_query": schema.StringAttribute{
				Description: "Raw query JSON string to pass directly to the Vast.ai search API. " +
					"When set, all structured filter attributes are ignored. " +
					"Use this for advanced queries not supported by the structured filters.",
				Optional: true,
			},

			// Result attributes (Computed)
			"offers": schema.ListNestedAttribute{
				Description: "List of GPU offers matching the search criteria, sorted by the `order_by` field.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: offerNestedAttributes(),
				},
			},
			"most_affordable": schema.SingleNestedAttribute{
				Description: "The most affordable offer from the results (first result when sorted by price). " +
					"Convenience attribute to avoid indexing into the offers list.",
				Computed:   true,
				Attributes: offerNestedAttributes(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *GpuOffersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read queries the Vast.ai API for GPU offers matching the configured filters.
func (d *GpuOffersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model GpuOffersDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build search params from model
	params := &client.OfferSearchParams{}

	if !model.GPUName.IsNull() && !model.GPUName.IsUnknown() {
		params.GPUName = model.GPUName.ValueString()
	}

	if !model.NumGPUs.IsNull() && !model.NumGPUs.IsUnknown() {
		v := int(model.NumGPUs.ValueInt64())
		params.NumGPUs = &v
	}

	if !model.GPURamGB.IsNull() && !model.GPURamGB.IsUnknown() {
		v := model.GPURamGB.ValueFloat64()
		params.GPURamGB = &v
	}

	if !model.MaxPricePerHour.IsNull() && !model.MaxPricePerHour.IsUnknown() {
		v := model.MaxPricePerHour.ValueFloat64()
		params.MaxPrice = &v
	}

	if !model.DatacenterOnly.IsNull() && !model.DatacenterOnly.IsUnknown() {
		v := model.DatacenterOnly.ValueBool()
		params.DatacenterOnly = &v
	}

	if !model.Region.IsNull() && !model.Region.IsUnknown() {
		params.Region = model.Region.ValueString()
	}

	if !model.OfferType.IsNull() && !model.OfferType.IsUnknown() {
		params.OfferType = model.OfferType.ValueString()
	}

	// Set defaults for order_by and limit
	orderBy := "dph_total"
	if !model.OrderBy.IsNull() && !model.OrderBy.IsUnknown() {
		orderBy = model.OrderBy.ValueString()
	}
	params.OrderBy = orderBy
	model.OrderBy = types.StringValue(orderBy)

	limit := int64(10)
	if !model.Limit.IsNull() && !model.Limit.IsUnknown() {
		limit = model.Limit.ValueInt64()
	}
	params.Limit = int(limit)
	model.Limit = types.Int64Value(limit)

	if !model.RawQuery.IsNull() && !model.RawQuery.IsUnknown() {
		params.RawQuery = model.RawQuery.ValueString()
	}

	// Call API
	tflog.Debug(ctx, "Searching GPU offers", map[string]interface{}{
		"gpu_name": params.GPUName,
		"limit":    params.Limit,
		"order_by": params.OrderBy,
	})

	offers, err := d.client.Offers.Search(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Search GPU Offers",
			fmt.Sprintf("An unexpected error occurred while searching GPU offers: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "GPU offers search complete", map[string]interface{}{
		"count": len(offers),
	})

	// Convert API offers to Terraform model
	offerModels := make([]OfferModel, len(offers))
	for i, o := range offers {
		offerModels[i] = apiOfferToModel(o)
	}

	// Set offers list
	offersList, diags := offerModelsToList(ctx, offerModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Offers = offersList

	// Set most_affordable (first result, already sorted by price)
	if len(offerModels) > 0 {
		mostAffordable, diags := offerModelToObject(ctx, offerModels[0])
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		model.MostAffordable = mostAffordable
	} else {
		model.MostAffordable = types.ObjectNull(offerModelAttrTypes())
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// apiOfferToModel converts a client.Offer to an OfferModel.
// Converts MB values to GB for user-friendly display.
func apiOfferToModel(o client.Offer) OfferModel {
	return OfferModel{
		ID:               types.Int64Value(int64(o.ID)),
		MachineID:        types.Int64Value(int64(o.MachineID)),
		GPUName:          types.StringValue(o.GPUName),
		NumGPUs:          types.Int64Value(int64(o.NumGPUs)),
		GPURamGB:         types.Float64Value(o.GPURAM / 1000.0),
		GPUTotalRAMGB:    types.Float64Value(o.GPUTotalRAM / 1000.0),
		CPUCores:         types.Float64Value(o.CPUCoresEffective),
		CPURamGB:         types.Float64Value(o.CPURAM / 1000.0),
		DiskSpaceGB:      types.Float64Value(o.DiskSpace),
		PricePerHour:     types.Float64Value(o.DPHTotal),
		DLPerf:           types.Float64Value(o.DLPerf),
		InetUp:           types.Float64Value(o.InetUp),
		InetDown:         types.Float64Value(o.InetDown),
		Reliability:      types.Float64Value(o.Reliability),
		Geolocation:      types.StringValue(o.Geolocation),
		DatacenterHosted: types.BoolValue(o.HostingType != 0),
		Verification:     types.StringValue(o.Verification),
		StaticIP:         types.BoolValue(o.StaticIP),
		DirectPortCount:  types.Int64Value(int64(o.DirectPortCount)),
		CUDAVersion:      types.Float64Value(o.CUDAMaxGood),
		MinBid:           types.Float64Value(o.MinBid),
		StorageCostPerGB: types.Float64Value(o.StorageCost),
	}
}

// offerModelToAttrValues converts an OfferModel to a map of attr.Value.
func offerModelToAttrValues(m OfferModel) map[string]attr.Value {
	return map[string]attr.Value{
		"id":                  m.ID,
		"machine_id":          m.MachineID,
		"gpu_name":            m.GPUName,
		"num_gpus":            m.NumGPUs,
		"gpu_ram_gb":          m.GPURamGB,
		"gpu_total_ram_gb":    m.GPUTotalRAMGB,
		"cpu_cores":           m.CPUCores,
		"cpu_ram_gb":          m.CPURamGB,
		"disk_space_gb":       m.DiskSpaceGB,
		"price_per_hour":      m.PricePerHour,
		"dl_perf":             m.DLPerf,
		"inet_up":             m.InetUp,
		"inet_down":           m.InetDown,
		"reliability":         m.Reliability,
		"geolocation":         m.Geolocation,
		"datacenter_hosted":   m.DatacenterHosted,
		"verification":        m.Verification,
		"static_ip":           m.StaticIP,
		"direct_port_count":   m.DirectPortCount,
		"cuda_version":        m.CUDAVersion,
		"min_bid":             m.MinBid,
		"storage_cost_per_gb": m.StorageCostPerGB,
	}
}

// offerModelsToList converts a slice of OfferModel to a types.List.
func offerModelsToList(ctx context.Context, models []OfferModel) (types.List, diag.Diagnostics) {
	attrTypes := offerModelAttrTypes()
	if len(models) == 0 {
		return types.ListValueMust(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{}), nil
	}

	elems := make([]attr.Value, len(models))
	for i, m := range models {
		obj, diags := types.ObjectValue(attrTypes, offerModelToAttrValues(m))
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}
		elems[i] = obj
	}

	return types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
}

// offerModelToObject converts a single OfferModel to a types.Object.
func offerModelToObject(_ context.Context, m OfferModel) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(offerModelAttrTypes(), offerModelToAttrValues(m))
}
