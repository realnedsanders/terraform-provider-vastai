package subaccount

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure SubaccountResource satisfies the required interfaces.
var (
	_ resource.Resource                = &SubaccountResource{}
	_ resource.ResourceWithConfigure   = &SubaccountResource{}
	_ resource.ResourceWithImportState = &SubaccountResource{}
)

// SubaccountResource defines the resource implementation.
type SubaccountResource struct {
	client *client.VastAIClient
}

// NewSubaccountResource returns a new subaccount resource instance.
func NewSubaccountResource() resource.Resource {
	return &SubaccountResource{}
}

// Metadata returns the resource type name.
func (r *SubaccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount"
}

// Schema defines the schema for the resource.
func (r *SubaccountResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai subaccount. NOTE: Subaccounts cannot be deleted via the API. Destroying this resource removes it from Terraform state only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique numeric identifier of the subaccount.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address for the subaccount. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username for the subaccount. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				Description: "Password for the subaccount. Write-only, never returned by API. Changing this forces a new resource.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host_only": schema.BoolAttribute{
				Description: "Whether this is a host-only subaccount. Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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
func (r *SubaccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new subaccount.
func (r *SubaccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model SubaccountResourceModel

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
	username := model.Username.ValueString()
	password := model.Password.ValueString()
	hostOnly := model.HostOnly.ValueBool()

	tflog.Debug(ctx, "Creating subaccount", map[string]interface{}{
		"email":    email,
		"username": username,
	})

	sub, err := r.client.Subaccounts.Create(ctx, email, username, password, hostOnly)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Subaccount",
			fmt.Sprintf("Could not create subaccount: %s", err),
		)
		return
	}

	// Map response to model
	model.ID = types.StringValue(strconv.Itoa(sub.ID))

	tflog.Debug(ctx, "Created subaccount", map[string]interface{}{
		"id": sub.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the subaccount state from the API.
func (r *SubaccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model SubaccountResourceModel

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
			"Error Parsing Subaccount ID",
			fmt.Sprintf("Could not parse subaccount ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading subaccount", map[string]interface{}{
		"id": id,
	})

	// List all subaccounts and find the one matching our ID
	subs, err := r.client.Subaccounts.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Subaccount",
			fmt.Sprintf("Could not list subaccounts: %s", err),
		)
		return
	}

	var found *client.Subaccount
	for i := range subs {
		if subs[i].ID == id {
			found = &subs[i]
			break
		}
	}

	if found == nil {
		tflog.Warn(ctx, "Subaccount not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response (password is write-only, preserve from state)
	model.Email = types.StringValue(found.Email)
	model.Username = types.StringValue(found.Username)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update returns an error because subaccounts are immutable after creation.
func (r *SubaccountResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Subaccounts Are Immutable",
		"Subaccounts cannot be updated after creation. Modify requires replacement.",
	)
}

// Delete is a no-op for subaccounts because the Vast.ai API does not support deleting them.
// The resource is removed from Terraform state with a warning.
func (r *SubaccountResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Subaccount Not Deleted",
		"The subaccount was removed from Terraform state but was NOT deleted from Vast.ai. "+
			"Subaccount deletion is not supported by the Vast.ai API.",
	)
}

// ImportState imports an existing subaccount by its numeric ID.
// Note: The password cannot be recovered after creation.
// Usage: terraform import vastai_subaccount.example <id>.
func (r *SubaccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
