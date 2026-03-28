package sshkey

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure SSHKeyResource satisfies the required interfaces.
var (
	_ resource.Resource                = &SSHKeyResource{}
	_ resource.ResourceWithConfigure   = &SSHKeyResource{}
	_ resource.ResourceWithImportState = &SSHKeyResource{}
)

// SSHKeyResource defines the resource implementation.
type SSHKeyResource struct {
	client *client.VastAIClient
}

// NewSSHKeyResource returns a new SSH key resource instance.
func NewSSHKeyResource() resource.Resource {
	return &SSHKeyResource{}
}

// Metadata returns the resource type name.
func (r *SSHKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

// Schema defines the schema for the resource.
func (r *SSHKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an SSH key in Vast.ai. SSH keys can be attached to instances for secure access.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique numeric identifier of the SSH key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ssh_key": schema.StringAttribute{
				Description: "SSH public key content (e.g., 'ssh-rsa AAAA...'). Sensitive value.",
				Required:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^ssh-(rsa|ed25519|ecdsa|dsa) `),
						"must be a valid SSH public key starting with ssh-rsa, ssh-ed25519, ssh-ecdsa, or ssh-dsa",
					),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the SSH key was created.",
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

// Configure adds the provider-configured client to the resource.
func (r *SSHKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new SSH key.
func (r *SSHKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model SSHKeyResourceModel

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

	tflog.Debug(ctx, "Creating SSH key")

	sshKey, err := r.client.SSHKeys.Create(ctx, model.SSHKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating SSH Key",
			fmt.Sprintf("Could not create SSH key: %s", err),
		)
		return
	}

	// Map response to model
	model.ID = types.StringValue(strconv.Itoa(sshKey.ID))
	model.CreatedAt = types.StringValue(sshKey.CreatedAt)

	tflog.Debug(ctx, "Created SSH key", map[string]interface{}{
		"id": sshKey.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the SSH key state from the API.
func (r *SSHKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model SSHKeyResourceModel

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
			"Error Parsing SSH Key ID",
			fmt.Sprintf("Could not parse SSH key ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading SSH key", map[string]interface{}{
		"id": id,
	})

	// List all keys and find the one matching our ID (no single-get endpoint)
	keys, err := r.client.SSHKeys.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SSH Key",
			fmt.Sprintf("Could not list SSH keys: %s", err),
		)
		return
	}

	var found *client.SSHKey
	for i := range keys {
		if keys[i].ID == id {
			found = &keys[i]
			break
		}
	}

	if found == nil {
		tflog.Warn(ctx, "SSH key not found, removing from state", map[string]interface{}{
			"id": id,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from API response
	model.SSHKey = types.StringValue(found.SSHKey)
	model.CreatedAt = types.StringValue(found.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update updates an existing SSH key.
func (r *SSHKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model SSHKeyResourceModel

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

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing SSH Key ID",
			fmt.Sprintf("Could not parse SSH key ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Updating SSH key", map[string]interface{}{
		"id": id,
	})

	sshKey, err := r.client.SSHKeys.Update(ctx, id, model.SSHKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating SSH Key",
			fmt.Sprintf("Could not update SSH key %d: %s", id, err),
		)
		return
	}

	// Update model from response
	model.CreatedAt = types.StringValue(sshKey.CreatedAt)

	tflog.Debug(ctx, "Updated SSH key", map[string]interface{}{
		"id": id,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete deletes an SSH key.
func (r *SSHKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model SSHKeyResourceModel

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
			"Error Parsing SSH Key ID",
			fmt.Sprintf("Could not parse SSH key ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting SSH key", map[string]interface{}{
		"id": id,
	})

	if err := r.client.SSHKeys.Delete(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SSH Key",
			fmt.Sprintf("Could not delete SSH key %d: %s", id, err),
		)
		return
	}

	tflog.Debug(ctx, "Deleted SSH key", map[string]interface{}{
		"id": id,
	})
}

// ImportState imports an existing SSH key by its numeric ID.
// Usage: terraform import vastai_ssh_key.example <id>.
func (r *SSHKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
