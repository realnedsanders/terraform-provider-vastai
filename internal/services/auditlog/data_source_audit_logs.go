package auditlog

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

// formatUnixTimestamp converts a Unix timestamp (float64) to a string representation.
func formatUnixTimestamp(ts float64) string {
	if ts == 0 {
		return ""
	}
	return strconv.FormatFloat(ts, 'f', -1, 64)
}

// Ensure AuditLogsDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &AuditLogsDataSource{}
var _ datasource.DataSourceWithConfigure = &AuditLogsDataSource{}

// AuditLogsDataSource defines the data source implementation.
type AuditLogsDataSource struct {
	client *client.VastAIClient
}

// NewAuditLogsDataSource returns a new audit logs data source instance.
func NewAuditLogsDataSource() datasource.DataSource {
	return &AuditLogsDataSource{}
}

// Metadata returns the data source type name.
func (d *AuditLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_logs"
}

// Schema defines the schema for the data source.
func (d *AuditLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves audit log entries for the authenticated Vast.ai user.",
		Attributes: map[string]schema.Attribute{
			"audit_logs": schema.ListNestedAttribute{
				Description: "List of audit log entries.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Description: "IP address from which the API call was made.",
							Computed:    true,
						},
						"api_key_id": schema.StringAttribute{
							Description: "ID of the API key used for the call.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the audit log entry was created.",
							Computed:    true,
						},
						"api_route": schema.StringAttribute{
							Description: "API route that was called.",
							Computed:    true,
						},
						"args": schema.StringAttribute{
							Description: "Arguments passed to the API call.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *AuditLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches all audit log entries from the API.
func (d *AuditLogsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading audit logs data source")

	entries, err := d.client.AuditLogs.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Audit Logs",
			fmt.Sprintf("Could not list audit logs: %s", err),
		)
		return
	}

	var model AuditLogsDataSourceModel
	model.AuditLogs = make([]AuditLogModel, len(entries))

	for i, entry := range entries {
		model.AuditLogs[i] = AuditLogModel{
			IPAddress: types.StringValue(entry.IPAddress),
			ApiKeyID:  types.StringValue(strconv.Itoa(entry.ApiKeyID)),
			CreatedAt: types.StringValue(formatUnixTimestamp(entry.CreatedAt)),
			ApiRoute:  types.StringValue(entry.ApiRoute),
			Args:      types.StringValue(fmt.Sprintf("%v", entry.Args)),
		}
	}

	tflog.Debug(ctx, "Read audit logs data source", map[string]interface{}{
		"count": len(entries),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
