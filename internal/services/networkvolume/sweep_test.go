package networkvolume

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
	resource.AddTestSweepers("vastai_network_volume", &resource.Sweeper{
		Name: "vastai_network_volume",
		F:    sweepNetworkVolumes,
	})
}

func sweepNetworkVolumes(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	volumes, err := client.NetworkVolumes.List(ctx)
	if err != nil {
		return fmt.Errorf("error listing network volumes: %s", err)
	}

	var errs []error
	for _, vol := range volumes {
		if strings.HasPrefix(vol.Label, testResourcePrefix) {
			log.Printf("[INFO] Deleting network volume %d (%s)", vol.ID, vol.Label)
			if err := client.NetworkVolumes.Delete(ctx, vol.ID); err != nil {
				errs = append(errs, fmt.Errorf("error deleting network volume %d (%s): %w", vol.ID, vol.Label, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %v", errs)
	}
	return nil
}
