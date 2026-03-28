package sshkey_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/realnedsanders/terraform-provider-vastai/internal/acctest"
	"github.com/realnedsanders/terraform-provider-vastai/internal/sweep"
)

// generateTestSSHKey generates a unique ed25519 SSH public key for testing.
func generateTestSSHKey(t *testing.T, comment string) string {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate SSH key: %s", err)
	}
	// Build the SSH wire format: string "ssh-ed25519" + string <32-byte key>
	keyType := "ssh-ed25519"
	wireKey := marshalSSHString([]byte(keyType))
	wireKey = append(wireKey, marshalSSHString([]byte(pub))...)
	return fmt.Sprintf("ssh-ed25519 %s %s", base64.StdEncoding.EncodeToString(wireKey), comment)
}

// marshalSSHString encodes a byte slice in SSH wire format (4-byte big-endian length + data).
func marshalSSHString(data []byte) []byte {
	result := make([]byte, 4+len(data))
	result[0] = byte(len(data) >> 24)
	result[1] = byte(len(data) >> 16)
	result[2] = byte(len(data) >> 8)
	result[3] = byte(len(data))
	copy(result[4:], data)
	return result
}

// TestAccSSHKey_basic verifies the full create, read, and destroy lifecycle of an SSH key.
func TestAccSSHKey_basic(t *testing.T) {
	sshKey := generateTestSSHKey(t, "tfacc-basic")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccSSHKeyConfig(sshKey),
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
	sshKey1 := generateTestSSHKey(t, "tfacc-basic")
	sshKey2 := generateTestSSHKey(t, "tfacc-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			// Create with initial key
			{
				Config: testAccSSHKeyConfig(sshKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vastai_ssh_key.test", "id"),
				),
			},
			// Update with new key value
			{
				Config: testAccSSHKeyConfig(sshKey2),
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
	sshKey := generateTestSSHKey(t, "tfacc-import")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: testAccSSHKeyConfig(sshKey),
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
	sshKey := generateTestSSHKey(t, "tfacc-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeysDataSourceConfig(sshKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vastai_ssh_keys.all", "ssh_keys.#"),
				),
			},
		},
	})
}

func testAccSSHKeyConfig(sshKey string) string {
	return fmt.Sprintf(`
resource "vastai_ssh_key" "test" {
  ssh_key = %q
}
`, sshKey)
}

func testAccSSHKeysDataSourceConfig(sshKey string) string {
	return fmt.Sprintf(`
resource "vastai_ssh_key" "test" {
  ssh_key = %q
}

data "vastai_ssh_keys" "all" {
  depends_on = [vastai_ssh_key.test]
}
`, sshKey)
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
