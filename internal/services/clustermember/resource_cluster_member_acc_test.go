package clustermember_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// TestAccClusterMember_basic verifies the full create, read, and destroy lifecycle
// of a cluster member (machine joined to a cluster).
// This test requires owned GPU hardware. Set VASTAI_MACHINE_ID and VASTAI_CLUSTER_ID to run.
func TestAccClusterMember_basic(t *testing.T) {
	machineID := os.Getenv("VASTAI_MACHINE_ID")
	if machineID == "" {
		t.Skip("Cluster member tests require owned GPU machines. Set VASTAI_MACHINE_ID to run.")
	}
	clusterID := os.Getenv("VASTAI_CLUSTER_ID")
	if clusterID == "" {
		t.Skip("Cluster member tests require an existing cluster. Set VASTAI_CLUSTER_ID to run.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClusterMemberDestroy,
		Steps: []resource.TestStep{
			// Create (join machine to cluster) and read
			{
				Config: testAccClusterMemberConfig_basic(clusterID, machineID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_cluster_member.test", "id"),
					resource.TestCheckResourceAttr("vastai_cluster_member.test", "cluster_id", clusterID),
					resource.TestCheckResourceAttr("vastai_cluster_member.test", "machine_id", machineID),
					resource.TestCheckResourceAttrSet("vastai_cluster_member.test", "is_cluster_manager"),
					resource.TestCheckResourceAttrSet("vastai_cluster_member.test", "local_ip"),
				),
			},
		},
	})
}

func testAccClusterMemberConfig_basic(clusterID, machineID string) string {
	return fmt.Sprintf(`
resource "vastai_cluster_member" "test" {
  cluster_id = %q
  machine_id = %q
}
`, clusterID, machineID)
}

// testAccCheckClusterMemberDestroy verifies that the machine has been removed from the cluster.
func testAccCheckClusterMemberDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_cluster_member" {
			continue
		}

		clusterID := rs.Primary.Attributes["cluster_id"]
		machineID := rs.Primary.Attributes["machine_id"]

		clusters, err := client.Clusters.List(context.Background())
		if err != nil {
			return fmt.Errorf("error listing clusters: %s", err)
		}

		cluster, ok := clusters[clusterID]
		if !ok {
			// Cluster itself is gone, so member is implicitly gone.
			continue
		}

		for _, node := range cluster.Nodes {
			if fmt.Sprintf("%d", node.MachineID) == machineID {
				return fmt.Errorf("machine %s still in cluster %s", machineID, clusterID)
			}
		}
	}
	return nil
}
