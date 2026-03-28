package team

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TeamResourceModel describes the resource data model for vastai_team.
type TeamResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	TeamName types.String   `tfsdk:"team_name"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
