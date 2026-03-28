package user

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TestUserDataSource_Metadata verifies the data source type name.
func TestUserDataSource_Metadata(t *testing.T) {
	ds := NewUserDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_user" {
		t.Errorf("expected type name vastai_user, got %s", resp.TypeName)
	}
}

// TestUserDataSource_Schema verifies the schema has all expected attributes.
func TestUserDataSource_Schema(t *testing.T) {
	ds := NewUserDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	expectedAttrs := []string{
		"id", "username", "email", "email_verified", "fullname",
		"balance", "credit", "has_billing", "ssh_key",
		"balance_threshold", "balance_threshold_enabled",
	}

	for _, name := range expectedAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing attribute: %s", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %s should be computed", name)
		}
	}
}

// TestUserDataSource_SchemaDescription verifies the schema has a non-empty description.
func TestUserDataSource_SchemaDescription(t *testing.T) {
	ds := NewUserDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema
	if s.Description == "" {
		t.Error("schema has empty description")
	}
}

// TestUserDataSource_SchemaAttributeDescriptions verifies all attributes have descriptions.
func TestUserDataSource_SchemaAttributeDescriptions(t *testing.T) {
	ds := NewUserDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	for name, attr := range s.Attributes {
		desc := getAttributeDescription(attr)
		if desc == "" {
			t.Errorf("attribute %s has empty description", name)
		}
	}
}

// TestUserDataSource_SSHKeySensitive verifies the ssh_key attribute is marked as sensitive.
func TestUserDataSource_SSHKeySensitive(t *testing.T) {
	ds := NewUserDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema
	sshKeyAttr, ok := s.Attributes["ssh_key"].(schema.StringAttribute)
	if !ok {
		t.Fatal("ssh_key is not a StringAttribute")
	}
	if !sshKeyAttr.Sensitive {
		t.Error("ssh_key should be marked as sensitive")
	}
}

// TestUserDataSource_ImplementsConfigure verifies configure is implemented.
func TestUserDataSource_ImplementsConfigure(t *testing.T) {
	ds := NewUserDataSource()
	_, ok := ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Error("UserDataSource should implement DataSourceWithConfigure")
	}
}

// getAttributeDescription extracts the description from a schema attribute.
func getAttributeDescription(attr schema.Attribute) string {
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
	default:
		return ""
	}
}
