package invoice

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TestInvoicesDataSource_Metadata verifies the data source type name.
func TestInvoicesDataSource_Metadata(t *testing.T) {
	ds := NewInvoicesDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_invoices" {
		t.Errorf("expected type name vastai_invoices, got %s", resp.TypeName)
	}
}

// TestInvoicesDataSource_Schema verifies the schema has all expected attributes.
func TestInvoicesDataSource_Schema(t *testing.T) {
	ds := NewInvoicesDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	// Check optional filter attributes
	optionalAttrs := []string{"start_date", "end_date", "limit", "type"}
	for _, name := range optionalAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing attribute: %s", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("attribute %s should be optional", name)
		}
	}

	// Check computed attributes
	computedAttrs := []string{"invoices", "total"}
	for _, name := range computedAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing attribute: %s", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %s should be computed", name)
		}
	}
}

// TestInvoicesDataSource_SchemaInvoicesNested verifies the invoices list has nested attributes.
func TestInvoicesDataSource_SchemaInvoicesNested(t *testing.T) {
	ds := NewInvoicesDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	invoicesAttr, ok := s.Attributes["invoices"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("invoices is not ListNestedAttribute")
	}

	nestedAttrs := invoicesAttr.NestedObject.Attributes

	expectedNested := []string{"id", "amount", "type", "description", "timestamp"}
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

// TestInvoicesDataSource_SchemaDescriptions verifies all attributes have non-empty descriptions.
func TestInvoicesDataSource_SchemaDescriptions(t *testing.T) {
	ds := NewInvoicesDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	if s.Description == "" {
		t.Error("schema has empty description")
	}

	for name, attr := range s.Attributes {
		desc := getInvoiceAttributeDescription(attr)
		if desc == "" {
			t.Errorf("attribute %s has empty description", name)
		}
	}
}

// TestInvoicesDataSource_ImplementsConfigure verifies configure is implemented.
func TestInvoicesDataSource_ImplementsConfigure(t *testing.T) {
	ds := NewInvoicesDataSource()
	_, ok := ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Error("InvoicesDataSource should implement DataSourceWithConfigure")
	}
}

// getInvoiceAttributeDescription extracts the description from a schema attribute.
func getInvoiceAttributeDescription(attr schema.Attribute) string {
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
