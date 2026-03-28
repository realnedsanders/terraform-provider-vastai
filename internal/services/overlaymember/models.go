package overlaymember

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OverlayMemberResourceModel describes the resource data model for vastai_overlay_member.
// Uses composite ID format: overlay_id/instance_id.
type OverlayMemberResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	OverlayName types.String   `tfsdk:"overlay_name"`
	OverlayID   types.String   `tfsdk:"overlay_id"`
	InstanceID  types.String   `tfsdk:"instance_id"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}
