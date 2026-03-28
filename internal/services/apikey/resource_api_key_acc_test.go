package apikey_test

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

// TestAccApiKey_basic verifies the full create, read, and destroy lifecycle of an API key.
func TestAccApiKey_basic(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApiKeyDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccApiKeyConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_api_key.test", "id"),
					resource.TestCheckResourceAttr("vastai_api_key.test", "name", name),
					resource.TestCheckResourceAttrSet("vastai_api_key.test", "key"),
				),
			},
		},
	})
}

// TestAccApiKey_import verifies that an API key can be imported by its numeric ID.
func TestAccApiKey_import(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApiKeyDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testAccApiKeyConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_api_key.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:            "vastai_api_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key", "timeouts"},
			},
		},
	})
}

func testAccApiKeyConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "vastai_api_key" "test" {
  name        = %q
  permissions = "{}"
}
`, name)
}

// testAccCheckApiKeyDestroy verifies that all API keys created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckApiKeyDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_api_key" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing API key ID %q: %s", rs.Primary.ID, err)
		}
		keys, err := client.ApiKeys.List(context.Background())
		if err != nil {
			return fmt.Errorf("error checking API key: %s", err)
		}
		for _, k := range keys {
			if k.ID == id {
				return fmt.Errorf("API key %d still exists", id)
			}
		}
	}
	return nil
}
