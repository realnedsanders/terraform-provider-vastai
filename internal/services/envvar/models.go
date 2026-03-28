package envvar

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvVarResourceModel describes the resource data model for vastai_environment_variable.
type EnvVarResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Key      types.String   `tfsdk:"key"`
	Value    types.String   `tfsdk:"value"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
