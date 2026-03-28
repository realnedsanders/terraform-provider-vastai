# List all serverless endpoints for the authenticated user
data "vastai_endpoints" "all" {}

output "endpoint_count" {
  value = length(data.vastai_endpoints.all.endpoints)
}
