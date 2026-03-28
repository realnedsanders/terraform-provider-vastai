package clustermember

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ClusterMemberResourceModel describes the resource data model for vastai_cluster_member.
// Uses composite ID format: cluster_id/machine_id.
type ClusterMemberResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	ClusterID        types.String   `tfsdk:"cluster_id"`
	MachineID        types.String   `tfsdk:"machine_id"`
	NewManagerID     types.String   `tfsdk:"new_manager_id"`
	IsClusterManager types.Bool     `tfsdk:"is_cluster_manager"`
	LocalIP          types.String   `tfsdk:"local_ip"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
