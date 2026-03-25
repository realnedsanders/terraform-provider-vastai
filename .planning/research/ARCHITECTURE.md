# Architecture Patterns

**Domain:** Terraform/OpenTofu provider for Vast.ai GPU compute
**Researched:** 2026-03-25

## Recommended Architecture

A Terraform provider is a standalone Go binary that communicates with Terraform CLI over gRPC using the Plugin Protocol (version 6 for Plugin Framework). The provider binary receives CRUD requests from Terraform, translates them into REST API calls against the Vast.ai API, and maps responses back into Terraform state.

The architecture follows a three-layer design: **Provider Shell** (Terraform integration) -> **Service Layer** (resource/data source implementations) -> **API Client** (HTTP REST communication with Vast.ai).

```
+-------------------+
|   Terraform CLI   |  (gRPC Plugin Protocol v6)
+--------+----------+
         |
+--------v----------+
|  Provider Server   |  main.go - entry point, providerserver.Serve()
+--------+----------+
         |
+--------v----------+
|  Provider Config   |  internal/provider/provider.go
|  - Schema          |  - Reads VASTAI_API_KEY from config/env
|  - Configure()     |  - Creates API client
|  - Resources()     |  - Registers all resources
|  - DataSources()   |  - Registers all data sources
+--------+----------+
         |
    +----+----+
    |         |
+---v---+ +---v---+
|Resources| |Data   |
|        | |Sources |
| CRUD   | | Read   |  internal/services/<resource>/
| methods| | only   |  One directory per resource type
+---+---+ +---+---+
    |         |
+---v---------v---+
|   API Client     |  internal/client/
|  - HTTP wrapper  |  - Base client with auth, retry, errors
|  - Auth (API key)|  - Per-resource method sets
|  - Retry logic   |  - Request/response models
|  - Error mapping |
+--------+--------+
         |
+--------v--------+
|  Vast.ai REST    |
|  API v0          |
|  console.vast.ai |
+-----------------+
```

### Component Boundaries

| Component | Responsibility | Communicates With | Package |
|-----------|---------------|-------------------|---------|
| **main.go** | Binary entry point, starts gRPC server | Provider via providerserver.Serve() | `main` |
| **Provider** | Auth configuration, client creation, resource/datasource registration | All resources and data sources via Configure() | `internal/provider` |
| **Resources** | CRUD lifecycle for a single Vast.ai resource type; schema definition; state mapping; plan modification; import | Provider (receives client), API Client (makes calls) | `internal/services/<name>` |
| **Data Sources** | Read-only queries against Vast.ai API; schema definition; state mapping | Provider (receives client), API Client (makes calls) | `internal/services/<name>` |
| **API Client** | HTTP transport, authentication, retry, rate limiting, error normalization | Vast.ai REST API at `console.vast.ai/api/v0/` | `internal/client` |
| **Models** | Go structs for API request/response payloads | API Client (serialization), Resources (deserialization) | `internal/client` (or `internal/models`) |
| **Acceptance Tests** | End-to-end testing against real Vast.ai API | All components via Terraform test framework | `*_test.go` files alongside implementations |
| **Documentation** | Generated provider/resource/datasource docs for Terraform Registry | None (build artifact) | `docs/`, `templates/`, `examples/` |

### Data Flow

**Terraform Apply (Create Resource):**
```
1. User writes HCL config:     vastai_instance { ... }
2. Terraform CLI parses config, calls provider via gRPC
3. Provider's Configure() creates VastAI API client (with API key)
4. Provider dispatches to resource's Create() method
5. Create() reads plan data into Go model struct (via tfsdk tags)
6. Create() calls API client: client.Instances.Create(ctx, params)
7. API client builds HTTP request: PUT /api/v0/asks/{id}/?api_key=...
8. API client sends request, receives JSON response
9. API client deserializes response into Go struct
10. Create() maps API response into Terraform state model
11. Create() writes state via resp.State.Set(ctx, &model)
12. Terraform persists state to terraform.tfstate
```

**Terraform Plan (Read/Refresh):**
```
1. Terraform reads prior state from terraform.tfstate
2. Terraform calls resource's Read() method
3. Read() extracts ID from prior state
4. Read() calls API client: client.Instances.Get(ctx, id)
5. API client sends: GET /api/v0/instances/{id}/?api_key=...
6. Read() maps response to state model, writes to resp.State
7. Terraform diffs plan state vs. desired config state
```

**Terraform Import:**
```
1. User runs: terraform import vastai_instance.mine 12345
2. Terraform calls resource's ImportState() method
3. ImportState() sets the ID attribute from import identifier
4. Terraform calls Read() to populate remaining state
5. Normal Read() flow fills in all attributes from API
```

## Recommended Project Layout

```
terraform-provider-vastai/
|
|-- main.go                          # Entry point: providerserver.Serve()
|-- go.mod                           # Module: github.com/<org>/terraform-provider-vastai
|-- go.sum
|-- GNUmakefile                      # build, test, lint, generate targets
|-- .goreleaser.yml                  # Multi-platform build + signing
|-- .golangci.yml                    # Linter config
|-- .github/
|   |-- workflows/
|       |-- test.yml                 # CI: lint, unit test, acc test
|       |-- release.yml              # CD: goreleaser on tag push
|
|-- internal/
|   |-- provider/
|   |   |-- provider.go              # VastaiProvider struct, Schema, Configure, Resources, DataSources
|   |   |-- provider_test.go         # Provider config tests
|   |
|   |-- client/
|   |   |-- client.go                # VastAIClient struct, NewClient(), base HTTP methods
|   |   |-- client_test.go           # Client unit tests (with httptest)
|   |   |-- auth.go                  # API key authentication (query param injection)
|   |   |-- errors.go                # API error types, error mapping from HTTP status codes
|   |   |-- retry.go                 # Retry with backoff for transient failures
|   |   |-- instance.go              # InstanceService: Create, Get, List, Update, Delete, Start, Stop, Reboot
|   |   |-- template.go              # TemplateService: CRUD
|   |   |-- ssh_key.go               # SSHKeyService: CRUD + attach/detach
|   |   |-- volume.go                # VolumeService + NetworkVolumeService
|   |   |-- endpoint.go              # EndpointService: CRUD
|   |   |-- worker_group.go          # WorkerGroupService: CRUD
|   |   |-- cluster.go               # ClusterService + OverlayService
|   |   |-- api_key.go               # APIKeyService: CRUD
|   |   |-- team.go                  # TeamService + TeamRoleService + SubaccountService
|   |   |-- env_var.go               # EnvVarService: CRUD
|   |   |-- search.go                # Offer search, volume search, template search
|   |   |-- user.go                  # UserService: show profile, invoices, audit logs
|   |
|   |-- services/
|   |   |-- instance/
|   |   |   |-- resource_instance.go         # vastai_instance resource (CRUD + start/stop/reboot)
|   |   |   |-- resource_instance_test.go    # Acceptance tests
|   |   |   |-- data_source_instances.go     # vastai_instances data source (list)
|   |   |   |-- data_source_instances_test.go
|   |   |   |-- models.go                   # Terraform model structs (tfsdk tags)
|   |   |
|   |   |-- template/
|   |   |   |-- resource_template.go
|   |   |   |-- resource_template_test.go
|   |   |   |-- data_source_templates.go
|   |   |   |-- data_source_templates_test.go
|   |   |   |-- models.go
|   |   |
|   |   |-- ssh_key/
|   |   |   |-- resource_ssh_key.go
|   |   |   |-- resource_ssh_key_test.go
|   |   |   |-- models.go
|   |   |
|   |   |-- volume/
|   |   |   |-- resource_volume.go
|   |   |   |-- resource_network_volume.go
|   |   |   |-- data_source_volume_offers.go
|   |   |   |-- data_source_network_volume_offers.go
|   |   |   |-- models.go
|   |   |
|   |   |-- endpoint/
|   |   |   |-- resource_endpoint.go
|   |   |   |-- data_source_endpoint_status.go
|   |   |   |-- models.go
|   |   |
|   |   |-- worker_group/
|   |   |   |-- resource_worker_group.go
|   |   |   |-- models.go
|   |   |
|   |   |-- cluster/
|   |   |   |-- resource_cluster.go
|   |   |   |-- resource_overlay.go
|   |   |   |-- models.go
|   |   |
|   |   |-- api_key/
|   |   |   |-- resource_api_key.go
|   |   |   |-- models.go
|   |   |
|   |   |-- team/
|   |   |   |-- resource_team.go
|   |   |   |-- resource_team_role.go
|   |   |   |-- resource_subaccount.go
|   |   |   |-- models.go
|   |   |
|   |   |-- env_var/
|   |   |   |-- resource_env_var.go
|   |   |   |-- models.go
|   |   |
|   |   |-- offer/
|   |   |   |-- data_source_gpu_offers.go
|   |   |   |-- data_source_gpu_offers_test.go
|   |   |   |-- models.go
|   |   |
|   |   |-- machine/
|   |   |   |-- data_source_machines.go
|   |   |   |-- models.go
|   |   |
|   |   |-- user/
|   |       |-- data_source_user.go
|   |       |-- data_source_invoices.go
|   |       |-- data_source_audit_logs.go
|   |       |-- models.go
|   |
|   |-- acctest/
|   |   |-- helpers.go               # Shared acceptance test helpers, provider factories
|   |   |-- sweep.go                 # Test resource cleanup (sweepers)
|   |
|   |-- validators/
|   |   |-- validators.go            # Custom validators (API key format, etc.)
|   |
|   |-- planmodifiers/
|       |-- planmodifiers.go          # Custom plan modifiers if needed
|
|-- docs/                            # Generated by tfplugindocs (do not edit manually)
|   |-- index.md                     # Provider overview
|   |-- resources/
|   |   |-- instance.md
|   |   |-- template.md
|   |   |-- ...
|   |-- data-sources/
|       |-- gpu_offers.md
|       |-- ...
|
|-- templates/                       # Templates for tfplugindocs generation
|   |-- index.md.tmpl
|   |-- resources/
|   |   |-- instance.md.tmpl
|   |-- data-sources/
|       |-- gpu_offers.md.tmpl
|
|-- examples/
|   |-- provider/
|   |   |-- provider.tf              # Provider configuration example
|   |-- resources/
|   |   |-- vastai_instance/
|   |   |   |-- resource.tf
|   |   |-- vastai_template/
|   |       |-- resource.tf
|   |-- data-sources/
|       |-- vastai_gpu_offers/
|           |-- data-source.tf
|
|-- terraform-registry-manifest.json  # {"version": 1, "metadata": {"protocol_versions": ["6.0"]}}
```

### Why This Layout

**`internal/client/` separated from `internal/services/`**: The API client is a pure Go HTTP wrapper with zero Terraform dependencies. This enables independent unit testing with `httptest`, potential reuse outside the provider, and clean separation of API concerns from Terraform state management. This follows the "bindings" pattern recommended by both HashiCorp and the provider architecture literature.

**`internal/services/<name>/` service-per-directory**: Follows the pattern used by terraform-provider-aws (263 service directories), terraform-provider-cloudflare (242 service directories), and terraform-provider-digitalocean (36 service directories). Each service directory is self-contained with its resource, data source, tests, and Terraform model structs. This scales well -- when adding a new resource type, you create a new directory without touching existing code.

**`internal/provider/` as thin registration layer**: The provider package should be thin -- it defines the provider schema (API key config), creates the API client in Configure(), and returns the list of resource/data source constructors. No business logic lives here. This matches the scaffolding template pattern.

**`models.go` per service directory**: Each service directory has its own models.go containing Terraform state model structs (with `tfsdk` struct tags). These are separate from the API client models (which live in `internal/client/`) because Terraform models use `types.String`, `types.Int64`, etc. while API models use plain Go types. The mapping between them happens in the resource CRUD methods.

## Patterns to Follow

### Pattern 1: Provider Configure -> Client Injection

The provider's `Configure()` method is the single place where the API client is created. Resources receive it via their own `Configure()` method (implementing `resource.ResourceWithConfigure`).

**What:** Create API client once in provider, inject into all resources/data sources.
**When:** Always -- this is the standard pattern.
**Example:**
```go
// internal/provider/provider.go
func (p *VastaiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var config VastaiProviderModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // API key from config or environment variable
    apiKey := os.Getenv("VASTAI_API_KEY")
    if !config.APIKey.IsNull() {
        apiKey = config.APIKey.ValueString()
    }
    if apiKey == "" {
        resp.Diagnostics.AddError(
            "Missing API Key",
            "Set VASTAI_API_KEY environment variable or api_key in provider config",
        )
        return
    }

    client := client.NewVastAIClient(apiKey)
    resp.DataSourceData = client
    resp.ResourceData = client
}

// internal/services/instance/resource_instance.go
type InstanceResource struct {
    client *client.VastAIClient
}

func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    c, ok := req.ProviderData.(*client.VastAIClient)
    if !ok {
        resp.Diagnostics.AddError("Unexpected Provider Data",
            fmt.Sprintf("Expected *client.VastAIClient, got: %T", req.ProviderData))
        return
    }
    r.client = c
}
```

### Pattern 2: Constructor Function Registration

Each resource/data source exposes a `New*()` constructor function. The provider registers them by reference, not by calling them.

**What:** Factory functions for lazy resource initialization.
**When:** Always -- prevents shared state between resource instances.
**Example:**
```go
// internal/services/instance/resource_instance.go
func NewInstanceResource() resource.Resource {
    return &InstanceResource{}
}

// internal/provider/provider.go
func (p *VastaiProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        instance.NewInstanceResource,
        template.NewTemplateResource,
        ssh_key.NewSSHKeyResource,
        // ...
    }
}
```

### Pattern 3: Terraform Model <-> API Model Mapping

Maintain separate structs for Terraform state (using `types.*` from the framework) and API requests/responses (using plain Go types). Map between them explicitly in CRUD methods.

**What:** Two model layers -- Terraform models with `tfsdk` tags, API models with `json` tags.
**When:** Always -- framework types handle null/unknown states that plain Go types cannot.
**Example:**
```go
// internal/services/instance/models.go (Terraform model)
type InstanceModel struct {
    ID          types.String  `tfsdk:"id"`
    Image       types.String  `tfsdk:"image"`
    NumGPUs     types.Int64   `tfsdk:"num_gpus"`
    DiskGB      types.Float64 `tfsdk:"disk_gb"`
    Status      types.String  `tfsdk:"status"`
    SSHHost     types.String  `tfsdk:"ssh_host"`
    SSHPort     types.Int64   `tfsdk:"ssh_port"`
}

// internal/client/instance.go (API model)
type Instance struct {
    ID       int     `json:"id"`
    Image    string  `json:"image_uuid"`
    NumGPUs  int     `json:"num_gpus"`
    DiskGB   float64 `json:"disk_space"`
    Status   string  `json:"actual_status"`
    SSHHost  string  `json:"ssh_host"`
    SSHPort  int     `json:"ssh_port"`
}
```

### Pattern 4: Compile-Time Interface Verification

Use blank identifier assignments to verify interface satisfaction at compile time.

**What:** Static assertions that types implement required interfaces.
**When:** Every resource, data source, and the provider itself.
**Example:**
```go
var _ resource.Resource = &InstanceResource{}
var _ resource.ResourceWithConfigure = &InstanceResource{}
var _ resource.ResourceWithImportState = &InstanceResource{}
```

### Pattern 5: API Client with Service Objects

Structure the API client with service-scoped method receivers, similar to how the GitHub, DigitalOcean, and Stripe Go clients organize their API surface.

**What:** A root client struct with service sub-objects, each providing typed methods.
**When:** APIs with 10+ resource types -- avoids a monolithic client file.
**Example:**
```go
// internal/client/client.go
type VastAIClient struct {
    httpClient  *http.Client
    baseURL     string
    apiKey      string

    Instances      *InstanceService
    Templates      *TemplateService
    SSHKeys        *SSHKeyService
    Volumes        *VolumeService
    NetworkVolumes *NetworkVolumeService
    Endpoints      *EndpointService
    WorkerGroups   *WorkerGroupService
    Clusters       *ClusterService
    Overlays       *OverlayService
    APIKeys        *APIKeyService
    Teams          *TeamService
    TeamRoles      *TeamRoleService
    Subaccounts    *SubaccountService
    EnvVars        *EnvVarService
    Offers         *OfferService
    Users          *UserService
}

func NewVastAIClient(apiKey string) *VastAIClient {
    c := &VastAIClient{
        httpClient: &http.Client{Timeout: 30 * time.Second},
        baseURL:    "https://console.vast.ai/api/v0",
        apiKey:     apiKey,
    }
    c.Instances = &InstanceService{client: c}
    c.Templates = &TemplateService{client: c}
    // ... initialize all services
    return c
}

// internal/client/instance.go
type InstanceService struct {
    client *VastAIClient
}

func (s *InstanceService) Get(ctx context.Context, id int) (*Instance, error) {
    var instance Instance
    err := s.client.doRequest(ctx, "GET", fmt.Sprintf("/instances/%d/", id), nil, &instance)
    return &instance, err
}
```

### Pattern 6: Resource State Management for Instance Lifecycle

Vast.ai instances have lifecycle states (running, stopped, loading, etc.) that must be handled carefully. The Terraform resource should use `status` as a computed attribute and handle start/stop as part of the resource lifecycle rather than separate resources.

**What:** In-place lifecycle management via a `status` or `running` attribute.
**When:** Resources with running/stopped states (instances).
**Example:** A `running` bool attribute with `Default: true`. On Create, start the instance. On Update, if `running` changes, call start/stop. On Delete, destroy (which auto-stops). This follows the AWS provider pattern for EC2 instances where running state is managed in the main resource.

### Pattern 7: Acceptance Test Pattern

Use `terraform-plugin-testing` with `ProtoV6ProviderFactories` for Plugin Framework providers. Each resource needs at minimum a basic creation test and a "disappears" test (resource deleted externally).

**What:** Standard acceptance test structure with check functions.
**When:** Every resource and data source.
**Example:**
```go
func TestAccInstanceResource_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccInstanceConfig_basic(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("vastai_instance.test", "image", "pytorch/pytorch:latest"),
                    resource.TestCheckResourceAttrSet("vastai_instance.test", "id"),
                ),
            },
            // ImportState test
            {
                ResourceName:      "vastai_instance.test",
                ImportState:        true,
                ImportStateVerify:  true,
            },
        },
    })
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Monolithic Provider Package
**What:** Putting all resource, data source, and client code in `internal/provider/` as a single package.
**Why bad:** The scaffolding template starts this way for simplicity, but it does not scale. With 15+ resources and data sources (as in this project), a single package becomes unwieldy. Circular dependencies emerge when resources need to reference each other. Test files become massive.
**Instead:** Use the service-per-directory pattern from day one. The scaffolding template is a starting point, not a production architecture.

### Anti-Pattern 2: API Client Embedded in Resource Code
**What:** Making HTTP calls directly inside resource CRUD methods.
**Why bad:** Duplicates HTTP handling, authentication, retry, and error mapping across every resource. Makes unit testing resources impossible without mocking HTTP. Changes to API auth (e.g., header-based auth) require touching every resource file.
**Instead:** Separate API client in `internal/client/` with service objects. Resources call typed client methods, never `http.Do()` directly.

### Anti-Pattern 3: Using SDKv2 Instead of Plugin Framework
**What:** Building new resources with `github.com/hashicorp/terraform-plugin-sdk/v2`.
**Why bad:** SDKv2 is legacy. HashiCorp explicitly recommends Plugin Framework for all new providers. SDKv2 lacks protocol version 6 features, has weaker type safety, and will receive only maintenance updates.
**Instead:** Plugin Framework exclusively. No muxing with SDKv2 unless migrating an existing provider.

### Anti-Pattern 4: Storing API Key in State
**What:** Including the API key as a regular string attribute in provider state.
**Why bad:** Terraform state is often stored in shared backends (S3, GCS). API keys in state leak credentials.
**Instead:** Accept API key via `Sensitive: true` attribute and/or environment variable `VASTAI_API_KEY`. The Plugin Framework's `Sensitive` flag prevents the value from appearing in logs and plan output.

### Anti-Pattern 5: Ignoring Unknown Values in Configure
**What:** Failing to handle `types.String.IsUnknown()` in the provider's Configure method.
**Why bad:** When a practitioner interpolates a resource output into the provider config, the value is "unknown" during plan. If Configure treats unknown as empty and errors, `terraform plan` breaks.
**Instead:** Check `IsUnknown()` and return early with a warning, or use default/env values when the config value is unknown.

### Anti-Pattern 6: Flat API Client (One Giant File)
**What:** Putting all API methods in a single `client.go` file.
**Why bad:** Vast.ai has ~126 API operations across 25+ resource types. A single file would be 3000+ lines.
**Instead:** Service objects pattern -- one file per API resource area (instance.go, template.go, etc.), each defining a service struct with methods.

## Scalability Considerations

| Concern | At 5 resources | At 15 resources | At 25+ resources |
|---------|----------------|-----------------|------------------|
| **Code organization** | Could use flat internal/provider/ | Must use service-per-directory | Consider code generation for boilerplate |
| **Test execution** | All tests run in parallel | Separate test tags per service area | Sweeper infrastructure essential |
| **API client** | Single file acceptable | Service objects pattern needed | Per-service files with shared base client |
| **Documentation** | Manual templates fine | tfplugindocs required | Template-driven with examples per resource |
| **CI time** | < 5 min | 15-30 min (acc tests) | Parallel acc test jobs, test matrix |

This provider targets ~15 resources and ~10 data sources, placing it firmly in the "must use service-per-directory" tier from the start.

## Build Order (Dependencies Between Components)

The components have a strict dependency order that dictates what must be built first:

```
Phase 1: Foundation
  internal/client/client.go      (base HTTP, auth, errors, retry)
  internal/client/errors.go
  internal/client/auth.go
  internal/client/retry.go
  main.go                        (minimal entry point)
  internal/provider/provider.go  (skeleton with Configure, empty Resources/DataSources)

Phase 2: First API Service + First Resource
  internal/client/instance.go    (InstanceService with Get, Create, Delete)
  internal/services/instance/    (resource + models + test)
  Proves the full stack works end-to-end

Phase 3: Remaining API Services + Resources (parallelizable)
  Each additional resource follows identical pattern:
    1. Add API service methods to internal/client/<service>.go
    2. Create internal/services/<name>/ with resource, models, tests
    3. Register constructor in provider.go Resources()/DataSources()

Phase 4: Data Sources (parallelizable with Phase 3)
  Data sources depend on API client methods but are simpler (Read only)
  Can be built alongside or after their corresponding resources

Phase 5: Cross-Cutting Concerns
  Custom validators (internal/validators/)
  Custom plan modifiers (internal/planmodifiers/)
  Import support for all resources
  Sweep infrastructure for test cleanup

Phase 6: Release Infrastructure
  .goreleaser.yml
  .github/workflows/release.yml
  GPG signing setup
  terraform-registry-manifest.json
  Documentation templates + generation

Phase 7: Polish
  tfplugindocs generation
  examples/ directory with working configs
  Acceptance test coverage gaps
```

**Critical path:** Phase 1 -> Phase 2 -> Phase 3 (all other phases branch from Phase 2 completion). Getting the client + provider + one resource working end-to-end is the highest-risk, most-informative milestone.

## Key Technology Interfaces

### Provider Interface (must implement)
```go
provider.Provider {
    Metadata(context.Context, MetadataRequest, *MetadataResponse)
    Schema(context.Context, SchemaRequest, *SchemaResponse)
    Configure(context.Context, ConfigureRequest, *ConfigureResponse)
    Resources(context.Context) []func() resource.Resource
    DataSources(context.Context) []func() datasource.DataSource
}
```

### Resource Interface (must implement per resource)
```go
resource.Resource {
    Metadata(context.Context, MetadataRequest, *MetadataResponse)
    Schema(context.Context, SchemaRequest, *SchemaResponse)
    Create(context.Context, CreateRequest, *CreateResponse)
    Read(context.Context, ReadRequest, *ReadResponse)
    Update(context.Context, UpdateRequest, *UpdateResponse)
    Delete(context.Context, DeleteRequest, *DeleteResponse)
}

// Optional interfaces:
resource.ResourceWithConfigure     // Receive API client from provider
resource.ResourceWithImportState   // Support terraform import
resource.ResourceWithModifyPlan    // Custom plan modification
resource.ResourceWithValidateConfig // Custom validation
```

### Data Source Interface (must implement per data source)
```go
datasource.DataSource {
    Metadata(context.Context, MetadataRequest, *MetadataResponse)
    Schema(context.Context, SchemaRequest, *SchemaResponse)
    Read(context.Context, ReadRequest, *ReadResponse)
}

// Optional:
datasource.DataSourceWithConfigure // Receive API client from provider
```

## Vast.ai-Specific Architecture Decisions

### Authentication: Query Parameter, Not Header
The Vast.ai API authenticates via `?api_key=<key>` query parameter, not an Authorization header. The API client must append this to every request URL. This is unusual but straightforward -- implement in the base `doRequest()` method.

### Instance Lifecycle is Two-Step
Creating a Vast.ai instance is a two-phase operation: (1) select an offer (machine), (2) create an instance from that offer. This maps well to Terraform because the offer selection is a data source (`vastai_gpu_offers`) and the instance creation is a resource (`vastai_instance`) that references the offer ID. The resource's `Create()` method calls the API to rent the selected offer.

### No Partial Update Support
Many Vast.ai resources appear to lack PATCH semantics -- updates are full PUTs. The resource's `Update()` method should send the complete desired state, not a diff. This simplifies implementation (just send everything) but means the schema must carefully track which attributes are settable vs. computed.

### Eventual Consistency on Instance Operations
Instance start/stop/reboot operations are asynchronous. The resource may need polling in `Create()` and `Update()` to wait for the instance to reach the desired state. Use `retry.RetryContext()` from the standard library or a custom poller with exponential backoff. This is a common pattern in cloud providers (AWS EC2 uses the same approach).

## Sources

- [Terraform Plugin Framework - Official Documentation](https://developer.hashicorp.com/terraform/plugin/framework) (HIGH confidence)
- [Terraform Plugin Framework - GitHub Repository](https://github.com/hashicorp/terraform-plugin-framework) (HIGH confidence)
- [HashiCorp Provider Design Principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles) (HIGH confidence)
- [Terraform Provider Scaffolding Framework Template](https://github.com/hashicorp/terraform-provider-scaffolding-framework) (HIGH confidence)
- [Plugin Framework Tutorial: Implement a Provider](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider) (HIGH confidence)
- [Plugin Framework Tutorial: Configure Provider Client](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider-configure) (HIGH confidence)
- [Plugin Framework Tutorial: Resource Create and Read](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-resource-create) (HIGH confidence)
- [Plugin Framework Resources Documentation](https://developer.hashicorp.com/terraform/plugin/framework/resources) (HIGH confidence)
- [Plugin Framework Data Sources Documentation](https://developer.hashicorp.com/terraform/plugin/framework/data-sources) (HIGH confidence)
- [Plugin Framework Resource Import](https://developer.hashicorp.com/terraform/plugin/framework/resources/import) (HIGH confidence)
- [Terraform Registry Publishing Requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing) (HIGH confidence)
- [terraform-provider-aws - Internal Structure](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal) (HIGH confidence)
- [terraform-provider-cloudflare - Internal Structure](https://github.com/cloudflare/terraform-provider-cloudflare/tree/main/internal/services) (HIGH confidence)
- [terraform-provider-digitalocean - Droplet Package](https://github.com/digitalocean/terraform-provider-digitalocean/tree/main/digitalocean/droplet) (HIGH confidence)
- [AWS Provider Contributor Guide: Add a New Resource](https://hashicorp.github.io/terraform-provider-aws/add-a-new-resource/) (HIGH confidence)
- [AWS Provider Contributor Guide: Provider Design](https://hashicorp.github.io/terraform-provider-aws/provider-design/) (HIGH confidence)
- [terraform-plugin-framework-validators](https://github.com/hashicorp/terraform-plugin-framework-validators) (HIGH confidence)
- [HashiCorp GitHub Action for Provider Release](https://github.com/hashicorp/ghaction-terraform-provider-release) (HIGH confidence)
- [Terraform Provider Architecture: 11 Components](https://shadow-soft.com/terraform-provider-architecture/) (MEDIUM confidence)
- [Plugin Framework Plan Modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification) (HIGH confidence)
- [Plugin Framework Validation](https://developer.hashicorp.com/terraform/plugin/framework/validation) (HIGH confidence)
- [Plugin Framework Custom Types](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom) (HIGH confidence)
