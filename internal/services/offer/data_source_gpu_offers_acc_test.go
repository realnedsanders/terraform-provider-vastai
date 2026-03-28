package offer_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccGpuOffersDataSource_basic verifies that a GPU offers search with a small limit
// returns results and populates the most_affordable convenience attribute.
func TestAccGpuOffersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGpuOffersDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.test", "offers.#"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.test", "offers.0.id"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.test", "offers.0.gpu_name"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.test", "offers.0.price_per_hour"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.test", "most_affordable.id"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.test", "most_affordable.gpu_name"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.test", "most_affordable.price_per_hour"),
				),
			},
		},
	})
}

// TestAccGpuOffersDataSource_filtered verifies that GPU offers search respects filter parameters
// such as num_gpus and max_price_per_hour.
func TestAccGpuOffersDataSource_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGpuOffersDataSourceConfig_filtered(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.filtered", "offers.#"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.filtered", "most_affordable.id"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.filtered", "most_affordable.num_gpus"),
					resource.TestCheckResourceAttrSet("data.vastai_gpu_offers.filtered", "most_affordable.price_per_hour"),
				),
			},
		},
	})
}

func testAccGpuOffersDataSourceConfig_basic() string {
	return `
data "vastai_gpu_offers" "test" {
  limit = 3
}
`
}

func testAccGpuOffersDataSourceConfig_filtered() string {
	return `
data "vastai_gpu_offers" "filtered" {
  num_gpus          = 1
  max_price_per_hour = 5.0
  limit             = 5
}
`
}
