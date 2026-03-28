package subaccount

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SubaccountResourceModel describes the resource data model for vastai_subaccount.
type SubaccountResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Email    types.String   `tfsdk:"email"`
	Username types.String   `tfsdk:"username"`
	Password types.String   `tfsdk:"password"`
	HostOnly types.Bool     `tfsdk:"host_only"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
