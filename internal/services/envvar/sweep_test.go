package envvar

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
	resource.AddTestSweepers("vastai_environment_variable", &resource.Sweeper{
		Name: "vastai_environment_variable",
		F:    sweepEnvVars,
	})
}

func sweepEnvVars(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	envVars, err := client.EnvVars.List(ctx)
	if err != nil {
		return fmt.Errorf("error listing environment variables: %s", err)
	}

	for key := range envVars {
		if strings.HasPrefix(key, testResourcePrefix) {
			log.Printf("[INFO] Deleting environment variable %s", key)
			if err := client.EnvVars.Delete(ctx, key); err != nil {
				log.Printf("[ERROR] Failed to delete environment variable %s: %s", key, err)
			}
		}
	}
	return nil
}
