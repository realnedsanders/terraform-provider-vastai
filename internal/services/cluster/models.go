package cluster

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ClusterResourceModel describes the resource data model for vastai_cluster.
type ClusterResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	Subnet    types.String   `tfsdk:"subnet"`
	ManagerID types.String   `tfsdk:"manager_id"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}
