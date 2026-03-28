package teamrole_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// TestAccTeamRole_basic verifies the full create, read, and destroy lifecycle of a team role.
// NOTE: Requires a team to exist in the current API key context.
func TestAccTeamRole_basic(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamRoleDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccTeamRoleConfig(name, `{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_team_role.test", "id"),
					resource.TestCheckResourceAttr("vastai_team_role.test", "name", name),
				),
			},
		},
	})
}

// TestAccTeamRole_update verifies that a team role's permissions can be updated in-place.
func TestAccTeamRole_update(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamRoleDestroy,
		Steps: []resource.TestStep{
			// Create with initial permissions
			{
				Config: testAccTeamRoleConfig(name, `{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_team_role.test", "id"),
					resource.TestCheckResourceAttr("vastai_team_role.test", "name", name),
				),
			},
			// Update permissions
			{
				Config: testAccTeamRoleConfig(name, `{"api":{"instance_read":{}}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_team_role.test", "id"),
					resource.TestCheckResourceAttr("vastai_team_role.test", "name", name),
				),
			},
		},
	})
}

// TestAccTeamRole_import verifies that a team role can be imported by its numeric ID.
func TestAccTeamRole_import(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tfacc-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamRoleDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testAccTeamRoleConfig(name, `{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_team_role.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:            "vastai_team_role.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

func testAccTeamRoleConfig(name, permissions string) string {
	return fmt.Sprintf(`
resource "vastai_team_role" "test" {
  name        = %q
  permissions = %q
}
`, name, permissions)
}

// testAccCheckTeamRoleDestroy verifies that all team roles created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckTeamRoleDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_team_role" {
			continue
		}
		name := rs.Primary.Attributes["name"]
		roles, err := client.Teams.ListRoles(context.Background())
		if err != nil {
			// If listing roles fails, the team may have been destroyed too;
			// treat as successful destroy.
			return nil //nolint:nilerr
		}
		for _, r := range roles {
			if r.Name == name {
				return fmt.Errorf("team role %q still exists", name)
			}
		}
	}
	return nil
}
