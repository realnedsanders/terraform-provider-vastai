package template_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccTemplate_basic verifies the full create, read, and destroy lifecycle of a template.
func TestAccTemplate_basic(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tf-acc-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccTemplateConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_template.test", "id"),
					resource.TestCheckResourceAttr("vastai_template.test", "name", name),
					resource.TestCheckResourceAttr("vastai_template.test", "image", "ubuntu:22.04"),
				),
			},
		},
	})
}

// TestAccTemplate_update verifies that a template can be updated in-place with a new name.
func TestAccTemplate_update(t *testing.T) {
	rInt := rand.Int()
	initialName := fmt.Sprintf("tf-acc-test-%d", rInt)
	updatedName := fmt.Sprintf("tf-acc-test-%d-updated", rInt)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			// Create with initial name
			{
				Config: testAccTemplateConfig_basic(initialName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_template.test", "id"),
					resource.TestCheckResourceAttr("vastai_template.test", "name", initialName),
					resource.TestCheckResourceAttr("vastai_template.test", "image", "ubuntu:22.04"),
				),
			},
			// Update name
			{
				Config: testAccTemplateConfig_basic(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_template.test", "id"),
					resource.TestCheckResourceAttr("vastai_template.test", "name", updatedName),
				),
			},
		},
	})
}

// TestAccTemplate_import verifies that a template can be imported by its hash_id.
func TestAccTemplate_import(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tf-acc-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testAccTemplateConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_template.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:            "vastai_template.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

// TestAccTemplatesDataSource_basic verifies the templates data source can search
// for templates created in the test.
func TestAccTemplatesDataSource_basic(t *testing.T) {
	rInt := rand.Int()
	name := fmt.Sprintf("tf-acc-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplatesDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_templates.search", "templates.#"),
				),
			},
		},
	})
}

func testAccTemplateConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "vastai_template" "test" {
  name  = %q
  image = "ubuntu:22.04"
}
`, name)
}

func testAccTemplatesDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "vastai_template" "test" {
  name  = %q
  image = "ubuntu:22.04"
}

data "vastai_templates" "search" {
  query = %q

  depends_on = [vastai_template.test]
}
`, name, name)
}

// testAccCheckTemplateDestroy verifies that all templates created during the test
// have been properly destroyed. For each vastai_template in the Terraform state,
// it confirms the resource no longer has an ID, indicating successful deletion.
func testAccCheckTemplateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_template" {
			continue
		}

		if rs.Primary.ID != "" {
			return fmt.Errorf("template %s still exists in state after destroy", rs.Primary.ID)
		}
	}
	return nil
}
