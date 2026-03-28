package endpoint

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
	resource.AddTestSweepers("vastai_endpoint", &resource.Sweeper{
		Name:         "vastai_endpoint",
		F:            sweepEndpoints,
		Dependencies: []string{"vastai_worker_group"},
	})
}

func sweepEndpoints(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	endpoints, err := client.Endpoints.List(ctx)
	if err != nil {
		return fmt.Errorf("error listing endpoints: %s", err)
	}

	var errs []error
	for _, ep := range endpoints {
		if strings.HasPrefix(ep.EndpointName, testResourcePrefix) {
			log.Printf("[INFO] Deleting endpoint %d (%s)", ep.ID, ep.EndpointName)
			if err := client.Endpoints.Delete(ctx, ep.ID); err != nil {
				errs = append(errs, fmt.Errorf("error deleting endpoint %d (%s): %w", ep.ID, ep.EndpointName, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %v", errs)
	}
	return nil
}
