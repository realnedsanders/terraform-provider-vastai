package workergroup_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// TestAccWorkerGroup_basic verifies the full create, read, and destroy lifecycle
// of a worker group bound to a serverless endpoint.
func TestAccWorkerGroup_basic(t *testing.T) {
	rInt := rand.Int()
	endpointName := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWorkerGroupDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccWorkerGroupConfig_basic(endpointName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_worker_group.test", "id"),
					resource.TestCheckResourceAttrSet("vastai_worker_group.test", "endpoint_id"),
					resource.TestCheckResourceAttrSet("vastai_worker_group.test", "endpoint_name"),
					resource.TestCheckResourceAttrSet("vastai_worker_group.test", "test_workers"),
				),
			},
		},
	})
}

// TestAccWorkerGroup_import verifies that a worker group can be imported by its ID.
func TestAccWorkerGroup_import(t *testing.T) {
	rInt := rand.Int()
	endpointName := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWorkerGroupDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testAccWorkerGroupConfig_basic(endpointName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_worker_group.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:      "vastai_worker_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"timeouts",
				},
			},
		},
	})
}

// testAccCheckWorkerGroupDestroy verifies that all worker groups created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckWorkerGroupDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_worker_group" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing worker group ID %q: %s", rs.Primary.ID, err)
		}
		groups, err := client.WorkerGroups.List(context.Background())
		if err != nil {
			return fmt.Errorf("error listing worker groups: %s", err)
		}
		for _, wg := range groups {
			if wg.ID == id {
				return fmt.Errorf("worker group %d still exists", id)
			}
		}
	}
	return nil
}

func testAccWorkerGroupConfig_basic(endpointName string) string {
	return fmt.Sprintf(`
resource "vastai_template" "test" {
  name  = "%s-tmpl"
  image = "ubuntu:22.04"
}

resource "vastai_endpoint" "test" {
  endpoint_name = %q

  timeouts {
    create = "5m"
    delete = "5m"
  }
}

resource "vastai_worker_group" "test" {
  endpoint_id   = vastai_endpoint.test.id
  template_hash = vastai_template.test.id
  test_workers  = 0
  cold_workers  = 0
  search_params = "gpu_ram>=1"

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, endpointName, endpointName)
}
