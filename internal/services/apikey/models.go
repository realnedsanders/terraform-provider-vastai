package apikey

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApiKeyResourceModel describes the resource data model for vastai_api_key.
type ApiKeyResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Key         types.String   `tfsdk:"key"`
	Permissions types.String   `tfsdk:"permissions"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}
