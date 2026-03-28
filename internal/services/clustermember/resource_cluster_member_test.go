package clustermember

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestClusterMemberResource_Metadata verifies the resource type name.
func TestClusterMemberResource_Metadata(t *testing.T) {
	r := NewClusterMemberResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_cluster_member" {
		t.Errorf("expected type name vastai_cluster_member, got %s", resp.TypeName)
	}
}

// TestClusterMemberResource_Schema verifies all expected attributes exist.
func TestClusterMemberResource_Schema(t *testing.T) {
	r := NewClusterMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	expectedAttrs := []string{
		"id", "cluster_id", "machine_id", "new_manager_id",
		"is_cluster_manager", "local_ip",
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

// TestClusterMemberResource_CompositeID verifies the id attribute description mentions composite format.
func TestClusterMemberResource_CompositeID(t *testing.T) {
	r := NewClusterMemberResource()
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

// TestClusterMemberResource_ClusterIDRequiresReplace verifies cluster_id triggers replacement.
func TestClusterMemberResource_ClusterIDRequiresReplace(t *testing.T) {
	r := NewClusterMemberResource()
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

// TestClusterMemberResource_MachineIDRequiresReplace verifies machine_id triggers replacement.
func TestClusterMemberResource_MachineIDRequiresReplace(t *testing.T) {
	r := NewClusterMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["machine_id"]
	if !ok {
		t.Fatal("machine_id attribute not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("machine_id attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("machine_id should be Required")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("machine_id should have plan modifiers (RequiresReplace)")
	}
}

// TestClusterMemberResource_NewManagerIDOptional verifies new_manager_id is optional.
func TestClusterMemberResource_NewManagerIDOptional(t *testing.T) {
	r := NewClusterMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["new_manager_id"]
	if !ok {
		t.Fatal("new_manager_id attribute not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("new_manager_id attribute is not a StringAttribute")
	}

	if !strAttr.Optional {
		t.Error("new_manager_id should be Optional")
	}
}

// TestClusterMemberResource_ComputedFields verifies is_cluster_manager and local_ip are computed.
func TestClusterMemberResource_ComputedFields(t *testing.T) {
	r := NewClusterMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	// is_cluster_manager should be Computed
	isManagerAttr, ok := resp.Schema.Attributes["is_cluster_manager"]
	if !ok {
		t.Fatal("is_cluster_manager attribute not found")
	}
	boolAttr, ok := isManagerAttr.(schema.BoolAttribute)
	if !ok {
		t.Fatal("is_cluster_manager attribute is not a BoolAttribute")
	}
	if !boolAttr.Computed {
		t.Error("is_cluster_manager should be Computed")
	}

	// local_ip should be Computed
	localIPAttr, ok := resp.Schema.Attributes["local_ip"]
	if !ok {
		t.Fatal("local_ip attribute not found")
	}
	strAttr, ok := localIPAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("local_ip attribute is not a StringAttribute")
	}
	if !strAttr.Computed {
		t.Error("local_ip should be Computed")
	}
}

// TestClusterMemberResource_DescriptionsPresent verifies all attributes have descriptions.
func TestClusterMemberResource_DescriptionsPresent(t *testing.T) {
	r := NewClusterMemberResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	for name, attr := range resp.Schema.Attributes {
		switch a := attr.(type) {
		case schema.StringAttribute:
			if a.Description == "" {
				t.Errorf("attribute %q has an empty Description", name)
			}
		case schema.BoolAttribute:
			if a.Description == "" {
				t.Errorf("attribute %q has an empty Description", name)
			}
		}
	}
}

// TestClusterMemberResource_SchemaDescription verifies the schema has a description.
func TestClusterMemberResource_SchemaDescription(t *testing.T) {
	r := NewClusterMemberResource()
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
