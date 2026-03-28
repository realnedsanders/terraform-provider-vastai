package team

import (
	"context"
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

// Ensure TeamResource satisfies the required interfaces.
var (
	_ resource.Resource                = &TeamResource{}
	_ resource.ResourceWithImportState = &TeamResource{}
)

// TeamResource defines the resource implementation.
type TeamResource struct {
	client *client.VastAIClient
}

// NewTeamResource returns a new team resource instance.
func NewTeamResource() resource.Resource {
	return &TeamResource{}
}

// Metadata returns the resource type name.
func (r *TeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

// Schema defines the schema for the resource.
func (r *TeamResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai team. Only one team can be managed per API key context. " +
			"The destroy operation deletes the team associated with the current API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique numeric identifier of the team.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_name": schema.StringAttribute{
				Description: "Name of the team. Changing this forces creation of a new team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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
func (r *TeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new team.
func (r *TeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model TeamResourceModel

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

	tflog.Debug(ctx, "Creating team", map[string]interface{}{
		"team_name": model.TeamName.ValueString(),
	})

	team, err := r.client.Teams.CreateTeam(ctx, model.TeamName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Team",
			fmt.Sprintf("Could not create team: %s", err),
		)
		return
	}

	// Map response to model
	model.ID = types.StringValue(strconv.Itoa(team.ID))
	model.TeamName = types.StringValue(team.TeamName)

	tflog.Debug(ctx, "Created team", map[string]interface{}{
		"id":        team.ID,
		"team_name": team.TeamName,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the team state.
// Note: There is no single-team GET endpoint. We verify the team exists by
// calling ListRoles (which requires team context). If the call fails, we
// assume the team was destroyed externally and remove it from state.
// The team_name and ID are stored from Create and treated as source of truth
// via UseStateForUnknown.
func (r *TeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model TeamResourceModel

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

	tflog.Debug(ctx, "Reading team", map[string]interface{}{
		"id": model.ID.ValueString(),
	})

	// Verify team exists by trying to list roles (team context required).
	// If this fails, the team was likely destroyed externally.
	_, err := r.client.Teams.ListRoles(ctx)
	if err != nil {
		tflog.Warn(ctx, "Team not found (ListRoles failed), removing from state", map[string]interface{}{
			"id":    model.ID.ValueString(),
			"error": err.Error(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Team exists -- state is already correct (UseStateForUnknown on computed fields)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is not supported -- team_name has RequiresReplace.
func (r *TeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Team resources cannot be updated in place. Changing the team name requires replacement.",
	)
}

// Delete destroys the team.
func (r *TeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model TeamResourceModel

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

	tflog.Debug(ctx, "Deleting team", map[string]interface{}{
		"id": model.ID.ValueString(),
	})

	if err := r.client.Teams.DestroyTeam(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Team",
			fmt.Sprintf("Could not delete team: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Deleted team", map[string]interface{}{
		"id": model.ID.ValueString(),
	})
}

// ImportState imports an existing team by its numeric ID.
// Usage: terraform import vastai_team.example <id>.
func (r *TeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
