package endpoint

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &EndpointsDataSource{}
	_ datasource.DataSourceWithConfigure = &EndpointsDataSource{}
)

// EndpointsDataSource defines the data source implementation.
type EndpointsDataSource struct {
	client *client.VastAIClient
}

// NewEndpointsDataSource creates a new endpoints data source instance.
func NewEndpointsDataSource() datasource.DataSource {
	return &EndpointsDataSource{}
}

// Metadata returns the data source type name.
func (d *EndpointsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoints"
}

// endpointModelAttrTypes returns the attribute types for EndpointModel,
// used for constructing types.List values.
func endpointModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.Int64Type,
		"endpoint_name":  types.StringType,
		"min_load":       types.Float64Type,
		"min_cold_load":  types.Float64Type,
		"target_util":    types.Float64Type,
		"cold_mult":      types.Float64Type,
		"cold_workers":   types.Int64Type,
		"max_workers":    types.Int64Type,
		"endpoint_state": types.StringType,
	}
}

// endpointNestedAttributes returns the schema attributes for a single endpoint
// in the data source list.
func endpointNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Description: "Unique endpoint ID.",
			Computed:    true,
		},
		"endpoint_name": schema.StringAttribute{
			Description: "Name of the serverless endpoint.",
			Computed:    true,
		},
		"min_load": schema.Float64Attribute{
			Description: "Minimum floor load in perf units/s.",
			Computed:    true,
		},
		"min_cold_load": schema.Float64Attribute{
			Description: "Minimum floor load allowing cold workers.",
			Computed:    true,
		},
		"target_util": schema.Float64Attribute{
			Description: "Target capacity utilization fraction.",
			Computed:    true,
		},
		"cold_mult": schema.Float64Attribute{
			Description: "Cold capacity as multiple of hot capacity.",
			Computed:    true,
		},
		"cold_workers": schema.Int64Attribute{
			Description: "Minimum cold workers when no load.",
			Computed:    true,
		},
		"max_workers": schema.Int64Attribute{
			Description: "Maximum workers the endpoint can have.",
			Computed:    true,
		},
		"endpoint_state": schema.StringAttribute{
			Description: "Endpoint runtime state (active, suspended, stopped).",
			Computed:    true,
		},
	}
}

// Schema defines the schema for the endpoints data source.
func (d *EndpointsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all serverless endpoints owned by the authenticated user. " +
			"Returns endpoint metadata including autoscaling configuration and runtime state.",

		Attributes: map[string]schema.Attribute{
			"endpoints": schema.ListNestedAttribute{
				Description: "List of all serverless endpoints.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: endpointNestedAttributes(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *EndpointsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read queries the Vast.ai API for all endpoints.
func (d *EndpointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model EndpointsDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing all endpoints")

	// Call API
	endpoints, err := d.client.Endpoints.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to List Endpoints",
			fmt.Sprintf("An unexpected error occurred while listing endpoints: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Endpoints list complete", map[string]interface{}{
		"count": len(endpoints),
	})

	// Convert API endpoints to Terraform models
	endpointModels := make([]EndpointModel, len(endpoints))
	for i, e := range endpoints {
		endpointModels[i] = apiEndpointToModel(e)
	}

	// Build typed list
	endpointsList, diags := endpointModelsToList(endpointModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Endpoints = endpointsList

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// apiEndpointToModel converts a client.Endpoint to an EndpointModel.
func apiEndpointToModel(e client.Endpoint) EndpointModel {
	return EndpointModel{
		ID:            types.Int64Value(int64(e.ID)),
		EndpointName:  types.StringValue(e.EndpointName),
		MinLoad:       types.Float64Value(e.MinLoad),
		MinColdLoad:   types.Float64Value(e.MinColdLoad),
		TargetUtil:    types.Float64Value(e.TargetUtil),
		ColdMult:      types.Float64Value(e.ColdMult),
		ColdWorkers:   types.Int64Value(int64(e.ColdWorkers)),
		MaxWorkers:    types.Int64Value(int64(e.MaxWorkers)),
		EndpointState: types.StringValue(e.EndpointState),
	}
}

// endpointModelToAttrValues converts an EndpointModel to a map of attr.Value.
func endpointModelToAttrValues(m EndpointModel) map[string]attr.Value {
	return map[string]attr.Value{
		"id":             m.ID,
		"endpoint_name":  m.EndpointName,
		"min_load":       m.MinLoad,
		"min_cold_load":  m.MinColdLoad,
		"target_util":    m.TargetUtil,
		"cold_mult":      m.ColdMult,
		"cold_workers":   m.ColdWorkers,
		"max_workers":    m.MaxWorkers,
		"endpoint_state": m.EndpointState,
	}
}

// endpointModelsToList converts a slice of EndpointModel to a types.List.
func endpointModelsToList(models []EndpointModel) (types.List, diag.Diagnostics) {
	attrTypes := endpointModelAttrTypes()
	if len(models) == 0 {
		return types.ListValueMust(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{}), nil
	}

	elems := make([]attr.Value, len(models))
	for i, m := range models {
		obj, diags := types.ObjectValue(attrTypes, endpointModelToAttrValues(m))
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}
		elems[i] = obj
	}

	return types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
}
