package instance

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// getInstanceDataSourceSchema is a test helper that retrieves the instance data source schema.
func getInstanceDataSourceSchema(t *testing.T) datasourceschema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	d := NewInstanceDataSource()
	ds, ok := d.(*InstanceDataSource)
	if !ok {
		t.Fatal("unexpected data source type")
	}
	ds.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

// getInstancesDataSourceSchema is a test helper that retrieves the instances data source schema.
func getInstancesDataSourceSchema(t *testing.T) datasourceschema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	d := NewInstancesDataSource()
	ds, ok := d.(*InstancesDataSource)
	if !ok {
		t.Fatal("unexpected data source type")
	}
	ds.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

// TestInstanceDataSourceSchema verifies the singular instance data source schema:
// - id is Required (user supplies the ID to look up)
// - all other attributes are Computed (read-only from API).
func TestInstanceDataSourceSchema(t *testing.T) {
	s := getInstanceDataSourceSchema(t)

	// Verify id is Required
	idAttr, ok := s.Attributes["id"]
	if !ok {
		t.Fatal("expected 'id' attribute to exist")
	}
	strAttr, ok := idAttr.(datasourceschema.StringAttribute)
	if !ok {
		t.Fatal("expected 'id' to be StringAttribute")
	}
	if !strAttr.Required {
		t.Error("expected 'id' to be Required")
	}

	// Verify all computed attributes exist and are Computed
	computedAttrs := []string{
		"machine_id", "gpu_name", "num_gpus", "gpu_ram_gb", "cpu_cores",
		"cpu_ram_gb", "disk_space_gb", "actual_status", "intended_status",
		"ssh_host", "ssh_port", "cost_per_hour", "label", "image",
		"geolocation", "is_bid", "reliability", "inet_up_mbps", "inet_down_mbps",
		"status_msg", "template_hash_id", "onstart",
	}

	for _, name := range computedAttrs {
		attr, exists := s.Attributes[name]
		if !exists {
			t.Errorf("expected attribute %q to exist", name)
			continue
		}

		// Check the attribute is Computed by type-asserting each possible type
		switch a := attr.(type) {
		case datasourceschema.StringAttribute:
			if !a.Computed {
				t.Errorf("expected attribute %q to be Computed", name)
			}
		case datasourceschema.Int64Attribute:
			if !a.Computed {
				t.Errorf("expected attribute %q to be Computed", name)
			}
		case datasourceschema.Float64Attribute:
			if !a.Computed {
				t.Errorf("expected attribute %q to be Computed", name)
			}
		case datasourceschema.BoolAttribute:
			if !a.Computed {
				t.Errorf("expected attribute %q to be Computed", name)
			}
		default:
			t.Errorf("unexpected attribute type for %q: %T", name, attr)
		}
	}
}

// TestInstancesDataSourceSchema verifies the plural instances data source schema:
// - instances is a Computed ListNestedAttribute
// - label is an Optional StringAttribute for filtering.
func TestInstancesDataSourceSchema(t *testing.T) {
	s := getInstancesDataSourceSchema(t)

	// Verify label attribute is Optional
	labelAttr, ok := s.Attributes["label"]
	if !ok {
		t.Fatal("expected 'label' attribute to exist")
	}
	labelStrAttr, ok := labelAttr.(datasourceschema.StringAttribute)
	if !ok {
		t.Fatal("expected 'label' to be StringAttribute")
	}
	if !labelStrAttr.Optional {
		t.Error("expected 'label' to be Optional")
	}

	// Verify instances attribute is Computed ListNestedAttribute
	instancesAttr, ok := s.Attributes["instances"]
	if !ok {
		t.Fatal("expected 'instances' attribute to exist")
	}
	listNestedAttr, ok := instancesAttr.(datasourceschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("expected 'instances' to be ListNestedAttribute, got %T", instancesAttr)
	}
	if !listNestedAttr.Computed {
		t.Error("expected 'instances' to be Computed")
	}

	// Verify nested attributes exist
	nestedAttrs := listNestedAttr.NestedObject.Attributes
	expectedNested := []string{
		"id", "machine_id", "gpu_name", "num_gpus", "gpu_ram_gb", "cpu_cores",
		"cpu_ram_gb", "disk_space_gb", "actual_status", "intended_status",
		"ssh_host", "ssh_port", "cost_per_hour", "label", "image",
		"geolocation", "is_bid", "reliability", "inet_up_mbps", "inet_down_mbps",
		"status_msg", "template_hash_id", "onstart",
	}

	for _, name := range expectedNested {
		if _, exists := nestedAttrs[name]; !exists {
			t.Errorf("expected nested attribute %q to exist in 'instances'", name)
		}
	}
}

// TestInstanceDataSource_DescriptionsPresent verifies that every attribute in the
// singular instance data source has a non-empty Description string per SCHM-04.
func TestInstanceDataSource_DescriptionsPresent(t *testing.T) {
	s := getInstanceDataSourceSchema(t)

	for name, attr := range s.Attributes {
		var desc string
		switch a := attr.(type) {
		case datasourceschema.StringAttribute:
			desc = a.Description
		case datasourceschema.Int64Attribute:
			desc = a.Description
		case datasourceschema.Float64Attribute:
			desc = a.Description
		case datasourceschema.BoolAttribute:
			desc = a.Description
		}
		if desc == "" {
			t.Errorf("attribute %q has empty Description", name)
		}
	}
}

// TestInstancesDataSource_DescriptionsPresent verifies that every attribute in the
// plural instances data source has a non-empty Description string per SCHM-04.
func TestInstancesDataSource_DescriptionsPresent(t *testing.T) {
	s := getInstancesDataSourceSchema(t)

	// Check top-level attributes
	for name, attr := range s.Attributes {
		switch a := attr.(type) {
		case datasourceschema.StringAttribute:
			if a.Description == "" {
				t.Errorf("attribute %q has empty Description", name)
			}
		case datasourceschema.ListNestedAttribute:
			if a.Description == "" {
				t.Errorf("attribute %q has empty Description", name)
			}
			// Check nested attributes
			for nestedName, nestedAttr := range a.NestedObject.Attributes {
				var desc string
				switch na := nestedAttr.(type) {
				case datasourceschema.StringAttribute:
					desc = na.Description
				case datasourceschema.Int64Attribute:
					desc = na.Description
				case datasourceschema.Float64Attribute:
					desc = na.Description
				case datasourceschema.BoolAttribute:
					desc = na.Description
				}
				if desc == "" {
					t.Errorf("nested attribute %q.%q has empty Description", name, nestedName)
				}
			}
		}
	}
}

// TestInstanceDataSourceMetadata verifies the type name is correct.
func TestInstanceDataSourceMetadata(t *testing.T) {
	d := NewInstanceDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "vastai"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "vastai_instance" {
		t.Errorf("expected type name 'vastai_instance', got %q", resp.TypeName)
	}
}

// TestInstancesDataSourceMetadata verifies the type name is correct.
func TestInstancesDataSourceMetadata(t *testing.T) {
	d := NewInstancesDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "vastai"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "vastai_instances" {
		t.Errorf("expected type name 'vastai_instances', got %q", resp.TypeName)
	}
}
