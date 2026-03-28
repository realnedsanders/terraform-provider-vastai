package auditlog

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AuditLogsDataSourceModel describes the data source data model for vastai_audit_logs.
type AuditLogsDataSourceModel struct {
	AuditLogs []AuditLogModel `tfsdk:"audit_logs"`
}

// AuditLogModel describes a single audit log entry in the data source list (read-only).
type AuditLogModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	ApiKeyID  types.String `tfsdk:"api_key_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	ApiRoute  types.String `tfsdk:"api_route"`
	Args      types.String `tfsdk:"args"`
}
