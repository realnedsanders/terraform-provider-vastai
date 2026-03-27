package endpoint

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TestEndpointsDataSource_Metadata verifies the data source type name.
func TestEndpointsDataSource_Metadata(t *testing.T) {
	ds := NewEndpointsDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_endpoints" {
		t.Errorf("expected type name vastai_endpoints, got %s", resp.TypeName)
	}
}

// TestEndpointsDataSource_Schema verifies the schema has an endpoints list attribute.
func TestEndpointsDataSource_Schema(t *testing.T) {
	ds := NewEndpointsDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	// Verify endpoints attribute exists
	endpointsAttr, ok := s.Attributes["endpoints"]
	if !ok {
		t.Fatal("missing endpoints attribute")
	}

	// Verify it is a ListNestedAttribute
	_, ok = endpointsAttr.(schema.ListNestedAttribute)
	if !ok {
		t.Error("endpoints should be ListNestedAttribute")
	}
}

// TestEndpointsDataSource_SchemaComputed verifies the endpoints attribute is Computed.
func TestEndpointsDataSource_SchemaComputed(t *testing.T) {
	ds := NewEndpointsDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	endpointsAttr, ok := s.Attributes["endpoints"]
	if !ok {
		t.Fatal("missing endpoints attribute")
	}
	if !endpointsAttr.IsComputed() {
		t.Error("endpoints should be computed")
	}
}

// TestEndpointsDataSource_SchemaNestedAttributes verifies the nested endpoint object
// has all expected fields.
func TestEndpointsDataSource_SchemaNestedAttributes(t *testing.T) {
	ds := NewEndpointsDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	endpointsAttr, ok := s.Attributes["endpoints"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("endpoints is not ListNestedAttribute")
	}

	nestedAttrs := endpointsAttr.NestedObject.Attributes

	expectedNested := []string{
		"id", "endpoint_name", "min_load", "min_cold_load",
		"target_util", "cold_mult", "cold_workers", "max_workers",
		"endpoint_state",
	}

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

// TestEndpointsDataSource_SchemaDescriptions verifies all attributes have non-empty descriptions.
func TestEndpointsDataSource_SchemaDescriptions(t *testing.T) {
	ds := NewEndpointsDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// Check top-level description
	endpointsAttr, ok := s.Attributes["endpoints"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("endpoints is not ListNestedAttribute")
	}
	if endpointsAttr.Description == "" {
		t.Error("endpoints attribute has empty description")
	}

	// Check nested descriptions
	for name, attr := range endpointsAttr.NestedObject.Attributes {
		desc := getDataSourceAttributeDescription(attr)
		if desc == "" {
			t.Errorf("nested attribute %s has empty description", name)
		}
	}
}

// TestEndpointsDataSource_ImplementsConfigure verifies configure is implemented.
func TestEndpointsDataSource_ImplementsConfigure(t *testing.T) {
	ds := NewEndpointsDataSource()
	_, ok := ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Error("EndpointsDataSource should implement DataSourceWithConfigure")
	}
}

// getDataSourceAttributeDescription extracts the description from a data source schema attribute.
func getDataSourceAttributeDescription(attr schema.Attribute) string {
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
