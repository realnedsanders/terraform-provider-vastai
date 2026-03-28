package user

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

// Ensure UserDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &UserDataSource{}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	client *client.VastAIClient
}

// NewUserDataSource returns a new user data source instance.
func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// Metadata returns the data source type name.
func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the data source.
func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the current authenticated user's profile information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique numeric identifier of the user.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username of the authenticated user.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address of the authenticated user.",
				Computed:    true,
			},
			"email_verified": schema.BoolAttribute{
				Description: "Whether the user's email address has been verified.",
				Computed:    true,
			},
			"fullname": schema.StringAttribute{
				Description: "Full name of the authenticated user.",
				Computed:    true,
			},
			"balance": schema.Float64Attribute{
				Description: "Current account balance in USD.",
				Computed:    true,
			},
			"credit": schema.Float64Attribute{
				Description: "Current account credit in USD.",
				Computed:    true,
			},
			"has_billing": schema.BoolAttribute{
				Description: "Whether the user has billing information configured.",
				Computed:    true,
			},
			"ssh_key": schema.StringAttribute{
				Description: "Default SSH public key associated with the user.",
				Computed:    true,
				Sensitive:   true,
			},
			"balance_threshold": schema.Float64Attribute{
				Description: "Balance threshold for notifications.",
				Computed:    true,
			},
			"balance_threshold_enabled": schema.BoolAttribute{
				Description: "Whether balance threshold notifications are enabled.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the current user's profile from the API.
func (d *UserDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading user data source")

	user, err := d.client.Users.GetCurrent(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User Profile",
			fmt.Sprintf("Could not read user profile: %s", err),
		)
		return
	}

	model := UserDataSourceModel{
		ID:                      types.StringValue(strconv.Itoa(user.ID)),
		Username:                types.StringValue(user.Username),
		Email:                   types.StringValue(user.Email),
		EmailVerified:           types.BoolValue(user.EmailVerified),
		Fullname:                types.StringValue(user.Fullname),
		Balance:                 types.Float64Value(user.Balance),
		Credit:                  types.Float64Value(user.Credit),
		HasBilling:              types.BoolValue(user.HasBilling),
		SSHKey:                  types.StringValue(user.SSHKey),
		BalanceThreshold:        types.Float64Value(user.BalanceThreshold),
		BalanceThresholdEnabled: types.BoolValue(user.BalanceThresholdEnabled),
	}

	tflog.Debug(ctx, "Read user data source", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
