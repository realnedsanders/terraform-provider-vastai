package teamrole

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TeamRoleResourceModel describes the resource data model for vastai_team_role.
type TeamRoleResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Permissions types.String   `tfsdk:"permissions"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}
