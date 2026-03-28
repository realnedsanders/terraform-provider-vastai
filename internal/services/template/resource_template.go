package template

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &TemplateResource{}
	_ resource.ResourceWithConfigure   = &TemplateResource{}
	_ resource.ResourceWithImportState = &TemplateResource{}
)

// TemplateResource defines the resource implementation.
type TemplateResource struct {
	client *client.VastAIClient
}

// NewTemplateResource creates a new template resource instance.
func NewTemplateResource() resource.Resource {
	return &TemplateResource{}
}

// Metadata returns the resource type name.
func (r *TemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

// Schema defines the schema for the template resource.
func (r *TemplateResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai template. Templates define reusable instance configurations " +
			"including Docker image, environment variables, startup commands, and access settings.",

		Attributes: map[string]schema.Attribute{
			// Primary identifier (hash_id)
			"id": schema.StringAttribute{
				Description: "Numeric template ID (stable across updates).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"hash_id": schema.StringAttribute{
				Description: "Content-addressed template hash ID. Changes when template content is updated.",
				Computed:    true,
			},

			// Secondary computed identifier
			"numeric_id": schema.Int64Attribute{
				Description: "Numeric template ID assigned by the Vast.ai API (same as id, typed as integer).",
				Computed:    true,
			},

			// Required attributes
			"name": schema.StringAttribute{
				Description: "Template name. Must be between 1 and 200 characters.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			"image": schema.StringAttribute{
				Description: "Docker image reference (e.g., 'pytorch/pytorch:latest', 'nvidia/cuda:12.0-devel'). " +
					"Must be a valid Docker image reference.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]+(:[a-zA-Z0-9._-]+)?$`),
						"must be a valid Docker image reference (e.g., 'pytorch/pytorch:latest')",
					),
				},
			},

			// Optional attributes with server-set defaults (Optional+Computed)
			"tag": schema.StringAttribute{
				Description: "Docker image tag. If not set, the API may assign a default based on the image reference.",
				Optional:    true,
				Computed:    true,
			},
			"ssh_direct": schema.BoolAttribute{
				Description: "Enable direct SSH access to instances created from this template.",
				Optional:    true,
				Computed:    true,
			},
			"jup_direct": schema.BoolAttribute{
				Description: "Enable direct Jupyter access to instances created from this template.",
				Optional:    true,
				Computed:    true,
			},
			"use_jupyter_lab": schema.BoolAttribute{
				Description: "Use JupyterLab instead of classic Jupyter Notebook interface.",
				Optional:    true,
				Computed:    true,
			},
			"use_ssh": schema.BoolAttribute{
				Description: "Enable SSH server in instances created from this template.",
				Optional:    true,
				Computed:    true,
			},
			"readme_visible": schema.BoolAttribute{
				Description: "Whether the readme is visible on the template listing.",
				Optional:    true,
				Computed:    true,
			},

			// Optional attributes
			"env": schema.StringAttribute{
				Description: "Environment variables and port mappings in Docker CLI format " +
					"(e.g., '-e KEY=VALUE -p 8080:8080'). Passed directly to the container runtime.",
				Optional: true,
			},
			"onstart": schema.StringAttribute{
				Description: "Bash script to run when the instance starts. Executed after the container is initialized.",
				Optional:    true,
			},
			"private": schema.BoolAttribute{
				Description: "Whether this template is private (only visible to the owner). Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"readme": schema.StringAttribute{
				Description: "Markdown readme content displayed on the template listing page.",
				Optional:    true,
			},
			"desc": schema.StringAttribute{
				Description: "Short description of the template.",
				Optional:    true,
			},
			"recommended_disk_space": schema.StringAttribute{
				Description: "Recommended disk space for instances using this template (e.g., '50GB').",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"docker_login_repo": schema.StringAttribute{
				Description: "Private Docker registry URL for authenticated image pulls. " +
					"This value is sensitive and will not be displayed in plan output.",
				Optional:  true,
				Sensitive: true,
			},
			"href": schema.StringAttribute{
				Description: "External URL associated with the template (e.g., documentation or project page). May be auto-populated from Docker image metadata.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repo": schema.StringAttribute{
				Description: "Source code repository URL for the template. May be auto-populated from Docker image metadata.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"jupyter_dir": schema.StringAttribute{
				Description: "Working directory for Jupyter when the instance starts.",
				Optional:    true,
			},

			// Computed attributes (read-only from API)
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the template was created (set by the API).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *TemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.VastAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.VastAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// Create creates a new template via the Vast.ai API.
func (r *TemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model TemplateResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	createTimeout, diags := model.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Build API request
	createReq := modelToCreateRequest(model)

	tflog.Debug(ctx, "Creating template", map[string]interface{}{
		"name":  model.Name.ValueString(),
		"image": model.Image.ValueString(),
	})

	// Call API
	tmpl, err := r.client.Templates.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Template",
			fmt.Sprintf("Could not create template: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Template created", map[string]interface{}{
		"hash_id": tmpl.HashID,
		"id":      tmpl.ID,
	})

	// Map response to model
	apiTemplateToModel(tmpl, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the template state from the Vast.ai API.
func (r *TemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model TemplateResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	readTimeout, diags := model.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	templateID := model.ID.ValueString()
	numID, err := strconv.Atoi(templateID)
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing Template ID",
			fmt.Sprintf("Could not parse template ID %q as integer: %s", templateID, err))
		return
	}

	tflog.Debug(ctx, "Reading template", map[string]interface{}{
		"id": numID,
	})

	// Search for template by numeric ID using API filter.
	filter := fmt.Sprintf(`{"id":{"eq":%d}}`, numID)
	templates, err := r.client.Templates.Search(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Template",
			fmt.Sprintf("Could not read template %s: %s", templateID, err),
		)
		return
	}

	// Find matching template in results.
	var found *client.Template
	for i := range templates {
		if templates[i].ID == numID {
			found = &templates[i]
			break
		}
	}

	if found == nil {
		// Template not found -- remove from state
		tflog.Warn(ctx, "Template not found, removing from state", map[string]interface{}{
			"id": templateID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	apiTemplateToModel(found, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update modifies an existing template via the Vast.ai API.
func (r *TemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model TemplateResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	updateTimeout, diags := model.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Get the current hash_id from state (not from plan, as it's computed)
	var state TemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hashID := state.HashID.ValueString()

	// Build API request
	updateReq := modelToCreateRequest(model)

	tflog.Debug(ctx, "Updating template", map[string]interface{}{
		"hash_id": hashID,
		"name":    model.Name.ValueString(),
	})

	// Call API
	tmpl, err := r.client.Templates.Update(ctx, hashID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Template",
			fmt.Sprintf("Could not update template %s: %s", hashID, err),
		)
		return
	}

	// Map response to model
	apiTemplateToModel(tmpl, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete removes a template via the Vast.ai API.
func (r *TemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model TemplateResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	deleteTimeout, diags := model.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	templateID := model.ID.ValueString()
	numericID, err := strconv.Atoi(templateID)
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing Template ID",
			fmt.Sprintf("Could not parse template ID %q: %s", templateID, err))
		return
	}

	tflog.Debug(ctx, "Deleting template", map[string]interface{}{
		"id":      numericID,
		"hash_id": model.HashID.ValueString(),
	})

	err = r.client.Templates.DeleteByID(ctx, numericID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Template",
			fmt.Sprintf("Could not delete template %d: %s", numericID, err),
		)
		return
	}

	tflog.Debug(ctx, "Template deleted", map[string]interface{}{
		"id":      numericID,
		"hash_id": model.HashID.ValueString(),
	})
}

// ImportState imports an existing template by numeric template ID.
// Usage: terraform import vastai_template.example <numeric_id>.
func (r *TemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// modelToCreateRequest converts a TemplateResourceModel to a client.CreateTemplateRequest.
func modelToCreateRequest(model TemplateResourceModel) *client.CreateTemplateRequest {
	req := &client.CreateTemplateRequest{}

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		req.Name = model.Name.ValueString()
	}
	if !model.Image.IsNull() && !model.Image.IsUnknown() {
		req.Image = model.Image.ValueString()
	}
	if !model.Tag.IsNull() && !model.Tag.IsUnknown() {
		req.Tag = model.Tag.ValueString()
	}
	if !model.Env.IsNull() && !model.Env.IsUnknown() {
		req.Env = model.Env.ValueString()
	}
	if !model.Onstart.IsNull() && !model.Onstart.IsUnknown() {
		req.Onstart = model.Onstart.ValueString()
	}
	if !model.SSHDirect.IsNull() && !model.SSHDirect.IsUnknown() {
		req.SSHDirect = model.SSHDirect.ValueBool()
	}
	if !model.JupDirect.IsNull() && !model.JupDirect.IsUnknown() {
		req.JupDirect = model.JupDirect.ValueBool()
	}
	if !model.UseJupyterLab.IsNull() && !model.UseJupyterLab.IsUnknown() {
		req.UseJupyterLab = model.UseJupyterLab.ValueBool()
	}
	if !model.UseSSH.IsNull() && !model.UseSSH.IsUnknown() {
		req.UseSSH = model.UseSSH.ValueBool()
	}
	if !model.Private.IsNull() && !model.Private.IsUnknown() {
		req.Private = model.Private.ValueBool()
	}
	if !model.Readme.IsNull() && !model.Readme.IsUnknown() {
		req.Readme = model.Readme.ValueString()
	}
	if !model.ReadmeVisible.IsNull() && !model.ReadmeVisible.IsUnknown() {
		req.ReadmeVisible = model.ReadmeVisible.ValueBool()
	}
	if !model.Desc.IsNull() && !model.Desc.IsUnknown() {
		req.Desc = model.Desc.ValueString()
	}
	if !model.RecommendedDiskSpace.IsNull() && !model.RecommendedDiskSpace.IsUnknown() {
		if v, err := strconv.ParseFloat(model.RecommendedDiskSpace.ValueString(), 64); err == nil {
			req.RecommendedDiskSpace = v
		}
	}
	if !model.DockerLoginRepo.IsNull() && !model.DockerLoginRepo.IsUnknown() {
		req.DockerLoginRepo = model.DockerLoginRepo.ValueString()
	}
	if !model.Href.IsNull() && !model.Href.IsUnknown() {
		req.Href = model.Href.ValueString()
	}
	if !model.Repo.IsNull() && !model.Repo.IsUnknown() {
		req.Repo = model.Repo.ValueString()
	}
	if !model.JupyterDir.IsNull() && !model.JupyterDir.IsUnknown() {
		req.JupyterDir = model.JupyterDir.ValueString()
	}

	return req
}

// apiTemplateToModel maps a client.Template to a TemplateResourceModel.
// Preserves the Timeouts value from the existing model.
func apiTemplateToModel(tmpl *client.Template, model *TemplateResourceModel) {
	model.ID = types.StringValue(strconv.Itoa(tmpl.ID))
	model.HashID = types.StringValue(tmpl.HashID)
	model.NumericID = types.Int64Value(int64(tmpl.ID))
	model.Name = types.StringValue(tmpl.Name)
	model.Image = types.StringValue(tmpl.Image)

	// Optional+Computed: store server values
	if tmpl.Tag != "" {
		model.Tag = types.StringValue(tmpl.Tag)
	} else if model.Tag.IsNull() || model.Tag.IsUnknown() {
		model.Tag = types.StringNull()
	}

	// String optional fields - preserve null if not set by API
	if tmpl.Env != "" {
		model.Env = types.StringValue(tmpl.Env)
	} else if model.Env.IsNull() || model.Env.IsUnknown() {
		model.Env = types.StringNull()
	}

	if tmpl.Onstart != "" {
		model.Onstart = types.StringValue(tmpl.Onstart)
	} else if model.Onstart.IsNull() || model.Onstart.IsUnknown() {
		model.Onstart = types.StringNull()
	}

	model.SSHDirect = types.BoolValue(tmpl.SSHDirect)
	model.JupDirect = types.BoolValue(tmpl.JupDirect)
	model.UseJupyterLab = types.BoolValue(tmpl.UseJupyterLab)
	model.UseSSH = types.BoolValue(tmpl.UseSSH)
	model.Private = types.BoolValue(tmpl.Private)
	model.ReadmeVisible = types.BoolValue(tmpl.ReadmeVisible)

	if tmpl.Readme != "" {
		model.Readme = types.StringValue(tmpl.Readme)
	} else if model.Readme.IsNull() || model.Readme.IsUnknown() {
		model.Readme = types.StringNull()
	}

	if tmpl.Desc != "" {
		model.Desc = types.StringValue(tmpl.Desc)
	} else if model.Desc.IsNull() || model.Desc.IsUnknown() {
		model.Desc = types.StringNull()
	}

	if rds := tmpl.RecommendedDiskSpaceString(); rds != "" {
		model.RecommendedDiskSpace = types.StringValue(rds)
	} else if model.RecommendedDiskSpace.IsNull() || model.RecommendedDiskSpace.IsUnknown() {
		model.RecommendedDiskSpace = types.StringNull()
	}

	if tmpl.DockerLoginRepo != "" {
		model.DockerLoginRepo = types.StringValue(tmpl.DockerLoginRepo)
	} else if model.DockerLoginRepo.IsNull() || model.DockerLoginRepo.IsUnknown() {
		model.DockerLoginRepo = types.StringNull()
	}

	if tmpl.Href != "" {
		model.Href = types.StringValue(tmpl.Href)
	} else if model.Href.IsNull() || model.Href.IsUnknown() {
		model.Href = types.StringNull()
	}

	if tmpl.Repo != "" {
		model.Repo = types.StringValue(tmpl.Repo)
	} else if model.Repo.IsNull() || model.Repo.IsUnknown() {
		model.Repo = types.StringNull()
	}

	if tmpl.JupyterDir != "" {
		model.JupyterDir = types.StringValue(tmpl.JupyterDir)
	} else if model.JupyterDir.IsNull() || model.JupyterDir.IsUnknown() {
		model.JupyterDir = types.StringNull()
	}

	if tmpl.CreatedAt != 0 {
		model.CreatedAt = types.StringValue(fmt.Sprintf("%.0f", tmpl.CreatedAt))
	} else if model.CreatedAt.IsNull() || model.CreatedAt.IsUnknown() {
		model.CreatedAt = types.StringNull()
	}
}
