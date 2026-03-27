# Phase 4: Serverless - Research

**Researched:** 2026-03-27
**Domain:** Vast.ai Serverless Inference API (endpoints, worker groups, autoscaling)
**Confidence:** HIGH

## Summary

Phase 4 implements Terraform resources for Vast.ai's serverless inference system. The API surface consists of two primary resource types -- endpoints (`/endptjobs/`) and worker groups (`/autojobs/`) -- each with full CRUD semantics. Endpoints own autoscaling configuration (min_load, target_util, cold_mult, cold_workers, max_workers) and serve as the parent entity. Worker groups bind to endpoints and define the compute configuration (template, search params, GPU RAM requirements). There is no standalone "autogroup" resource -- the SDK's `create__autogroup` does not exist in the current codebase. The SRVL-03 requirement for `vastai_autogroup` should be folded into the endpoint resource's autoscaling configuration.

The API contracts are clearly extractable from the Python SDK. Both resources follow the same HTTP verb pattern as existing provider resources: POST to create, GET to list (no single-GET endpoint), PUT to update, DELETE with body to destroy. The response format uses `{"success": true, "results": [...]}` wrapping for list operations. Delete operations send a JSON body with the entity ID (`endptjob_id` or `autojob_id`), consistent with the template delete pattern already implemented.

**Primary recommendation:** Implement two resources (`vastai_endpoint`, `vastai_worker_group`) and one data source (`vastai_endpoints`). Model endpoints as the parent resource with all autoscaling parameters. Worker groups reference endpoints via `endpoint_id`. No standalone autogroup resource is needed -- autoscaling behavior is configured at the endpoint level.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- D-01: Claude's Discretion on resource separation (all separate vs combined) -- based on API structure and dependency patterns
- D-02: Explicit deletion required -- worker groups must be deleted before endpoint. Terraform manages dependencies, no cascade.
- D-03: Claude's Discretion on endpoint status data source depth -- metadata only vs full worker health/load details based on API response
- D-04: Some params required (min_load, max_workers -- must-think-about), rest optional with sensible defaults
- D-05: Strict validation ranges on all autoscaling params -- target_util 0-1, cold_mult >= 1, workers >= 0, etc.
- Inherited: Service-per-directory, service pattern on client, always snake_case, Optional+Computed for server defaults, comprehensive validators, per-resource models, API models in client, dual test strategy

### Claude's Discretion
- Resource model separation (endpoint, worker group, autogroup)
- Endpoint status data source depth
- Specific default values for optional autoscaling params
- Timeout defaults for endpoint/worker operations

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SRVL-01 | `vastai_endpoint` resource with CRUD and autoscaling parameters (min_load, target_util, cold_mult, cold_workers, max_workers) | API contract fully extracted: POST /endptjobs/ for create, PUT /endptjobs/{id}/ for update, DELETE /endptjobs/{id}/ for delete, GET /endptjobs/ for list. All autoscaling params documented with types and defaults. |
| SRVL-02 | `vastai_worker_group` resource with CRUD (bound to endpoint, template, search params, autoscaling config) | API contract fully extracted: POST /autojobs/ for create, PUT /autojobs/{id}/ for update, DELETE /autojobs/{id}/ for delete, GET /autojobs/ for list. Worker groups bind to endpoints via endpoint_id/endpoint_name. |
| SRVL-03 | `vastai_autogroup` resource with CRUD for autoscaling groups | Research finding: no standalone autogroup API exists in the SDK. `create__autogroup` is absent. Autoscaling is configured directly on endpoints. Recommend: fulfill SRVL-03 by documenting that autoscaling config lives on the endpoint resource, not as a separate resource. |
| DATA-09 | `vastai_endpoints` data source (list serverless endpoints) | API contract: GET /endptjobs/ returns `{"success": true, "results": [...]}`. Response includes endpoint metadata plus autoscaling params. Data source strips `api_key`, `auto_delete_in_seconds`, `auto_delete_due_24h` from response per SDK pattern. |
</phase_requirements>

## Standard Stack

No new libraries needed for Phase 4. All dependencies are already in `go.mod`.

### Core (already in go.mod)
| Library | Version | Purpose | Status |
|---------|---------|---------|--------|
| terraform-plugin-framework | v1.19.0 | Resource/data source definitions | In go.mod |
| terraform-plugin-framework-timeouts | v0.5.0 | Configurable CRUD timeouts | In go.mod |
| terraform-plugin-framework-validators | v0.19.0 | Float64/Int64/String validators | In go.mod |
| terraform-plugin-log (tflog) | v0.10.0 | Structured logging | In go.mod |
| terraform-plugin-testing | v1.15.0 | Acceptance tests | In go.mod |

**Installation:** No new packages needed. `go mod tidy` is sufficient.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  client/
    endpoints.go           # EndpointService + API models
    endpoints_test.go      # Client unit tests (httptest mocks)
    worker_groups.go       # WorkerGroupService + API models
    worker_groups_test.go  # Client unit tests (httptest mocks)
  services/
    endpoint/
      resource_endpoint.go          # vastai_endpoint resource
      resource_endpoint_test.go     # Schema unit tests
      data_source_endpoints.go      # vastai_endpoints data source
      data_source_endpoints_test.go # Schema unit tests
      models.go                     # TF resource/data models
    workergroup/
      resource_worker_group.go          # vastai_worker_group resource
      resource_worker_group_test.go     # Schema unit tests
      models.go                         # TF resource models
  provider/
    provider.go            # Add Endpoints, WorkerGroups services + register resources
```

### Pattern 1: Client Service (per existing convention)

Each API domain gets a service struct attached to VastAIClient. Follow the exact pattern from `InstanceService`, `VolumeService`, etc.

**What:** Service sub-object on VastAIClient with typed methods for CRUD.
**When to use:** Every new API domain.
**Example:**
```go
// In client/endpoints.go
type EndpointService struct {
    client *VastAIClient
}

type CreateEndpointRequest struct {
    ClientID           string  `json:"client_id"`
    EndpointName       string  `json:"endpoint_name"`
    MinLoad            float64 `json:"min_load"`
    MinColdLoad        float64 `json:"min_cold_load"`
    TargetUtil         float64 `json:"target_util"`
    ColdMult           float64 `json:"cold_mult"`
    ColdWorkers        int     `json:"cold_workers"`
    MaxWorkers         int     `json:"max_workers"`
    AutoscalerInstance string  `json:"autoscaler_instance"`
}

// In client/client.go -- add to VastAIClient struct:
Endpoints    *EndpointService
WorkerGroups *WorkerGroupService

// In NewVastAIClient -- initialize:
c.Endpoints = &EndpointService{client: c}
c.WorkerGroups = &WorkerGroupService{client: c}
```

### Pattern 2: Resource with CRUD (per existing convention)

Follow the exact structure from `resource_volume.go` and `resource_instance.go`:
- Implement `Resource`, `ResourceWithConfigure`, `ResourceWithImportState` interfaces
- Schema with validators, plan modifiers, timeouts block
- Configure method extracting `*client.VastAIClient`
- Create/Read/Update/Delete/ImportState methods
- Helper function `read*IntoModel` for API-to-TF state mapping

### Pattern 3: Read-via-List (per volume/SSH key pattern)

The serverless API has no single-GET endpoint. Both `show__endpoints` and `show__workergroups` use list endpoints (`GET /endptjobs/` and `GET /autojobs/`). Read operations must:
1. Call the list endpoint
2. Filter by ID in Go code
3. Return "not found" (remove from state) if absent

```go
func (s *EndpointService) List(ctx context.Context) ([]Endpoint, error) {
    var resp endpointListResponse
    if err := s.client.Get(ctx, "/endptjobs/", &resp); err != nil {
        return nil, fmt.Errorf("listing endpoints: %w", err)
    }
    if !resp.Success {
        return nil, fmt.Errorf("listing endpoints: API returned success=false: %s", resp.Message)
    }
    return resp.Results, nil
}
```

### Pattern 4: Delete-with-Body (per template/endpoint SDK pattern)

Both endpoint and worker group delete operations send a JSON body with the entity ID. Use `DeleteWithBody` (already available on VastAIClient).

```go
func (s *EndpointService) Delete(ctx context.Context, id int) error {
    path := fmt.Sprintf("/endptjobs/%d/", id)
    body := map[string]interface{}{
        "client_id":  "me",
        "endptjob_id": id,
    }
    if err := s.client.DeleteWithBody(ctx, path, body, nil); err != nil {
        return fmt.Errorf("deleting endpoint %d: %w", id, err)
    }
    return nil
}
```

### Anti-Patterns to Avoid
- **Creating a standalone autogroup resource:** The SDK has no `create__autogroup` function. Autoscaling is embedded in the endpoint. Do not invent a resource for something the API does not support.
- **Cascading delete on endpoint:** Per D-02, worker groups must be deleted before their parent endpoint. Terraform's dependency graph handles this naturally via `endpoint_id` reference, but the resource should NOT attempt to auto-delete child worker groups.
- **Sending nil/zero for optional autoscaling params on update:** The Python SDK always sends all fields. When updating, send the full current state, not just changed fields, to avoid the API resetting unspecified fields to defaults.

## Resource Model Recommendation (Claude's Discretion: D-01)

### Decision: Two Separate Resources + One Data Source

Based on the API structure:

1. **`vastai_endpoint`** -- Maps to `/endptjobs/` API. Owns all autoscaling configuration.
2. **`vastai_worker_group`** -- Maps to `/autojobs/` API. References endpoint, defines compute configuration.
3. **`vastai_endpoints`** -- Data source listing all endpoints.

**Rationale:**
- The API has clearly separate endpoints for each resource type (`/endptjobs/` vs `/autojobs/`)
- Worker groups reference endpoints via `endpoint_id`/`endpoint_name` -- this is a natural Terraform dependency
- One endpoint can have multiple worker groups (the API explicitly supports this)
- No standalone autogroup API exists; autoscaling params belong to the endpoint

**No `vastai_autogroup` resource.** SRVL-03 is satisfied by the endpoint resource's autoscaling configuration. The term "autogroup" in the Vast.ai API refers to the combination of an endpoint with its autoscaling config -- it is not a separate entity.

## API Contracts (Extracted from Python SDK)

### Endpoint API (`/endptjobs/`)

**Create: POST /endptjobs/**
```json
{
    "client_id": "me",
    "endpoint_name": "string",
    "min_load": 0.0,
    "min_cold_load": 0.0,
    "target_util": 0.9,
    "cold_mult": 2.5,
    "cold_workers": 5,
    "max_workers": 20,
    "autoscaler_instance": "prod"
}
```
Response: `{"success": true, ...}` (with ID in response -- SDK prints `r.json()`)

**List: GET /endptjobs/**
Request body: `{"client_id": "me", "api_key": "<key>"}`
Response:
```json
{
    "success": true,
    "results": [
        {
            "id": 123,
            "endpoint_name": "LLama",
            "min_load": 0.0,
            "min_cold_load": 0.0,
            "target_util": 0.9,
            "cold_mult": 2.5,
            "cold_workers": 5,
            "max_workers": 20,
            "endpoint_state": "active",
            "api_key": "<stripped>",
            "auto_delete_in_seconds": "<stripped>",
            "auto_delete_due_24h": "<stripped>"
        }
    ]
}
```
Note: SDK strips `api_key`, `auto_delete_in_seconds`, `auto_delete_due_24h` from display.

**Update: PUT /endptjobs/{id}/**
```json
{
    "client_id": "me",
    "endptjob_id": 123,
    "min_load": 0.0,
    "min_cold_load": 0.0,
    "target_util": 0.9,
    "cold_mult": 2.5,
    "cold_workers": 5,
    "max_workers": 20,
    "endpoint_name": "LLama",
    "endpoint_state": "active",
    "autoscaler_instance": "prod"
}
```
Note: `endpoint_state` supports values: "active", "suspended", "stopped"

**Delete: DELETE /endptjobs/{id}/**
Request body: `{"client_id": "me", "endptjob_id": 123}`

### Worker Group API (`/autojobs/`)

**Create: POST /autojobs/**
```json
{
    "client_id": "me",
    "endpoint_name": "string",
    "endpoint_id": 123,
    "template_hash": "string",
    "template_id": 456,
    "search_params": "gpu_ram>=23 num_gpus=2 gpu_name=RTX_4090 inet_down>200 direct_port_count>2 disk_space>=64",
    "launch_args": "string",
    "gpu_ram": 32.0,
    "min_load": 0.0,
    "target_util": 0.9,
    "cold_mult": 2.0,
    "cold_workers": 5,
    "test_workers": 3,
    "autoscaler_instance": "prod"
}
```
Note: SDK adds default search params `"verified=True rentable=True rented=False"` unless `--no-default` flag is set.
Note: Per SDK help text, `min_load`, `target_util`, `cold_mult` are "not currently used at the workergroup level" -- they exist for API compatibility but autoscaling is driven by the endpoint.

**List: GET /autojobs/**
Request body: `{"client_id": "me", "api_key": "<key>"}`
Response: `{"success": true, "results": [...]}`

**Update: PUT /autojobs/{id}/**
```json
{
    "client_id": "me",
    "autojob_id": 123,
    "endpoint_name": "string",
    "endpoint_id": 456,
    "template_hash": "string",
    "template_id": 789,
    "search_params": "string",
    "launch_args": "string",
    "gpu_ram": 32.0,
    "min_load": 0.0,
    "target_util": 0.9,
    "cold_mult": 2.0,
    "cold_workers": 5,
    "test_workers": 3
}
```

**Delete: DELETE /autojobs/{id}/**
Request body: `{"client_id": "me", "autojob_id": 123}`
Note: "deleting a workergroup doesn't automatically destroy all the instances that are associated with your workergroup" (per SDK help)

### Endpoint Logs: POST https://run.vast.ai/get_endpoint_logs/
Request body: `{"id": 123, "api_key": "<key>", "tail": 100}`
Note: Goes to `run.vast.ai`, NOT the standard API. Out of scope for core CRUD resources.

### Worker Group Logs: POST https://run.vast.ai/get_autogroup_logs/
Same pattern as endpoint logs. Out of scope.

## Autoscaling Parameters Reference

### Endpoint-Level Parameters
| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `min_load` | float64 | 0.0 | >= 0.0 | Minimum floor load in perf units/s (token/s for LLMs) |
| `min_cold_load` | float64 | 0.0 | >= 0.0 | Min floor load allowing cold workers to handle |
| `target_util` | float64 | 0.9 | 0.0-1.0 | Target capacity utilization fraction |
| `cold_mult` | float64 | 2.5 | >= 1.0 | Cold/stopped capacity as multiple of hot capacity |
| `cold_workers` | int | 5 | >= 0 | Min cold workers when no load |
| `max_workers` | int | 20 | >= 0 | Max workers the endpoint can have |
| `endpoint_state` | string | "active" | active/suspended/stopped | Runtime state (update only) |

### Worker-Group-Level Parameters
| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `template_hash` | string | required | - | Template hash for worker instances |
| `template_id` | int | optional | >= 1 | Numeric template ID (alternative to hash) |
| `search_params` | string | "" + defaults | - | Offer search query string |
| `launch_args` | string | optional | - | Instance launch args string |
| `gpu_ram` | float64 | optional | > 0 | Estimated GPU RAM requirement |
| `test_workers` | int | 3 | >= 0 | Workers to create for initial performance estimate |
| `cold_workers` | int | optional | >= 0 | Min cold workers for this specific group |
| `endpoint_name` | string | required* | - | Parent endpoint name |
| `endpoint_id` | int | optional | >= 1 | Parent endpoint ID (alternative to name) |

*Note: Either `endpoint_name` or `endpoint_id` must be provided to link a worker group to an endpoint.

## Terraform Schema Design

### `vastai_endpoint` Resource Schema

```go
// Required
"endpoint_name" // string, Required. Name/identifier for the endpoint.

// Optional+Computed (autoscaling params with server defaults per D-04, D-05)
"min_load"       // float64, Optional+Computed, default 0.0, AtLeast(0)
"min_cold_load"  // float64, Optional+Computed, default 0.0, AtLeast(0)
"target_util"    // float64, Optional+Computed, default 0.9, Between(0, 1)
"cold_mult"      // float64, Optional+Computed, default 2.5, AtLeast(1.0)
"cold_workers"   // int64, Optional+Computed, default 5, AtLeast(0)
"max_workers"    // int64, Optional+Computed, default 20, AtLeast(0)

// Optional (update-only)
"endpoint_state" // string, Optional+Computed, OneOf("active","suspended","stopped")

// Computed
"id"             // string, Computed, UseStateForUnknown

// Timeouts block
"timeouts"       // create, read, update, delete
```

### `vastai_worker_group` Resource Schema

```go
// Required
"endpoint_id"    // int64, Required (binds to parent endpoint). RequiresReplace.
"template_hash"  // string, Required (or Optional if template_id provided -- use AtLeastOneOf)

// Optional
"template_id"    // int64, Optional (alternative to template_hash)
"search_params"  // string, Optional (offer search filter string)
"launch_args"    // string, Optional (instance launch arguments)
"gpu_ram"        // float64, Optional, AtLeast(0)
"test_workers"   // int64, Optional+Computed, default 3, AtLeast(0)
"cold_workers"   // int64, Optional, AtLeast(0)
"endpoint_name"  // string, Optional (inferred from endpoint_id in most cases)

// Computed
"id"             // string, Computed, UseStateForUnknown

// Timeouts block
"timeouts"       // create, read, update, delete
```

### `vastai_endpoints` Data Source Schema

```go
// Computed (list result)
"endpoints"      // List of objects containing endpoint metadata

// Per endpoint object:
"id"             // int64
"endpoint_name"  // string
"min_load"       // float64
"min_cold_load"  // float64
"target_util"    // float64
"cold_mult"      // float64
"cold_workers"   // int64
"max_workers"    // int64
"endpoint_state" // string
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Float64 range validation | Custom validator for 0-1 range | `float64validator.Between(0, 1)` | Already in framework-validators |
| Int64 minimum validation | Custom >= 0 check | `int64validator.AtLeast(0)` | Standard validator |
| String enum validation | Custom string matching | `stringvalidator.OneOf("active","suspended","stopped")` | Standard validator |
| At-least-one-of validation | Custom config validator | `path.MatchRoot("template_hash"), path.MatchRoot("template_id")` with `AtLeastOneOf` | Framework-native |
| Timeout configuration | Manual context timeout | `timeouts.Block` + `model.Timeouts.Create(ctx, default)` | Already used across all resources |
| HTTP retry/backoff | Custom retry loop | `go-retryablehttp` via VastAIClient | Already configured |

## Common Pitfalls

### Pitfall 1: No Single-GET Endpoint
**What goes wrong:** Attempting `GET /endptjobs/{id}/` or `GET /autojobs/{id}/` returns 404 or unexpected response.
**Why it happens:** The Vast.ai API only supports list operations for serverless resources (same pattern as SSH keys, volumes).
**How to avoid:** Always use the list endpoint and filter by ID in Go code. Cache the list result within a single CRUD operation if multiple reads are needed.
**Warning signs:** 404 errors on read operations.

### Pitfall 2: Delete Requires JSON Body
**What goes wrong:** `DELETE /endptjobs/{id}/` with no body fails or returns error.
**Why it happens:** The API expects `{"client_id": "me", "endptjob_id": id}` in the request body (same pattern as template delete).
**How to avoid:** Use `client.DeleteWithBody()` instead of `client.Delete()`.
**Warning signs:** Delete operations returning 400 or "missing required field" errors.

### Pitfall 3: Worker Group Autoscaling Params Are No-Ops
**What goes wrong:** Setting `min_load`, `target_util`, `cold_mult` on a worker group and expecting them to control scaling.
**Why it happens:** Per the SDK help text, these fields are "not currently used at the workergroup level". Autoscaling is driven by the endpoint.
**How to avoid:** Expose these fields as Optional on the worker group (the API accepts them) but document that they are endpoint-level concerns. Consider omitting them from the worker group schema entirely if they truly have no effect.
**Warning signs:** Setting different autoscaling params on worker group vs endpoint and seeing only endpoint values take effect.

### Pitfall 4: Default Search Params on Worker Group Create
**What goes wrong:** Created worker groups match different offers than expected.
**Why it happens:** The Python SDK appends `"verified=True rentable=True rented=False"` to search_params by default. Our Go client must decide whether to replicate this behavior.
**How to avoid:** Document clearly in the schema description that the API may apply default search filters. Consider offering a `no_default_search_params` bool attribute or always appending the defaults in the client.
**Warning signs:** Worker groups not finding suitable offers.

### Pitfall 5: autoscaler_instance Field
**What goes wrong:** Omitting `autoscaler_instance` from create/update requests causes unexpected behavior.
**Why it happens:** This field is a hidden/suppressed arg in the SDK that defaults to `"prod"`. It must be included in API requests.
**How to avoid:** Always send `"autoscaler_instance": "prod"` in create and update requests. Do NOT expose this as a user-facing attribute -- it is an internal API detail.
**Warning signs:** API errors or requests being routed to wrong autoscaler instance.

### Pitfall 6: endpoint_state Only Available on Update
**What goes wrong:** Trying to set `endpoint_state` during create fails or is ignored.
**Why it happens:** The create endpoint doesn't include `endpoint_state` in its request body. It is only available on the update endpoint.
**How to avoid:** Make `endpoint_state` Optional+Computed. If set in config, apply it via a post-create update call. Or simply document that new endpoints start as "active" and state can be changed via updates.
**Warning signs:** endpoint_state being ignored on initial creation.

### Pitfall 7: Ordering Dependency -- Worker Groups Before Endpoints
**What goes wrong:** Deleting an endpoint before its worker groups causes orphaned resources.
**Why it happens:** Per D-02, there is no cascade delete.
**How to avoid:** Terraform's dependency graph handles this naturally when worker groups reference `vastai_endpoint.example.id` as their `endpoint_id`. Terraform will destroy worker groups first. Document this dependency clearly.
**Warning signs:** Errors when destroying endpoints that still have worker groups.

### Pitfall 8: GET Requests with JSON Body
**What goes wrong:** `show__endpoints` and `show__workergroups` send `{"client_id": "me", "api_key": "<key>"}` as the request body on a GET request.
**Why it happens:** Vast.ai API accepts GET requests with JSON bodies (non-standard but functional).
**How to avoid:** The existing VastAIClient.Get method does not send a body. For list operations, we may need to either: (a) use query parameters instead, (b) send the API key via the Authorization header (which we already do), or (c) use Post for list operations. Since Bearer auth is already configured, the `client_id` and `api_key` body fields may be redundant. Test with just the Authorization header first.
**Warning signs:** List operations returning empty results or auth errors.

## Code Examples

### Client Service -- Endpoint CRUD
```go
// Source: Extracted from vast-sdk/vastai/vast.py create__endpoint, show__endpoints,
// update__endpoint, delete__endpoint

// endpoints.go
package client

import (
    "context"
    "fmt"
)

type EndpointService struct {
    client *VastAIClient
}

type CreateEndpointRequest struct {
    ClientID           string  `json:"client_id"`
    EndpointName       string  `json:"endpoint_name"`
    MinLoad            float64 `json:"min_load"`
    MinColdLoad        float64 `json:"min_cold_load"`
    TargetUtil         float64 `json:"target_util"`
    ColdMult           float64 `json:"cold_mult"`
    ColdWorkers        int     `json:"cold_workers"`
    MaxWorkers         int     `json:"max_workers"`
    AutoscalerInstance string  `json:"autoscaler_instance"`
}

type UpdateEndpointRequest struct {
    ClientID           string   `json:"client_id"`
    EndptJobID         int      `json:"endptjob_id"`
    EndpointName       string   `json:"endpoint_name,omitempty"`
    MinLoad            *float64 `json:"min_load,omitempty"`
    MinColdLoad        *float64 `json:"min_cold_load,omitempty"`
    TargetUtil         *float64 `json:"target_util,omitempty"`
    ColdMult           *float64 `json:"cold_mult,omitempty"`
    ColdWorkers        *int     `json:"cold_workers,omitempty"`
    MaxWorkers         *int     `json:"max_workers,omitempty"`
    EndpointState      string   `json:"endpoint_state,omitempty"`
    AutoscalerInstance string   `json:"autoscaler_instance"`
}

// Endpoint represents an endpoint from the API response
type Endpoint struct {
    ID            int     `json:"id"`
    EndpointName  string  `json:"endpoint_name"`
    MinLoad       float64 `json:"min_load"`
    MinColdLoad   float64 `json:"min_cold_load"`
    TargetUtil    float64 `json:"target_util"`
    ColdMult      float64 `json:"cold_mult"`
    ColdWorkers   int     `json:"cold_workers"`
    MaxWorkers    int     `json:"max_workers"`
    EndpointState string  `json:"endpoint_state"`
}

type endpointListResponse struct {
    Success bool       `json:"success"`
    Results []Endpoint `json:"results"`
    Message string     `json:"msg,omitempty"`
}

func (s *EndpointService) Create(ctx context.Context, req *CreateEndpointRequest) (*Endpoint, error) {
    req.ClientID = "me"
    req.AutoscalerInstance = "prod"
    // POST to /endptjobs/ -- response should include the new endpoint ID
    var resp map[string]interface{}
    if err := s.client.Post(ctx, "/endptjobs/", req, &resp); err != nil {
        return nil, fmt.Errorf("creating endpoint: %w", err)
    }
    // After create, list to find the newly created endpoint (create-then-read pattern)
    endpoints, err := s.List(ctx)
    if err != nil {
        return nil, fmt.Errorf("reading endpoint after create: %w", err)
    }
    // Find by name (most recently created matching endpoint)
    for i := len(endpoints) - 1; i >= 0; i-- {
        if endpoints[i].EndpointName == req.EndpointName {
            return &endpoints[i], nil
        }
    }
    return nil, fmt.Errorf("endpoint %q not found after creation", req.EndpointName)
}

func (s *EndpointService) List(ctx context.Context) ([]Endpoint, error) {
    var resp endpointListResponse
    if err := s.client.Get(ctx, "/endptjobs/", &resp); err != nil {
        return nil, fmt.Errorf("listing endpoints: %w", err)
    }
    if !resp.Success {
        return nil, fmt.Errorf("listing endpoints: API returned success=false: %s", resp.Message)
    }
    return resp.Results, nil
}

func (s *EndpointService) Update(ctx context.Context, id int, req *UpdateEndpointRequest) error {
    req.ClientID = "me"
    req.EndptJobID = id
    req.AutoscalerInstance = "prod"
    path := fmt.Sprintf("/endptjobs/%d/", id)
    if err := s.client.Put(ctx, path, req, nil); err != nil {
        return fmt.Errorf("updating endpoint %d: %w", id, err)
    }
    return nil
}

func (s *EndpointService) Delete(ctx context.Context, id int) error {
    path := fmt.Sprintf("/endptjobs/%d/", id)
    body := map[string]interface{}{
        "client_id":   "me",
        "endptjob_id": id,
    }
    if err := s.client.DeleteWithBody(ctx, path, body, nil); err != nil {
        return fmt.Errorf("deleting endpoint %d: %w", id, err)
    }
    return nil
}
```

### Client Service -- Worker Group CRUD
```go
// Source: Extracted from vast-sdk/vastai/vast.py create__workergroup, show__workergroups,
// update__workergroup, delete__workergroup

type WorkerGroupService struct {
    client *VastAIClient
}

type CreateWorkerGroupRequest struct {
    ClientID           string  `json:"client_id"`
    EndpointName       string  `json:"endpoint_name,omitempty"`
    EndpointID         int     `json:"endpoint_id,omitempty"`
    TemplateHash       string  `json:"template_hash,omitempty"`
    TemplateID         int     `json:"template_id,omitempty"`
    SearchParams       string  `json:"search_params,omitempty"`
    LaunchArgs         string  `json:"launch_args,omitempty"`
    GpuRAM             float64 `json:"gpu_ram,omitempty"`
    MinLoad            float64 `json:"min_load"`
    TargetUtil         float64 `json:"target_util"`
    ColdMult           float64 `json:"cold_mult"`
    ColdWorkers        int     `json:"cold_workers"`
    TestWorkers        int     `json:"test_workers"`
    AutoscalerInstance string  `json:"autoscaler_instance"`
}

type WorkerGroup struct {
    ID           int     `json:"id"`
    EndpointName string  `json:"endpoint_name"`
    EndpointID   int     `json:"endpoint_id"`
    TemplateHash string  `json:"template_hash"`
    TemplateID   int     `json:"template_id"`
    SearchParams string  `json:"search_params"`
    LaunchArgs   string  `json:"launch_args"`
    GpuRAM       float64 `json:"gpu_ram"`
    MinLoad      float64 `json:"min_load"`
    TargetUtil   float64 `json:"target_util"`
    ColdMult     float64 `json:"cold_mult"`
    ColdWorkers  int     `json:"cold_workers"`
    TestWorkers  int     `json:"test_workers"`
}

type workerGroupListResponse struct {
    Success bool          `json:"success"`
    Results []WorkerGroup `json:"results"`
    Message string        `json:"msg,omitempty"`
}

func (s *WorkerGroupService) Create(ctx context.Context, req *CreateWorkerGroupRequest) (*WorkerGroup, error) {
    req.ClientID = "me"
    req.AutoscalerInstance = "prod"
    var resp map[string]interface{}
    if err := s.client.Post(ctx, "/autojobs/", req, &resp); err != nil {
        return nil, fmt.Errorf("creating worker group: %w", err)
    }
    // Create-then-read: list to find the new worker group
    groups, err := s.List(ctx)
    if err != nil {
        return nil, fmt.Errorf("reading worker group after create: %w", err)
    }
    // Find by highest ID (most recently created)
    var newest *WorkerGroup
    for i := range groups {
        if newest == nil || groups[i].ID > newest.ID {
            newest = &groups[i]
        }
    }
    if newest == nil {
        return nil, fmt.Errorf("worker group not found after creation")
    }
    return newest, nil
}

func (s *WorkerGroupService) Delete(ctx context.Context, id int) error {
    path := fmt.Sprintf("/autojobs/%d/", id)
    body := map[string]interface{}{
        "client_id":   "me",
        "autojob_id":  id,
    }
    if err := s.client.DeleteWithBody(ctx, path, body, nil); err != nil {
        return fmt.Errorf("deleting worker group %d: %w", id, err)
    }
    return nil
}
```

### Resource Registration Pattern
```go
// In provider.go -- add imports:
"github.com/realnedsanders/terraform-provider-vastai/internal/services/endpoint"
"github.com/realnedsanders/terraform-provider-vastai/internal/services/workergroup"

// In Resources():
endpoint.NewEndpointResource,
workergroup.NewWorkerGroupResource,

// In DataSources():
endpoint.NewEndpointsDataSource,
```

## Timeout Defaults Recommendation (Claude's Discretion)

| Operation | Default | Rationale |
|-----------|---------|-----------|
| Endpoint Create | 5 min | Server-side creation; no polling needed |
| Endpoint Read | 2 min | Simple list + filter |
| Endpoint Update | 5 min | May trigger autoscaler reconfiguration |
| Endpoint Delete | 5 min | May need to wait for worker drain |
| Worker Group Create | 10 min | May trigger initial test_workers spin-up |
| Worker Group Read | 2 min | Simple list + filter |
| Worker Group Update | 5 min | May trigger worker reprovisioning |
| Worker Group Delete | 5 min | Note: does NOT auto-destroy instances |

## Data Source Depth Recommendation (Claude's Discretion: D-03)

Expose all fields from the list response except sensitive/internal ones (`api_key`, `auto_delete_in_seconds`, `auto_delete_due_24h`). The `vastai_endpoints` data source should include:
- Endpoint metadata: id, endpoint_name, endpoint_state
- Autoscaling config: min_load, min_cold_load, target_util, cold_mult, cold_workers, max_workers

This gives users enough information for monitoring and conditional logic without exposing internal API details. Worker health/load details are not available from the list API response and would require separate log endpoints (out of scope).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single combined "autoscale" resource | Separate endpoint + worker group | Current API design | Resources are naturally separate in the API |
| Autogroup as separate entity | Autoscaling config on endpoint | No change detected | `create__autogroup` never existed in SDK |

## Open Questions

1. **Exact create response format for endpoints and worker groups**
   - What we know: SDK does `r.json()` and prints result. Response likely includes `{"success": true, "id": 123}` or similar.
   - What's unclear: Whether create returns the full object or just an ID+success.
   - Recommendation: Implement create-then-read pattern (create, then list to get full object). This is already the proven pattern from volumes (Phase 3).

2. **GET with JSON body behavior**
   - What we know: SDK sends `{"client_id": "me", "api_key": "<key>"}` in GET request body for list operations.
   - What's unclear: Whether the API requires this body or if Bearer auth header is sufficient.
   - Recommendation: Try with just Bearer auth first (consistent with all other list operations in the provider). Fall back to POST for list if needed.

3. **Worker group - endpoint binding strength**
   - What we know: Worker groups reference endpoints via both `endpoint_name` and `endpoint_id`.
   - What's unclear: Whether changing endpoint_name on an existing endpoint breaks worker group bindings.
   - Recommendation: Use `endpoint_id` as the primary binding (immutable after creation). Make `endpoint_id` RequiresReplace on worker groups to be safe.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | terraform-plugin-testing v1.15.0 + Go testing |
| Config file | go.mod (tests run via `go test`) |
| Quick run command | `go test ./internal/services/endpoint/... ./internal/services/workergroup/... ./internal/client/... -count=1 -v` |
| Full suite command | `go test ./... -count=1 -v` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SRVL-01 | Endpoint CRUD + autoscaling params | unit | `go test ./internal/services/endpoint/... -run TestEndpoint -count=1 -v` | Wave 0 |
| SRVL-01 | Endpoint client methods | unit | `go test ./internal/client/... -run TestEndpoint -count=1 -v` | Wave 0 |
| SRVL-02 | Worker group CRUD + bindings | unit | `go test ./internal/services/workergroup/... -run TestWorkerGroup -count=1 -v` | Wave 0 |
| SRVL-02 | Worker group client methods | unit | `go test ./internal/client/... -run TestWorkerGroup -count=1 -v` | Wave 0 |
| SRVL-03 | Autoscaling via endpoint (no standalone resource) | unit | Covered by SRVL-01 endpoint tests | Wave 0 |
| DATA-09 | Endpoints data source | unit | `go test ./internal/services/endpoint/... -run TestEndpointsDataSource -count=1 -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/services/endpoint/... ./internal/services/workergroup/... ./internal/client/... -count=1 -v`
- **Per wave merge:** `go test ./... -count=1 -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/services/endpoint/resource_endpoint_test.go` -- covers SRVL-01 schema validation
- [ ] `internal/services/endpoint/data_source_endpoints_test.go` -- covers DATA-09 schema
- [ ] `internal/services/workergroup/resource_worker_group_test.go` -- covers SRVL-02 schema
- [ ] `internal/client/endpoints_test.go` -- covers endpoint API methods
- [ ] `internal/client/worker_groups_test.go` -- covers worker group API methods

## Project Constraints (from CLAUDE.md)

- **Language:** Go -- required by Terraform provider ecosystem
- **Framework:** Terraform Plugin Framework v1.19.0 -- HashiCorp's current recommendation
- **API Reference:** Python SDK at `vast-sdk/` is the source of truth for API behavior
- **Testing:** Dual strategy -- httptest mocks for client, schema unit tests for resources, TF_ACC for acceptance
- **Auth:** Bearer header authentication via `VASTAI_API_KEY`
- **Service pattern:** Service-per-directory (`internal/services/<resource>/`), service sub-objects on client
- **Naming:** Always snake_case for Terraform attributes
- **Schema quality:** Optional+Computed for server defaults, comprehensive validators, per-resource models

## Sources

### Primary (HIGH confidence)
- `vast-sdk/vastai/vast.py` lines 2233-2325 (create__workergroup, create__endpoint)
- `vast-sdk/vastai/vast.py` lines 2914-2969 (delete__workergroup, delete__endpoint)
- `vast-sdk/vastai/vast.py` lines 5003-5062 (show__workergroups, show__endpoints)
- `vast-sdk/vastai/vast.py` lines 6641-6723 (update__workergroup, update__endpoint)
- `vast-sdk/vastai/vast.py` lines 3228-3300 (get__endpt_logs, get__wrkgrp_logs)
- `internal/client/client.go` -- VastAIClient service registration pattern
- `internal/client/instances.go` -- InstanceService CRUD pattern
- `internal/client/volumes.go` -- VolumeService + create-then-read pattern
- `internal/services/instance/resource_instance.go` -- Complex resource pattern
- `internal/services/volume/resource_volume.go` -- CRUD resource pattern
- `internal/services/volume/models.go` -- Model separation pattern
- `internal/provider/provider.go` -- Resource/data source registration

### Secondary (MEDIUM confidence)
- SDK help text for autoscaling parameter descriptions and defaults
- SDK argparse definitions for parameter types and default values

### Tertiary (LOW confidence)
- Exact response format from endpoint/worker group create operations (SDK only prints `r.json()`, does not document response structure)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all patterns established in prior phases
- Architecture: HIGH -- API contracts extracted directly from Python SDK source code
- Pitfalls: HIGH -- identified from SDK code analysis and prior phase patterns
- API response format: MEDIUM -- list responses verified (`{"success":true,"results":[...]}`), create response format inferred

**Research date:** 2026-03-27
**Valid until:** 2026-04-27 (stable -- Vast.ai API unlikely to change rapidly)
