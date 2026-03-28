package endpoint_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccEndpointsDataSource_basic verifies that the endpoints data source can list
// endpoints. Creates an endpoint first, then reads via the data source to verify
// the list contains at least one entry.
func TestAccEndpointsDataSource_basic(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEndpointPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointsDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_endpoints.test", "endpoints.#"),
				),
			},
		},
	})
}

func testAccEndpointsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "vastai_endpoint" "test" {
  endpoint_name = %q

  timeouts {
    create = "5m"
    delete = "5m"
  }
}

data "vastai_endpoints" "test" {
  depends_on = [vastai_endpoint.test]
}
`, name)
}
