package volume

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
	resource.AddTestSweepers("vastai_volume", &resource.Sweeper{
		Name: "vastai_volume",
		F:    sweepVolumes,
	})
}

func sweepVolumes(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	volumes, err := client.Volumes.List(ctx, "local_volume")
	if err != nil {
		return fmt.Errorf("error listing volumes: %s", err)
	}

	var errs []error
	for _, vol := range volumes {
		if strings.HasPrefix(vol.Label, testResourcePrefix) {
			log.Printf("[INFO] Deleting volume %d (%s)", vol.ID, vol.Label)
			if err := client.Volumes.Delete(ctx, vol.ID); err != nil {
				errs = append(errs, fmt.Errorf("error deleting volume %d (%s): %w", vol.ID, vol.Label, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %v", errs)
	}
	return nil
}
