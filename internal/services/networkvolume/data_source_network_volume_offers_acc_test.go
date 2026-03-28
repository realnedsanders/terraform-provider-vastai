package networkvolume_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccNetworkVolumeOffersDataSource_basic verifies that the network volume offers
// data source returns a non-empty list with the most_affordable convenience attribute set.
func TestAccNetworkVolumeOffersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkVolumeOffersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_network_volume_offers.test", "offers.#"),
					resource.TestCheckResourceAttrSet("data.vastai_network_volume_offers.test", "most_affordable.id"),
					resource.TestCheckResourceAttrSet("data.vastai_network_volume_offers.test", "most_affordable.storage_cost"),
					resource.TestCheckResourceAttrSet("data.vastai_network_volume_offers.test", "most_affordable.disk_space"),
				),
			},
		},
	})
}

func testAccNetworkVolumeOffersDataSourceConfig() string {
	return `
data "vastai_network_volume_offers" "test" {
  order_by          = "storage_cost"
  allocated_storage = 1
  limit             = 5
}
`
}
