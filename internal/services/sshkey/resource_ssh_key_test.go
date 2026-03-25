package sshkey

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// getResourceSchema is a test helper that retrieves the SSH key resource schema.
func getResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := fwresource.SchemaRequest{}
	schemaResp := &fwresource.SchemaResponse{}

	r := NewSSHKeyResource()
	r.(*SSHKeyResource).Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

// getDataSourceSchema is a test helper that retrieves the SSH keys data source schema.
func getDataSourceSchema(t *testing.T) datasourceschema.Schema {
	t.Helper()
	ctx := context.Background()
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	d := NewSSHKeysDataSource()
	d.(*SSHKeysDataSource).Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

func TestSSHKeyResourceSchema(t *testing.T) {
	s := getResourceSchema(t)

	// Verify all expected attributes exist
	expectedAttrs := []string{"id", "ssh_key", "created_at"}
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

func TestSSHKeyResource_SSHKeyIsSensitive(t *testing.T) {
	s := getResourceSchema(t)

	sshKeyAttr, ok := s.Attributes["ssh_key"]
	if !ok {
		t.Fatal("ssh_key attribute not found in schema")
	}

	strAttr, ok := sshKeyAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("ssh_key attribute is not a StringAttribute")
	}

	if !strAttr.Sensitive {
		t.Error("ssh_key attribute should be marked as Sensitive")
	}
}

func TestSSHKeyResource_DescriptionsPresent(t *testing.T) {
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

func TestSSHKeyResource_PlanModifiers(t *testing.T) {
	s := getResourceSchema(t)

	// id and created_at should have UseStateForUnknown
	for _, attrName := range []string{"id", "created_at"} {
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

func TestSSHKeyResource_Validators(t *testing.T) {
	s := getResourceSchema(t)

	sshKeyAttr, ok := s.Attributes["ssh_key"]
	if !ok {
		t.Fatal("ssh_key attribute not found in schema")
	}

	strAttr, ok := sshKeyAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("ssh_key attribute is not a StringAttribute")
	}

	// Should have at least 2 validators: LengthAtLeast and RegexMatches
	if len(strAttr.Validators) < 2 {
		t.Errorf("ssh_key attribute has %d validators, expected at least 2 (LengthAtLeast, RegexMatches)", len(strAttr.Validators))
	}
}

func TestSSHKeyResource_Metadata(t *testing.T) {
	r := NewSSHKeyResource()
	metaReq := fwresource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	metaResp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), metaReq, metaResp)

	if metaResp.TypeName != "vastai_ssh_key" {
		t.Errorf("Expected type name 'vastai_ssh_key', got %q", metaResp.TypeName)
	}
}

func TestSSHKeysDataSourceSchema(t *testing.T) {
	s := getDataSourceSchema(t)

	// Verify ssh_keys list attribute exists
	sshKeysAttr, ok := s.Attributes["ssh_keys"]
	if !ok {
		t.Fatal("ssh_keys attribute not found in data source schema")
	}

	// Verify it's a ListNestedAttribute
	_, ok = sshKeysAttr.(datasourceschema.ListNestedAttribute)
	if !ok {
		t.Fatal("ssh_keys attribute is not a ListNestedAttribute")
	}
}

func TestSSHKeysDataSource_Metadata(t *testing.T) {
	d := NewSSHKeysDataSource()
	metaReq := datasource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	metaResp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), metaReq, metaResp)

	if metaResp.TypeName != "vastai_ssh_keys" {
		t.Errorf("Expected type name 'vastai_ssh_keys', got %q", metaResp.TypeName)
	}
}

func TestSSHKeysDataSource_SSHKeyIsSensitive(t *testing.T) {
	s := getDataSourceSchema(t)

	sshKeysAttr, ok := s.Attributes["ssh_keys"]
	if !ok {
		t.Fatal("ssh_keys attribute not found")
	}

	listAttr, ok := sshKeysAttr.(datasourceschema.ListNestedAttribute)
	if !ok {
		t.Fatal("ssh_keys is not a ListNestedAttribute")
	}

	nestedAttrs := listAttr.NestedObject.Attributes
	sshKeyNested, ok := nestedAttrs["ssh_key"]
	if !ok {
		t.Fatal("ssh_key nested attribute not found")
	}

	strAttr, ok := sshKeyNested.(datasourceschema.StringAttribute)
	if !ok {
		t.Fatal("ssh_key nested attribute is not a StringAttribute")
	}

	if !strAttr.Sensitive {
		t.Error("ssh_key nested attribute should be marked as Sensitive")
	}
}

func TestSSHKeysDataSource_DescriptionsPresent(t *testing.T) {
	s := getDataSourceSchema(t)

	// Check top-level description
	if s.Description == "" {
		t.Error("Data source schema has an empty Description")
	}

	// Check ssh_keys attribute
	sshKeysAttr, ok := s.Attributes["ssh_keys"]
	if !ok {
		t.Fatal("ssh_keys attribute not found")
	}

	listAttr, ok := sshKeysAttr.(datasourceschema.ListNestedAttribute)
	if !ok {
		t.Fatal("ssh_keys is not a ListNestedAttribute")
	}

	if listAttr.Description == "" {
		t.Error("ssh_keys attribute has an empty Description")
	}

	// Check nested attributes
	for name, attr := range listAttr.NestedObject.Attributes {
		strAttr, ok := attr.(datasourceschema.StringAttribute)
		if !ok {
			t.Errorf("Nested attribute %q is not a StringAttribute", name)
			continue
		}

		if strAttr.Description == "" {
			t.Errorf("Nested attribute %q has an empty Description", name)
		}
	}
}
