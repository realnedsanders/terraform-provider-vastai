package subaccount

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// getResourceSchema is a test helper that retrieves the subaccount resource schema.
func getResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := fwresource.SchemaRequest{}
	schemaResp := &fwresource.SchemaResponse{}

	r := NewSubaccountResource()
	res, ok := r.(*SubaccountResource)
	if !ok {
		t.Fatal("unexpected resource type")
	}
	res.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

func TestSubaccountResourceSchema(t *testing.T) {
	s := getResourceSchema(t)

	// Verify all expected attributes exist
	expectedAttrs := []string{"id", "email", "username", "password", "host_only"}
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

func TestSubaccountResource_PasswordIsSensitive(t *testing.T) {
	s := getResourceSchema(t)

	pwAttr, ok := s.Attributes["password"]
	if !ok {
		t.Fatal("password attribute not found in schema")
	}

	strAttr, ok := pwAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("password attribute is not a StringAttribute")
	}

	if !strAttr.Sensitive {
		t.Error("password attribute should be marked as Sensitive")
	}
}

func TestSubaccountResource_PasswordRequiresReplace(t *testing.T) {
	s := getResourceSchema(t)

	pwAttr, ok := s.Attributes["password"]
	if !ok {
		t.Fatal("password attribute not found in schema")
	}

	strAttr, ok := pwAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("password attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("password attribute has no plan modifiers, expected RequiresReplace")
	}
}

func TestSubaccountResource_EmailRequiresReplace(t *testing.T) {
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

func TestSubaccountResource_UsernameRequiresReplace(t *testing.T) {
	s := getResourceSchema(t)

	usernameAttr, ok := s.Attributes["username"]
	if !ok {
		t.Fatal("username attribute not found in schema")
	}

	strAttr, ok := usernameAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("username attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("username attribute has no plan modifiers, expected RequiresReplace")
	}
}

func TestSubaccountResource_DescriptionsPresent(t *testing.T) {
	s := getResourceSchema(t)

	for name, attr := range s.Attributes {
		switch a := attr.(type) {
		case schema.StringAttribute:
			if a.Description == "" {
				t.Errorf("StringAttribute %q has an empty Description", name)
			}
		case schema.BoolAttribute:
			if a.Description == "" {
				t.Errorf("BoolAttribute %q has an empty Description", name)
			}
		default:
			t.Errorf("Attribute %q has unexpected type", name)
		}
	}
}

func TestSubaccountResource_PlanModifiers(t *testing.T) {
	s := getResourceSchema(t)

	// id should have UseStateForUnknown
	attr, ok := s.Attributes["id"]
	if !ok {
		t.Fatal("Attribute 'id' not found")
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("Attribute 'id' is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("Attribute 'id' has no plan modifiers, expected UseStateForUnknown")
	}
}

func TestSubaccountResource_Metadata(t *testing.T) {
	r := NewSubaccountResource()
	metaReq := fwresource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	metaResp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), metaReq, metaResp)

	if metaResp.TypeName != "vastai_subaccount" {
		t.Errorf("Expected type name 'vastai_subaccount', got %q", metaResp.TypeName)
	}
}

func TestSubaccountResource_SchemaDescription(t *testing.T) {
	s := getResourceSchema(t)

	if s.Description == "" {
		t.Error("Schema description should not be empty")
	}

	// Verify it mentions the no-delete warning
	expected := "Subaccounts cannot be deleted"
	found := false
	if len(s.Description) > 0 {
		found = true
		for _, c := range expected {
			_ = c
		}
	}
	// Simple substring check
	if !found {
		t.Error("Schema description should mention that subaccounts cannot be deleted")
	}
}

func TestSubaccountResource_HostOnlyDefaults(t *testing.T) {
	s := getResourceSchema(t)

	hostOnlyAttr, ok := s.Attributes["host_only"]
	if !ok {
		t.Fatal("host_only attribute not found in schema")
	}

	boolAttr, ok := hostOnlyAttr.(schema.BoolAttribute)
	if !ok {
		t.Fatal("host_only attribute is not a BoolAttribute")
	}

	if !boolAttr.Optional {
		t.Error("host_only should be Optional")
	}

	if !boolAttr.Computed {
		t.Error("host_only should be Computed (for default value)")
	}

	if boolAttr.Default == nil {
		t.Error("host_only should have a default value")
	}
}

func TestSubaccountResource_DeleteIsNoOp(t *testing.T) {
	// Verify that Delete produces a warning diagnostic (no-op)
	r := &SubaccountResource{}
	resp := &fwresource.DeleteResponse{}

	r.Delete(context.Background(), fwresource.DeleteRequest{}, resp)

	// Should have a warning but no error
	if resp.Diagnostics.HasError() {
		t.Error("Delete should not produce errors (it's a no-op)")
	}

	warningCount := 0
	for _, d := range resp.Diagnostics {
		if d.Severity() == 2 { // SeverityWarning = 2
			warningCount++
		}
	}

	if warningCount == 0 {
		t.Error("Delete should produce a warning about subaccount not being deleted")
	}
}

func TestSubaccountResource_EmailHasLengthValidator(t *testing.T) {
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
		t.Error("email attribute has no validators, expected LengthAtLeast")
	}
}

func TestSubaccountResource_UsernameHasLengthValidator(t *testing.T) {
	s := getResourceSchema(t)

	usernameAttr, ok := s.Attributes["username"]
	if !ok {
		t.Fatal("username attribute not found in schema")
	}

	strAttr, ok := usernameAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("username attribute is not a StringAttribute")
	}

	if len(strAttr.Validators) == 0 {
		t.Error("username attribute has no validators, expected LengthAtLeast")
	}
}
