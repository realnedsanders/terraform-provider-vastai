# Team Management Workflow
#
# This example demonstrates managing Vast.ai teams:
# 1. Create a team
# 2. Define roles with permissions
# 3. Invite team members

terraform {
  required_providers {
    vastai = {
      source = "realnedsanders/vastai"
    }
  }
}

provider "vastai" {}

# Step 1: Create a team
resource "vastai_team" "engineering" {
  name = "engineering-team"
}

# Step 2: Define roles with permissions
resource "vastai_team_role" "developer" {
  name        = "developer"
  permissions = jsonencode({
    api = {
      instance_read  = {}
      instance_write = {}
      template_read  = {}
      template_write = {}
    }
  })
}

resource "vastai_team_role" "viewer" {
  name        = "viewer"
  permissions = jsonencode({
    api = {
      instance_read = {}
      template_read = {}
    }
  })
}

# Step 3: Invite team members
resource "vastai_team_member" "alice" {
  email = "alice@example.com"
  role  = vastai_team_role.developer.name
}

resource "vastai_team_member" "bob" {
  email = "bob@example.com"
  role  = vastai_team_role.viewer.name
}
