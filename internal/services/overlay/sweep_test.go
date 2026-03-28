package overlay

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

const testResourcePrefix = "tfacc-"

func init() {
	resource.AddTestSweepers("vastai_overlay", &resource.Sweeper{
		Name: "vastai_overlay",
		F:    sweepOverlays,
	})
}

func sweepOverlays(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	overlays, err := client.Overlays.List(ctx)
	if err != nil {
		return fmt.Errorf("error listing overlays: %s", err)
	}

	for _, ov := range overlays {
		if strings.HasPrefix(ov.Name, testResourcePrefix) {
			log.Printf("[INFO] Deleting overlay %d (%s)", ov.OverlayID, ov.Name)
			if err := client.Overlays.Delete(ctx, ov.OverlayID); err != nil {
				log.Printf("[ERROR] Failed to delete overlay %d: %s", ov.OverlayID, err)
			}
		}
	}
	return nil
}
