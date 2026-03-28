package user_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccUserDataSource_basic verifies the user data source returns the current
// authenticated user's profile with all expected fields populated.
func TestAccUserDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_user.me", "id"),
					resource.TestCheckResourceAttrSet("data.vastai_user.me", "username"),
					resource.TestCheckResourceAttrSet("data.vastai_user.me", "email"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig() string {
	return `
data "vastai_user" "me" {}
`
}
