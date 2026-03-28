package teammember

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// getResourceSchema is a test helper that retrieves the team member resource schema.
func getResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := fwresource.SchemaRequest{}
	schemaResp := &fwresource.SchemaResponse{}

	r := NewTeamMemberResource()
	r.(*TeamMemberResource).Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

func TestTeamMemberResourceSchema(t *testing.T) {
	s := getResourceSchema(t)

	// Verify all expected attributes exist
	expectedAttrs := []string{"id", "email", "role"}
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

func TestTeamMemberResource_EmailIsRequired(t *testing.T) {
	s := getResourceSchema(t)

	emailAttr, ok := s.Attributes["email"]
	if !ok {
		t.Fatal("email attribute not found in schema")
	}

	strAttr, ok := emailAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("email attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("email attribute should be Required")
	}
}

func TestTeamMemberResource_EmailIsForceNew(t *testing.T) {
	s := getResourceSchema(t)

	emailAttr, ok := s.Attributes["email"]
	if !ok {
		t.Fatal("email attribute not found in schema")
	}

	strAttr, ok := emailAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("email attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("email attribute has no plan modifiers, expected RequiresReplace")
	}
}

func TestTeamMemberResource_RoleIsRequired(t *testing.T) {
	s := getResourceSchema(t)

	roleAttr, ok := s.Attributes["role"]
	if !ok {
		t.Fatal("role attribute not found in schema")
	}

	strAttr, ok := roleAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("role attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("role attribute should be Required")
	}
}

func TestTeamMemberResource_RoleIsForceNew(t *testing.T) {
	s := getResourceSchema(t)

	roleAttr, ok := s.Attributes["role"]
	if !ok {
		t.Fatal("role attribute not found in schema")
	}

	strAttr, ok := roleAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("role attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("role attribute has no plan modifiers, expected RequiresReplace")
	}
}

func TestTeamMemberResource_IDIsComputed(t *testing.T) {
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

func TestTeamMemberResource_IDHasUseStateForUnknown(t *testing.T) {
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

func TestTeamMemberResource_DescriptionsPresent(t *testing.T) {
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

func TestTeamMemberResource_EmailValidators(t *testing.T) {
	s := getResourceSchema(t)

	emailAttr, ok := s.Attributes["email"]
	if !ok {
		t.Fatal("email attribute not found in schema")
	}

	strAttr, ok := emailAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("email attribute is not a StringAttribute")
	}

	if len(strAttr.Validators) == 0 {
		t.Error("email attribute has no validators, expected LengthAtLeast(1)")
	}
}

func TestTeamMemberResource_Metadata(t *testing.T) {
	r := NewTeamMemberResource()
	metaReq := fwresource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	metaResp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), metaReq, metaResp)

	if metaResp.TypeName != "vastai_team_member" {
		t.Errorf("Expected type name 'vastai_team_member', got %q", metaResp.TypeName)
	}
}

func TestTeamMemberResource_ImplementsResourceInterface(t *testing.T) {
	r := NewTeamMemberResource()

	// Verify it implements resource.Resource
	if _, ok := r.(fwresource.Resource); !ok {
		t.Error("TeamMemberResource does not implement resource.Resource")
	}

	// Verify it implements resource.ResourceWithImportState
	if _, ok := r.(fwresource.ResourceWithImportState); !ok {
		t.Error("TeamMemberResource does not implement resource.ResourceWithImportState")
	}
}

func TestTeamMemberResource_InviteAsCreatePattern(t *testing.T) {
	// This is a structural documentation test verifying D-01:
	// Create = invite, Delete = remove. There is no update endpoint,
	// so both email and role have RequiresReplace.
	s := getResourceSchema(t)

	emailAttr := s.Attributes["email"].(schema.StringAttribute)
	if len(emailAttr.PlanModifiers) == 0 {
		t.Error("email should have RequiresReplace (no update mechanism)")
	}

	roleAttr := s.Attributes["role"].(schema.StringAttribute)
	if len(roleAttr.PlanModifiers) == 0 {
		t.Error("role should have RequiresReplace (no update mechanism)")
	}
}

func TestTeamMemberResource_TimeoutsBlock(t *testing.T) {
	s := getResourceSchema(t)

	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("Expected 'timeouts' block not found in schema")
	}
}
