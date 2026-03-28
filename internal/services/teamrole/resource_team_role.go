package teamrole

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

// Ensure TeamRoleResource satisfies the required interfaces.
var (
	_ resource.Resource                = &TeamRoleResource{}
	_ resource.ResourceWithImportState = &TeamRoleResource{}
)

// TeamRoleResource defines the resource implementation.
type TeamRoleResource struct {
	client *client.VastAIClient
}

// NewTeamRoleResource returns a new team role resource instance.
func NewTeamRoleResource() resource.Resource {
	return &TeamRoleResource{}
}

// Metadata returns the resource type name.
func (r *TeamRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_role"
}

// jsonValidator validates that a string is valid JSON.
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

	value := req.ConfigValue.ValueString()
	if !json.Valid([]byte(value)) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON",
			fmt.Sprintf("The value %q is not valid JSON.", value),
		)
	}
}

// Schema defines the schema for the resource.
func (r *TeamRoleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai team role with permission configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric ID of the role (used for updates).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Role name (used for read/delete operations).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"permissions": schema.StringAttribute{
				Description: "JSON string of permission configuration. The API uses nested JSON objects " +
					"(e.g., {\"api\":{\"instance_read\":{...}}}), not a flat string set.",
				Required: true,
				Validators: []validator.String{
					jsonValidator{},
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

// Configure adds the provider-configured client to the resource.
func (r *TeamRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new team role.
func (r *TeamRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model TeamRoleResourceModel

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

	name := model.Name.ValueString()
	perms := json.RawMessage(model.Permissions.ValueString())

	tflog.Debug(ctx, "Creating team role", map[string]interface{}{
		"name": name,
	})

	role, err := r.client.Teams.CreateRole(ctx, name, perms)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Team Role",
			fmt.Sprintf("Could not create team role: %s", err),
		)
		return
	}

	// Map response to model
	model.ID = types.StringValue(strconv.Itoa(role.ID))
	model.Name = types.StringValue(role.Name)
	if role.Permissions != nil {
		model.Permissions = types.StringValue(string(role.Permissions))
	}

	tflog.Debug(ctx, "Created team role", map[string]interface{}{
		"id":   role.ID,
		"name": role.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the team role state from the API.
// Read uses role NAME per Pitfall 3 (asymmetric API).
func (r *TeamRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model TeamRoleResourceModel

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

	// If we only have an ID (from import), resolve the name first by listing roles
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		tflog.Debug(ctx, "Resolving role name from ID via ListRoles", map[string]interface{}{
			"id": model.ID.ValueString(),
		})

		id, err := strconv.Atoi(model.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Team Role ID",
				fmt.Sprintf("Could not parse team role ID %q as integer: %s", model.ID.ValueString(), err),
			)
			return
		}

		roles, err := r.client.Teams.ListRoles(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Team Roles",
				fmt.Sprintf("Could not list team roles to resolve name for ID %d: %s", id, err),
			)
			return
		}

		var found *client.TeamRole
		for i := range roles {
			if roles[i].ID == id {
				found = &roles[i]
				break
			}
		}

		if found == nil {
			tflog.Warn(ctx, "Team role not found by ID, removing from state", map[string]interface{}{
				"id": id,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		model.Name = types.StringValue(found.Name)
	}

	name := model.Name.ValueString()

	tflog.Debug(ctx, "Reading team role", map[string]interface{}{
		"name": name,
	})

	// Read uses NAME per Pitfall 3
	role, err := r.client.Teams.GetRole(ctx, name)
	if err != nil {
		tflog.Warn(ctx, "Team role not found, removing from state", map[string]interface{}{
			"name":  name,
			"error": err.Error(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	model.ID = types.StringValue(strconv.Itoa(role.ID))
	model.Name = types.StringValue(role.Name)
	if role.Permissions != nil {
		model.Permissions = types.StringValue(string(role.Permissions))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update updates a team role.
// Update uses role ID per Pitfall 3 (asymmetric API).
func (r *TeamRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model TeamRoleResourceModel

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

	// Parse the numeric ID for the update call
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Team Role ID",
			fmt.Sprintf("Could not parse team role ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	name := model.Name.ValueString()
	perms := json.RawMessage(model.Permissions.ValueString())

	tflog.Debug(ctx, "Updating team role", map[string]interface{}{
		"id":   id,
		"name": name,
	})

	// Update uses ID per Pitfall 3
	role, err := r.client.Teams.UpdateRole(ctx, id, name, perms)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Team Role",
			fmt.Sprintf("Could not update team role %d: %s", id, err),
		)
		return
	}

	// Update model from response
	model.ID = types.StringValue(strconv.Itoa(role.ID))
	model.Name = types.StringValue(role.Name)
	if role.Permissions != nil {
		model.Permissions = types.StringValue(string(role.Permissions))
	}

	tflog.Debug(ctx, "Updated team role", map[string]interface{}{
		"id":   role.ID,
		"name": role.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete deletes a team role.
// Delete uses role NAME per Pitfall 3 (asymmetric API).
func (r *TeamRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model TeamRoleResourceModel

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

	name := model.Name.ValueString()

	tflog.Debug(ctx, "Deleting team role", map[string]interface{}{
		"name": name,
	})

	// Delete uses NAME per Pitfall 3
	if err := r.client.Teams.DeleteRole(ctx, name); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Team Role",
			fmt.Sprintf("Could not delete team role %q: %s", name, err),
		)
		return
	}

	tflog.Debug(ctx, "Deleted team role", map[string]interface{}{
		"name": name,
	})
}

// ImportState imports an existing team role by its numeric ID.
// The Read function will resolve the role name from the ID via ListRoles.
// Usage: terraform import vastai_team_role.example <id>
func (r *TeamRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
