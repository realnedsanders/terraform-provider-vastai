package envvar_test

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

// TestAccEnvVar_basic verifies the full create, read, and destroy lifecycle of an environment variable.
func TestAccEnvVar_basic(t *testing.T) {
	rInt := rand.Int()
	key := fmt.Sprintf("TFACC_%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEnvVarDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccEnvVarConfig(key, "initial-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vastai_environment_variable.test", "id", key),
					resource.TestCheckResourceAttr("vastai_environment_variable.test", "key", key),
					resource.TestCheckResourceAttr("vastai_environment_variable.test", "value", "initial-value"),
				),
			},
		},
	})
}

// TestAccEnvVar_update verifies that an environment variable value can be updated in-place.
func TestAccEnvVar_update(t *testing.T) {
	rInt := rand.Int()
	key := fmt.Sprintf("TFACC_%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEnvVarDestroy,
		Steps: []resource.TestStep{
			// Create with initial value
			{
				Config: testAccEnvVarConfig(key, "initial-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vastai_environment_variable.test", "key", key),
					resource.TestCheckResourceAttr("vastai_environment_variable.test", "value", "initial-value"),
				),
			},
			// Update value
			{
				Config: testAccEnvVarConfig(key, "updated-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vastai_environment_variable.test", "key", key),
					resource.TestCheckResourceAttr("vastai_environment_variable.test", "value", "updated-value"),
				),
			},
		},
	})
}

// TestAccEnvVar_import verifies that an environment variable can be imported by its key name.
func TestAccEnvVar_import(t *testing.T) {
	rInt := rand.Int()
	key := fmt.Sprintf("TFACC_%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEnvVarDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testAccEnvVarConfig(key, "import-test-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vastai_environment_variable.test", "key", key),
				),
			},
			// Import the resource
			{
				ResourceName:            "vastai_environment_variable.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

func testAccEnvVarConfig(key, value string) string {
	return fmt.Sprintf(`
resource "vastai_environment_variable" "test" {
  key   = %q
  value = %q
}
`, key, value)
}

// testAccCheckEnvVarDestroy verifies that all environment variables created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckEnvVarDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_environment_variable" {
			continue
		}
		key := rs.Primary.ID
		envVars, err := client.EnvVars.List(context.Background())
		if err != nil {
			return fmt.Errorf("error checking environment variable: %s", err)
		}
		if _, found := envVars[key]; found {
			return fmt.Errorf("environment variable %q still exists", key)
		}
	}
	return nil
}
