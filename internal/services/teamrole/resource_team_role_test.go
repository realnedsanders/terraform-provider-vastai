package teamrole

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// getResourceSchema is a test helper that retrieves the team role resource schema.
func getResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := fwresource.SchemaRequest{}
	schemaResp := &fwresource.SchemaResponse{}

	r := NewTeamRoleResource()
	res, ok := r.(*TeamRoleResource)
	if !ok {
		t.Fatal("unexpected resource type")
	}
	res.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

func TestTeamRoleResourceSchema(t *testing.T) {
	s := getResourceSchema(t)

	// Verify all expected attributes exist
	expectedAttrs := []string{"id", "name", "permissions"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Errorf("Expected attribute %q not found in schema", attr)
		}
	}

	// Verify timeouts block exists
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("Expected 'timeouts' block not found in schema")
	}
}

func TestTeamRoleResource_NameIsRequired(t *testing.T) {
	s := getResourceSchema(t)

	nameAttr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found in schema")
	}

	strAttr, ok := nameAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("name attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("name attribute should be Required")
	}
}

func TestTeamRoleResource_PermissionsIsRequired(t *testing.T) {
	s := getResourceSchema(t)

	permsAttr, ok := s.Attributes["permissions"]
	if !ok {
		t.Fatal("permissions attribute not found in schema")
	}

	strAttr, ok := permsAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("permissions attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("permissions attribute should be Required")
	}
}

func TestTeamRoleResource_IDIsComputed(t *testing.T) {
	s := getResourceSchema(t)

	idAttr, ok := s.Attributes["id"]
	if !ok {
		t.Fatal("id attribute not found in schema")
	}

	strAttr, ok := idAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("id attribute is not a StringAttribute")
	}

	if !strAttr.Computed {
		t.Error("id attribute should be Computed")
	}
}

func TestTeamRoleResource_IDHasUseStateForUnknown(t *testing.T) {
	s := getResourceSchema(t)

	idAttr, ok := s.Attributes["id"]
	if !ok {
		t.Fatal("id attribute not found in schema")
	}

	strAttr, ok := idAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("id attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("id attribute has no plan modifiers, expected UseStateForUnknown")
	}
}

func TestTeamRoleResource_DescriptionsPresent(t *testing.T) {
	s := getResourceSchema(t)

	if s.Description == "" {
		t.Error("Resource schema has an empty Description")
	}

	for name, attr := range s.Attributes {
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Errorf("Attribute %q is not a StringAttribute", name)
			continue
		}

		if strAttr.Description == "" {
			t.Errorf("Attribute %q has an empty Description", name)
		}
	}
}

func TestTeamRoleResource_NameValidators(t *testing.T) {
	s := getResourceSchema(t)

	nameAttr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found in schema")
	}

	strAttr, ok := nameAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("name attribute is not a StringAttribute")
	}

	if len(strAttr.Validators) == 0 {
		t.Error("name attribute has no validators, expected LengthAtLeast(1)")
	}
}

func TestTeamRoleResource_PermissionsHasJSONValidator(t *testing.T) {
	s := getResourceSchema(t)

	permsAttr, ok := s.Attributes["permissions"]
	if !ok {
		t.Fatal("permissions attribute not found in schema")
	}

	strAttr, ok := permsAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("permissions attribute is not a StringAttribute")
	}

	if len(strAttr.Validators) == 0 {
		t.Error("permissions attribute has no validators, expected JSON validator")
	}
}

func TestTeamRoleResource_Metadata(t *testing.T) {
	r := NewTeamRoleResource()
	metaReq := fwresource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	metaResp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), metaReq, metaResp)

	if metaResp.TypeName != "vastai_team_role" {
		t.Errorf("Expected type name 'vastai_team_role', got %q", metaResp.TypeName)
	}
}

func TestTeamRoleResource_ImplementsResourceInterface(t *testing.T) {
	r := NewTeamRoleResource()

	// Verify it implements resource.ResourceWithImportState
	if _, ok := r.(fwresource.ResourceWithImportState); !ok {
		t.Error("TeamRoleResource does not implement resource.ResourceWithImportState")
	}
}

func TestTeamRoleResource_AsymmetricAPIPattern(t *testing.T) {
	// This is a structural documentation test verifying that the resource
	// handles the asymmetric API correctly:
	// - Read/Delete use role NAME (GetRole, DeleteRole)
	// - Update uses role ID (UpdateRole)
	// The actual API calls are tested via acceptance tests.
	// Here we verify the schema supports both identifiers.
	s := getResourceSchema(t)

	// ID must be computed (assigned by API on create, used for updates).
	idAttr, ok := s.Attributes["id"].(schema.StringAttribute)
	if !ok {
		t.Fatal("id attribute is not a StringAttribute")
	}
	if !idAttr.Computed {
		t.Error("id must be Computed (assigned by API, used for updates)")
	}

	// Name must be required (used for read/delete).
	nameAttr, ok := s.Attributes["name"].(schema.StringAttribute)
	if !ok {
		t.Fatal("name attribute is not a StringAttribute")
	}
	if !nameAttr.Required {
		t.Error("name must be Required (used for read/delete)")
	}
}

func TestTeamRoleResource_TimeoutsBlock(t *testing.T) {
	s := getResourceSchema(t)

	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("Expected 'timeouts' block not found in schema")
	}
}
