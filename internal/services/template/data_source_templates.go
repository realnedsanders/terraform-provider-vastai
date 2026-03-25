package template

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TemplatesDataSource{}
var _ datasource.DataSourceWithConfigure = &TemplatesDataSource{}

// TemplatesDataSource defines the templates data source implementation.
type TemplatesDataSource struct {
	client *client.VastAIClient
}

// NewTemplatesDataSource creates a new templates data source instance.
func NewTemplatesDataSource() datasource.DataSource {
	return &TemplatesDataSource{}
}

// Metadata returns the data source type name.
func (d *TemplatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_templates"
}

// templateDataSourceAttrTypes returns the attribute types for TemplateDataSourceModel.
func templateDataSourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"numeric_id":             types.Int64Type,
		"name":                   types.StringType,
		"image":                  types.StringType,
		"tag":                    types.StringType,
		"env":                    types.StringType,
		"onstart":                types.StringType,
		"ssh_direct":             types.BoolType,
		"jup_direct":             types.BoolType,
		"use_jupyter_lab":        types.BoolType,
		"use_ssh":                types.BoolType,
		"private":                types.BoolType,
		"readme":                 types.StringType,
		"readme_visible":         types.BoolType,
		"desc":                   types.StringType,
		"recommended_disk_space": types.StringType,
		"created_at":             types.StringType,
	}
}

// templateNestedAttributes returns the schema attributes for a single template
// in the data source results (read-only view).
func templateNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Template hash ID (primary identifier).",
			Computed:    true,
		},
		"numeric_id": schema.Int64Attribute{
			Description: "Numeric template ID assigned by the Vast.ai API.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "Template name.",
			Computed:    true,
		},
		"image": schema.StringAttribute{
			Description: "Docker image reference used by this template.",
			Computed:    true,
		},
		"tag": schema.StringAttribute{
			Description: "Docker image tag.",
			Computed:    true,
		},
		"env": schema.StringAttribute{
			Description: "Environment variables and port mappings in Docker CLI format.",
			Computed:    true,
		},
		"onstart": schema.StringAttribute{
			Description: "Bash script executed when the instance starts.",
			Computed:    true,
		},
		"ssh_direct": schema.BoolAttribute{
			Description: "Whether direct SSH access is enabled.",
			Computed:    true,
		},
		"jup_direct": schema.BoolAttribute{
			Description: "Whether direct Jupyter access is enabled.",
			Computed:    true,
		},
		"use_jupyter_lab": schema.BoolAttribute{
			Description: "Whether JupyterLab is used instead of classic Jupyter.",
			Computed:    true,
		},
		"use_ssh": schema.BoolAttribute{
			Description: "Whether SSH server is enabled.",
			Computed:    true,
		},
		"private": schema.BoolAttribute{
			Description: "Whether this template is private.",
			Computed:    true,
		},
		"readme": schema.StringAttribute{
			Description: "Markdown readme content for the template.",
			Computed:    true,
		},
		"readme_visible": schema.BoolAttribute{
			Description: "Whether the readme is visible on the template listing.",
			Computed:    true,
		},
		"desc": schema.StringAttribute{
			Description: "Short description of the template.",
			Computed:    true,
		},
		"recommended_disk_space": schema.StringAttribute{
			Description: "Recommended disk space for instances using this template.",
			Computed:    true,
		},
		"created_at": schema.StringAttribute{
			Description: "Timestamp when the template was created.",
			Computed:    true,
		},
	}
}

// Schema defines the schema for the templates data source.
func (d *TemplatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Search for Vast.ai templates by query string. " +
			"Returns a list of templates matching the search criteria.",

		Attributes: map[string]schema.Attribute{
			"query": schema.StringAttribute{
				Description: "Search query string to filter templates. Matches against template name and description.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"templates": schema.ListNestedAttribute{
				Description: "List of templates matching the search query.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: templateNestedAttributes(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TemplatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.VastAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.VastAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

// Read queries the Vast.ai API for templates matching the search query.
func (d *TemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model TemplatesDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := model.Query.ValueString()

	tflog.Debug(ctx, "Searching templates", map[string]interface{}{
		"query": query,
	})

	// Call API
	templates, err := d.client.Templates.Search(ctx, query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Search Templates",
			fmt.Sprintf("An unexpected error occurred while searching templates: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Templates search complete", map[string]interface{}{
		"count": len(templates),
	})

	// Convert API templates to list
	templatesList, diags := apiTemplatesToList(templates)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Templates = templatesList

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// apiTemplateToDataSourceAttrValues converts a client.Template to a map of attr.Value.
func apiTemplateToDataSourceAttrValues(tmpl client.Template) map[string]attr.Value {
	return map[string]attr.Value{
		"id":                     types.StringValue(tmpl.HashID),
		"numeric_id":             types.Int64Value(int64(tmpl.ID)),
		"name":                   types.StringValue(tmpl.Name),
		"image":                  types.StringValue(tmpl.Image),
		"tag":                    types.StringValue(tmpl.Tag),
		"env":                    types.StringValue(tmpl.Env),
		"onstart":                types.StringValue(tmpl.Onstart),
		"ssh_direct":             types.BoolValue(tmpl.SSHDirect),
		"jup_direct":             types.BoolValue(tmpl.JupDirect),
		"use_jupyter_lab":        types.BoolValue(tmpl.UseJupyterLab),
		"use_ssh":                types.BoolValue(tmpl.UseSSH),
		"private":                types.BoolValue(tmpl.Private),
		"readme":                 types.StringValue(tmpl.Readme),
		"readme_visible":         types.BoolValue(tmpl.ReadmeVisible),
		"desc":                   types.StringValue(tmpl.Desc),
		"recommended_disk_space": types.StringValue(tmpl.RecommendedDiskSpace),
		"created_at":             types.StringValue(tmpl.CreatedAt),
	}
}

// apiTemplatesToList converts a slice of client.Template to a types.List.
func apiTemplatesToList(templates []client.Template) (types.List, diag.Diagnostics) {
	attrTypes := templateDataSourceAttrTypes()
	if len(templates) == 0 {
		return types.ListValueMust(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{}), nil
	}

	elems := make([]attr.Value, len(templates))
	for i, tmpl := range templates {
		obj, diags := types.ObjectValue(attrTypes, apiTemplateToDataSourceAttrValues(tmpl))
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}
		elems[i] = obj
	}

	return types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
}
