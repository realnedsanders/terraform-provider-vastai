package template

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
	resource.AddTestSweepers("vastai_template", &resource.Sweeper{
		Name: "vastai_template",
		F:    sweepTemplates,
	})
}

func sweepTemplates(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	// Search with empty query to list all user templates
	templates, err := client.Templates.Search(ctx, "")
	if err != nil {
		return fmt.Errorf("error listing templates: %s", err)
	}

	var errs []error
	for _, tmpl := range templates {
		if strings.HasPrefix(tmpl.Name, testResourcePrefix) {
			log.Printf("[INFO] Deleting template %s (id=%d, %s)", tmpl.HashID, tmpl.ID, tmpl.Name)
			if err := client.Templates.DeleteByID(ctx, tmpl.ID); err != nil {
				errs = append(errs, fmt.Errorf("error deleting template %s (id=%d, %s): %w", tmpl.HashID, tmpl.ID, tmpl.Name, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %v", errs)
	}
	return nil
}
