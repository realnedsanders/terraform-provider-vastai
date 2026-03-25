package instance

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &InstanceResource{}
	_ resource.ResourceWithConfigure   = &InstanceResource{}
	_ resource.ResourceWithImportState = &InstanceResource{}
)

// InstanceResource defines the instance resource implementation.
type InstanceResource struct {
	client *client.VastAIClient
}

// NewInstanceResource creates a new instance resource.
func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

// Metadata returns the resource type name.
func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Schema defines the schema for the instance resource.
// Follows D-14 (snake_case), D-15 (Optional+Computed for server defaults),
// D-16 (plan-time validators), SCHM-01 through SCHM-06.
func (r *InstanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vast.ai GPU instance. Instances are created from GPU offers and support " +
			"full lifecycle management including start/stop, label and bid price updates, template changes, " +
			"SSH key attachment, and spot instance preemption handling.",

		Attributes: map[string]schema.Attribute{
			// ========== Computed, stable (UseStateForUnknown) ==========

			"id": schema.StringAttribute{
				Description: "Unique contract ID of the instance. Used as the primary identifier for all operations.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"machine_id": schema.Int64Attribute{
				Description: "Physical machine ID hosting this instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ssh_host": schema.StringAttribute{
				Description: "SSH connection hostname for this instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ssh_port": schema.Int64Attribute{
				Description: "SSH connection port for this instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"num_gpus": schema.Int64Attribute{
				Description: "Number of GPUs allocated to this instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"gpu_name": schema.StringAttribute{
				Description: "GPU model name (e.g., 'RTX 4090', 'A100 SXM4').",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the instance was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// ========== Required, immutable (RequiresReplace) ==========

			"offer_id": schema.Int64Attribute{
				Description: "ID of the GPU offer to create this instance from. Obtain from the vastai_gpu_offers " +
					"data source. Changing this forces a new resource.",
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"disk_gb": schema.Float64Attribute{
				Description: "Local disk partition size in GB. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Float64{
					float64validator.AtLeast(1.0),
				},
			},

			// ========== Optional+Computed (server-set defaults per D-15) ==========

			"image": schema.StringAttribute{
				Description: "Docker image to launch (e.g., 'pytorch/pytorch:latest'). Can be set via template. " +
					"If not specified, the template's image is used.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"status": schema.StringAttribute{
				Description: "Desired instance state. Set to 'running' or 'stopped' to start/stop the instance " +
					"without destroying and recreating it. Defaults to 'running' on creation.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("running", "stopped"),
				},
			},
			"use_ssh": schema.BoolAttribute{
				Description: "Enable SSH access to this instance. Defaults to the template setting if not specified.",
				Optional:    true,
				Computed:    true,
			},
			"use_jupyter_lab": schema.BoolAttribute{
				Description: "Enable JupyterLab access to this instance. Defaults to the template setting if not specified.",
				Optional:    true,
				Computed:    true,
			},

			// ========== Optional, mutable ==========

			"label": schema.StringAttribute{
				Description: "Human-readable label for the instance. Can be updated in-place.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(200),
				},
			},
			"bid_price": schema.Float64Attribute{
				Description: "Bid price per GPU per hour in USD. Set to create an interruptible (spot) instance. " +
					"Omit for on-demand pricing. Can be updated in-place to change bid amount.",
				Optional: true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0.001),
				},
			},
			"onstart": schema.StringAttribute{
				Description: "Startup script executed when the instance boots. Can be updated via template update.",
				Optional:    true,
			},
			"env": schema.MapAttribute{
				Description: "Environment variables as key-value pairs passed to the instance container. " +
					"Can be updated via template update.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"template_hash_id": schema.StringAttribute{
				Description: "Hash ID of a Vast.ai template to apply to this instance. Can be updated in-place.",
				Optional:    true,
			},
			"ssh_key_ids": schema.SetAttribute{
				Description: "Set of SSH key IDs to attach to this instance. Keys are attached/detached " +
					"incrementally when the set changes. Use IDs from vastai_ssh_key resources.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"image_login": schema.StringAttribute{
				Description: "Docker registry credentials for private image pulls (format: '-u user -p pass registry'). " +
					"This value is sensitive and will not be displayed in plan output.",
				Optional:  true,
				Sensitive: true,
			},
			"cancel_unavail": schema.BoolAttribute{
				Description: "Cancel instance creation if the selected offer becomes unavailable.",
				Optional:    true,
			},

			// ========== Computed, dynamic (no UseStateForUnknown) ==========

			"actual_status": schema.StringAttribute{
				Description: "Current actual status of the instance as reported by the API (e.g., 'running', 'loading', 'exited').",
				Computed:    true,
			},
			"cost_per_hour": schema.Float64Attribute{
				Description: "Current cost per hour in USD for this instance.",
				Computed:    true,
			},
			"gpu_ram_gb": schema.Float64Attribute{
				Description: "Total GPU VRAM in GB.",
				Computed:    true,
			},
			"cpu_ram_gb": schema.Float64Attribute{
				Description: "Total CPU RAM in GB.",
				Computed:    true,
			},
			"cpu_cores": schema.Float64Attribute{
				Description: "Number of effective CPU cores allocated to this instance.",
				Computed:    true,
			},
			"inet_up_mbps": schema.Float64Attribute{
				Description: "Internet upload speed in Mbps.",
				Computed:    true,
			},
			"inet_down_mbps": schema.Float64Attribute{
				Description: "Internet download speed in Mbps.",
				Computed:    true,
			},
			"reliability": schema.Float64Attribute{
				Description: "Host machine reliability score (0.0 to 1.0).",
				Computed:    true,
			},
			"geolocation": schema.StringAttribute{
				Description: "Geographic location of the host machine.",
				Computed:    true,
			},
			"is_bid": schema.BoolAttribute{
				Description: "Whether this instance uses bid (spot/interruptible) pricing.",
				Computed:    true,
			},
			"status_msg": schema.StringAttribute{
				Description: "Current status message from the instance (e.g., progress information during loading).",
				Computed:    true,
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
func (r *InstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// defaultCreateTimeout is the default timeout for instance creation.
// GPU provisioning can take several minutes, so 10 minutes is a safe default per D-10.
var defaultCreateTimeout = 10 * time.Minute

// defaultOperationTimeout is the default timeout for read, update, and delete operations.
var defaultOperationTimeout = 5 * time.Minute

// Create creates a new instance from a GPU offer.
// Sends PUT /asks/{offerID}/ and polls until the instance reaches "running" status per COMP-05.
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model InstanceResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout per SCHM-06
	createTimeout, diags := model.Timeouts.Create(ctx, defaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	offerID := int(model.OfferID.ValueInt64())

	// Build create request
	createReq := &client.CreateInstanceRequest{
		ClientID: "me",
		Disk:     model.DiskGB.ValueFloat64(),
	}

	// Image
	if !model.Image.IsNull() && !model.Image.IsUnknown() {
		createReq.Image = model.Image.ValueString()
	}

	// Label
	if !model.Label.IsNull() && !model.Label.IsUnknown() {
		createReq.Label = model.Label.ValueString()
	}

	// Bid price: nil = on-demand, &value = spot
	if !model.BidPrice.IsNull() && !model.BidPrice.IsUnknown() {
		price := model.BidPrice.ValueFloat64()
		createReq.Price = &price
	}

	// Onstart script
	if !model.Onstart.IsNull() && !model.Onstart.IsUnknown() {
		createReq.Onstart = model.Onstart.ValueString()
	}

	// Template hash ID
	if !model.TemplateHashID.IsNull() && !model.TemplateHashID.IsUnknown() {
		createReq.TemplateHashID = model.TemplateHashID.ValueString()
	}

	// Image login credentials
	if !model.ImageLogin.IsNull() && !model.ImageLogin.IsUnknown() {
		createReq.ImageLogin = model.ImageLogin.ValueString()
	}

	// Cancel if unavailable
	if !model.CancelUnavail.IsNull() && !model.CancelUnavail.IsUnknown() {
		createReq.CancelUnavail = model.CancelUnavail.ValueBool()
	}

	// JupyterLab
	if !model.UseJupyterLab.IsNull() && !model.UseJupyterLab.IsUnknown() {
		createReq.UseJupyterLab = model.UseJupyterLab.ValueBool()
	}

	// Build Runtype from use_ssh and use_jupyter_lab per Pitfall 7
	createReq.Runtype = buildRuntype(model.UseSSH, model.UseJupyterLab)

	// Environment variables: convert Terraform map to Go map
	if !model.Env.IsNull() && !model.Env.IsUnknown() {
		envMap := make(map[string]string)
		diags := model.Env.ElementsAs(ctx, &envMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Env = envMap
	}

	tflog.Debug(ctx, "Creating instance from offer", map[string]interface{}{
		"offer_id": offerID,
		"image":    createReq.Image,
		"disk_gb":  createReq.Disk,
	})

	// Call API to create instance
	createResp, err := r.client.Instances.Create(ctx, offerID, createReq)
	if err != nil {
		// Handle offer expiry per D-06
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 404 || apiErr.StatusCode == 400) {
			resp.Diagnostics.AddError(
				"Offer No Longer Available",
				fmt.Sprintf("Offer %d is no longer available. GPU offers are ephemeral and can be claimed by "+
					"other users. Run `terraform plan` again to search for current offers.", offerID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Creating Instance",
			fmt.Sprintf("Could not create instance from offer %d: %s", offerID, err),
		)
		return
	}

	contractID := createResp.NewContract

	tflog.Debug(ctx, "Instance creation initiated", map[string]interface{}{
		"contract_id": contractID,
		"offer_id":    offerID,
	})

	// Poll until running per COMP-05
	instance, err := r.client.Instances.WaitForStatus(ctx, contractID, "running", createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Instance",
			fmt.Sprintf("Instance %d was created but failed to reach 'running' status: %s", contractID, err),
		)
		return
	}

	tflog.Debug(ctx, "Instance is running", map[string]interface{}{
		"contract_id": contractID,
	})

	// Map API response to model
	mapInstanceToModel(instance, &model)

	// Attach SSH keys if specified per COMP-08
	if !model.SSHKeyIDs.IsNull() && !model.SSHKeyIDs.IsUnknown() {
		var keyIDs []string
		diags := model.SSHKeyIDs.ElementsAs(ctx, &keyIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if err := r.attachSSHKeys(ctx, contractID, keyIDs); err != nil {
			resp.Diagnostics.AddError(
				"Error Attaching SSH Keys",
				fmt.Sprintf("Instance %d was created but SSH key attachment failed: %s", contractID, err),
			)
			return
		}
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the instance state from the API.
// Handles preemption detection per D-09: spot instances that were preempted are
// silently removed from state, while normal state changes are reflected normally.
func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model InstanceResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	readTimeout, diags := model.Timeouts.Read(ctx, defaultOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Instance ID",
			fmt.Sprintf("Could not parse instance ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Reading instance", map[string]interface{}{
		"instance_id": id,
	})

	instance, err := r.client.Instances.Get(ctx, id)
	if err != nil {
		// Handle 404: instance was destroyed externally
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			tflog.Warn(ctx, "Instance not found, removing from state", map[string]interface{}{
				"instance_id": id,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Instance",
			fmt.Sprintf("Could not read instance %d: %s", id, err),
		)
		return
	}

	// Preemption detection per D-09
	// Only remove from state when instance was ACTUALLY preempted:
	// is_bid=true AND intended_status="running" AND actual_status indicates preemption
	if isPreempted(instance) {
		tflog.Warn(ctx, "Instance appears to have been preempted, removing from state", map[string]interface{}{
			"instance_id":     id,
			"is_bid":          instance.IsBid,
			"intended_status": instance.IntendedStatus,
			"actual_status":   instance.ActualStatus,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Normal state update
	mapInstanceToModel(instance, &model)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update modifies an existing instance.
// Supports status change (start/stop per COMP-02), label update, bid price change,
// template update, and SSH key attachment changes per COMP-03/COMP-08.
func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InstanceResourceModel
	var state InstanceResourceModel

	// Read plan and state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Instance ID",
			fmt.Sprintf("Could not parse instance ID %q as integer: %s", state.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Updating instance", map[string]interface{}{
		"instance_id": id,
	})

	// (a) Status change: start/stop per COMP-02
	if !plan.Status.Equal(state.Status) {
		newStatus := plan.Status.ValueString()
		tflog.Debug(ctx, "Changing instance status", map[string]interface{}{
			"instance_id": id,
			"old_status":  state.Status.ValueString(),
			"new_status":  newStatus,
		})

		if newStatus == "running" {
			if err := r.client.Instances.Start(ctx, id); err != nil {
				resp.Diagnostics.AddError(
					"Error Starting Instance",
					fmt.Sprintf("Could not start instance %d: %s", id, err),
				)
				return
			}
			if _, err := r.client.Instances.WaitForStatus(ctx, id, "running", updateTimeout); err != nil {
				resp.Diagnostics.AddError(
					"Error Waiting for Instance Start",
					fmt.Sprintf("Instance %d did not reach 'running' status: %s", id, err),
				)
				return
			}
		} else if newStatus == "stopped" {
			if err := r.client.Instances.Stop(ctx, id); err != nil {
				resp.Diagnostics.AddError(
					"Error Stopping Instance",
					fmt.Sprintf("Could not stop instance %d: %s", id, err),
				)
				return
			}
			if _, err := r.client.Instances.WaitForStatus(ctx, id, "stopped", updateTimeout); err != nil {
				resp.Diagnostics.AddError(
					"Error Waiting for Instance Stop",
					fmt.Sprintf("Instance %d did not reach 'stopped' status: %s", id, err),
				)
				return
			}
		}
	}

	// (b) Label changed per COMP-03
	if !plan.Label.Equal(state.Label) {
		newLabel := ""
		if !plan.Label.IsNull() && !plan.Label.IsUnknown() {
			newLabel = plan.Label.ValueString()
		}
		tflog.Debug(ctx, "Updating instance label", map[string]interface{}{
			"instance_id": id,
			"new_label":   newLabel,
		})
		if err := r.client.Instances.SetLabel(ctx, id, newLabel); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Instance Label",
				fmt.Sprintf("Could not set label on instance %d: %s", id, err),
			)
			return
		}
	}

	// (c) Bid price changed per COMP-03, D-11
	if !plan.BidPrice.Equal(state.BidPrice) {
		if !plan.BidPrice.IsNull() && !plan.BidPrice.IsUnknown() {
			newPrice := plan.BidPrice.ValueFloat64()
			tflog.Debug(ctx, "Changing instance bid price", map[string]interface{}{
				"instance_id": id,
				"new_price":   newPrice,
			})
			if err := r.client.Instances.ChangeBid(ctx, id, newPrice); err != nil {
				resp.Diagnostics.AddError(
					"Error Changing Bid Price",
					fmt.Sprintf("Could not change bid price on instance %d: %s", id, err),
				)
				return
			}
		}
	}

	// (d) Template/image/env/onstart changed per COMP-03
	templateChanged := !plan.Image.Equal(state.Image) ||
		!plan.Onstart.Equal(state.Onstart) ||
		!plan.Env.Equal(state.Env) ||
		!plan.TemplateHashID.Equal(state.TemplateHashID)

	if templateChanged {
		updateReq := &client.UpdateTemplateRequest{
			ID: id,
		}

		if !plan.Image.IsNull() && !plan.Image.IsUnknown() {
			updateReq.Image = plan.Image.ValueString()
		}
		if !plan.Onstart.IsNull() && !plan.Onstart.IsUnknown() {
			updateReq.Onstart = plan.Onstart.ValueString()
		}
		if !plan.TemplateHashID.IsNull() && !plan.TemplateHashID.IsUnknown() {
			updateReq.TemplateHashID = plan.TemplateHashID.ValueString()
		}
		if !plan.Env.IsNull() && !plan.Env.IsUnknown() {
			envMap := make(map[string]string)
			diags := plan.Env.ElementsAs(ctx, &envMap, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			updateReq.Env = envMap
		}

		tflog.Debug(ctx, "Updating instance template", map[string]interface{}{
			"instance_id": id,
		})

		if err := r.client.Instances.UpdateTemplate(ctx, id, updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Instance Template",
				fmt.Sprintf("Could not update template on instance %d: %s", id, err),
			)
			return
		}
	}

	// (e) SSH key IDs changed per COMP-08
	if !plan.SSHKeyIDs.Equal(state.SSHKeyIDs) {
		if err := r.reconcileSSHKeys(ctx, id, state.SSHKeyIDs, plan.SSHKeyIDs); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating SSH Keys",
				fmt.Sprintf("Could not update SSH keys on instance %d: %s", id, err),
			)
			return
		}
	}

	// Re-read instance state from API
	instance, err := r.client.Instances.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Instance After Update",
			fmt.Sprintf("Could not read instance %d after update: %s", id, err),
		)
		return
	}

	// Map to model and preserve plan values for user-settable attributes
	mapInstanceToModel(instance, &plan)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete destroys an instance.
func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model InstanceResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure timeout
	deleteTimeout, diags := model.Timeouts.Delete(ctx, defaultOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Instance ID",
			fmt.Sprintf("Could not parse instance ID %q as integer: %s", model.ID.ValueString(), err),
		)
		return
	}

	tflog.Debug(ctx, "Destroying instance", map[string]interface{}{
		"instance_id": id,
	})

	if err := r.client.Instances.Destroy(ctx, id); err != nil {
		// If already gone (404), that's fine
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Instance already destroyed", map[string]interface{}{
				"instance_id": id,
			})
			return
		}
		resp.Diagnostics.AddError(
			"Error Destroying Instance",
			fmt.Sprintf("Could not destroy instance %d: %s", id, err),
		)
		return
	}

	// Optionally wait for the instance to be fully destroyed
	_, _ = r.client.Instances.WaitForStatus(ctx, id, "destroyed", deleteTimeout)

	tflog.Debug(ctx, "Instance destroyed", map[string]interface{}{
		"instance_id": id,
	})
}

// ImportState imports an existing instance by its contract ID.
// Usage: terraform import vastai_instance.example <contract_id>
func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// isPreempted determines whether an instance was preempted (outbid/evicted).
// Per D-09: only spot instances (is_bid=true) that intended to be running
// but were forced to stop are considered preempted.
func isPreempted(instance *client.Instance) bool {
	if !instance.IsBid {
		return false
	}
	if instance.IntendedStatus != "running" {
		return false
	}
	// Preemption indicators: the instance was supposed to be running
	// but is actually in a stopped/offline state, indicating it was
	// outbid or evicted by the host.
	preemptionStatuses := map[string]bool{
		"stopped": true,
		"offline": true,
	}
	return preemptionStatuses[instance.ActualStatus]
}

// mapInstanceToModel converts an API Instance struct to a Terraform InstanceResourceModel.
// RAM values are divided by 1000 for GB conversion per Pitfall 6 in research.
func mapInstanceToModel(instance *client.Instance, model *InstanceResourceModel) {
	model.ID = types.StringValue(strconv.Itoa(instance.ID))
	model.MachineID = types.Int64Value(int64(instance.MachineID))
	model.ActualStatus = types.StringValue(instance.ActualStatus)
	model.Status = types.StringValue(instance.IntendedStatus)
	model.NumGPUs = types.Int64Value(int64(instance.NumGPUs))
	model.GPUName = types.StringValue(instance.GPUName)
	model.DPHTotal = types.Float64Value(instance.DPHTotal)
	model.IsBid = types.BoolValue(instance.IsBid)

	// SSH connection info
	if instance.SSHHost != "" {
		model.SSHHost = types.StringValue(instance.SSHHost)
	} else {
		model.SSHHost = types.StringNull()
	}
	if instance.SSHPort != 0 {
		model.SSHPort = types.Int64Value(int64(instance.SSHPort))
	} else {
		model.SSHPort = types.Int64Null()
	}

	// RAM values: API returns in MB, convert to GB per Pitfall 6
	model.GPURamGB = types.Float64Value(instance.GPUTotalRAM / 1000.0)
	model.CPURamGB = types.Float64Value(instance.CPURAM / 1000.0)
	model.CPUCores = types.Float64Value(instance.CPUCoresEffective)

	// Network
	model.InetUp = types.Float64Value(instance.InetUp)
	model.InetDown = types.Float64Value(instance.InetDown)

	// Reliability
	model.Reliability = types.Float64Value(instance.Reliability2)

	// Location
	if instance.Geolocation != "" {
		model.Geolocation = types.StringValue(instance.Geolocation)
	} else {
		model.Geolocation = types.StringNull()
	}

	// Status message
	if instance.StatusMsg != "" {
		model.StatusMsg = types.StringValue(instance.StatusMsg)
	} else {
		model.StatusMsg = types.StringNull()
	}

	// Image (from the instance's image UUID)
	if instance.ImageUUID != "" {
		model.Image = types.StringValue(instance.ImageUUID)
	}

	// Onstart
	if instance.Onstart != "" {
		model.Onstart = types.StringValue(instance.Onstart)
	} else if model.Onstart.IsNull() || model.Onstart.IsUnknown() {
		model.Onstart = types.StringNull()
	}

	// Template hash ID
	if instance.TemplateHashID != "" {
		model.TemplateHashID = types.StringValue(instance.TemplateHashID)
	} else if model.TemplateHashID.IsNull() || model.TemplateHashID.IsUnknown() {
		model.TemplateHashID = types.StringNull()
	}

	// Label
	if instance.Label != "" {
		model.Label = types.StringValue(instance.Label)
	} else if model.Label.IsNull() || model.Label.IsUnknown() {
		model.Label = types.StringNull()
	}

	// Created at: convert start_date (Unix timestamp float) to RFC3339
	if instance.StartDate > 0 {
		t := time.Unix(int64(instance.StartDate), 0)
		model.CreatedAt = types.StringValue(t.UTC().Format(time.RFC3339))
	} else if model.CreatedAt.IsNull() || model.CreatedAt.IsUnknown() {
		model.CreatedAt = types.StringNull()
	}

	// Use SSH / Use JupyterLab: infer from runtype/onstart if available
	// These are set on create and reflected back; the API doesn't expose them directly,
	// so we preserve the current model values unless we can infer from the instance.
	// The is_bid status is already set above.
}

// buildRuntype constructs the runtype string from use_ssh and use_jupyter_lab flags.
// Per Pitfall 7 from research: runtype is a space-separated string of enabled features.
func buildRuntype(useSSH, useJupyterLab types.Bool) string {
	var parts []string

	if !useSSH.IsNull() && !useSSH.IsUnknown() && useSSH.ValueBool() {
		parts = append(parts, "ssh_direc", "ssh_proxy")
	}

	if !useJupyterLab.IsNull() && !useJupyterLab.IsUnknown() && useJupyterLab.ValueBool() {
		parts = append(parts, "jupyter_direc")
	}

	return strings.Join(parts, " ")
}

// attachSSHKeys attaches SSH keys to an instance by resolving key IDs to public keys.
// Per COMP-08: The attach endpoint requires the full SSH key content, not just the ID.
func (r *InstanceResource) attachSSHKeys(ctx context.Context, instanceID int, keyIDs []string) error {
	// List all SSH keys to resolve IDs to content
	allKeys, err := r.client.SSHKeys.List(ctx)
	if err != nil {
		return fmt.Errorf("listing SSH keys to resolve IDs: %w", err)
	}

	keyMap := make(map[string]string) // ID -> public key content
	for _, k := range allKeys {
		keyMap[strconv.Itoa(k.ID)] = k.SSHKey
	}

	for _, keyID := range keyIDs {
		publicKey, ok := keyMap[keyID]
		if !ok {
			return fmt.Errorf("SSH key %s not found in account", keyID)
		}
		if err := r.client.SSHKeys.AttachToInstance(ctx, instanceID, publicKey); err != nil {
			return fmt.Errorf("attaching SSH key %s: %w", keyID, err)
		}
		tflog.Debug(ctx, "Attached SSH key to instance", map[string]interface{}{
			"instance_id": instanceID,
			"ssh_key_id":  keyID,
		})
	}
	return nil
}

// reconcileSSHKeys computes the diff between old and new SSH key sets and
// attaches/detaches keys accordingly per COMP-08.
func (r *InstanceResource) reconcileSSHKeys(ctx context.Context, instanceID int, oldSet, newSet types.Set) error {
	oldIDs := extractStringSet(oldSet)
	newIDs := extractStringSet(newSet)

	// Compute added and removed
	added := setDifference(newIDs, oldIDs)
	removed := setDifference(oldIDs, newIDs)

	// Attach new keys
	if len(added) > 0 {
		if err := r.attachSSHKeys(ctx, instanceID, added); err != nil {
			return err
		}
	}

	// Detach removed keys
	for _, keyIDStr := range removed {
		keyID, err := strconv.Atoi(keyIDStr)
		if err != nil {
			return fmt.Errorf("parsing SSH key ID %q: %w", keyIDStr, err)
		}
		if err := r.client.SSHKeys.DetachFromInstance(ctx, instanceID, keyID); err != nil {
			return fmt.Errorf("detaching SSH key %d: %w", keyID, err)
		}
		tflog.Debug(ctx, "Detached SSH key from instance", map[string]interface{}{
			"instance_id": instanceID,
			"ssh_key_id":  keyID,
		})
	}

	return nil
}

// extractStringSet extracts string values from a types.Set.
func extractStringSet(s types.Set) []string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	var result []string
	for _, elem := range s.Elements() {
		if sv, ok := elem.(types.String); ok {
			result = append(result, sv.ValueString())
		}
	}
	return result
}

// setDifference returns elements in a that are not in b.
func setDifference(a, b []string) []string {
	bSet := make(map[string]bool, len(b))
	for _, v := range b {
		bSet[v] = true
	}
	var diff []string
	for _, v := range a {
		if !bSet[v] {
			diff = append(diff, v)
		}
	}
	return diff
}

// stringSetValue creates a types.Set of strings from a slice of strings.
// This is a helper for constructing Set values in tests and model mapping.
func stringSetValue(vals []string) types.Set {
	if vals == nil {
		return types.SetNull(types.StringType)
	}
	elems := make([]attr.Value, len(vals))
	for i, v := range vals {
		elems[i] = types.StringValue(v)
	}
	s, _ := types.SetValue(types.StringType, elems)
	return s
}
