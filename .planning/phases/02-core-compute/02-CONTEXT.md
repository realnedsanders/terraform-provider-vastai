# Phase 2: Core Compute - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Complete compute workflow end-to-end: GPU offer search data source with rich filtering, instance resource with full lifecycle (create/start/stop/update/destroy), template resource, SSH key resource, and associated data sources. Establishes schema quality patterns (validators, plan modifiers, import) used by all subsequent phases.

</domain>

<decisions>
## Implementation Decisions

### Offer Search Data Source
- **D-01:** Claude's Discretion on filter approach — structured typed attributes for common filters (gpu_name, num_gpus, gpu_ram, price, region, datacenter_only) with optional raw_query escape hatch for power users
- **D-02:** Return all matching offers as a list, with a computed `most_affordable` convenience attribute for the cheapest one
- **D-03:** Configurable `order_by` attribute — user picks sort field (price, score, gpu_ram, etc.)
- **D-04:** Default limit of 10 results, configurable via `limit` attribute
- **D-05:** Standard Terraform behavior — data source re-queries live on every plan/apply (no caching)
- **D-06:** Claude's Discretion on offer expiry handling — error with guidance vs auto-retry based on Terraform's declarative model
- **D-07:** Full offer details exposed — GPU name, VRAM, CPU cores, RAM, disk, DL performance, reliability score, location, hosting type

### Instance Lifecycle
- **D-08:** Claude's Discretion on start/stop modeling (status attribute vs separate resource) based on cloud provider patterns
- **D-09:** Preemption handling: only silent removal from state when instance was ACTUALLY preempted (outbid/evicted). Successful exits on spot instances should NOT trigger silent removal — they should surface as a normal state change. Must distinguish preemption from normal termination via API status.
- **D-10:** Claude's Discretion on create timeout default based on GPU provisioning times
- **D-11:** Claude's Discretion on bid price updates (in-place vs replace) based on what the Vast.ai API supports
- **D-12:** Claude's Discretion on SSH key attachment model (inline vs separate resource vs both) based on Terraform patterns
- **D-13:** Claude's Discretion on immutable vs updatable attributes based on Vast.ai API behavior

### Schema Conventions
- **D-14:** Always snake_case for all Terraform attributes regardless of API naming. Mapping handled in model conversion.
- **D-15:** Server-set defaults use Optional+Computed classification — user can set them, otherwise server default stored in state. Prevents noisy diffs.
- **D-16:** Comprehensive plan-time validators — validate GPU names, regions, image format, port ranges, etc. Catch errors before API call.
- **D-17:** Per-resource model structs (InstanceModel, TemplateModel) — no coupling between resources
- **D-18:** API response/request types live in `internal/client/`, Terraform state models live in service directories — clean separation per standalone client decision (D-05 from Phase 1)
- **D-19:** Use SetAttribute where ordering doesn't matter (SSH keys, tags), ListAttribute where it does (env vars, ports). Prevents spurious diffs.
- **D-20:** Include Vast.ai API documentation URLs in attribute descriptions where helpful

### Acceptance Tests
- **D-21:** Dual strategy — httptest mocks for schema/logic validation (fast, free) + real API tests with cheapest available offers for integration testing
- **D-22:** Claude's Discretion on test parallelism based on Vast.ai API rate limit behavior
- **D-23:** Full acceptance test coverage for ALL resources in Phase 2 — every resource/data source gets create, read, update, import, destroy tests

### Claude's Discretion
- Offer filter approach details (structured + raw escape hatch balance)
- Offer expiry error handling strategy
- Start/stop modeling (status attribute vs separate resource)
- Create timeout default
- Bid price update behavior (in-place vs replace)
- SSH key attachment model
- Immutable vs updatable attribute classification
- Test parallelism strategy

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Vast.ai API Reference
- `vast-sdk/vastai/vast.py` — Primary API client, all endpoint paths, query params, response shapes. Critical sections: `search__offers` (line ~2800+), `create__instance` (line ~3000+), `show__instance`, `start__instance`, `stop__instance`, `create__template`, `create__ssh_key`, `attach__ssh`, `detach__ssh`
- `vast-sdk/vastai/vastai_sdk.py` — SDK wrapper with method signatures

### Existing Provider Code
- `internal/client/client.go` — VastAIClient with HTTP methods, retry logic, Bearer auth
- `internal/client/auth.go` — Request construction, auth header injection
- `internal/client/errors.go` — APIError type for structured diagnostics
- `internal/provider/provider.go` — Provider shell with schema, Configure method wiring client

### Project Research
- `.planning/research/STACK.md` — Tech stack with versions
- `.planning/research/ARCHITECTURE.md` — Provider architecture, service-per-directory pattern
- `.planning/research/PITFALLS.md` — Schema misconfiguration, async creation, preemption handling
- `.planning/research/FEATURES.md` — Feature priorities, dependency graph

### Phase 1 Context
- `.planning/phases/01-foundation/01-CONTEXT.md` — Decisions D-01 through D-15 (module path, client architecture, project layout, CI/CD)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `VastAIClient` in `internal/client/client.go` — HTTP client with Get/Post/Put/Delete, retry, auth. Phase 2 adds service sub-objects (Instances, Offers, Templates, SSHKeys) to this client.
- `APIError` in `internal/client/errors.go` — Structured error type. Use for Terraform diagnostic mapping.
- `VastaiProvider` in `internal/provider/provider.go` — Provider with Configure that creates client. Resources register here via `Resources()` and `DataSources()`.

### Established Patterns
- Bearer auth via `Authorization` header (never query param)
- 150ms/1.5x/5 retry on 429/5xx
- tflog at DEBUG (request/response status) and TRACE (response body)
- `VastaiProviderModel` with `tfsdk` struct tags for schema binding

### Integration Points
- New service directories: `internal/services/instance/`, `internal/services/template/`, `internal/services/sshkey/`
- New client services: `client.Instances`, `client.Offers`, `client.Templates`, `client.SSHKeys`
- Register resources in `provider.go` `Resources()` and `DataSources()` methods
- API models in `internal/client/` (e.g., `client/instances.go`, `client/offers.go`)

</code_context>

<specifics>
## Specific Ideas

- Preemption handling is nuanced: ONLY silent state removal for actual preemption (outbid/evicted). Normal spot instance exits should surface as state changes, not silent removal. The API status field must be checked to distinguish these cases.
- Offer search should be the "entry point" for all compute workflows — users search for offers first, then create instances from the best match.
- Validators should catch common mistakes at plan time (wrong GPU name, invalid region) rather than waiting for API errors at apply time.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-core-compute*
*Context gathered: 2026-03-25*
