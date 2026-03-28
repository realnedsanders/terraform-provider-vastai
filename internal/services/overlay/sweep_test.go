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

	var errs []error
	for _, ov := range overlays {
		if strings.HasPrefix(ov.Name, testResourcePrefix) {
			log.Printf("[INFO] Deleting overlay %d (%s)", ov.OverlayID, ov.Name)
			if err := client.Overlays.Delete(ctx, ov.OverlayID); err != nil {
				errs = append(errs, fmt.Errorf("error deleting overlay %d (%s): %w", ov.OverlayID, ov.Name, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %v", errs)
	}
	return nil
}
