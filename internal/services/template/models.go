package template

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TemplateResourceModel describes the resource data model for vastai_template.
// The primary identifier is hash_id (stored as "id"), with numeric_id as a
// secondary computed identifier.
type TemplateResourceModel struct {
	// Primary identifier (hash_id from API)
	ID types.String `tfsdk:"id"`

	// Secondary computed identifier (template integer ID)
	NumericID types.Int64 `tfsdk:"numeric_id"`

	// Required attributes
	Name  types.String `tfsdk:"name"`
	Image types.String `tfsdk:"image"`

	// Optional attributes
	Tag              types.String `tfsdk:"tag"`
	Env              types.String `tfsdk:"env"`
	Onstart          types.String `tfsdk:"onstart"`
	SSHDirect        types.Bool   `tfsdk:"ssh_direct"`
	JupDirect        types.Bool   `tfsdk:"jup_direct"`
	UseJupyterLab    types.Bool   `tfsdk:"use_jupyter_lab"`
	UseSSH           types.Bool   `tfsdk:"use_ssh"`
	Private          types.Bool   `tfsdk:"private"`
	Readme           types.String `tfsdk:"readme"`
	ReadmeVisible    types.Bool   `tfsdk:"readme_visible"`
	Desc             types.String `tfsdk:"desc"`
	RecommendedDiskSpace types.String `tfsdk:"recommended_disk_space"`
	DockerLoginRepo  types.String `tfsdk:"docker_login_repo"`
	Href             types.String `tfsdk:"href"`
	Repo             types.String `tfsdk:"repo"`

	// Computed attributes (read-only from API)
	CreatedAt types.String `tfsdk:"created_at"`

	// Timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// TemplatesDataSourceModel describes the data source data model for vastai_templates.
type TemplatesDataSourceModel struct {
	Query     types.String `tfsdk:"query"`
	Templates types.List   `tfsdk:"templates"`
}

// TemplateDataSourceModel represents a single template in the data source results.
// This is a read-only view of template fields.
type TemplateDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	NumericID            types.Int64  `tfsdk:"numeric_id"`
	Name                 types.String `tfsdk:"name"`
	Image                types.String `tfsdk:"image"`
	Tag                  types.String `tfsdk:"tag"`
	Env                  types.String `tfsdk:"env"`
	Onstart              types.String `tfsdk:"onstart"`
	SSHDirect            types.Bool   `tfsdk:"ssh_direct"`
	JupDirect            types.Bool   `tfsdk:"jup_direct"`
	UseJupyterLab        types.Bool   `tfsdk:"use_jupyter_lab"`
	UseSSH               types.Bool   `tfsdk:"use_ssh"`
	Private              types.Bool   `tfsdk:"private"`
	Readme               types.String `tfsdk:"readme"`
	ReadmeVisible        types.Bool   `tfsdk:"readme_visible"`
	Desc                 types.String `tfsdk:"desc"`
	RecommendedDiskSpace types.String `tfsdk:"recommended_disk_space"`
	CreatedAt            types.String `tfsdk:"created_at"`
}
