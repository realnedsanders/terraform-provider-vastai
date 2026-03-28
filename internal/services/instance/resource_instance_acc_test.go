package instance_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// testInstanceBaseConfig provides a GPU offers data source that finds the cheapest
// available offer, used as a foundation for all instance acceptance tests.
// Per D-21: use cheapest available offer to minimize cost.
const testInstanceBaseConfig = `
data "vastai_gpu_offers" "cheapest" {
  num_gpus           = 1
  max_price_per_hour = 0.50
  order_by           = "dph_total"
  limit              = 1
}
`

// TestAccInstance_basic verifies the full create, read, and destroy lifecycle
// of an instance provisioned from the cheapest available GPU offer.
func TestAccInstance_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testInstanceBaseConfig + testAccInstanceConfig_basic("tfacc-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_instance.test", "id"),
					resource.TestCheckResourceAttrSet("vastai_instance.test", "actual_status"),
					resource.TestCheckResourceAttrSet("vastai_instance.test", "ssh_host"),
					resource.TestCheckResourceAttrSet("vastai_instance.test", "gpu_name"),
					resource.TestCheckResourceAttrSet("vastai_instance.test", "machine_id"),
					resource.TestCheckResourceAttrSet("vastai_instance.test", "cost_per_hour"),
				),
			},
		},
	})
}

// TestAccInstance_update verifies that an instance label can be updated in-place
// without recreating the instance.
func TestAccInstance_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// Create with initial label
			{
				Config: testInstanceBaseConfig + testAccInstanceConfig_basic("tfacc-initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_instance.test", "id"),
					resource.TestCheckResourceAttr("vastai_instance.test", "label", "tfacc-initial"),
				),
			},
			// Update label in-place
			{
				Config: testInstanceBaseConfig + testAccInstanceConfig_basic("tfacc-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_instance.test", "id"),
					resource.TestCheckResourceAttr("vastai_instance.test", "label", "tfacc-updated"),
				),
			},
		},
	})
}

// TestAccInstance_import verifies that an instance can be imported by its contract ID.
func TestAccInstance_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testInstanceBaseConfig + testAccInstanceConfig_basic("tfacc-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_instance.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:      "vastai_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"offer_id",
					"disk_gb",
					"timeouts",
					"ssh_key_ids",
					"image_login",
					"cancel_unavail",
					"env",
					"use_ssh",
					"use_jupyter_lab",
				},
			},
		},
	})
}

// testAccCheckInstanceDestroy verifies that all instances created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckInstanceDestroy(s *terraform.State) error {
	apiClient, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_instance" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing instance ID %q: %s", rs.Primary.ID, err)
		}
		inst, err := apiClient.Instances.Get(context.Background(), id)
		if err != nil {
			// A 404 means the instance is gone, which is the expected outcome.
			var apiErr *client.APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("error checking instance %d: %s", id, err)
		}
		// If the instance still exists, check if it's in a destroyed/terminal state.
		if inst.ActualStatus != "destroyed" && inst.ActualStatus != "exited" {
			return fmt.Errorf("instance %d still exists with status %q", id, inst.ActualStatus)
		}
	}
	return nil
}

func testAccInstanceConfig_basic(label string) string {
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
`, label)
}
