# Search for templates matching a query
data "vastai_templates" "pytorch" {
  query = "pytorch"
}

output "template_count" {
  value = length(data.vastai_templates.pytorch.templates)
}
