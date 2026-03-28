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

	for _, tmpl := range templates {
		if strings.HasPrefix(tmpl.Name, testResourcePrefix) {
			log.Printf("[INFO] Deleting template %s (%s)", tmpl.HashID, tmpl.Name)
			if err := client.Templates.Delete(ctx, tmpl.HashID); err != nil {
				log.Printf("[ERROR] Failed to delete template %s: %s", tmpl.HashID, err)
			}
		}
	}
	return nil
}
