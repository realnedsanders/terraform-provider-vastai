package envvar

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure EnvVarResource satisfies the required interfaces.
var (
	_ resource.Resource                = &EnvVarResource{}
	_ resource.ResourceWithImportState = &EnvVarResource{}
)

// EnvVarResource defines the resource implementation.
type EnvVarResource struct {
	client *client.VastAIClient
}

// NewEnvVarResource returns a new environment variable resource instance.
func NewEnvVarResource() resource.Resource {
	return &EnvVarResource{}
}

// Metadata returns the resource type name.
func (r *EnvVarResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}

// Schema defines the schema for the resource.
func (r *EnvVarResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai environment variable (secret). Environment variables are identified by their key name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The environment variable name (same as key). Used as the unique identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "Environment variable name. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"value": schema.StringAttribute{
				Description: "Environment variable value.",
				Required:    true,
				Sensitive:   true,
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

// Configure adds the provider-configured client to the resource.
func (r *EnvVarResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new environment variable.
func (r *EnvVarResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model EnvVarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set create timeout
	createTimeout, diags := model.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	key := model.Key.ValueString()
	value := model.Value.ValueString()

	tflog.Debug(ctx, "Creating environment variable", map[string]interface{}{
		"key": key,
	})

	if err := r.client.EnvVars.Create(ctx, key, value); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Environment Variable",
			fmt.Sprintf("Could not create environment variable %q: %s", key, err),
		)
		return
	}

	// ID is the key name for env vars
	model.ID = types.StringValue(key)

	tflog.Debug(ctx, "Created environment variable", map[string]interface{}{
		"key": key,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the environment variable state from the API.
func (r *EnvVarResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model EnvVarResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set read timeout
	readTimeout, diags := model.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	key := model.Key.ValueString()

	tflog.Debug(ctx, "Reading environment variable", map[string]interface{}{
		"key": key,
	})

	// List all env vars and find the one matching our key
	envVars, err := r.client.EnvVars.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Environment Variable",
			fmt.Sprintf("Could not list environment variables: %s", err),
		)
		return
	}

	value, found := envVars[key]
	if !found {
		tflog.Warn(ctx, "Environment variable not found, removing from state", map[string]interface{}{
			"key": key,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	model.Value = types.StringValue(value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update updates an existing environment variable.
func (r *EnvVarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model EnvVarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set update timeout
	updateTimeout, diags := model.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	key := model.Key.ValueString()
	value := model.Value.ValueString()

	tflog.Debug(ctx, "Updating environment variable", map[string]interface{}{
		"key": key,
	})

	if err := r.client.EnvVars.Update(ctx, key, value); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Environment Variable",
			fmt.Sprintf("Could not update environment variable %q: %s", key, err),
		)
		return
	}

	tflog.Debug(ctx, "Updated environment variable", map[string]interface{}{
		"key": key,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete deletes an environment variable.
func (r *EnvVarResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model EnvVarResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set delete timeout
	deleteTimeout, diags := model.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	key := model.Key.ValueString()

	tflog.Debug(ctx, "Deleting environment variable", map[string]interface{}{
		"key": key,
	})

	if err := r.client.EnvVars.Delete(ctx, key); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Environment Variable",
			fmt.Sprintf("Could not delete environment variable %q: %s", key, err),
		)
		return
	}

	tflog.Debug(ctx, "Deleted environment variable", map[string]interface{}{
		"key": key,
	})
}

// ImportState imports an existing environment variable by its key name.
// Usage: terraform import vastai_environment_variable.example <key_name>
func (r *EnvVarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// For env vars, the ID is the key name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), req.ID)...)
}
