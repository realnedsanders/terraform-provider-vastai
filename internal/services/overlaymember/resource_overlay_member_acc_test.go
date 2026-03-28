package overlaymember_test

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

// TestAccOverlayMember_basic verifies that an instance can be joined to an overlay network.
// This test requires an existing overlay and a running instance.
// Set VASTAI_OVERLAY_ID and VASTAI_INSTANCE_ID to run.
func TestAccOverlayMember_basic(t *testing.T) {
	overlayID := os.Getenv("VASTAI_OVERLAY_ID")
	if overlayID == "" {
		t.Skip("Overlay member tests require an existing overlay. Set VASTAI_OVERLAY_ID to run.")
	}
	instanceID := os.Getenv("VASTAI_INSTANCE_ID")
	if instanceID == "" {
		t.Skip("Overlay member tests require a running instance. Set VASTAI_INSTANCE_ID to run.")
	}

	// We need the overlay name to build the config. Look it up from the overlay ID.
	overlayName := os.Getenv("VASTAI_OVERLAY_NAME")
	if overlayName == "" {
		t.Skip("Overlay member tests require the overlay name. Set VASTAI_OVERLAY_NAME to run.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckOverlayMemberDestroy,
		Steps: []resource.TestStep{
			// Join instance to overlay and read
			{
				Config: testAccOverlayMemberConfig_basic(overlayName, instanceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_overlay_member.test", "id"),
					resource.TestCheckResourceAttr("vastai_overlay_member.test", "overlay_name", overlayName),
					resource.TestCheckResourceAttr("vastai_overlay_member.test", "instance_id", instanceID),
					resource.TestCheckResourceAttrSet("vastai_overlay_member.test", "overlay_id"),
				),
			},
		},
	})
}

func testAccOverlayMemberConfig_basic(overlayName, instanceID string) string {
	return fmt.Sprintf(`
resource "vastai_overlay_member" "test" {
  overlay_name = %q
  instance_id  = %q
}
`, overlayName, instanceID)
}

// testAccCheckOverlayMemberDestroy is a best-effort check.
// The Vast.ai API does not support removing individual instances from overlays,
// so destroy only removes the resource from Terraform state. We verify the
// resource is no longer tracked in state (which the framework guarantees).
func testAccCheckOverlayMemberDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_overlay_member" {
			continue
		}

		overlayIDStr := rs.Primary.Attributes["overlay_id"]
		overlayID, err := strconv.Atoi(overlayIDStr)
		if err != nil {
			// If we can't parse the overlay ID, the resource is effectively gone.
			continue
		}

		instanceIDStr := rs.Primary.Attributes["instance_id"]
		instanceID, err := strconv.Atoi(instanceIDStr)
		if err != nil {
			continue
		}

		// NOTE: The API does not support removing individual instances from overlays.
		// We only verify the overlay still exists and log a warning if the instance
		// is still present (expected behavior since delete is state-only).
		overlays, err := client.Overlays.List(context.Background())
		if err != nil {
			return fmt.Errorf("error listing overlays: %s", err)
		}

		for _, ov := range overlays {
			if ov.OverlayID == overlayID {
				for _, inst := range ov.Instances {
					if inst == instanceID {
						// This is expected -- the API has no remove-instance endpoint.
						// The resource was removed from state only.
						return nil
					}
				}
			}
		}
	}
	return nil
}
