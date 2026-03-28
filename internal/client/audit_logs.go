package client

import (
	"context"
	"fmt"
)

// AuditLogService handles audit log-related API operations.
type AuditLogService struct {
	client *VastAIClient
}

// AuditLogEntry represents a single audit log entry from the Vast.ai API.
type AuditLogEntry struct {
	IPAddress string  `json:"ip_address"`
	ApiKeyID  int     `json:"api_key_id"`
	CreatedAt float64 `json:"created_at"`
	ApiRoute  string  `json:"api_route"`
	Args      string  `json:"args"`
}

// List retrieves all audit log entries.
// Sends GET /audit_logs/.
// Per D-07: no filtering needed, return all entries.
func (s *AuditLogService) List(ctx context.Context) ([]AuditLogEntry, error) {
	var resp []AuditLogEntry
	if err := s.client.Get(ctx, "/audit_logs/", &resp); err != nil {
		return nil, fmt.Errorf("listing audit logs: %w", err)
	}
	return resp, nil
}
