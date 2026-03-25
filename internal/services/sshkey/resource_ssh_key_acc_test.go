package sshkey_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
)

// TestAccSSHKey_basic verifies the full create, read, and destroy lifecycle of an SSH key.
func TestAccSSHKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccSSHKeyConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_ssh_key.test", "id"),
					resource.TestCheckResourceAttrSet("vastai_ssh_key.test", "created_at"),
				),
			},
		},
	})
}

// TestAccSSHKey_update verifies that an SSH key can be updated in-place with a new key value.
func TestAccSSHKey_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with initial key
			{
				Config: testAccSSHKeyConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_ssh_key.test", "id"),
				),
			},
			// Update with new key value
			{
				Config: testAccSSHKeyConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_ssh_key.test", "id"),
					resource.TestCheckResourceAttrSet("vastai_ssh_key.test", "created_at"),
				),
			},
		},
	})
}

// TestAccSSHKey_import verifies that an SSH key can be imported by its numeric ID.
func TestAccSSHKey_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testAccSSHKeyConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_ssh_key.test", "id"),
				),
			},
			// Import the resource
			{
				ResourceName:            "vastai_ssh_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

// TestAccSSHKeysDataSource_basic verifies the SSH keys data source returns at least one key
// after creating a key resource.
func TestAccSSHKeysDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeysDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_ssh_keys.all", "ssh_keys.#"),
				),
			},
		},
	})
}

func testAccSSHKeyConfig_basic() string {
	return fmt.Sprintf(`
resource "vastai_ssh_key" "test" {
  ssh_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBnqKWPJdBeFdZCHmJGHfONMfOqbmmVOi9WpJAxKmLiQ acc-test-basic"
}
`)
}

func testAccSSHKeyConfig_updated() string {
	return fmt.Sprintf(`
resource "vastai_ssh_key" "test" {
  ssh_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHUgVGbn2rkTEJYFVEPaJVBGMGOIVkW6fnOfsPYVfBmI acc-test-updated"
}
`)
}

func testAccSSHKeysDataSourceConfig() string {
	return `
resource "vastai_ssh_key" "test" {
  ssh_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBnqKWPJdBeFdZCHmJGHfONMfOqbmmVOi9WpJAxKmLiQ acc-test-ds"
}

data "vastai_ssh_keys" "all" {
  depends_on = [vastai_ssh_key.test]
}
`
}
