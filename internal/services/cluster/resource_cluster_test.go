package cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestClusterResource_Metadata verifies the resource type name.
func TestClusterResource_Metadata(t *testing.T) {
	r := NewClusterResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_cluster" {
		t.Errorf("expected type name vastai_cluster, got %s", resp.TypeName)
	}
}

// TestClusterResource_Schema verifies all expected attributes exist.
func TestClusterResource_Schema(t *testing.T) {
	r := NewClusterResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	expectedAttrs := []string{"id", "subnet", "manager_id"}
	for _, name := range expectedAttrs {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing expected attribute: %s", name)
		}
	}

	// Verify timeouts block exists
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("expected 'timeouts' block not found in schema")
	}
}

// TestClusterResource_DescriptionsPresent verifies all attributes have descriptions.
func TestClusterResource_DescriptionsPresent(t *testing.T) {
	r := NewClusterResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	for name, attr := range resp.Schema.Attributes {
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			continue
		}
		if strAttr.Description == "" {
			t.Errorf("attribute %q has an empty Description", name)
		}
	}
}

// TestClusterResource_SubnetRequiresReplace verifies that subnet triggers replacement.
func TestClusterResource_SubnetRequiresReplace(t *testing.T) {
	r := NewClusterResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	subnetAttr, ok := resp.Schema.Attributes["subnet"]
	if !ok {
		t.Fatal("subnet attribute not found")
	}

	strAttr, ok := subnetAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("subnet attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("subnet attribute has no plan modifiers, expected RequiresReplace")
	}
}

// TestClusterResource_ManagerIDRequiresReplace verifies that manager_id triggers replacement.
func TestClusterResource_ManagerIDRequiresReplace(t *testing.T) {
	r := NewClusterResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	managerIDAttr, ok := resp.Schema.Attributes["manager_id"]
	if !ok {
		t.Fatal("manager_id attribute not found")
	}

	strAttr, ok := managerIDAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("manager_id attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("manager_id attribute has no plan modifiers, expected RequiresReplace")
	}
}

// TestClusterResource_IDComputed verifies that id is computed with UseStateForUnknown.
func TestClusterResource_IDComputed(t *testing.T) {
	r := NewClusterResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("id attribute not found")
	}

	strAttr, ok := idAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("id attribute is not a StringAttribute")
	}

	if !strAttr.Computed {
		t.Error("id attribute should be Computed")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("id attribute has no plan modifiers, expected UseStateForUnknown")
	}
}

// TestClusterResource_SchemaDescription verifies the schema has a description.
func TestClusterResource_SchemaDescription(t *testing.T) {
	r := NewClusterResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	if resp.Schema.Description == "" {
		t.Error("schema has an empty Description")
	}
}
