package networkvolume_test

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

// testNetworkVolumeBaseConfig provides a network volume offers data source that finds
// the cheapest available offer, used as a foundation for all network volume acceptance tests.
const testNetworkVolumeBaseConfig = `
data "vastai_network_volume_offers" "cheapest" {
  order_by          = "storage_cost"
  allocated_storage = 1
  limit             = 1
}
`

// TestAccNetworkVolume_basic verifies the full create, read, and destroy lifecycle
// of a network volume provisioned from the cheapest available offer.
func TestAccNetworkVolume_basic(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNetworkVolumeDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testNetworkVolumeBaseConfig + testAccNetworkVolumeConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_network_volume.test", "id"),
					resource.TestCheckResourceAttrSet("vastai_network_volume.test", "status"),
					resource.TestCheckResourceAttrSet("vastai_network_volume.test", "disk_space"),
					resource.TestCheckResourceAttrSet("vastai_network_volume.test", "machine_id"),
				),
			},
		},
	})
}

// TestAccNetworkVolume_import verifies that a network volume can be imported by its contract ID.
func TestAccNetworkVolume_import(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNetworkVolumeDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testNetworkVolumeBaseConfig + testAccNetworkVolumeConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_network_volume.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:      "vastai_network_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"offer_id",
					"size",
					"name",
					"timeouts",
				},
			},
		},
	})
}

// testAccCheckNetworkVolumeDestroy verifies that all network volumes created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckNetworkVolumeDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_network_volume" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing network volume ID %q: %s", rs.Primary.ID, err)
		}
		volumes, err := client.NetworkVolumes.List(context.Background())
		if err != nil {
			return fmt.Errorf("error listing network volumes: %s", err)
		}
		for _, vol := range volumes {
			if vol.ID == id {
				return fmt.Errorf("network volume %d still exists", id)
			}
		}
	}
	return nil
}

func testAccNetworkVolumeConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "vastai_network_volume" "test" {
  offer_id = data.vastai_network_volume_offers.cheapest.most_affordable.id
  size     = 1
  name     = %q

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, name)
}
