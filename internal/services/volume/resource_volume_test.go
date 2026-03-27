package volume

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestVolumeResource_Metadata verifies the resource type name.
func TestVolumeResource_Metadata(t *testing.T) {
	r := NewVolumeResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_volume" {
		t.Errorf("expected type name vastai_volume, got %s", resp.TypeName)
	}
}

// TestVolumeResource_Schema_HasExpectedAttributes verifies all expected attribute names exist.
func TestVolumeResource_Schema_HasExpectedAttributes(t *testing.T) {
	r := NewVolumeResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	expectedAttrs := []string{
		"id", "offer_id", "size", "name", "clone_from_id", "disable_compression",
		"status", "disk_space", "machine_id", "geolocation",
		"inet_up", "inet_down", "reliability", "disk_name",
		"driver_version", "host_id", "verification",
	}

	for _, name := range expectedAttrs {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing expected attribute: %s", name)
		}
	}
}

// TestVolumeResource_Schema_OfferID_Required verifies offer_id is Required with AtLeast(1) validator.
func TestVolumeResource_Schema_OfferID_Required(t *testing.T) {
	r := NewVolumeResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	offerIDAttr, ok := s.Attributes["offer_id"]
	if !ok {
		t.Fatal("missing offer_id attribute")
	}
	if !offerIDAttr.IsRequired() {
		t.Error("offer_id should be required")
	}

	// Check it has validators
	int64Attr, ok := offerIDAttr.(schema.Int64Attribute)
	if !ok {
		t.Fatal("offer_id is not Int64Attribute")
	}
	if len(int64Attr.Validators) == 0 {
		t.Error("offer_id should have validators (AtLeast(1))")
	}

	// Check it has RequiresReplace plan modifier
	if len(int64Attr.PlanModifiers) == 0 {
		t.Error("offer_id should have RequiresReplace plan modifier")
	}
}

// TestVolumeResource_Schema_Size_Required verifies size is Required, RequiresReplace, AtLeast(1).
func TestVolumeResource_Schema_Size_Required(t *testing.T) {
	r := NewVolumeResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	sizeAttr, ok := s.Attributes["size"]
	if !ok {
		t.Fatal("missing size attribute")
	}
	if !sizeAttr.IsRequired() {
		t.Error("size should be required")
	}

	int64Attr, ok := sizeAttr.(schema.Int64Attribute)
	if !ok {
		t.Fatal("size is not Int64Attribute")
	}
	if len(int64Attr.Validators) == 0 {
		t.Error("size should have validators (AtLeast(1))")
	}
	if len(int64Attr.PlanModifiers) == 0 {
		t.Error("size should have RequiresReplace plan modifier")
	}
}

// TestVolumeResource_Schema_CloneFromID_Optional verifies clone_from_id is Optional, RequiresReplace.
func TestVolumeResource_Schema_CloneFromID_Optional(t *testing.T) {
	r := NewVolumeResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	cloneAttr, ok := s.Attributes["clone_from_id"]
	if !ok {
		t.Fatal("missing clone_from_id attribute")
	}
	if !cloneAttr.IsOptional() {
		t.Error("clone_from_id should be optional")
	}

	int64Attr, ok := cloneAttr.(schema.Int64Attribute)
	if !ok {
		t.Fatal("clone_from_id is not Int64Attribute")
	}
	if len(int64Attr.PlanModifiers) == 0 {
		t.Error("clone_from_id should have RequiresReplace plan modifier")
	}
	if len(int64Attr.Validators) == 0 {
		t.Error("clone_from_id should have validators (AtLeast(1))")
	}
}

// TestVolumeResource_Schema_ComputedFields verifies status, disk_space, machine_id etc are Computed.
func TestVolumeResource_Schema_ComputedFields(t *testing.T) {
	r := NewVolumeResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	computedAttrs := []string{
		"id", "status", "disk_space", "machine_id", "geolocation",
		"inet_up", "inet_down", "reliability", "disk_name",
		"driver_version", "host_id", "verification",
	}

	for _, name := range computedAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing computed attribute: %s", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %s should be computed", name)
		}
	}
}

// TestVolumeResource_Schema_PlanModifiers verifies UseStateForUnknown on stable computed fields.
func TestVolumeResource_Schema_PlanModifiers(t *testing.T) {
	r := NewVolumeResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// String attributes with UseStateForUnknown
	stringAttrsWithModifier := []string{"id", "geolocation", "disk_name"}
	for _, name := range stringAttrsWithModifier {
		attr, ok := s.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Errorf("%s is not StringAttribute", name)
			continue
		}
		if len(attr.PlanModifiers) == 0 {
			t.Errorf("attribute %s should have UseStateForUnknown plan modifier", name)
		}
	}

	// Int64 attributes with UseStateForUnknown
	int64AttrsWithModifier := []string{"machine_id", "host_id"}
	for _, name := range int64AttrsWithModifier {
		attr, ok := s.Attributes[name].(schema.Int64Attribute)
		if !ok {
			t.Errorf("%s is not Int64Attribute", name)
			continue
		}
		if len(attr.PlanModifiers) == 0 {
			t.Errorf("attribute %s should have UseStateForUnknown plan modifier", name)
		}
	}
}

// TestVolumeResource_Schema_HasTimeouts verifies the timeouts block exists.
func TestVolumeResource_Schema_HasTimeouts(t *testing.T) {
	r := NewVolumeResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
}

// TestVolumeResource_Schema_Descriptions verifies all attributes have non-empty descriptions.
func TestVolumeResource_Schema_Descriptions(t *testing.T) {
	r := NewVolumeResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	for name, attr := range s.Attributes {
		desc := getResourceAttributeDescription(attr)
		if desc == "" {
			t.Errorf("attribute %s has empty description", name)
		}
	}
}

// TestVolumeResource_ImplementsImportState verifies that import state is implemented.
func TestVolumeResource_ImplementsImportState(t *testing.T) {
	r := NewVolumeResource()
	_, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Error("VolumeResource should implement ResourceWithImportState")
	}
}

// TestVolumeResource_ImplementsConfigure verifies that configure is implemented.
func TestVolumeResource_ImplementsConfigure(t *testing.T) {
	r := NewVolumeResource()
	_, ok := r.(resource.ResourceWithConfigure)
	if !ok {
		t.Error("VolumeResource should implement ResourceWithConfigure")
	}
}

// getResourceAttributeDescription extracts the description from a resource schema attribute.
func getResourceAttributeDescription(attr schema.Attribute) string {
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
	case schema.SingleNestedAttribute:
		return a.Description
	default:
		return ""
	}
}
