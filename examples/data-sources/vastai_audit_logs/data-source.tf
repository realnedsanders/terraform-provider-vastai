# List recent account activity from audit logs
data "vastai_audit_logs" "recent" {}

output "log_count" {
  value = length(data.vastai_audit_logs.recent.audit_logs)
}
