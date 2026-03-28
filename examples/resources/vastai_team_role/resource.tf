# Define a team role with specific permissions
resource "vastai_team_role" "developer" {
  name = "developer"
  permissions = jsonencode({
    "api" = {
      "instance_read"  = {}
      "instance_write" = {}
      "template_read"  = {}
      "ssh_key_read"   = {}
    }
  })
}
