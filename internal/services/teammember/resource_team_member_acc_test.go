package teammember_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// TestAccTeamMember_basic verifies the full create, read, and destroy lifecycle
// of a team member invitation.
// This test requires a valid email address to invite. Set VASTAI_TEST_EMAIL to run.
func TestAccTeamMember_basic(t *testing.T) {
	email := os.Getenv("VASTAI_TEST_EMAIL")
	if email == "" {
		t.Skip("Team member tests require a valid email to invite. Set VASTAI_TEST_EMAIL to run.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamMemberDestroy,
		Steps: []resource.TestStep{
			// Create team, role, and member invitation
			{
				Config: testAccTeamMemberConfig_basic(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_team_member.test", "id"),
					resource.TestCheckResourceAttr("vastai_team_member.test", "email", email),
					resource.TestCheckResourceAttr("vastai_team_member.test", "role", "tfacc-member-role"),
				),
			},
		},
	})
}

func testAccTeamMemberConfig_basic(email string) string {
	return fmt.Sprintf(`
resource "vastai_team" "test" {
  team_name = "tfacc-member-team"
}

resource "vastai_team_role" "test" {
  name        = "tfacc-member-role"
  permissions = "{}"

  depends_on = [vastai_team.test]
}

resource "vastai_team_member" "test" {
  email = %q
  role  = vastai_team_role.test.name

  depends_on = [vastai_team_role.test]
}
`, email)
}

// testAccCheckTeamMemberDestroy verifies that the team member has been removed
// from the team after the test resources are destroyed.
func testAccCheckTeamMemberDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	members, err := client.Teams.ListMembers(context.Background())
	if err != nil {
		// If listing fails (e.g., team was also destroyed), member is gone.
		return nil //nolint:nilerr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_team_member" {
			continue
		}
		email := rs.Primary.Attributes["email"]
		for _, m := range members {
			if m.Email == email {
				return fmt.Errorf("team member %q still exists with ID %d", email, m.ID)
			}
		}
	}
	return nil
}
