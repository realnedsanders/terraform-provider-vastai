package invoice

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

// Ensure InvoicesDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &InvoicesDataSource{}

// InvoicesDataSource defines the data source implementation.
type InvoicesDataSource struct {
	client *client.VastAIClient
}

// NewInvoicesDataSource returns a new invoices data source instance.
func NewInvoicesDataSource() datasource.DataSource {
	return &InvoicesDataSource{}
}

// Metadata returns the data source type name.
func (d *InvoicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invoices"
}

// Schema defines the schema for the data source.
func (d *InvoicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves billing invoices for the authenticated Vast.ai user. Uses the v1 API endpoint.",
		Attributes: map[string]schema.Attribute{
			"start_date": schema.StringAttribute{
				Description: "Start date for filtering (YYYY-MM-DD format).",
				Optional:    true,
			},
			"end_date": schema.StringAttribute{
				Description: "End date for filtering (YYYY-MM-DD format).",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum number of invoices to return. Defaults to 100.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Invoice type filter.",
				Optional:    true,
			},
			"invoices": schema.ListNestedAttribute{
				Description: "List of billing invoices.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique numeric identifier of the invoice.",
							Computed:    true,
						},
						"amount": schema.Float64Attribute{
							Description: "Invoice amount in USD.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of invoice entry.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the invoice entry.",
							Computed:    true,
						},
						"timestamp": schema.StringAttribute{
							Description: "Timestamp when the invoice entry was created.",
							Computed:    true,
						},
					},
				},
			},
			"total": schema.Int64Attribute{
				Description: "Total number of matching invoices.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *InvoicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches invoices from the API with optional filtering.
func (d *InvoicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading invoices data source")

	var model InvoicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build query parameters from model.
	params := client.InvoiceListParams{}
	if !model.StartDate.IsNull() && !model.StartDate.IsUnknown() {
		params.StartDate = model.StartDate.ValueString()
	}
	if !model.EndDate.IsNull() && !model.EndDate.IsUnknown() {
		params.EndDate = model.EndDate.ValueString()
	}
	if !model.Limit.IsNull() && !model.Limit.IsUnknown() {
		params.Limit = int(model.Limit.ValueInt64())
	}
	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		params.Type = model.Type.ValueString()
	}

	result, err := d.client.Invoices.List(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Invoices",
			fmt.Sprintf("Could not list invoices: %s", err),
		)
		return
	}

	model.Invoices = make([]InvoiceModel, len(result.Results))
	for i, inv := range result.Results {
		model.Invoices[i] = InvoiceModel{
			ID:          types.StringValue(strconv.Itoa(inv.ID)),
			Amount:      types.Float64Value(inv.Amount),
			Type:        types.StringValue(inv.Type),
			Description: types.StringValue(inv.Description),
			Timestamp:   types.StringValue(inv.Timestamp),
		}
	}
	model.Total = types.Int64Value(int64(result.Total))

	tflog.Debug(ctx, "Read invoices data source", map[string]interface{}{
		"count": len(result.Results),
		"total": result.Total,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
