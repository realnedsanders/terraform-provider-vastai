package envvar

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// getResourceSchema is a test helper that retrieves the environment variable resource schema.
func getResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := fwresource.SchemaRequest{}
	schemaResp := &fwresource.SchemaResponse{}

	r := NewEnvVarResource()
	r.(*EnvVarResource).Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

func TestEnvVarResourceSchema(t *testing.T) {
	s := getResourceSchema(t)

	// Verify all expected attributes exist
	expectedAttrs := []string{"id", "key", "value"}
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

func TestEnvVarResource_ValueIsSensitive(t *testing.T) {
	s := getResourceSchema(t)

	valueAttr, ok := s.Attributes["value"]
	if !ok {
		t.Fatal("value attribute not found in schema")
	}

	strAttr, ok := valueAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("value attribute is not a StringAttribute")
	}

	if !strAttr.Sensitive {
		t.Error("value attribute should be marked as Sensitive")
	}
}

func TestEnvVarResource_KeyRequiresReplace(t *testing.T) {
	s := getResourceSchema(t)

	keyAttr, ok := s.Attributes["key"]
	if !ok {
		t.Fatal("key attribute not found in schema")
	}

	strAttr, ok := keyAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("key attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("key attribute has no plan modifiers, expected RequiresReplace")
	}
}

func TestEnvVarResource_DescriptionsPresent(t *testing.T) {
	s := getResourceSchema(t)

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

func TestEnvVarResource_PlanModifiers(t *testing.T) {
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

func TestEnvVarResource_Metadata(t *testing.T) {
	r := NewEnvVarResource()
	metaReq := fwresource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	metaResp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), metaReq, metaResp)

	if metaResp.TypeName != "vastai_environment_variable" {
		t.Errorf("Expected type name 'vastai_environment_variable', got %q", metaResp.TypeName)
	}
}

func TestEnvVarResource_KeyHasLengthValidator(t *testing.T) {
	s := getResourceSchema(t)

	keyAttr, ok := s.Attributes["key"]
	if !ok {
		t.Fatal("key attribute not found in schema")
	}

	strAttr, ok := keyAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("key attribute is not a StringAttribute")
	}

	if len(strAttr.Validators) == 0 {
		t.Error("key attribute has no validators, expected LengthAtLeast")
	}
}

func TestEnvVarResource_ValueIsRequired(t *testing.T) {
	s := getResourceSchema(t)

	valueAttr, ok := s.Attributes["value"]
	if !ok {
		t.Fatal("value attribute not found in schema")
	}

	strAttr, ok := valueAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("value attribute is not a StringAttribute")
	}

	if !strAttr.Required {
		t.Error("value attribute should be Required")
	}
}

func TestEnvVarResource_SchemaDescription(t *testing.T) {
	s := getResourceSchema(t)

	if s.Description == "" {
		t.Error("Schema description should not be empty")
	}
}
