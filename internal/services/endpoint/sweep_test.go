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

	for _, ep := range endpoints {
		if strings.HasPrefix(ep.EndpointName, testResourcePrefix) {
			log.Printf("[INFO] Deleting endpoint %d (%s)", ep.ID, ep.EndpointName)
			if err := client.Endpoints.Delete(ctx, ep.ID); err != nil {
				log.Printf("[ERROR] Failed to delete endpoint %d: %s", ep.ID, err)
			}
		}
	}
	return nil
}
