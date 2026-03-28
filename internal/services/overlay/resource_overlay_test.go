package overlay

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestOverlayResource_Metadata verifies the resource type name.
func TestOverlayResource_Metadata(t *testing.T) {
	r := NewOverlayResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_overlay" {
		t.Errorf("expected type name vastai_overlay, got %s", resp.TypeName)
	}
}

// TestOverlayResource_Schema verifies all expected attributes exist.
func TestOverlayResource_Schema(t *testing.T) {
	r := NewOverlayResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	expectedAttrs := []string{"id", "name", "cluster_id", "internal_subnet"}
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

// TestOverlayResource_DescriptionsPresent verifies all attributes have descriptions.
func TestOverlayResource_DescriptionsPresent(t *testing.T) {
	r := NewOverlayResource()
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

// TestOverlayResource_NameRequiresReplace verifies that name triggers replacement.
func TestOverlayResource_NameRequiresReplace(t *testing.T) {
	r := NewOverlayResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	nameAttr, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found")
	}

	strAttr, ok := nameAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("name attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("name should be Required")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("name should have plan modifiers (RequiresReplace)")
	}

	if len(strAttr.Validators) == 0 {
		t.Error("name should have validators (LengthAtLeast)")
	}
}

// TestOverlayResource_ClusterIDRequiresReplace verifies cluster_id triggers replacement.
func TestOverlayResource_ClusterIDRequiresReplace(t *testing.T) {
	r := NewOverlayResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["cluster_id"]
	if !ok {
		t.Fatal("cluster_id attribute not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("cluster_id attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("cluster_id should be Required")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("cluster_id should have plan modifiers (RequiresReplace)")
	}
}

// TestOverlayResource_InternalSubnetComputed verifies internal_subnet is computed.
func TestOverlayResource_InternalSubnetComputed(t *testing.T) {
	r := NewOverlayResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["internal_subnet"]
	if !ok {
		t.Fatal("internal_subnet attribute not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("internal_subnet attribute is not a StringAttribute")
	}

	if !strAttr.Computed {
		t.Error("internal_subnet should be Computed")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("internal_subnet should have plan modifiers (UseStateForUnknown)")
	}
}

// TestOverlayResource_IDComputed verifies id is computed with UseStateForUnknown.
func TestOverlayResource_IDComputed(t *testing.T) {
	r := NewOverlayResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("id attribute not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("id attribute is not a StringAttribute")
	}

	if !strAttr.Computed {
		t.Error("id should be Computed")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("id should have plan modifiers (UseStateForUnknown)")
	}
}

// TestOverlayResource_SchemaDescription verifies the schema has a description.
func TestOverlayResource_SchemaDescription(t *testing.T) {
	r := NewOverlayResource()
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
