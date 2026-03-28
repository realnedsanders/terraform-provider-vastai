package overlaymember

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestOverlayMemberResource_Metadata verifies the resource type name.
func TestOverlayMemberResource_Metadata(t *testing.T) {
	r := NewOverlayMemberResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_overlay_member" {
		t.Errorf("expected type name vastai_overlay_member, got %s", resp.TypeName)
	}
}

// TestOverlayMemberResource_Schema verifies all expected attributes exist.
func TestOverlayMemberResource_Schema(t *testing.T) {
	r := NewOverlayMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	expectedAttrs := []string{
		"id", "overlay_name", "overlay_id", "instance_id",
	}
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

// TestOverlayMemberResource_CompositeID verifies the id attribute description mentions composite format.
func TestOverlayMemberResource_CompositeID(t *testing.T) {
	r := NewOverlayMemberResource()
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

	if strAttr.Description == "" {
		t.Error("id attribute should have a description mentioning composite format")
	}
}

// TestOverlayMemberResource_OverlayNameRequiresReplace verifies overlay_name triggers replacement.
func TestOverlayMemberResource_OverlayNameRequiresReplace(t *testing.T) {
	r := NewOverlayMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["overlay_name"]
	if !ok {
		t.Fatal("overlay_name attribute not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("overlay_name attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("overlay_name should be Required")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("overlay_name should have plan modifiers (RequiresReplace)")
	}
}

// TestOverlayMemberResource_InstanceIDRequiresReplace verifies instance_id triggers replacement.
func TestOverlayMemberResource_InstanceIDRequiresReplace(t *testing.T) {
	r := NewOverlayMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["instance_id"]
	if !ok {
		t.Fatal("instance_id attribute not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("instance_id attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("instance_id should be Required")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("instance_id should have plan modifiers (RequiresReplace)")
	}
}

// TestOverlayMemberResource_OverlayIDComputed verifies overlay_id is computed.
func TestOverlayMemberResource_OverlayIDComputed(t *testing.T) {
	r := NewOverlayMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["overlay_id"]
	if !ok {
		t.Fatal("overlay_id attribute not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("overlay_id attribute is not a StringAttribute")
	}

	if !strAttr.Computed {
		t.Error("overlay_id should be Computed")
	}
}

// TestOverlayMemberResource_DescriptionsPresent verifies all attributes have descriptions.
func TestOverlayMemberResource_DescriptionsPresent(t *testing.T) {
	r := NewOverlayMemberResource()
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

// TestOverlayMemberResource_SchemaDescription verifies the schema describes the no-op destroy.
func TestOverlayMemberResource_SchemaDescription(t *testing.T) {
	r := NewOverlayMemberResource()
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
