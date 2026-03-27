package endpoint

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestEndpointResource_Metadata verifies the resource type name.
func TestEndpointResource_Metadata(t *testing.T) {
	r := NewEndpointResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_endpoint" {
		t.Errorf("expected type name vastai_endpoint, got %s", resp.TypeName)
	}
}

// TestEndpointResource_Schema verifies all expected attributes exist.
func TestEndpointResource_Schema(t *testing.T) {
	r := NewEndpointResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	expectedAttrs := []string{
		"id", "endpoint_name", "min_load", "min_cold_load",
		"target_util", "cold_mult", "cold_workers", "max_workers",
		"endpoint_state",
	}

	for _, name := range expectedAttrs {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing expected attribute: %s", name)
		}
	}
}

// TestEndpointResource_SchemaRequired verifies endpoint_name is Required.
func TestEndpointResource_SchemaRequired(t *testing.T) {
	r := NewEndpointResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	nameAttr, ok := s.Attributes["endpoint_name"]
	if !ok {
		t.Fatal("missing endpoint_name attribute")
	}
	if !nameAttr.IsRequired() {
		t.Error("endpoint_name should be required")
	}
}

// TestEndpointResource_SchemaComputed verifies id is Computed with UseStateForUnknown;
// autoscaling attrs are Optional+Computed.
func TestEndpointResource_SchemaComputed(t *testing.T) {
	r := NewEndpointResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// ID should be Computed with UseStateForUnknown plan modifier
	idAttr, ok := s.Attributes["id"]
	if !ok {
		t.Fatal("missing id attribute")
	}
	if !idAttr.IsComputed() {
		t.Error("id should be computed")
	}
	strAttr, ok := idAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("id is not StringAttribute")
	}
	if len(strAttr.PlanModifiers) == 0 {
		t.Error("id should have UseStateForUnknown plan modifier")
	}

	// Autoscaling attributes should be Optional+Computed
	optionalComputedAttrs := []string{
		"min_load", "min_cold_load", "target_util", "cold_mult",
		"cold_workers", "max_workers", "endpoint_state",
	}

	for _, name := range optionalComputedAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing attribute: %s", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("attribute %s should be optional", name)
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %s should be computed", name)
		}
	}
}

// TestEndpointResource_SchemaValidators verifies validators on constrained attributes.
func TestEndpointResource_SchemaValidators(t *testing.T) {
	r := NewEndpointResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// target_util should have Between(0, 1) validator
	targetUtilAttr, ok := s.Attributes["target_util"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("target_util is not Float64Attribute")
	}
	if len(targetUtilAttr.Validators) == 0 {
		t.Error("target_util should have validators (Between(0, 1))")
	}

	// cold_mult should have AtLeast(1.0) validator
	coldMultAttr, ok := s.Attributes["cold_mult"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("cold_mult is not Float64Attribute")
	}
	if len(coldMultAttr.Validators) == 0 {
		t.Error("cold_mult should have validators (AtLeast(1.0))")
	}

	// min_load should have AtLeast(0) validator
	minLoadAttr, ok := s.Attributes["min_load"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("min_load is not Float64Attribute")
	}
	if len(minLoadAttr.Validators) == 0 {
		t.Error("min_load should have validators (AtLeast(0))")
	}

	// min_cold_load should have AtLeast(0) validator
	minColdLoadAttr, ok := s.Attributes["min_cold_load"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("min_cold_load is not Float64Attribute")
	}
	if len(minColdLoadAttr.Validators) == 0 {
		t.Error("min_cold_load should have validators (AtLeast(0))")
	}

	// cold_workers should have AtLeast(0) validator
	coldWorkersAttr, ok := s.Attributes["cold_workers"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("cold_workers is not Int64Attribute")
	}
	if len(coldWorkersAttr.Validators) == 0 {
		t.Error("cold_workers should have validators (AtLeast(0))")
	}

	// max_workers should have AtLeast(0) validator
	maxWorkersAttr, ok := s.Attributes["max_workers"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("max_workers is not Int64Attribute")
	}
	if len(maxWorkersAttr.Validators) == 0 {
		t.Error("max_workers should have validators (AtLeast(0))")
	}

	// endpoint_state should have OneOf validator
	endpointStateAttr, ok := s.Attributes["endpoint_state"].(schema.StringAttribute)
	if !ok {
		t.Fatal("endpoint_state is not StringAttribute")
	}
	if len(endpointStateAttr.Validators) == 0 {
		t.Error("endpoint_state should have validators (OneOf)")
	}
}

// TestEndpointResource_Schema_HasTimeouts verifies the timeouts block exists.
func TestEndpointResource_Schema_HasTimeouts(t *testing.T) {
	r := NewEndpointResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
}

// TestEndpointResource_Schema_Descriptions verifies all attributes have non-empty descriptions.
func TestEndpointResource_Schema_Descriptions(t *testing.T) {
	r := NewEndpointResource()
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

// TestEndpointResource_ImplementsImportState verifies that import state is implemented.
func TestEndpointResource_ImplementsImportState(t *testing.T) {
	r := NewEndpointResource()
	_, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Error("EndpointResource should implement ResourceWithImportState")
	}
}

// TestEndpointResource_ImplementsConfigure verifies that configure is implemented.
func TestEndpointResource_ImplementsConfigure(t *testing.T) {
	r := NewEndpointResource()
	_, ok := r.(resource.ResourceWithConfigure)
	if !ok {
		t.Error("EndpointResource should implement ResourceWithConfigure")
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
