package instance

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
	resource.AddTestSweepers("vastai_instance", &resource.Sweeper{
		Name: "vastai_instance",
		F:    sweepInstances,
	})
}

func sweepInstances(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	instances, err := client.Instances.List(ctx)
	if err != nil {
		return fmt.Errorf("error listing instances: %s", err)
	}

	var errs []error
	for _, inst := range instances {
		if strings.HasPrefix(inst.Label, testResourcePrefix) {
			log.Printf("[INFO] Destroying instance %d (%s)", inst.ID, inst.Label)
			if err := client.Instances.Destroy(ctx, inst.ID); err != nil {
				errs = append(errs, fmt.Errorf("error destroying instance %d (%s): %w", inst.ID, inst.Label, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %v", errs)
	}
	return nil
}
