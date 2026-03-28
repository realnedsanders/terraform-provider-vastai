# Create a team role (required dependency)
resource "vastai_team_role" "developer" {
  name        = "developer"
  permissions = ["view_instances", "create_instances"]
}

# Invite a team member with a specific role
resource "vastai_team_member" "alice" {
  email = "alice@example.com"
  role  = vastai_team_role.developer.name
}
