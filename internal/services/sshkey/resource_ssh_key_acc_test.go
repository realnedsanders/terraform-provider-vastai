package sshkey_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// TestAccSSHKey_basic verifies the full create, read, and destroy lifecycle of an SSH key.
func TestAccSSHKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy,
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
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy,
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
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy,
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
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy,
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
	return `
resource "vastai_ssh_key" "test" {
  ssh_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBnqKWPJdBeFdZCHmJGHfONMfOqbmmVOi9WpJAxKmLiQ tfacc-basic"
}
`
}

func testAccSSHKeyConfig_updated() string {
	return `
resource "vastai_ssh_key" "test" {
  ssh_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHUgVGbn2rkTEJYFVEPaJVBGMGOIVkW6fnOfsPYVfBmI tfacc-updated"
}
`
}

func testAccSSHKeysDataSourceConfig() string {
	return `
resource "vastai_ssh_key" "test" {
  ssh_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBnqKWPJdBeFdZCHmJGHfONMfOqbmmVOi9WpJAxKmLiQ tfacc-ds"
}

data "vastai_ssh_keys" "all" {
  depends_on = [vastai_ssh_key.test]
}
`
}

// testAccCheckSSHKeyDestroy verifies that all SSH keys created during the test
// have been properly destroyed by querying the Vast.ai API.
func testAccCheckSSHKeyDestroy(s *terraform.State) error {
	client, err := sweep.SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vastai_ssh_key" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing SSH key ID %q: %s", rs.Primary.ID, err)
		}
		keys, err := client.SSHKeys.List(context.Background())
		if err != nil {
			return fmt.Errorf("error checking SSH key: %s", err)
		}
		for _, k := range keys {
			if k.ID == id {
				return fmt.Errorf("SSH key %d still exists", id)
			}
		}
	}
	return nil
}
