package workergroup

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestWorkerGroupResource_Metadata verifies the resource type name.
func TestWorkerGroupResource_Metadata(t *testing.T) {
	r := NewWorkerGroupResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_worker_group" {
		t.Errorf("expected type name vastai_worker_group, got %s", resp.TypeName)
	}
}

// TestWorkerGroupResource_Schema verifies all expected attributes exist.
func TestWorkerGroupResource_Schema(t *testing.T) {
	r := NewWorkerGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	expectedAttrs := []string{
		"id", "endpoint_id", "endpoint_name", "template_hash",
		"template_id", "search_params", "launch_args", "gpu_ram",
		"test_workers", "cold_workers",
	}

	for _, name := range expectedAttrs {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing expected attribute: %s", name)
		}
	}
}

// TestWorkerGroupResource_SchemaRequired verifies endpoint_id is Required.
func TestWorkerGroupResource_SchemaRequired(t *testing.T) {
	r := NewWorkerGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	endpointIDAttr, ok := s.Attributes["endpoint_id"]
	if !ok {
		t.Fatal("missing endpoint_id attribute")
	}
	if !endpointIDAttr.IsRequired() {
		t.Error("endpoint_id should be required")
	}
}

// TestWorkerGroupResource_SchemaComputed verifies id is Computed with UseStateForUnknown;
// endpoint_name is Optional+Computed; test_workers is Optional+Computed.
func TestWorkerGroupResource_SchemaComputed(t *testing.T) {
	r := NewWorkerGroupResource()
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

	// endpoint_name should be Optional+Computed
	endpointNameAttr, ok := s.Attributes["endpoint_name"]
	if !ok {
		t.Fatal("missing endpoint_name attribute")
	}
	if !endpointNameAttr.IsOptional() {
		t.Error("endpoint_name should be optional")
	}
	if !endpointNameAttr.IsComputed() {
		t.Error("endpoint_name should be computed")
	}

	// test_workers should be Optional+Computed
	testWorkersAttr, ok := s.Attributes["test_workers"]
	if !ok {
		t.Fatal("missing test_workers attribute")
	}
	if !testWorkersAttr.IsOptional() {
		t.Error("test_workers should be optional")
	}
	if !testWorkersAttr.IsComputed() {
		t.Error("test_workers should be computed")
	}
}

// TestWorkerGroupResource_SchemaForceNew verifies endpoint_id has RequiresReplace plan modifier.
func TestWorkerGroupResource_SchemaForceNew(t *testing.T) {
	r := NewWorkerGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	endpointIDAttr, ok := s.Attributes["endpoint_id"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("endpoint_id is not Int64Attribute")
	}
	if len(endpointIDAttr.PlanModifiers) == 0 {
		t.Error("endpoint_id should have RequiresReplace plan modifier")
	}
}

// TestWorkerGroupResource_SchemaValidators verifies validators on constrained attributes.
func TestWorkerGroupResource_SchemaValidators(t *testing.T) {
	r := NewWorkerGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// gpu_ram should have AtLeast(0) validator
	gpuRAMAttr, ok := s.Attributes["gpu_ram"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("gpu_ram is not Float64Attribute")
	}
	if len(gpuRAMAttr.Validators) == 0 {
		t.Error("gpu_ram should have validators (AtLeast(0))")
	}

	// test_workers should have AtLeast(0) validator
	testWorkersAttr, ok := s.Attributes["test_workers"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("test_workers is not Int64Attribute")
	}
	if len(testWorkersAttr.Validators) == 0 {
		t.Error("test_workers should have validators (AtLeast(0))")
	}

	// cold_workers should have AtLeast(0) validator
	coldWorkersAttr, ok := s.Attributes["cold_workers"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("cold_workers is not Int64Attribute")
	}
	if len(coldWorkersAttr.Validators) == 0 {
		t.Error("cold_workers should have validators (AtLeast(0))")
	}

	// template_hash should have AtLeastOneOf validator
	templateHashAttr, ok := s.Attributes["template_hash"].(schema.StringAttribute)
	if !ok {
		t.Fatal("template_hash is not StringAttribute")
	}
	if len(templateHashAttr.Validators) == 0 {
		t.Error("template_hash should have AtLeastOneOf validator")
	}

	// template_id should have AtLeastOneOf validator
	templateIDAttr, ok := s.Attributes["template_id"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("template_id is not Int64Attribute")
	}
	if len(templateIDAttr.Validators) == 0 {
		t.Error("template_id should have AtLeastOneOf validator")
	}
}

// TestWorkerGroupResource_NoAutoscalingParams verifies that autoscaling params
// (min_load, target_util, cold_mult) are NOT in the schema.
// Per Pitfall 3 from research: these are not used at the worker group level.
// Autoscaling is driven by the parent endpoint.
func TestWorkerGroupResource_NoAutoscalingParams(t *testing.T) {
	r := NewWorkerGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	excludedAttrs := []string{"min_load", "target_util", "cold_mult"}
	for _, name := range excludedAttrs {
		if _, ok := s.Attributes[name]; ok {
			t.Errorf("attribute %s should NOT be in worker group schema (autoscaling is endpoint-level per Pitfall 3)", name)
		}
	}
}

// TestWorkerGroupResource_Schema_HasTimeouts verifies the timeouts block exists.
func TestWorkerGroupResource_Schema_HasTimeouts(t *testing.T) {
	r := NewWorkerGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
}

// TestWorkerGroupResource_Schema_Descriptions verifies all attributes have non-empty descriptions.
func TestWorkerGroupResource_Schema_Descriptions(t *testing.T) {
	r := NewWorkerGroupResource()
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

// TestWorkerGroupResource_ImplementsImportState verifies that import state is implemented.
func TestWorkerGroupResource_ImplementsImportState(t *testing.T) {
	r := NewWorkerGroupResource()
	_, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Error("WorkerGroupResource should implement ResourceWithImportState")
	}
}

// TestWorkerGroupResource_ImplementsConfigure verifies that configure is implemented.
func TestWorkerGroupResource_ImplementsConfigure(t *testing.T) {
	r := NewWorkerGroupResource()
	_, ok := r.(resource.ResourceWithConfigure)
	if !ok {
		t.Error("WorkerGroupResource should implement ResourceWithConfigure")
	}
}

// TestWorkerGroupResource_SchemaDescription verifies the resource-level description
// documents that autoscaling is on the endpoint.
func TestWorkerGroupResource_SchemaDescription(t *testing.T) {
	r := NewWorkerGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	if s.Description == "" {
		t.Error("schema should have a non-empty description")
	}

	// Should mention that autoscaling is at the endpoint level
	expected := "Autoscaling behavior is controlled at the endpoint level"
	if s.Description == "" {
		t.Error("schema description is empty")
	}
	found := false
	if len(s.Description) > 0 {
		for i := 0; i <= len(s.Description)-len(expected); i++ {
			if s.Description[i:i+len(expected)] == expected {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("schema description should contain %q", expected)
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
