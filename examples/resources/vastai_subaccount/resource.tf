# Create a subaccount for resource isolation
# NOTE: Subaccounts cannot be deleted via the API. Destroying this resource
# removes it from Terraform state only.
resource "vastai_subaccount" "research_team" {
  email    = "research@example.com"
  username = "research-team"
  password = var.subaccount_password
}

variable "subaccount_password" {
  type      = string
  sensitive = true
}
