# List all instances for the authenticated user
data "vastai_instances" "all" {}

output "instance_count" {
  value = length(data.vastai_instances.all.instances)
}
