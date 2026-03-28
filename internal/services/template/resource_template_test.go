package template

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// TestTemplateResourceMetadata verifies the resource type name.
func TestTemplateResourceMetadata(t *testing.T) {
	r := NewTemplateResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_template" {
		t.Errorf("expected type name vastai_template, got %s", resp.TypeName)
	}
}

// TestTemplateResourceSchema verifies the schema has all expected attributes
// with correct types, descriptions, and validators.
func TestTemplateResourceSchema(t *testing.T) {
	r := NewTemplateResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	// Verify required attributes
	requiredAttrs := []string{"name", "image"}
	for _, name := range requiredAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing required attribute: %s", name)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("attribute %s should be required", name)
		}
	}

	// Verify computed attributes
	computedAttrs := []string{"id", "numeric_id", "created_at"}
	for _, name := range computedAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing computed attribute: %s", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %s should be computed", name)
		}
	}

	// Verify optional+computed attributes (server-set defaults)
	optionalComputedAttrs := []string{"tag", "ssh_direct", "jup_direct", "use_jupyter_lab", "use_ssh", "readme_visible", "private"}
	for _, name := range optionalComputedAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing optional+computed attribute: %s", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("attribute %s should be optional", name)
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %s should be computed", name)
		}
	}

	// Verify optional attributes
	optionalAttrs := []string{"env", "onstart", "readme", "desc", "recommended_disk_space", "docker_login_repo", "href", "repo"}
	for _, name := range optionalAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing optional attribute: %s", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("attribute %s should be optional", name)
		}
	}

	// Verify all attributes have non-empty descriptions
	for name, attr := range s.Attributes {
		desc := getResourceAttributeDescription(attr)
		if desc == "" {
			t.Errorf("attribute %s has empty description", name)
		}
	}

	// Verify timeouts block exists
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
}

// TestTemplateResourceSchemaValidators verifies that validators are present
// on constrained attributes.
func TestTemplateResourceSchemaValidators(t *testing.T) {
	r := NewTemplateResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// Check name has LengthBetween validator
	nameAttr, ok := s.Attributes["name"].(schema.StringAttribute)
	if !ok {
		t.Fatal("name is not StringAttribute")
	}
	if len(nameAttr.Validators) == 0 {
		t.Error("name should have validators (LengthBetween(1, 200))")
	}

	// Check image has validators (LengthAtLeast + RegexMatches)
	imageAttr, ok := s.Attributes["image"].(schema.StringAttribute)
	if !ok {
		t.Fatal("image is not StringAttribute")
	}
	if len(imageAttr.Validators) < 2 {
		t.Errorf("image should have at least 2 validators, got %d", len(imageAttr.Validators))
	}
}

// TestTemplateResourceSensitiveAttributes verifies that sensitive fields are marked.
func TestTemplateResourceSensitiveAttributes(t *testing.T) {
	r := NewTemplateResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// Verify docker_login_repo is marked sensitive
	dockerLogin, ok := s.Attributes["docker_login_repo"].(schema.StringAttribute)
	if !ok {
		t.Fatal("docker_login_repo is not StringAttribute")
	}
	if !dockerLogin.Sensitive {
		t.Error("docker_login_repo should be marked as sensitive")
	}
}

// TestTemplateResourcePlanModifiers verifies UseStateForUnknown on stable computed fields.
func TestTemplateResourcePlanModifiers(t *testing.T) {
	r := NewTemplateResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// Check id has UseStateForUnknown plan modifier
	idAttr, ok := s.Attributes["id"].(schema.StringAttribute)
	if !ok {
		t.Fatal("id is not StringAttribute")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("id should have UseStateForUnknown plan modifier")
	}

	// Check created_at has UseStateForUnknown plan modifier
	createdAtAttr, ok := s.Attributes["created_at"].(schema.StringAttribute)
	if !ok {
		t.Fatal("created_at is not StringAttribute")
	}
	if len(createdAtAttr.PlanModifiers) == 0 {
		t.Error("created_at should have UseStateForUnknown plan modifier")
	}
}

// TestTemplateResourceImplementsImportState verifies that import state is implemented.
func TestTemplateResourceImplementsImportState(t *testing.T) {
	r := NewTemplateResource()
	_, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Error("TemplateResource should implement ResourceWithImportState")
	}
}

// TestModelToCreateRequest verifies the conversion from Terraform model to API request.
func TestModelToCreateRequest(t *testing.T) {
	model := TemplateResourceModel{
		Name:          types.StringValue("my-template"),
		Image:         types.StringValue("pytorch/pytorch:latest"),
		Tag:           types.StringValue("latest"),
		Env:           types.StringValue("-e KEY=VALUE"),
		Onstart:       types.StringValue("echo hello"),
		SSHDirect:     types.BoolValue(true),
		JupDirect:     types.BoolValue(false),
		UseJupyterLab: types.BoolValue(true),
		UseSSH:        types.BoolValue(true),
		Private:       types.BoolValue(false),
	}

	req := modelToCreateRequest(model)

	if req.Name != "my-template" {
		t.Errorf("expected name 'my-template', got %s", req.Name)
	}
	if req.Image != "pytorch/pytorch:latest" {
		t.Errorf("expected image 'pytorch/pytorch:latest', got %s", req.Image)
	}
	if req.Tag != "latest" {
		t.Errorf("expected tag 'latest', got %s", req.Tag)
	}
	if req.Env != "-e KEY=VALUE" {
		t.Errorf("expected env '-e KEY=VALUE', got %s", req.Env)
	}
	if req.Onstart != "echo hello" {
		t.Errorf("expected onstart 'echo hello', got %s", req.Onstart)
	}
	if !req.SSHDirect {
		t.Error("expected ssh_direct to be true")
	}
	if req.JupDirect {
		t.Error("expected jup_direct to be false")
	}
	if !req.UseJupyterLab {
		t.Error("expected use_jupyter_lab to be true")
	}
}

// TestModelToCreateRequestNullFields verifies null fields are not set in the request.
func TestModelToCreateRequestNullFields(t *testing.T) {
	model := TemplateResourceModel{
		Name:  types.StringValue("minimal-template"),
		Image: types.StringValue("ubuntu:22.04"),
		// All other fields are null (zero value of types.String/Bool/Int64 is null)
	}

	req := modelToCreateRequest(model)

	if req.Env != "" {
		t.Errorf("expected empty env for null model, got %s", req.Env)
	}
	if req.Onstart != "" {
		t.Errorf("expected empty onstart for null model, got %s", req.Onstart)
	}
}

// TestApiTemplateToModel verifies the conversion from API Template to Terraform model.
func TestApiTemplateToModel(t *testing.T) {
	tmpl := &client.Template{
		ID:            42,
		HashID:        "abc123def",
		Name:          "test-template",
		Image:         "pytorch/pytorch",
		Tag:           "2.0",
		CreatedAt:     1704067200.0,
		Env:           "-e FOO=BAR",
		Onstart:       "pip install torch",
		SSHDirect:     true,
		JupDirect:     false,
		UseSSH:        true,
		UseJupyterLab: false,
		Private:       true,
		ReadmeVisible: true,
	}

	model := &TemplateResourceModel{}
	apiTemplateToModel(tmpl, model)

	if model.ID.ValueString() != "42" {
		t.Errorf("expected id '42', got %s", model.ID.ValueString())
	}
	if model.HashID.ValueString() != "abc123def" {
		t.Errorf("expected hash_id 'abc123def', got %s", model.HashID.ValueString())
	}
	if model.NumericID.ValueInt64() != 42 {
		t.Errorf("expected numeric_id 42, got %d", model.NumericID.ValueInt64())
	}
	if model.Name.ValueString() != "test-template" {
		t.Errorf("expected name 'test-template', got %s", model.Name.ValueString())
	}
	if model.Image.ValueString() != "pytorch/pytorch" {
		t.Errorf("expected image 'pytorch/pytorch', got %s", model.Image.ValueString())
	}
	if model.Tag.ValueString() != "2.0" {
		t.Errorf("expected tag '2.0', got %s", model.Tag.ValueString())
	}
	if !model.SSHDirect.ValueBool() {
		t.Error("expected ssh_direct to be true")
	}
	if model.JupDirect.ValueBool() {
		t.Error("expected jup_direct to be false")
	}
	if !model.Private.ValueBool() {
		t.Error("expected private to be true")
	}
}

// TestTemplatesDataSourceMetadata verifies the data source type name.
func TestTemplatesDataSourceMetadata(t *testing.T) {
	ds := NewTemplatesDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_templates" {
		t.Errorf("expected type name vastai_templates, got %s", resp.TypeName)
	}
}

// TestTemplatesDataSourceSchema verifies the data source schema.
func TestTemplatesDataSourceSchema(t *testing.T) {
	ds := NewTemplatesDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	// Verify query attribute
	queryAttr, ok := s.Attributes["query"]
	if !ok {
		t.Fatal("missing query attribute")
	}
	if !queryAttr.IsRequired() {
		t.Error("query should be required")
	}

	// Verify templates attribute
	templatesAttr, ok := s.Attributes["templates"]
	if !ok {
		t.Fatal("missing templates attribute")
	}
	if !templatesAttr.IsComputed() {
		t.Error("templates should be computed")
	}

	// Verify all attributes have descriptions
	for name, attr := range s.Attributes {
		desc := getDSAttributeDescription(attr)
		if desc == "" {
			t.Errorf("attribute %s has empty description", name)
		}
	}

	// Verify nested template attributes have descriptions
	templatesListAttr, ok := s.Attributes["templates"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("templates attribute is not ListNestedAttribute")
	}
	for name, attr := range templatesListAttr.NestedObject.Attributes {
		desc := getDSAttributeDescription(attr)
		if desc == "" {
			t.Errorf("nested template attribute %s has empty description", name)
		}
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

// getDSAttributeDescription extracts the description from a data source schema attribute.
func getDSAttributeDescription(attr dsschema.Attribute) string {
	switch a := attr.(type) {
	case dsschema.StringAttribute:
		return a.Description
	case dsschema.Int64Attribute:
		return a.Description
	case dsschema.Float64Attribute:
		return a.Description
	case dsschema.BoolAttribute:
		return a.Description
	case dsschema.ListNestedAttribute:
		return a.Description
	case dsschema.SingleNestedAttribute:
		return a.Description
	default:
		return ""
	}
}
