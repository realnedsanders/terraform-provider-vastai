# Upload an SSH public key for instance access
resource "vastai_ssh_key" "deploy_key" {
  ssh_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBxYEL5G7fH5KdDpqcMdEhwowkPsGQPha3sQvJbp3kXD deploy@ci"
}
