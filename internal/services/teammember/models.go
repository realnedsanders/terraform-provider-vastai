package teammember

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TeamMemberResourceModel describes the resource data model for vastai_team_member.
type TeamMemberResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Email    types.String   `tfsdk:"email"`
	Role     types.String   `tfsdk:"role"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
