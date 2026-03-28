# List all SSH keys for the authenticated user
data "vastai_ssh_keys" "all" {}

output "key_count" {
  value = length(data.vastai_ssh_keys.all.ssh_keys)
}
