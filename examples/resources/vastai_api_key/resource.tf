# Create a scoped API key with limited permissions
resource "vastai_api_key" "ci_deploy" {
  name        = "ci-deploy-key"
  permissions = jsonencode({
    "instances" : "full",
    "ssh_keys" : "read",
    "templates" : "read"
  })
}

output "api_key_id" {
  value = vastai_api_key.ci_deploy.id
}
