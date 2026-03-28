package subaccount_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// TestAccSubaccount_basic verifies the create and read lifecycle of a subaccount.
// NOTE: The Vast.ai API does not support deleting subaccounts, so CheckDestroy is
// intentionally a no-op. Each test run creates a subaccount with a unique email
// using a timestamp to avoid conflicts.
func TestAccSubaccount_basic(t *testing.T) {
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("tfacc-%d@test.example.com", timestamp)
	username := fmt.Sprintf("tfacc-%d", timestamp)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSubaccountDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccSubaccountConfig_basic(email, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_subaccount.test", "id"),
					resource.TestCheckResourceAttr("vastai_subaccount.test", "email", email),
					resource.TestCheckResourceAttr("vastai_subaccount.test", "username", username),
				),
			},
		},
	})
}

func testAccSubaccountConfig_basic(email, username string) string {
	return fmt.Sprintf(`
resource "vastai_subaccount" "test" {
  email    = %q
  username = %q
  password = "tfacc-T3stP@ss!"
}
`, email, username)
}

// testAccCheckSubaccountDestroy is intentionally a no-op.
// The Vast.ai API does not provide a delete endpoint for subaccounts.
// Destroying the resource only removes it from Terraform state.
// We verify the subaccount still exists (expected) rather than asserting removal.
func testAccCheckSubaccountDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_subaccount" {
			continue
		}

		// The subaccount should still exist since there is no delete API.
		// We just verify the API is reachable; presence is expected.
		_, err := client.Subaccounts.List(context.Background())
		if err != nil {
			return fmt.Errorf("error listing subaccounts during destroy check: %s", err)
		}

		// No error means API is working. Subaccount persistence is expected.
	}
	return nil
}
