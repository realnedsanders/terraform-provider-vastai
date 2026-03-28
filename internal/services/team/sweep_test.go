package team

// Team sweeper is intentionally limited to team roles only.
//
// The Team API is account-scoped: there is no List endpoint to enumerate teams,
// and DestroyTeam() is a parameterless delete that destroys the entire team
// associated with the current API key. There is no safe way to filter by
// "tfacc-" prefix since teams don't support listing.
//
// However, team roles DO support listing and have a name field, so we sweep
// roles that were created by tests.
//
// Acceptance tests that create teams should clean up in their own
// CheckDestroy functions.

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
	resource.AddTestSweepers("vastai_team_role", &resource.Sweeper{
		Name: "vastai_team_role",
		F:    sweepTeamRoles,
	})
}

func sweepTeamRoles(_ string) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	roles, err := client.Teams.ListRoles(ctx)
	if err != nil {
		// Teams may not exist; treat list failure as non-fatal
		log.Printf("[WARN] Failed to list team roles (team may not exist): %s", err)
		return nil
	}

	for _, role := range roles {
		if strings.HasPrefix(role.Name, testResourcePrefix) {
			log.Printf("[INFO] Deleting team role %d (%s)", role.ID, role.Name)
			if err := client.Teams.DeleteRole(ctx, role.Name); err != nil {
				log.Printf("[ERROR] Failed to delete team role %s: %s", role.Name, err)
			}
		}
	}
	return nil
}
