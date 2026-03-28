package overlay_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// TestAccOverlay_basic verifies the full create, read, and destroy lifecycle of an overlay.
// This test requires an existing cluster backed by owned hardware. Set VASTAI_CLUSTER_ID to run.
func TestAccOverlay_basic(t *testing.T) {
	clusterID := os.Getenv("VASTAI_CLUSTER_ID")
	if clusterID == "" {
		t.Skip("Overlay tests require an existing cluster backed by owned hardware. Set VASTAI_CLUSTER_ID to run.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckOverlayDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccOverlayConfig_basic(clusterID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_overlay.test", "id"),
					resource.TestCheckResourceAttr("vastai_overlay.test", "name", "tfacc-overlay-basic"),
					resource.TestCheckResourceAttr("vastai_overlay.test", "cluster_id", clusterID),
					resource.TestCheckResourceAttrSet("vastai_overlay.test", "internal_subnet"),
				),
			},
		},
	})
}

func testAccOverlayConfig_basic(clusterID string) string {
	return fmt.Sprintf(`
resource "vastai_overlay" "test" {
  name       = "tfacc-overlay-basic"
  cluster_id = %q
}
`, clusterID)
}

// testAccCheckOverlayDestroy verifies that all overlays created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckOverlayDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_overlay" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing overlay ID %q: %s", rs.Primary.ID, err)
		}
		overlays, err := client.Overlays.List(context.Background())
		if err != nil {
			return fmt.Errorf("error listing overlays: %s", err)
		}
		for _, ov := range overlays {
			if ov.OverlayID == id {
				return fmt.Errorf("overlay %d still exists", id)
			}
		}
	}
	return nil
}
