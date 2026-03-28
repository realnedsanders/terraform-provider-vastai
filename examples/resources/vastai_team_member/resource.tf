# Invite a team member with a specific role
resource "vastai_team_member" "alice" {
  email = "alice@example.com"
  role  = vastai_team_role.developer.name
}
