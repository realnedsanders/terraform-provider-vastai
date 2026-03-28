package workergroup

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
	resource.AddTestSweepers("vastai_worker_group", &resource.Sweeper{
		Name: "vastai_worker_group",
		F:    sweepWorkerGroups,
	})
}

func sweepWorkerGroups(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	groups, err := client.WorkerGroups.List(ctx)
	if err != nil {
		return fmt.Errorf("error listing worker groups: %s", err)
	}

	var errs []error
	for _, wg := range groups {
		// Worker groups don't have their own name, but belong to named endpoints.
		// Delete worker groups associated with test endpoints (tfacc- prefix on endpoint name).
		if strings.HasPrefix(wg.EndpointName, testResourcePrefix) {
			log.Printf("[INFO] Deleting worker group %d (endpoint: %s)", wg.ID, wg.EndpointName)
			if err := client.WorkerGroups.Delete(ctx, wg.ID); err != nil {
				errs = append(errs, fmt.Errorf("error deleting worker group %d (endpoint: %s): %w", wg.ID, wg.EndpointName, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %v", errs)
	}
	return nil
}
