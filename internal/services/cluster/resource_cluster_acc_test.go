package cluster_test

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

// TestAccCluster_basic verifies the full create, read, and destroy lifecycle of a cluster.
// This test requires owned GPU hardware. Set VASTAI_MACHINE_ID to run.
func TestAccCluster_basic(t *testing.T) {
	machineID := os.Getenv("VASTAI_MACHINE_ID")
	if machineID == "" {
		t.Skip("Cluster tests require owned GPU machines (machine IDs). Set VASTAI_MACHINE_ID to run.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccClusterConfig_basic(machineID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_cluster.test", "id"),
					resource.TestCheckResourceAttr("vastai_cluster.test", "subnet", "10.99.0.0/24"),
					resource.TestCheckResourceAttr("vastai_cluster.test", "manager_id", machineID),
				),
			},
		},
	})
}

func testAccClusterConfig_basic(machineID string) string {
	return fmt.Sprintf(`
resource "vastai_cluster" "test" {
  subnet     = "10.99.0.0/24"
  manager_id = %q
}
`, machineID)
}

// testAccCheckClusterDestroy verifies that all clusters created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckClusterDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_cluster" {
			continue
		}
		clusters, err := client.Clusters.List(context.Background())
		if err != nil {
			return fmt.Errorf("error listing clusters: %s", err)
		}
		if _, ok := clusters[rs.Primary.ID]; ok {
			return fmt.Errorf("cluster %s still exists", rs.Primary.ID)
		}
	}
	return nil
}
