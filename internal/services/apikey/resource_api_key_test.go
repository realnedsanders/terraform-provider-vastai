package apikey

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// getResourceSchema is a test helper that retrieves the API key resource schema.
func getResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := fwresource.SchemaRequest{}
	schemaResp := &fwresource.SchemaResponse{}

	r := NewApiKeyResource()
	res, ok := r.(*ApiKeyResource)
	if !ok {
		t.Fatal("unexpected resource type")
	}
	res.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

func TestApiKeyResourceSchema(t *testing.T) {
	s := getResourceSchema(t)

	// Verify all expected attributes exist
	expectedAttrs := []string{"id", "name", "key", "permissions", "created_at"}
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

func TestApiKeyResource_KeyIsSensitive(t *testing.T) {
	s := getResourceSchema(t)

	keyAttr, ok := s.Attributes["key"]
	if !ok {
		t.Fatal("key attribute not found in schema")
	}

	strAttr, ok := keyAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("key attribute is not a StringAttribute")
	}

	if !strAttr.Sensitive {
		t.Error("key attribute should be marked as Sensitive")
	}
}

func TestApiKeyResource_KeyHasUseStateForUnknown(t *testing.T) {
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
		t.Error("key attribute has no plan modifiers, expected UseStateForUnknown")
	}
}

func TestApiKeyResource_NameRequiresReplace(t *testing.T) {
	s := getResourceSchema(t)

	nameAttr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found in schema")
	}

	strAttr, ok := nameAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("name attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("name attribute has no plan modifiers, expected RequiresReplace")
	}
}

func TestApiKeyResource_PermissionsRequiresReplace(t *testing.T) {
	s := getResourceSchema(t)

	permAttr, ok := s.Attributes["permissions"]
	if !ok {
		t.Fatal("permissions attribute not found in schema")
	}

	strAttr, ok := permAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("permissions attribute is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) == 0 {
		t.Error("permissions attribute has no plan modifiers, expected RequiresReplace")
	}
}

func TestApiKeyResource_DescriptionsPresent(t *testing.T) {
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

func TestApiKeyResource_PlanModifiers(t *testing.T) {
	s := getResourceSchema(t)

	// id, key, and created_at should have UseStateForUnknown
	for _, attrName := range []string{"id", "key", "created_at"} {
		attr, ok := s.Attributes[attrName]
		if !ok {
			t.Errorf("Attribute %q not found", attrName)
			continue
		}

		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Errorf("Attribute %q is not a StringAttribute", attrName)
			continue
		}

		if len(strAttr.PlanModifiers) == 0 {
			t.Errorf("Attribute %q has no plan modifiers, expected UseStateForUnknown", attrName)
		}
	}
}

func TestApiKeyResource_PermissionsHasValidator(t *testing.T) {
	s := getResourceSchema(t)

	permAttr, ok := s.Attributes["permissions"]
	if !ok {
		t.Fatal("permissions attribute not found in schema")
	}

	strAttr, ok := permAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("permissions attribute is not a StringAttribute")
	}

	if len(strAttr.Validators) == 0 {
		t.Error("permissions attribute has no validators, expected JSON validator")
	}
}

func TestApiKeyResource_Metadata(t *testing.T) {
	r := NewApiKeyResource()
	metaReq := fwresource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	metaResp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), metaReq, metaResp)

	if metaResp.TypeName != "vastai_api_key" {
		t.Errorf("Expected type name 'vastai_api_key', got %q", metaResp.TypeName)
	}
}

func TestApiKeyResource_NameHasLengthValidator(t *testing.T) {
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
		t.Error("name attribute has no validators, expected LengthAtLeast")
	}
}
