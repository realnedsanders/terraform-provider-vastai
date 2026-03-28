package endpoint_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// testAccEndpointPreCheck skips the test if the VASTAI_ENDPOINT_TEST env var is not set.
// Endpoint creation requires specific account permissions or a paid plan.
func testAccEndpointPreCheck(t *testing.T) {
	t.Helper()
	acctest.TestAccPreCheck(t)
	if os.Getenv("VASTAI_ENDPOINT_TEST") == "" {
		t.Skip("VASTAI_ENDPOINT_TEST must be set to run endpoint acceptance tests (requires paid plan)")
	}
}

// TestAccEndpoint_basic verifies the full create, read, and destroy lifecycle
// of a serverless endpoint with basic autoscaling parameters.
func TestAccEndpoint_basic(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEndpointPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccEndpointConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_endpoint.test", "id"),
					resource.TestCheckResourceAttr("vastai_endpoint.test", "endpoint_name", name),
					resource.TestCheckResourceAttrSet("vastai_endpoint.test", "target_util"),
					resource.TestCheckResourceAttrSet("vastai_endpoint.test", "cold_mult"),
					resource.TestCheckResourceAttrSet("vastai_endpoint.test", "max_workers"),
				),
			},
		},
	})
}

// TestAccEndpoint_update verifies that endpoint autoscaling parameters can be
// updated in-place without recreating the endpoint.
func TestAccEndpoint_update(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)
	updatedName := fmt.Sprintf("tfacc-%d-updated", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEndpointPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			// Create with initial config
			{
				Config: testAccEndpointConfig_withParams(name, 5, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_endpoint.test", "id"),
					resource.TestCheckResourceAttr("vastai_endpoint.test", "endpoint_name", name),
					resource.TestCheckResourceAttr("vastai_endpoint.test", "cold_workers", "5"),
					resource.TestCheckResourceAttr("vastai_endpoint.test", "max_workers", "10"),
				),
			},
			// Update autoscaling params
			{
				Config: testAccEndpointConfig_withParams(updatedName, 2, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_endpoint.test", "id"),
					resource.TestCheckResourceAttr("vastai_endpoint.test", "endpoint_name", updatedName),
					resource.TestCheckResourceAttr("vastai_endpoint.test", "cold_workers", "2"),
					resource.TestCheckResourceAttr("vastai_endpoint.test", "max_workers", "5"),
				),
			},
		},
	})
}

// TestAccEndpoint_import verifies that an endpoint can be imported by its ID.
func TestAccEndpoint_import(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEndpointPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testAccEndpointConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_endpoint.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:            "vastai_endpoint.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

// testAccCheckEndpointDestroy verifies that all endpoints created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckEndpointDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_endpoint" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing endpoint ID %q: %s", rs.Primary.ID, err)
		}
		endpoints, err := client.Endpoints.List(context.Background())
		if err != nil {
			return fmt.Errorf("error listing endpoints: %s", err)
		}
		for _, ep := range endpoints {
			if ep.ID == id {
				return fmt.Errorf("endpoint %d still exists", id)
			}
		}
	}
	return nil
}

func testAccEndpointConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "vastai_endpoint" "test" {
  endpoint_name = %q

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, name)
}

func testAccEndpointConfig_withParams(name string, coldWorkers, maxWorkers int) string {
	return fmt.Sprintf(`
resource "vastai_endpoint" "test" {
  endpoint_name = %q
  cold_workers  = %d
  max_workers   = %d
  target_util   = 0.8

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}
`, name, coldWorkers, maxWorkers)
}
