package overlay

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OverlayResourceModel describes the resource data model for vastai_overlay.
type OverlayResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	Name           types.String   `tfsdk:"name"`
	ClusterID      types.String   `tfsdk:"cluster_id"`
	InternalSubnet types.String   `tfsdk:"internal_subnet"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}
