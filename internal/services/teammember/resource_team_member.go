package teammember

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

// Ensure TeamMemberResource satisfies the required interfaces.
var (
	_ resource.Resource                = &TeamMemberResource{}
	_ resource.ResourceWithImportState = &TeamMemberResource{}
)

// TeamMemberResource defines the resource implementation.
type TeamMemberResource struct {
	client *client.VastAIClient
}

// NewTeamMemberResource returns a new team member resource instance.
func NewTeamMemberResource() resource.Resource {
	return &TeamMemberResource{}
}

// Metadata returns the resource type name.
func (r *TeamMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_member"
}

// Schema defines the schema for the resource.
func (r *TeamMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai team member. Creating this resource sends an invitation " +
			"to the specified email address. The member's user ID becomes the resource ID " +
			"once they accept or when the invite is registered.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Member ID assigned by Vast.ai.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address to invite. Changing this requires removing and re-inviting the member.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"role": schema.StringAttribute{
				Description: "Role name to assign to the member. Must match an existing team role. " +
					"Changing the role requires removing and re-inviting the member.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
func (r *TeamMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create invites a new team member (invite IS the create operation per D-01).
func (r *TeamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model TeamMemberResourceModel

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

	email := model.Email.ValueString()
	role := model.Role.ValueString()

	tflog.Debug(ctx, "Inviting team member", map[string]interface{}{
		"email": email,
		"role":  role,
	})

	// Invite the member (uses query params per Pitfall 5)
	if err := r.client.Teams.InviteMember(ctx, email, role); err != nil {
		resp.Diagnostics.AddError(
			"Error Inviting Team Member",
			fmt.Sprintf("Could not invite team member %q: %s", email, err),
		)
		return
	}

	// After invite, list members to find the new member by email and get their ID
	members, err := r.client.Teams.ListMembers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Team Members",
			fmt.Sprintf("Could not list team members after inviting %q: %s", email, err),
		)
		return
	}

	var found *client.TeamMember
	for i := range members {
		if members[i].Email == email {
			found = &members[i]
			break
		}
	}

	if found == nil {
		// If member not yet visible (async invite), use a synthetic ID.
		// The next Read will resolve the real ID once the member appears.
		resp.Diagnostics.AddWarning(
			"Member Not Yet Visible",
			fmt.Sprintf("Team member %q was invited but not yet visible in member list. "+
				"The resource will use a placeholder ID until the next refresh.", email),
		)
		model.ID = types.StringValue("pending-" + email)
	} else {
		model.ID = types.StringValue(strconv.Itoa(found.ID))
		model.Role = types.StringValue(found.Role)
	}

	tflog.Debug(ctx, "Invited team member", map[string]interface{}{
		"email": email,
		"id":    model.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the team member state from the API.
func (r *TeamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model TeamMemberResourceModel

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

	tflog.Debug(ctx, "Reading team member", map[string]interface{}{
		"id": model.ID.ValueString(),
	})

	members, err := r.client.Teams.ListMembers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Team Members",
			fmt.Sprintf("Could not list team members: %s", err),
		)
		return
	}

	// Try to find member by ID first, then by email
	var found *client.TeamMember
	idStr := model.ID.ValueString()

	// If we have a numeric ID, search by ID
	if id, err := strconv.Atoi(idStr); err == nil {
		for i := range members {
			if members[i].ID == id {
				found = &members[i]
				break
			}
		}
	}

	// If not found by ID (or pending ID), try by email
	if found == nil && !model.Email.IsNull() {
		email := model.Email.ValueString()
		for i := range members {
			if members[i].Email == email {
				found = &members[i]
				break
			}
		}
	}

	if found == nil {
		tflog.Warn(ctx, "Team member not found, removing from state", map[string]interface{}{
			"id": idStr,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	model.ID = types.StringValue(strconv.Itoa(found.ID))
	model.Email = types.StringValue(found.Email)
	model.Role = types.StringValue(found.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is not supported -- email and role both have RequiresReplace.
func (r *TeamMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Team member resources cannot be updated in place. Changing email or role requires replacement.",
	)
}

// Delete removes a team member.
func (r *TeamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model TeamMemberResourceModel

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
			"Error Parsing Team Member ID",
			fmt.Sprintf("Could not parse team member ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Removing team member", map[string]interface{}{
		"id": id,
	})

	if err := r.client.Teams.RemoveMember(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Team Member",
			fmt.Sprintf("Could not remove team member %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Removed team member", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing team member by their numeric ID.
// Usage: terraform import vastai_team_member.example <id>.
func (r *TeamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
