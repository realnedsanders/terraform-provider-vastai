package auditlog

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TestAuditLogsDataSource_Metadata verifies the data source type name.
func TestAuditLogsDataSource_Metadata(t *testing.T) {
	ds := NewAuditLogsDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_audit_logs" {
		t.Errorf("expected type name vastai_audit_logs, got %s", resp.TypeName)
	}
}

// TestAuditLogsDataSource_Schema verifies the schema has the audit_logs list attribute.
func TestAuditLogsDataSource_Schema(t *testing.T) {
	ds := NewAuditLogsDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	// Verify audit_logs attribute exists and is computed
	auditLogsAttr, ok := s.Attributes["audit_logs"]
	if !ok {
		t.Fatal("missing audit_logs attribute")
	}
	if !auditLogsAttr.IsComputed() {
		t.Error("audit_logs should be computed")
	}

	// Verify it is a ListNestedAttribute
	_, ok = auditLogsAttr.(schema.ListNestedAttribute)
	if !ok {
		t.Error("audit_logs should be ListNestedAttribute")
	}
}

// TestAuditLogsDataSource_SchemaNestedAttributes verifies the nested audit log object.
func TestAuditLogsDataSource_SchemaNestedAttributes(t *testing.T) {
	ds := NewAuditLogsDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	auditLogsAttr, ok := s.Attributes["audit_logs"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("audit_logs is not ListNestedAttribute")
	}

	nestedAttrs := auditLogsAttr.NestedObject.Attributes

	expectedNested := []string{"ip_address", "api_key_id", "created_at", "api_route", "args"}
	for _, name := range expectedNested {
		attr, ok := nestedAttrs[name]
		if !ok {
			t.Errorf("missing nested attribute: %s", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("nested attribute %s should be computed", name)
		}
	}
}

// TestAuditLogsDataSource_SchemaDescriptions verifies all attributes have descriptions.
func TestAuditLogsDataSource_SchemaDescriptions(t *testing.T) {
	ds := NewAuditLogsDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	if s.Description == "" {
		t.Error("schema has empty description")
	}

	auditLogsAttr, ok := s.Attributes["audit_logs"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("audit_logs is not ListNestedAttribute")
	}
	if auditLogsAttr.Description == "" {
		t.Error("audit_logs attribute has empty description")
	}

	for name, attr := range auditLogsAttr.NestedObject.Attributes {
		desc := getAuditLogAttributeDescription(attr)
		if desc == "" {
			t.Errorf("nested attribute %s has empty description", name)
		}
	}
}

// TestAuditLogsDataSource_ImplementsConfigure verifies configure is implemented.
func TestAuditLogsDataSource_ImplementsConfigure(t *testing.T) {
	ds := NewAuditLogsDataSource()
	_, ok := ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Error("AuditLogsDataSource should implement DataSourceWithConfigure")
	}
}

// getAuditLogAttributeDescription extracts the description from a schema attribute.
func getAuditLogAttributeDescription(attr schema.Attribute) string {
	switch a := attr.(type) {
	case schema.StringAttribute:
		return a.Description
	case schema.Int64Attribute:
		return a.Description
	case schema.Float64Attribute:
		return a.Description
	case schema.BoolAttribute:
		return a.Description
	case schema.ListNestedAttribute:
		return a.Description
	default:
		return ""
	}
}
