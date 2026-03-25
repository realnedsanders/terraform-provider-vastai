package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

// Create creates a new instance from a GPU offer.
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Implemented in Task 2
	_ = ctx
	_ = req
	_ = resp
}

// Read refreshes the instance state from the API.
func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Implemented in Task 2
	_ = ctx
	_ = req
	_ = resp
}

// Update modifies an existing instance.
func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Implemented in Task 2
	_ = ctx
	_ = req
	_ = resp
}

// Delete destroys an instance.
func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Implemented in Task 2
	_ = ctx
	_ = req
	_ = resp
}

// ImportState imports an existing instance by its contract ID.
func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// defaultCreateTimeout is the default timeout for instance creation.
// GPU provisioning can take several minutes, so 10 minutes is a safe default per D-10.
var defaultCreateTimeout = 10 * time.Minute

// defaultOperationTimeout is the default timeout for read, update, and delete operations.
var defaultOperationTimeout = 5 * time.Minute
