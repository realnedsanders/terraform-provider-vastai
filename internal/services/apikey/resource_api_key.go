package apikey

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

// Ensure ApiKeyResource satisfies the required interfaces.
var (
	_ resource.Resource                = &ApiKeyResource{}
	_ resource.ResourceWithConfigure   = &ApiKeyResource{}
	_ resource.ResourceWithImportState = &ApiKeyResource{}
)

// ApiKeyResource defines the resource implementation.
type ApiKeyResource struct {
	client *client.VastAIClient
}

// NewApiKeyResource returns a new API key resource instance.
func NewApiKeyResource() resource.Resource {
	return &ApiKeyResource{}
}

// Metadata returns the resource type name.
func (r *ApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

// jsonValidator validates that a string value is valid JSON.
type jsonValidator struct{}

func (v jsonValidator) Description(_ context.Context) string {
	return "value must be valid JSON"
}

func (v jsonValidator) MarkdownDescription(_ context.Context) string {
	return "value must be valid JSON"
}

func (v jsonValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if !json.Valid([]byte(req.ConfigValue.ValueString())) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON",
			"The value must be a valid JSON string.",
		)
	}
}

// Schema defines the schema for the resource.
func (r *ApiKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai API key. API keys are immutable after creation -- changing name or permissions requires replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique numeric identifier of the API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the API key. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"key": schema.StringAttribute{
				Description: "The API key value. Only available on creation, never returned by subsequent reads. The key value cannot be recovered after creation.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"permissions": schema.StringAttribute{
				Description: "JSON string of permission configuration. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					jsonValidator{},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the API key was created.",
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
				Delete: true,
			}),
		},
	}
}

// Configure adds the provider-configured client to the resource.
func (r *ApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new API key.
func (r *ApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model ApiKeyResourceModel

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

	tflog.Debug(ctx, "Creating API key", map[string]interface{}{
		"name": model.Name.ValueString(),
	})

	// Parse permissions JSON string into raw JSON
	permissions := json.RawMessage(model.Permissions.ValueString())

	apiKey, err := r.client.ApiKeys.Create(ctx, model.Name.ValueString(), permissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating API Key",
			fmt.Sprintf("Could not create API key: %s", err),
		)
		return
	}

	// Map response to model
	model.ID = types.StringValue(strconv.Itoa(apiKey.ID))
	model.Key = types.StringValue(apiKey.Key)
	model.CreatedAt = types.StringValue(apiKey.CreatedAt)

	// Store permissions as normalized JSON string
	if apiKey.Permissions != nil {
		model.Permissions = types.StringValue(string(apiKey.Permissions))
	}

	tflog.Debug(ctx, "Created API key", map[string]interface{}{
		"id": apiKey.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the API key state from the API.
func (r *ApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ApiKeyResourceModel

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

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing API Key ID",
			fmt.Sprintf("Could not parse API key ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading API key", map[string]interface{}{
		"id": id,
	})

	// List all keys and find the one matching our ID (no single-get endpoint)
	keys, err := r.client.ApiKeys.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading API Key",
			fmt.Sprintf("Could not list API keys: %s", err),
		)
		return
	}

	var found *client.ApiKey
	for i := range keys {
		if keys[i].ID == id {
			found = &keys[i]
			break
		}
	}

	if found == nil {
		tflog.Warn(ctx, "API key not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response -- do NOT touch the key field (UseStateForUnknown preserves it)
	model.Name = types.StringValue(found.Name)
	if found.Permissions != nil {
		model.Permissions = types.StringValue(string(found.Permissions))
	}
	if found.CreatedAt != "" {
		model.CreatedAt = types.StringValue(found.CreatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update returns an error because API keys are immutable.
func (r *ApiKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"API Keys Are Immutable",
		"API keys cannot be updated. Modify requires replacement.",
	)
}

// Delete deletes an API key.
func (r *ApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ApiKeyResourceModel

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

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing API Key ID",
			fmt.Sprintf("Could not parse API key ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting API key", map[string]interface{}{
		"id": id,
	})

	if err := r.client.ApiKeys.Delete(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting API Key",
			fmt.Sprintf("Could not delete API key %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Deleted API key", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing API key by its numeric ID.
// Note: The key value cannot be recovered after creation.
// Usage: terraform import vastai_api_key.example <id>.
func (r *ApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
