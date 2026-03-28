package instance_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccInstanceDataSource_basic verifies that an instance can be looked up by ID
// via the vastai_instance data source after creating it as a resource.
func TestAccInstanceDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testInstanceBaseConfig + testAccInstanceDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_instance.lookup", "id"),
					resource.TestCheckResourceAttrSet("data.vastai_instance.lookup", "gpu_name"),
					resource.TestCheckResourceAttrSet("data.vastai_instance.lookup", "actual_status"),
					resource.TestCheckResourceAttrSet("data.vastai_instance.lookup", "machine_id"),
				),
			},
		},
	})
}

// TestAccInstancesDataSource_basic verifies that the vastai_instances data source
// can list instances and filter by label.
func TestAccInstancesDataSource_basic(t *testing.T) {
	rInt := rand.Int()
	label := fmt.Sprintf("tfacc-ds-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testInstanceBaseConfig + testAccInstancesDataSourceConfig_basic(label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_instances.filtered", "instances.#"),
				),
			},
		},
	})
}

func testAccInstanceDataSourceConfig_basic() string {
	return `
resource "vastai_instance" "test" {
  offer_id = data.vastai_gpu_offers.cheapest.offers[0].id
  image    = "ubuntu:22.04"
  disk_gb  = 10
  label    = "tfacc-ds-lookup"

  timeouts {
    create = "15m"
    delete = "5m"
  }
}

data "vastai_instance" "lookup" {
  id = vastai_instance.test.id
}
`
}

func testAccInstancesDataSourceConfig_basic(label string) string {
	return fmt.Sprintf(`
resource "vastai_instance" "test" {
  offer_id = data.vastai_gpu_offers.cheapest.offers[0].id
  image    = "ubuntu:22.04"
  disk_gb  = 10
  label    = %q

  timeouts {
    create = "15m"
    delete = "5m"
  }
}

data "vastai_instances" "filtered" {
  label = %q

  depends_on = [vastai_instance.test]
}
`, label, label)
}
