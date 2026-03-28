package sshkey

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

// Ensure SSHKeysDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &SSHKeysDataSource{}
var _ datasource.DataSourceWithConfigure = &SSHKeysDataSource{}

// SSHKeysDataSource defines the data source implementation.
type SSHKeysDataSource struct {
	client *client.VastAIClient
}

// NewSSHKeysDataSource returns a new SSH keys data source instance.
func NewSSHKeysDataSource() datasource.DataSource {
	return &SSHKeysDataSource{}
}

// Metadata returns the data source type name.
func (d *SSHKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys"
}

// Schema defines the schema for the data source.
func (d *SSHKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all SSH keys for the authenticated Vast.ai user.",
		Attributes: map[string]schema.Attribute{
			"ssh_keys": schema.ListNestedAttribute{
				Description: "List of SSH keys associated with the authenticated user.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique numeric identifier of the SSH key.",
							Computed:    true,
						},
						"ssh_key": schema.StringAttribute{
							Description: "SSH public key content.",
							Computed:    true,
							Sensitive:   true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the SSH key was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *SSHKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches all SSH keys from the API.
func (d *SSHKeysDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading SSH keys data source")

	keys, err := d.client.SSHKeys.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SSH Keys",
			fmt.Sprintf("Could not list SSH keys: %s", err),
		)
		return
	}

	var model SSHKeysDataSourceModel
	model.SSHKeys = make([]SSHKeyModel, len(keys))

	for i, key := range keys {
		model.SSHKeys[i] = SSHKeyModel{
			ID:        types.StringValue(strconv.Itoa(key.ID)),
			SSHKey:    types.StringValue(key.SSHKey),
			CreatedAt: types.StringValue(key.CreatedAt),
		}
	}

	tflog.Debug(ctx, "Read SSH keys data source", map[string]interface{}{
		"count": len(keys),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
