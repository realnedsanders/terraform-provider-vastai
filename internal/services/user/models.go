package user

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UserDataSourceModel describes the data source data model for vastai_user.
type UserDataSourceModel struct {
	ID                      types.String  `tfsdk:"id"`
	Username                types.String  `tfsdk:"username"`
	Email                   types.String  `tfsdk:"email"`
	EmailVerified           types.Bool    `tfsdk:"email_verified"`
	Fullname                types.String  `tfsdk:"fullname"`
	Balance                 types.Float64 `tfsdk:"balance"`
	Credit                  types.Float64 `tfsdk:"credit"`
	HasBilling              types.Bool    `tfsdk:"has_billing"`
	SSHKey                  types.String  `tfsdk:"ssh_key"`
	BalanceThreshold        types.Float64 `tfsdk:"balance_threshold"`
	BalanceThresholdEnabled types.Bool    `tfsdk:"balance_threshold_enabled"`
}
