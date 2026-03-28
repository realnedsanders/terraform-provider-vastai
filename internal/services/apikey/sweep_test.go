package apikey

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
	resource.AddTestSweepers("vastai_api_key", &resource.Sweeper{
		Name: "vastai_api_key",
		F:    sweepApiKeys,
	})
}

func sweepApiKeys(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	keys, err := client.ApiKeys.List(ctx)
	if err != nil {
		return fmt.Errorf("error listing API keys: %s", err)
	}

	var errs []error
	for _, key := range keys {
		if strings.HasPrefix(key.Name, testResourcePrefix) {
			log.Printf("[INFO] Deleting API key %d (%s)", key.ID, key.Name)
			if err := client.ApiKeys.Delete(ctx, key.ID); err != nil {
				errs = append(errs, fmt.Errorf("error deleting API key %d (%s): %w", key.ID, key.Name, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %v", errs)
	}
	return nil
}
