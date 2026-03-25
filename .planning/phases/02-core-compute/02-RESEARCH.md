# Phase 2: Core Compute - Research

**Researched:** 2026-03-25
**Domain:** Vast.ai compute API contracts, Terraform Plugin Framework resource/data-source patterns, instance lifecycle management
**Confidence:** HIGH

## Summary

Phase 2 implements the complete compute workflow: GPU offer search, instance lifecycle (create/start/stop/update/destroy), template management, SSH key management, and establishes the schema quality patterns (validators, plan modifiers, import) reused by all subsequent phases. This is the most technically complex phase of the project.

The Vast.ai API has been thoroughly extracted from the Python SDK at `vast-sdk/vastai/vast.py`. The API uses a marketplace model where instances are created from ephemeral offer IDs, creation is asynchronous (returns a contract ID, not a full instance), and lifecycle management uses PUT with a `state` field. The API has four status tracking fields (`actual_status`, `intended_status`, `cur_state`, `next_state`) and approximately 90 fields per instance.

**Primary recommendation:** Build the API client service layer first (offers, instances, templates, SSH keys), then the Terraform resources/data-sources on top. Use httptest mocks for all unit tests. Establish the schema pattern (validators, plan modifiers, import) on the instance resource first and replicate it to all subsequent resources.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Claude's Discretion on filter approach -- structured typed attributes for common filters (gpu_name, num_gpus, gpu_ram, price, region, datacenter_only) with optional raw_query escape hatch for power users
- **D-02:** Return all matching offers as a list, with a computed `most_affordable` convenience attribute for the cheapest one
- **D-03:** Configurable `order_by` attribute -- user picks sort field (price, score, gpu_ram, etc.)
- **D-04:** Default limit of 10 results, configurable via `limit` attribute
- **D-05:** Standard Terraform behavior -- data source re-queries live on every plan/apply (no caching)
- **D-06:** Claude's Discretion on offer expiry handling -- error with guidance vs auto-retry based on Terraform's declarative model
- **D-07:** Full offer details exposed -- GPU name, VRAM, CPU cores, RAM, disk, DL performance, reliability score, location, hosting type
- **D-08:** Claude's Discretion on start/stop modeling (status attribute vs separate resource) based on cloud provider patterns
- **D-09:** Preemption handling: only silent removal from state when instance was ACTUALLY preempted (outbid/evicted). Successful exits on spot instances should NOT trigger silent removal -- they should surface as a normal state change. Must distinguish preemption from normal termination via API status.
- **D-10:** Claude's Discretion on create timeout default based on GPU provisioning times
- **D-11:** Claude's Discretion on bid price updates (in-place vs replace) based on what the Vast.ai API supports
- **D-12:** Claude's Discretion on SSH key attachment model (inline vs separate resource vs both) based on Terraform patterns
- **D-13:** Claude's Discretion on immutable vs updatable attributes based on Vast.ai API behavior
- **D-14:** Always snake_case for all Terraform attributes regardless of API naming. Mapping handled in model conversion.
- **D-15:** Server-set defaults use Optional+Computed classification -- user can set them, otherwise server default stored in state. Prevents noisy diffs.
- **D-16:** Comprehensive plan-time validators -- validate GPU names, regions, image format, port ranges, etc. Catch errors before API call.
- **D-17:** Per-resource model structs (InstanceModel, TemplateModel) -- no coupling between resources
- **D-18:** API response/request types live in `internal/client/`, Terraform state models live in service directories -- clean separation per standalone client decision (D-05 from Phase 1)
- **D-19:** Use SetAttribute where ordering doesn't matter (SSH keys, tags), ListAttribute where it does (env vars, ports). Prevents spurious diffs.
- **D-20:** Include Vast.ai API documentation URLs in attribute descriptions where helpful
- **D-21:** Dual strategy -- httptest mocks for schema/logic validation (fast, free) + real API tests with cheapest available offers for integration testing
- **D-22:** Claude's Discretion on test parallelism based on Vast.ai API rate limit behavior
- **D-23:** Full acceptance test coverage for ALL resources in Phase 2 -- every resource/data source gets create, read, update, import, destroy tests

### Claude's Discretion
- Offer filter approach details (structured + raw escape hatch balance)
- Offer expiry error handling strategy
- Start/stop modeling (status attribute vs separate resource)
- Create timeout default
- Bid price update behavior (in-place vs replace)
- SSH key attachment model
- Immutable vs updatable attribute classification
- Test parallelism strategy

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| COMP-01 | `vastai_instance` resource with full CRUD (create from offer ID, read, update, destroy) | API contracts extracted: PUT /asks/{id}/ for create, GET /instances/{id}/ for read, PUT /instances/update_template/{id}/ for update, DELETE /instances/{id}/ for destroy |
| COMP-02 | `vastai_instance` supports start/stop via `status` attribute without destroy/recreate | API uses PUT /instances/{id}/ with `{"state": "running"}` or `{"state": "stopped"}` -- supports in-place status toggle |
| COMP-03 | `vastai_instance` supports label, bid price change, and template update | Label: PUT /instances/{id}/ with `{"label": "..."}`. Bid: PUT /instances/bid_price/{id}/ with `{"price": ...}`. Template: PUT /instances/update_template/{id}/ |
| COMP-04 | `vastai_instance` handles spot/interruptible preemption gracefully | Instance has `actual_status`, `intended_status`, `cur_state`, `next_state` fields. Interruptible instances can be "paused if outbid" per docs. Read must check for preemption vs normal stop. |
| COMP-05 | `vastai_instance` creation polls until running state (async create with waiter) | Create returns `{"success": true, "new_contract": 7835610}`. Must poll GET /instances/{id}/ until `actual_status` == "running" |
| COMP-06 | `vastai_template` resource with full CRUD | Create: POST /template/ with image, env, runtype, etc. Update: PUT /template/ with hash_id. Delete: DELETE /template/ with hash_id or template_id. Search: GET /template/ |
| COMP-07 | `vastai_ssh_key` resource with full CRUD | Create: POST /ssh/ with `{"ssh_key": "..."}`. List: GET /ssh/. Update: PUT /ssh/{id}/ with `{"ssh_key": "..."}`. Delete: DELETE /ssh/{id}/ |
| COMP-08 | `vastai_ssh_key` supports attach/detach to instances | Attach: POST /instances/{id}/ssh/ with `{"ssh_key": "..."}`. Detach: DELETE /instances/{id}/ssh/{ssh_key_id}/ |
| DATA-01 | `vastai_gpu_offers` data source with structured filter attributes | POST /bundles/ with query JSON. Supports filter operators (eq, gt, lt, gte, lte, in, notin). Returns `{"offers": [...]}` |
| DATA-02 | `vastai_instance` data source (singular by ID) | GET /instances/{id}/?owner=me returns `{"instances": {...}}` (single instance object) |
| DATA-03 | `vastai_instances` data source (list with optional filtering) | GET /instances?owner=me returns `{"instances": [...]}` |
| DATA-04 | `vastai_templates` data source (search by query) | GET /template/?select_cols=[*]&select_filters={query} returns `{"templates": [...]}` |
| DATA-08 | `vastai_ssh_keys` data source (list all keys) | GET /ssh/ returns list of SSH keys |
| SCHM-01 | Attribute validators on all constrained fields | Use terraform-plugin-framework-validators: stringvalidator, int64validator, float64validator. Custom validators for GPU names (via /gpu_names/unique/ API) |
| SCHM-02 | Sensitive flag on all secret attributes | Mark ssh_key values, bid_price (optional), image_login credentials |
| SCHM-03 | Correct Required/Optional/Computed classification | Detailed attribute classification tables provided below |
| SCHM-04 | Meaningful description on every attribute | Reference Vast.ai docs URLs where available per D-20 |
| SCHM-05 | Plan modifiers: UseStateForUnknown for stable computed fields, RequiresReplace for immutable fields | Immutable: offer_id/machine_id on instance. Stable computed: id, ssh_host, machine_id |
| SCHM-06 | Configurable timeouts per resource via terraform-plugin-framework-timeouts | Use timeouts.Attributes() with create (10min), read (5min), update (5min), delete (5min) defaults |
| IMPT-01 | terraform import support for all managed resources via resource ID | Instance by contract ID, template by hash_id or ID, SSH key by ID |
| IMPT-02 | Import documentation with example commands | Generate via tfplugindocs templates in Phase 6, but ImportState must be implemented now |
| TEST-01 | Acceptance tests for all resources (create, read, update, import, destroy) | Dual strategy: httptest mocks (fast) + real API tests (gated behind TF_ACC) |
| TEST-02 | Unit tests for validators, plan modifiers, and API client logic | Pure Go tests with httptest mock server for client, standard unit tests for validators/modifiers |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

- **Language:** Go (required by Terraform provider ecosystem)
- **Framework:** Terraform Plugin Framework (HashiCorp's current recommendation)
- **API Reference:** Python SDK at `vast-sdk/` is the source of truth for API behavior
- **Auth:** Bearer header only (never query parameters) -- established in Phase 1
- **Registry:** Must follow HashiCorp's publishing requirements
- **GSD Workflow:** All changes through GSD commands

## Standard Stack

### Core (already installed from Phase 1)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| terraform-plugin-framework | v1.19.0 | Provider SDK | Already in go.mod |
| hashicorp/go-retryablehttp | v0.7.8 | HTTP client with retry | Already in go.mod |
| terraform-plugin-log (tflog) | v0.10.0 | Structured logging | Already indirect dep |

### New Dependencies for Phase 2
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| terraform-plugin-framework-validators | v0.19.0 | Pre-built validators (string, int, float, set, list) | SCHM-01: All constrained attributes |
| terraform-plugin-framework-timeouts | v0.5.0 | Configurable resource timeouts | SCHM-06: Instance create/delete waiters |
| terraform-plugin-testing | v1.15.0 | Acceptance test framework | TEST-01, TEST-02: All acceptance tests |

**Installation:**
```bash
go get github.com/hashicorp/terraform-plugin-framework-validators@v0.19.0
go get github.com/hashicorp/terraform-plugin-framework-timeouts@v0.5.0
go get github.com/hashicorp/terraform-plugin-testing@v1.15.0
```

## Architecture Patterns

### Recommended Project Structure (Phase 2 additions)
```
internal/
  client/
    client.go           # (exists) Base HTTP client
    auth.go             # (exists) Request construction
    errors.go           # (exists) APIError type
    instances.go        # NEW: InstanceService CRUD + start/stop/bid
    offers.go           # NEW: OfferService search
    templates.go        # NEW: TemplateService CRUD
    ssh_keys.go         # NEW: SSHKeyService CRUD + attach/detach
  provider/
    provider.go         # (exists) Updated to register new resources/datasources
  services/
    instance/
      resource_instance.go       # vastai_instance resource
      resource_instance_test.go  # Acceptance + unit tests
      data_source_instance.go    # vastai_instance data source (singular)
      data_source_instances.go   # vastai_instances data source (plural)
      models.go                  # InstanceModel, InstanceDataSourceModel
    offer/
      data_source_gpu_offers.go       # vastai_gpu_offers data source
      data_source_gpu_offers_test.go
      models.go                       # OfferModel, GpuOffersDataSourceModel
    template/
      resource_template.go       # vastai_template resource
      resource_template_test.go
      data_source_templates.go   # vastai_templates data source
      models.go                  # TemplateModel
    sshkey/
      resource_ssh_key.go        # vastai_ssh_key resource
      resource_ssh_key_test.go
      data_source_ssh_keys.go    # vastai_ssh_keys data source
      models.go                  # SSHKeyModel
  acctest/
    helpers.go                   # ProtoV6ProviderFactories, test config helpers
```

### Pattern 1: API Client Service Layer

Each resource domain gets a service file in `internal/client/` with typed request/response structs and methods.

**What:** Service objects on VastAIClient with domain-specific methods.
**When:** Every API domain (instances, offers, templates, ssh_keys).
**Example:**
```go
// internal/client/instances.go

type InstanceService struct {
    client *VastAIClient
}

// CreateInstanceRequest is the JSON body for PUT /asks/{offer_id}/
type CreateInstanceRequest struct {
    ClientID       string            `json:"client_id"`
    Image          string            `json:"image"`
    Env            map[string]string `json:"env,omitempty"`
    Price          *float64          `json:"price,omitempty"`       // nil = on-demand
    Disk           float64           `json:"disk"`
    Label          string            `json:"label,omitempty"`
    Onstart        string            `json:"onstart,omitempty"`
    Runtype        string            `json:"runtype,omitempty"`     // ssh, jupyter, args
    TemplateHashID string            `json:"template_hash_id,omitempty"`
    ImageLogin     string            `json:"image_login,omitempty"`
    CancelUnavail  bool              `json:"cancel_unavail,omitempty"`
    // ... additional fields
}

// CreateInstanceResponse is the JSON response from instance creation.
type CreateInstanceResponse struct {
    Success     bool `json:"success"`
    NewContract int  `json:"new_contract"`
}

// Instance represents the full instance object from GET /instances/{id}/
type Instance struct {
    ID               int               `json:"id"`
    MachineID        int               `json:"machine_id"`
    ActualStatus     string            `json:"actual_status"`
    IntendedStatus   string            `json:"intended_status"`
    CurState         string            `json:"cur_state"`
    NextState        string            `json:"next_state"`
    NumGPUs          int               `json:"num_gpus"`
    GPUName          string            `json:"gpu_name"`
    GPUUtil          float64           `json:"gpu_util"`
    GPURAM           float64           `json:"gpu_ram"`
    GPUTotalRAM      float64           `json:"gpu_totalram"`
    CPUCoresEffective float64          `json:"cpu_cores_effective"`
    CPURAM           float64           `json:"cpu_ram"`
    DiskSpace        float64           `json:"disk_space"`
    SSHHost          string            `json:"ssh_host"`
    SSHPort          int               `json:"ssh_port"`
    DPHTotal         float64           `json:"dph_total"`
    ImageUUID        string            `json:"image_uuid"`
    Label            string            `json:"label"`
    InetUp           float64           `json:"inet_up"`
    InetDown         float64           `json:"inet_down"`
    Reliability      float64           `json:"reliability2"`
    StartDate        float64           `json:"start_date"`
    IsBid            bool              `json:"is_bid"`
    MinBid           float64           `json:"min_bid"`
    Geolocation      string            `json:"geolocation"`
    HostingType      int               `json:"hosting_type"`
    TemplateID       int               `json:"template_id"`
    TemplateHashID   string            `json:"template_hash_id"`
    StatusMsg        string            `json:"status_msg"`
    ExtraEnv         [][]string        `json:"extra_env"`
    Onstart          string            `json:"onstart"`
    Verification     string            `json:"verification"`
    DirectPortCount  int               `json:"direct_port_count"`
    // ... many more fields (~90 total)
}

func (s *InstanceService) Create(ctx context.Context, offerID int, req *CreateInstanceRequest) (*CreateInstanceResponse, error) {
    var resp CreateInstanceResponse
    err := s.client.Put(ctx, fmt.Sprintf("/asks/%d/", offerID), req, &resp)
    return &resp, err
}

func (s *InstanceService) Get(ctx context.Context, id int) (*Instance, error) {
    var wrapper struct {
        Instances Instance `json:"instances"`
    }
    err := s.client.Get(ctx, fmt.Sprintf("/instances/%d/?owner=me", id), &wrapper)
    return &wrapper.Instances, err
}

func (s *InstanceService) List(ctx context.Context) ([]Instance, error) {
    var wrapper struct {
        Instances []Instance `json:"instances"`
    }
    err := s.client.Get(ctx, "/instances?owner=me", &wrapper)
    return wrapper.Instances, err
}

func (s *InstanceService) Start(ctx context.Context, id int) error {
    var resp struct {
        Success bool   `json:"success"`
        Msg     string `json:"msg"`
    }
    return s.client.Put(ctx, fmt.Sprintf("/instances/%d/", id), map[string]string{"state": "running"}, &resp)
}

func (s *InstanceService) Stop(ctx context.Context, id int) error {
    var resp struct {
        Success bool   `json:"success"`
        Msg     string `json:"msg"`
    }
    return s.client.Put(ctx, fmt.Sprintf("/instances/%d/", id), map[string]string{"state": "stopped"}, &resp)
}

func (s *InstanceService) Destroy(ctx context.Context, id int) error {
    var resp struct {
        Success bool   `json:"success"`
        Msg     string `json:"msg"`
    }
    return s.client.Delete(ctx, fmt.Sprintf("/instances/%d/", id), &resp)
}

func (s *InstanceService) SetLabel(ctx context.Context, id int, label string) error {
    var resp struct{ Success bool `json:"success"` }
    return s.client.Put(ctx, fmt.Sprintf("/instances/%d/", id), map[string]string{"label": label}, &resp)
}

func (s *InstanceService) ChangeBid(ctx context.Context, id int, price float64) error {
    var resp struct{ Success bool `json:"success"` }
    body := map[string]interface{}{"client_id": "me", "price": price}
    return s.client.Put(ctx, fmt.Sprintf("/instances/bid_price/%d/", id), body, &resp)
}

func (s *InstanceService) UpdateTemplate(ctx context.Context, id int, req *UpdateTemplateRequest) error {
    var resp struct {
        Success         bool        `json:"success"`
        UpdatedInstance interface{} `json:"updated_instance"`
    }
    return s.client.Put(ctx, fmt.Sprintf("/instances/update_template/%d/", id), req, &resp)
}
```

### Pattern 2: Waiter/Poller for Async Operations

**What:** Poll instance status after create until target state reached or timeout.
**When:** Instance Create (wait for "running"), Instance Destroy (wait for gone/404).
**Recommendation for D-10 (create timeout):** Default 10 minutes. GPU provisioning includes image download, container startup, and network setup. Vast.ai docs mention instances can be stuck in "scheduling" state. 10 minutes covers large image pulls on slow connections.

```go
// internal/client/instances.go

func (s *InstanceService) WaitForStatus(ctx context.Context, id int, targetStatus string, timeout time.Duration) (*Instance, error) {
    deadline := time.Now().Add(timeout)
    pollInterval := 5 * time.Second

    for {
        if time.Now().After(deadline) {
            return nil, fmt.Errorf("timed out waiting for instance %d to reach status %q", id, targetStatus)
        }

        instance, err := s.Get(ctx, id)
        if err != nil {
            // 404 means destroyed -- if we're waiting for destruction, that's success
            var apiErr *APIError
            if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
                if targetStatus == "destroyed" {
                    return nil, nil
                }
            }
            return nil, fmt.Errorf("polling instance %d: %w", id, err)
        }

        if instance.ActualStatus == targetStatus {
            return instance, nil
        }

        // Check for terminal error states
        if instance.ActualStatus == "exited" && targetStatus == "running" {
            return instance, fmt.Errorf("instance %d exited unexpectedly (status_msg: %s)", id, instance.StatusMsg)
        }

        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(pollInterval):
        }
    }
}
```

### Pattern 3: Start/Stop via Status Attribute (Recommendation for D-08)

**Recommendation:** Use a `status` attribute (not a separate resource). This follows the AWS EC2 pattern where `instance_state` is managed in the main resource. Reasons:
1. The Vast.ai API uses the same endpoint (PUT /instances/{id}/) for both status changes and other updates
2. A separate resource would create awkward cross-resource dependencies
3. Users expect `vastai_instance` to control whether their instance is running

```go
// In resource_instance.go Update method:
// If status changed from "running" to "stopped" -> call Stop
// If status changed from "stopped" to "running" -> call Start
// These are in-place updates, no destroy/recreate needed
```

### Pattern 4: Offer Expiry Handling (Recommendation for D-06)

**Recommendation:** Error with guidance message. Rationale:
- Terraform's declarative model expects deterministic plan-to-apply
- Auto-retry would hide the fact that the infrastructure spec changed
- Users need to re-plan to get fresh offers
- Error message: "Offer {id} is no longer available. Run `terraform plan` again to search for current offers."

### Pattern 5: Bid Price Updates (Recommendation for D-11)

**Recommendation:** In-place update. The Vast.ai API supports `PUT /instances/bid_price/{id}/` with a new price, confirming bid price can be changed without recreating the instance. Mark `bid_price` as Optional+Computed (no RequiresReplace needed).

### Pattern 6: SSH Key Attachment Model (Recommendation for D-12)

**Recommendation:** Inline `ssh_key_ids` attribute on the instance resource (SetAttribute) for the common case, with independent `vastai_ssh_key` resource for key CRUD. The attach/detach API endpoints (POST/DELETE /instances/{id}/ssh/) support modifying keys on existing instances. This means:
- `vastai_ssh_key` resource manages key CRUD (create/read/update/delete)
- `vastai_instance` has an optional `ssh_key_ids` set attribute
- Changes to `ssh_key_ids` trigger attach/detach in the Update method

Do NOT create a separate `vastai_ssh_key_attachment` resource -- it adds complexity without benefit since the attach/detach operations are simple and naturally part of instance management.

### Pattern 7: Immutable vs Updatable (Recommendation for D-13)

Based on API analysis:

| Attribute | Mutable? | How | Plan Modifier |
|-----------|----------|-----|---------------|
| offer_id | Immutable | Only set at creation | RequiresReplace |
| machine_id | Immutable | Assigned by API at creation | UseStateForUnknown |
| image | Mutable | Via update_template endpoint | None |
| disk | Immutable | Set at creation, cannot resize | RequiresReplace |
| num_gpus | Immutable | Determined by offer | UseStateForUnknown |
| gpu_name | Immutable | Determined by offer | UseStateForUnknown |
| label | Mutable | PUT /instances/{id}/ with label | None |
| bid_price | Mutable | PUT /instances/bid_price/{id}/ | None |
| template_hash_id | Mutable | Via update_template endpoint | None |
| onstart | Mutable | Via update_template endpoint | None |
| env | Mutable | Via update_template endpoint | None |
| status (running/stopped) | Mutable | PUT with state field | None |
| ssh_host | Computed, stable | Assigned by API | UseStateForUnknown |
| ssh_port | Computed, stable | Assigned by API | UseStateForUnknown |
| id | Computed, stable | Returned on creation | UseStateForUnknown |

## Vast.ai API Contract Reference

### Extracted API Endpoints

#### Offer Search
| Operation | Method | Path | Request | Response |
|-----------|--------|------|---------|----------|
| Search offers | POST | /bundles/ | JSON query object (see below) | `{"offers": [...]}`  |
| Search offers (new) | PUT | /search/asks/ | `{"select_cols": ["*"], "q": {query}}` | `{"offers": [...]}` |
| GPU names list | GET | /gpu_names/unique/ | None (no auth required) | `{"gpu_names": ["RTX 3090", ...]}` |

**Offer search query format:**
```json
{
    "verified": {"eq": true},
    "external": {"eq": false},
    "rentable": {"eq": true},
    "rented": {"eq": false},
    "gpu_name": {"eq": "RTX_4090"},
    "num_gpus": {"gte": 1},
    "gpu_ram": {"gte": 24000},
    "dph_total": {"lte": 1.0},
    "order": [["dph_total", "asc"]],
    "type": "on-demand",
    "limit": 10,
    "allocated_storage": 50
}
```

**Filter operators:** `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `in`, `notin`

**Offer type values:** `"on-demand"`, `"bid"` (interruptible), `"reserved"`

**Available offer filter fields (from offers_fields):**
bw_nvlink, compute_cap, cpu_arch, cpu_cores, cpu_cores_effective, cpu_ghz, cpu_ram, cuda_max_good, datacenter, direct_port_count, driver_version, disk_bw, disk_space, dlperf, dlperf_per_dphtotal, dph_total, duration, external, flops_per_dphtotal, gpu_arch, gpu_display_active, gpu_frac, gpu_mem_bw, gpu_name, gpu_ram, gpu_total_ram, gpu_max_power, gpu_max_temp, has_avx, host_id, id, inet_down, inet_down_cost, inet_up, inet_up_cost, machine_id, min_bid, mobo_name, num_gpus, pci_gen, pcie_bw, reliability, rentable, rented, storage_cost, static_ip, total_flops, ubuntu_version, verification, verified, vms_enabled, geolocation, cluster_id

**Field aliases:** cuda_vers -> cuda_max_good, display_active -> gpu_display_active, dlperf_usd -> dlperf_per_dphtotal, dph -> dph_total, flops_usd -> flops_per_dphtotal

**Field multipliers (values in API are x1000 for RAM fields):**
cpu_ram: x1000 (MB in API, user specifies GB), gpu_ram: x1000, gpu_total_ram: x1000, duration: x86400 (seconds in API, user specifies days)

#### Instance CRUD
| Operation | Method | Path | Request Body | Response |
|-----------|--------|------|-------------|----------|
| Create | PUT | /asks/{offer_id}/ | CreateInstanceRequest JSON | `{"success": true, "new_contract": 7835610}` |
| Read (single) | GET | /instances/{id}/?owner=me | None | `{"instances": {instance_object}}` |
| Read (list) | GET | /instances?owner=me | None | `{"instances": [instance_objects]}` |
| Start | PUT | /instances/{id}/ | `{"state": "running"}` | `{"success": true}` |
| Stop | PUT | /instances/{id}/ | `{"state": "stopped"}` | `{"success": true}` |
| Set label | PUT | /instances/{id}/ | `{"label": "my-label"}` | `{"success": true}` |
| Change bid | PUT | /instances/bid_price/{id}/ | `{"client_id": "me", "price": 0.50}` | Success |
| Update template | PUT | /instances/update_template/{id}/ | `{"id": id, "template_id": ..., "image": ..., "env": ..., "onstart": ...}` | `{"success": true, "updated_instance": {...}}` |
| Destroy | DELETE | /instances/{id}/ | `{}` | `{"success": true}` |
| Attach SSH | POST | /instances/{id}/ssh/ | `{"ssh_key": "ssh-rsa AAAA..."}` | JSON response |
| Detach SSH | DELETE | /instances/{id}/ssh/{ssh_key_id}/ | None | JSON response |

**CRITICAL: Create instance request fields:**
```json
{
    "client_id": "me",
    "image": "pytorch/pytorch:latest",
    "env": {"-p 8080:8080": "1", "MODEL": "llama2"},
    "price": 0.50,
    "disk": 50.0,
    "label": "my-training-run",
    "onstart": "#!/bin/bash\necho hello",
    "runtype": "ssh_direc ssh_proxy",
    "template_hash_id": "abc123",
    "image_login": "-u user -p pass docker.io",
    "python_utf8": false,
    "lang_utf8": false,
    "use_jupyter_lab": false,
    "jupyter_dir": null,
    "force": false,
    "cancel_unavail": false,
    "user": null
}
```

**Instance `actual_status` values (observed from SDK and docs):**
- `"running"` -- Instance is running normally
- `"stopped"` -- Manually stopped, data preserved
- `"exited"` -- Container exited (crash or normal completion)
- `"loading"` -- Docker image being downloaded
- `"creating"` -- Vast initiating instance creation
- `"scheduling"` -- Waiting for GPU availability on restart
- `"connecting"` -- Docker running, connection unverified
- `"offline"` -- Machine disconnected from Vast servers
- `null` -- Instance not yet started or state unknown

**Preemption detection strategy (D-09):**
Interruptible instances are "paused if outbid or if on-demand requested." The API does not have a dedicated "preempted" status. Detection must use:
1. Check `is_bid` field (true = interruptible instance)
2. Check `actual_status` vs `intended_status` mismatch: if `intended_status` is "running" but `actual_status` is "stopped" or "offline" for a bid instance, it was likely preempted
3. Check `status_msg` for preemption indicators
4. In Read: if instance returns 404, always remove from state (regardless of type)
5. If `is_bid=true` AND `actual_status` changed to stopped/offline while `intended_status` was "running" -> this is preemption -> remove from state silently
6. If `is_bid=false` (on-demand) or `intended_status` was already "stopped" -> normal state change -> do NOT remove from state

#### Template CRUD
| Operation | Method | Path | Request Body | Response |
|-----------|--------|------|-------------|----------|
| Create | POST | /template/ | Template JSON (see below) | `{"success": true, "template": {...}}` |
| Update | PUT | /template/ | Template JSON with hash_id | `{"success": true, "template": {...}}` |
| Delete | DELETE | /template/ | `{"hash_id": "..."}` or `{"template_id": N}` | `{"msg": "..."}` |
| Search | GET | /template/?select_cols=[*]&select_filters={query} | None (filters in query params) | `{"templates": [...]}` |

**Template create/update fields:**
```json
{
    "name": "my-template",
    "image": "pytorch/pytorch:latest",
    "tag": "latest",
    "href": "https://example.com",
    "repo": "https://github.com/...",
    "env": "-e KEY=VALUE -p 8080:8080",
    "onstart": "#!/bin/bash\necho hello",
    "jup_direct": false,
    "ssh_direct": true,
    "use_jupyter_lab": false,
    "runtype": "ssh_direc ssh_proxy",
    "use_ssh": true,
    "jupyter_dir": null,
    "docker_login_repo": "docker.io/myrepo",
    "extra_filters": {},
    "recommended_disk_space": "50",
    "readme": "# My Template",
    "readme_visible": true,
    "desc": "Description text",
    "private": false
}
```

**Template search fields:** creator_id, created_at, count_created, default_tag, docker_login_repo, id, image, jup_direct, hash_id, private, name, recent_create_date, recommended_disk_space, recommended, ssh_direct, tag, use_ssh

#### SSH Key CRUD
| Operation | Method | Path | Request Body | Response |
|-----------|--------|------|-------------|----------|
| Create | POST | /ssh/ | `{"ssh_key": "ssh-rsa AAAA..."}` | JSON with key details |
| List | GET | /ssh/ | None | JSON array of keys |
| Update | PUT | /ssh/{id}/ | `{"id": N, "ssh_key": "ssh-rsa NEW..."}` | JSON response |
| Delete | DELETE | /ssh/{id}/ | None | JSON response |
| Attach to instance | POST | /instances/{id}/ssh/ | `{"ssh_key": "ssh-rsa AAAA..."}` | JSON response |
| Detach from instance | DELETE | /instances/{id}/ssh/{ssh_key_id}/ | None | JSON response |

**NOTE on attach:** The attach endpoint takes the full SSH key content, not a key ID. The detach endpoint takes the key ID.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| String length/format validation | Custom regex validators | `stringvalidator.LengthBetween()`, `stringvalidator.RegexMatches()` from terraform-plugin-framework-validators | Comprehensive, tested, handles null/unknown correctly |
| Integer range validation | Custom min/max checks | `int64validator.Between()`, `int64validator.AtLeast()` | Framework-integrated, proper diagnostics |
| Float range validation | Custom bounds checking | `float64validator.Between()`, `float64validator.AtLeast()` | Same as above |
| Enum validation | Manual switch statements | `stringvalidator.OneOf()` | Type-safe, proper error messages |
| Resource timeouts | Manual context.WithTimeout | `terraform-plugin-framework-timeouts` | Integrates with Terraform timeouts block, user-configurable |
| UseStateForUnknown plan mod | Custom ModifyPlan | `resource/schema.UseStateForUnknown()` from framework | Built-in, correct semantics |
| RequiresReplace plan mod | Custom ModifyPlan | `resource/schema.RequiresReplace()` from framework | Built-in, correct semantics |
| Acceptance test framework | Manual terraform CLI calls | `terraform-plugin-testing` resource.Test | Official framework, handles provider lifecycle |
| SSH key content validation | Custom parsing | Validate prefix starts with "ssh-" | Simple check, matches SDK validation |

**Key insight:** The Terraform Plugin Framework ecosystem provides pre-built validators and plan modifiers for nearly every common case. Custom implementations should be limited to Vast.ai-specific validation (GPU name validation against the live API, region validation against the known region list).

## Common Pitfalls

### Pitfall 1: Instance Create Returns Contract ID, Not Full Instance
**What goes wrong:** Create returns `{"success": true, "new_contract": 7835610}` but the resource needs the full instance state. Developers try to use the response directly.
**Why it happens:** The Vast.ai API is asynchronous. Creation queues the instance; the actual instance data is only available after polling.
**How to avoid:** After Create, immediately poll `GET /instances/{new_contract}/` with the waiter until `actual_status` is "running" (or timeout). Set full state from the poll response, not the creation response.
**Warning signs:** "instance not found" errors immediately after creation. Null computed attributes in state.

### Pitfall 2: Offer IDs Are Ephemeral and Non-Reusable
**What goes wrong:** User plans with offer ID 12345, but by apply time the offer is gone. Or user tries to create two instances from the same offer ID.
**Why it happens:** Offers represent available GPU slots. Once rented, the offer is consumed. Offers also expire when hosts go offline.
**How to avoid:** Mark `offer_id` as RequiresReplace (immutable after creation). After creation, store `machine_id` as the stable reference. Error clearly on 404/400 from the create endpoint: "Offer {id} is no longer available."

### Pitfall 3: The "env" Field Has a Special Format
**What goes wrong:** Developers pass env vars as simple key-value pairs, but the API expects Docker-CLI-style format with `-e`, `-p`, `-h`, `-v` prefixes parsed into a specific map structure.
**Why it happens:** The Python SDK's `parse_env()` function (line 7833) parses `-e KEY=VALUE -p 8080:8080` into `{"KEY": "VALUE", "-p 8080:8080": "1"}`. The API stores this format.
**How to avoid:** In the Terraform schema, expose env vars as a map (key-value) and ports as a list of port mappings. The client layer translates these into the API's format. Do NOT expose the raw Docker-CLI string format to Terraform users.

### Pitfall 4: Instance Response Wraps in "instances" Key (Singular Endpoint)
**What goes wrong:** `GET /instances/{id}/` returns `{"instances": {single_object}}` not `{"instance": {...}}` or just the object. Developers unmarshal into the wrong structure.
**Why it happens:** Inconsistent API naming -- the plural key name is used even for a single instance.
**How to avoid:** Always use the wrapper struct: `struct { Instances Instance `json:"instances"` }` for singular, `struct { Instances []Instance `json:"instances"` }` for list.

### Pitfall 5: Template Delete Uses Body Not Path
**What goes wrong:** Developers try `DELETE /template/{id}/` but the API expects `DELETE /template/` with `{"hash_id": "..."}` or `{"template_id": N}` in the request body.
**Why it happens:** Unusual API design where delete uses a body payload.
**How to avoid:** The client's Delete method must support sending a request body. The current `VastAIClient.Delete()` already takes no body parameter -- it needs updating to support optional body for template deletion.

### Pitfall 6: GPU RAM Units Are in Megabytes, Not Gigabytes
**What goes wrong:** User specifies `gpu_ram >= 24` meaning 24 GB, but the API field `gpu_ram` is in MB (24000).
**Why it happens:** The Python SDK has `offers_mult` that multiplies gpu_ram by 1000 before sending to API. The Terraform provider must do the same conversion.
**How to avoid:** In the data source schema, expose `gpu_ram_gb` as the user-facing attribute (in GB). Convert to MB in the client layer query construction. Document this clearly.

### Pitfall 7: Runtype String Is Complex, Not a Simple Enum
**What goes wrong:** Developers expect runtype to be "ssh" or "jupyter" but it's actually composite: "ssh_direc ssh_proxy", "jupyter_direc ssh_direc ssh_proxy", etc.
**Why it happens:** The Python SDK's `get_runtype()` (line 2327) combines SSH/Jupyter/direct settings into a space-separated string.
**How to avoid:** In the Terraform schema, expose `ssh`, `jupyter`, and `direct` as separate boolean attributes. The client layer combines them into the correct runtype string. This is much better UX than exposing the raw runtype.

## Code Examples

### Offer Data Source Schema (verified pattern from framework docs)
```go
func (d *GpuOffersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Search for GPU offers on the Vast.ai marketplace. " +
            "See https://docs.vast.ai/api-reference/offers/search-offers",
        Attributes: map[string]schema.Attribute{
            // Filter attributes
            "gpu_name": schema.StringAttribute{
                Description: "Filter by GPU model name (e.g., 'RTX_4090', 'A100_SXM4').",
                Optional:    true,
                Validators: []validator.String{
                    stringvalidator.LengthAtLeast(1),
                },
            },
            "num_gpus": schema.Int64Attribute{
                Description: "Filter by minimum number of GPUs.",
                Optional:    true,
                Validators: []validator.Int64{
                    int64validator.Between(1, 16),
                },
            },
            "gpu_ram_gb": schema.Float64Attribute{
                Description: "Filter by minimum GPU VRAM in GB.",
                Optional:    true,
                Validators: []validator.Float64{
                    float64validator.AtLeast(1.0),
                },
            },
            "max_price_per_hour": schema.Float64Attribute{
                Description: "Maximum price per hour in USD.",
                Optional:    true,
                Validators: []validator.Float64{
                    float64validator.AtLeast(0.01),
                },
            },
            "datacenter_only": schema.BoolAttribute{
                Description: "Only return offers from datacenter-hosted machines.",
                Optional:    true,
            },
            "region": schema.StringAttribute{
                Description: "Filter by region (e.g., 'North_America', 'Europe', or country codes '[US,CA]').",
                Optional:    true,
            },
            "offer_type": schema.StringAttribute{
                Description: "Type of offer: 'on-demand', 'bid' (interruptible), or 'reserved'.",
                Optional:    true,
                Validators: []validator.String{
                    stringvalidator.OneOf("on-demand", "bid", "reserved"),
                },
            },
            "order_by": schema.StringAttribute{
                Description: "Sort field for results. Prefix with '-' for descending. Default: 'dph_total' (price ascending).",
                Optional:    true,
                Computed:    true,
            },
            "limit": schema.Int64Attribute{
                Description: "Maximum number of offers to return. Default: 10.",
                Optional:    true,
                Computed:    true,
            },
            "raw_query": schema.StringAttribute{
                Description: "Advanced: raw Vast.ai query DSL string. Overrides structured filters when set.",
                Optional:    true,
            },
            // Result attributes
            "offers": schema.ListNestedAttribute{
                Description: "List of matching GPU offers.",
                Computed:    true,
                NestedObject: schema.NestedAttributeObject{
                    Attributes: offerAttributes(), // extracted for reuse
                },
            },
            "most_affordable": schema.SingleNestedAttribute{
                Description: "The cheapest matching offer (convenience attribute).",
                Computed:    true,
                Attributes:  offerAttributes(),
            },
        },
    }
}
```

### Instance Resource Schema (key attributes)
```go
// Instance resource key attributes (abbreviated)
"id": schema.StringAttribute{
    Description: "Unique identifier of the instance (contract ID).",
    Computed:    true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
"offer_id": schema.Int64Attribute{
    Description: "ID of the GPU offer to create this instance from. " +
        "Obtain from the vastai_gpu_offers data source.",
    Required: true,
    PlanModifiers: []planmodifier.Int64{
        int64planmodifier.RequiresReplace(),
    },
},
"image": schema.StringAttribute{
    Description: "Docker image to launch (e.g., 'pytorch/pytorch:latest').",
    Optional:    true,
    Computed:    true,  // Can come from template
},
"disk_gb": schema.Float64Attribute{
    Description: "Size of local disk partition in GB.",
    Required:    true,
    PlanModifiers: []planmodifier.Float64{
        float64planmodifier.RequiresReplace(),
    },
},
"status": schema.StringAttribute{
    Description: "Desired instance status: 'running' or 'stopped'. " +
        "Changing this triggers start/stop without destroying the instance.",
    Optional: true,
    Computed: true,
    Validators: []validator.String{
        stringvalidator.OneOf("running", "stopped"),
    },
},
"bid_price": schema.Float64Attribute{
    Description: "Bid price per GPU per hour in USD. " +
        "Set to create an interruptible (spot) instance. " +
        "Omit for on-demand pricing.",
    Optional: true,
    Validators: []validator.Float64{
        float64validator.AtLeast(0.001),
    },
},
```

### Acceptance Test Helper Setup
```go
// internal/acctest/helpers.go
package acctest

import (
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
    "github.com/realnedsanders/terraform-provider-vastai/internal/provider"
)

var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "vastai": providerserver.NewProtocol6WithError(provider.New("test")()),
}
```

### Import State Implementation
```go
// resource_instance.go
func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| SDKv2 testing | terraform-plugin-testing v1.15.0 | 2024 | Standalone test module, works with both Framework and SDKv2 |
| Schema-level timeouts | terraform-plugin-framework-timeouts with nested attributes | 2024 | Nested attributes preferred over blocks for new providers |
| Manual validators | terraform-plugin-framework-validators v0.19.0 | Ongoing | Comprehensive pre-built validators, avoid custom implementations |
| Plan modifiers via ModifyPlan | Attribute-level plan modifiers | Plugin Framework 1.x | UseStateForUnknown(), RequiresReplace() directly on schema attributes |

## Open Questions

1. **Instance extra_env format**
   - What we know: The API stores env vars as `[[key1, val1], [key2, val2]]` (array of pairs). The Python SDK's parse_env() converts Docker CLI format (`-e KEY=VALUE -p 8080:8080`) into a map.
   - What's unclear: The exact response format when reading back -- is it always `[[k,v], ...]` or can it be a map?
   - Recommendation: Implement as `map(string)` in Terraform for env vars and `list(string)` for port mappings. Convert formats in the client layer. Test with live API during first acceptance test to verify.

2. **Template identification: hash_id vs template_id**
   - What we know: Templates have both a numeric `id` and a string `hash_id`. Create returns both. Delete accepts either.
   - What's unclear: Which is the stable identifier? Can hash_id change on update?
   - Recommendation: Use `hash_id` as the primary Terraform identifier since it is used by `template_hash_id` in instance creation. Store both in state.

3. **Preemption vs normal termination detection**
   - What we know: No explicit "preempted" status exists in the API. Must infer from `is_bid`, `actual_status`, `intended_status`, and `status_msg`.
   - What's unclear: Exact `status_msg` values when preempted vs other failure modes.
   - Recommendation: Implement conservative detection: only remove from state on 404 (instance gone) or when `is_bid=true` AND `intended_status=running` AND `actual_status` is in `["stopped", "offline", "exited"]`. Log at WARN level when removing preempted instances.

4. **Delete method with request body**
   - What we know: Template delete needs a body (`{"hash_id": "..."}`) but the current `VastAIClient.Delete()` does not accept a body parameter.
   - Recommendation: Add a `DeleteWithBody()` method or modify Delete to accept an optional body interface{}.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | terraform-plugin-testing v1.15.0 + Go testing |
| Config file | None needed -- uses Go test conventions |
| Quick run command | `go test -v -count=1 -parallel=4 -timeout 120s ./internal/...` |
| Full suite command | `TF_ACC=1 go test -v -count=1 -parallel=4 -timeout 120m ./internal/...` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| COMP-01 | Instance CRUD | acceptance | `TF_ACC=1 go test ./internal/services/instance/ -run TestAccInstance -v` | Wave 0 |
| COMP-02 | Instance start/stop | acceptance | `TF_ACC=1 go test ./internal/services/instance/ -run TestAccInstanceStartStop -v` | Wave 0 |
| COMP-03 | Instance label/bid/template update | acceptance | `TF_ACC=1 go test ./internal/services/instance/ -run TestAccInstanceUpdate -v` | Wave 0 |
| COMP-04 | Preemption handling | unit | `go test ./internal/services/instance/ -run TestPreemption -v` | Wave 0 |
| COMP-05 | Instance create polling | unit | `go test ./internal/client/ -run TestWaitForStatus -v` | Wave 0 |
| COMP-06 | Template CRUD | acceptance | `TF_ACC=1 go test ./internal/services/template/ -run TestAccTemplate -v` | Wave 0 |
| COMP-07 | SSH key CRUD | acceptance | `TF_ACC=1 go test ./internal/services/sshkey/ -run TestAccSSHKey -v` | Wave 0 |
| COMP-08 | SSH attach/detach | acceptance | `TF_ACC=1 go test ./internal/services/sshkey/ -run TestAccSSHKeyAttach -v` | Wave 0 |
| DATA-01 | GPU offers data source | acceptance+unit | `go test ./internal/services/offer/ -run TestGpuOffers -v` | Wave 0 |
| DATA-02 | Instance data source | acceptance | `TF_ACC=1 go test ./internal/services/instance/ -run TestAccInstanceDataSource -v` | Wave 0 |
| DATA-03 | Instances data source | acceptance | `TF_ACC=1 go test ./internal/services/instance/ -run TestAccInstancesDataSource -v` | Wave 0 |
| DATA-04 | Templates data source | acceptance | `TF_ACC=1 go test ./internal/services/template/ -run TestAccTemplatesDataSource -v` | Wave 0 |
| DATA-08 | SSH keys data source | acceptance | `TF_ACC=1 go test ./internal/services/sshkey/ -run TestAccSSHKeysDataSource -v` | Wave 0 |
| SCHM-01 | Validators | unit | `go test ./internal/services/... -run TestValidator -v` | Wave 0 |
| SCHM-02 | Sensitive flags | unit | `go test ./internal/services/... -run TestSensitive -v` | Wave 0 |
| SCHM-03 | Attribute classification | unit | Schema tests verify Required/Optional/Computed flags | Wave 0 |
| SCHM-04 | Descriptions | unit | Schema tests verify all descriptions non-empty | Wave 0 |
| SCHM-05 | Plan modifiers | unit | `go test ./internal/services/... -run TestPlanModifier -v` | Wave 0 |
| SCHM-06 | Timeouts | unit | Schema tests verify timeouts block exists | Wave 0 |
| IMPT-01 | Import support | acceptance | ImportState test steps in acceptance tests | Wave 0 |
| IMPT-02 | Import docs | manual | Verify example import commands in templates/ | Wave 0 |
| TEST-01 | Acceptance tests | acceptance | `TF_ACC=1 go test ./internal/... -v -timeout 120m` | Wave 0 |
| TEST-02 | Unit tests | unit | `go test ./internal/... -v -timeout 120s` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test -v -count=1 -parallel=4 -timeout 120s ./internal/...`
- **Per wave merge:** `go test -v -count=1 -parallel=4 -timeout 120s ./internal/...` (unit only)
- **Phase gate:** Full suite green before verification

### Wave 0 Gaps
- [ ] `internal/acctest/helpers.go` -- ProtoV6ProviderFactories shared helper
- [ ] Framework deps: `go get github.com/hashicorp/terraform-plugin-testing@v1.15.0`
- [ ] Framework deps: `go get github.com/hashicorp/terraform-plugin-framework-validators@v0.19.0`
- [ ] Framework deps: `go get github.com/hashicorp/terraform-plugin-framework-timeouts@v0.5.0`

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | All compilation | Verify at plan time | 1.25.x | -- |
| terraform (or tofu) | Acceptance tests | Verify at plan time | >= 1.0 | -- |
| golangci-lint | Linting | Verify at plan time | v2.11.x | -- |
| VASTAI_API_KEY | Acceptance tests | Verify at plan time | -- | Skip acc tests |

## Sources

### Primary (HIGH confidence)
- `vast-sdk/vastai/vast.py` -- All API endpoints, request/response formats, query DSL, status values. Lines 781-803 (instance_fields), 973-1046 (offers_fields/alias/mult), 2386-2450 (create_instance), 3078-3098 (destroy_instance), 3881-3906 (start_instance), 3948-3973 (stop_instance), 4326-4438 (search_offers), 2705-2759 (create_template), 3029-3057 (delete_template), 6758-6888 (update_instance/template), 2102-2117 (create_ssh_key), 4994-5001 (show_ssh_keys), 2870-2874 (delete_ssh_key), 6896-6908 (update_ssh_key), 1351-1357 (attach_ssh), 3147-3151 (detach_ssh), 1501-1524 (change_bid), 3381-3399 (label_instance), 5613-5634 (show_instance), 5641-5667 (show_instances)
- `internal/client/client.go` -- Existing VastAIClient with Get/Post/Put/Delete, retry logic
- `internal/client/auth.go` -- Bearer auth, request construction
- `internal/provider/provider.go` -- Provider shell with Configure, empty Resources/DataSources

### Secondary (MEDIUM confidence)
- [Vast.ai Instance Management Docs](https://docs.vast.ai/documentation/instances/manage-instances) -- Instance lifecycle states
- [Vast.ai Instance Types](https://docs.vast.ai/documentation/instances/choosing/instance-types) -- On-demand, interruptible, reserved behavior
- [Vast.ai Show Instance API](https://docs.vast.ai/api-reference/instances/show-instance) -- ~90 response fields
- [Terraform Plugin Framework Timeouts](https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts) -- Timeouts library usage
- [Terraform Plugin Framework Validation](https://developer.hashicorp.com/terraform/plugin/framework/validation) -- Validator patterns
- [terraform-plugin-framework-validators GitHub](https://github.com/hashicorp/terraform-plugin-framework-validators) -- Pre-built validators

### Tertiary (LOW confidence)
- Instance preemption detection logic -- inferred from API fields, not explicitly documented by Vast.ai

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- All libraries verified, versions confirmed from Phase 1 research
- Architecture: HIGH -- Patterns follow HashiCorp best practices and match existing Phase 1 code
- API contracts: HIGH -- Extracted directly from Python SDK source code, not training data
- Pitfalls: HIGH -- Based on actual API quirks discovered during SDK analysis
- Preemption handling: MEDIUM -- Inferred from multiple fields; no explicit preemption status documented

**Research date:** 2026-03-25
**Valid until:** 2026-04-25 (stable API, framework versions unlikely to change)
