package volume_test

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

// testVolumeBaseConfig provides a volume offers data source that finds the cheapest
// available offer, used as a foundation for all volume acceptance tests.
const testVolumeBaseConfig = `
data "vastai_volume_offers" "cheapest" {
  order_by          = "storage_cost"
  allocated_storage = 1
  limit             = 1
}
`

// TestAccVolume_basic verifies the full create, read, and destroy lifecycle
// of a local volume provisioned from the cheapest available volume offer.
func TestAccVolume_basic(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testVolumeBaseConfig + testAccVolumeConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_volume.test", "id"),
					resource.TestCheckResourceAttrSet("vastai_volume.test", "status"),
					resource.TestCheckResourceAttrSet("vastai_volume.test", "disk_space"),
					resource.TestCheckResourceAttrSet("vastai_volume.test", "machine_id"),
				),
			},
		},
	})
}

// TestAccVolume_import verifies that a volume can be imported by its contract ID.
func TestAccVolume_import(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testVolumeBaseConfig + testAccVolumeConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_volume.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:      "vastai_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"offer_id",
					"size",
					"name",
					"clone_from_id",
					"disable_compression",
					"timeouts",
				},
			},
		},
	})
}

// testAccCheckVolumeDestroy verifies that all volumes created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckVolumeDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_volume" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing volume ID %q: %s", rs.Primary.ID, err)
		}
		volumes, err := client.Volumes.List(context.Background(), "local_volume")
		if err != nil {
			return fmt.Errorf("error listing volumes: %s", err)
		}
		for _, vol := range volumes {
			if vol.ID == id {
				return fmt.Errorf("volume %d still exists", id)
			}
		}
	}
	return nil
}

func testAccVolumeConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "vastai_volume" "test" {
  offer_id = data.vastai_volume_offers.cheapest.most_affordable.id
  size     = 1
  name     = %q

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, name)
}
