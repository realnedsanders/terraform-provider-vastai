# Phase 3: Storage - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Persistent volume and network volume resources with full CRUD, marketplace listing, cloning, and offer search data sources. Reuses all schema quality patterns established in Phase 2.

</domain>

<decisions>
## Implementation Decisions

### Volume Lifecycle
- **D-01:** Volume cloning via `clone_from_id` optional attribute on the volume resource — when set, creates by cloning instead of from offer. ForceNew on change.
- **D-02:** Claude's Discretion on marketplace list/unlist modeling (boolean attribute vs separate resource) — based on API behavior fit
- **D-03:** Separate resources for volumes and network volumes: `vastai_volume` and `vastai_network_volume` — different API paths, different capabilities

### Offer Search
- **D-04:** Claude's Discretion on offer search pattern — mirror GPU offers structure where API supports it, simplify where volume search API is more limited
- **D-05:** Claude's Discretion on whether volume and network volume offers are separate or combined data sources — based on API endpoint differences

### Inherited from Phase 1 & 2 (DO NOT re-decide)
- Service-per-directory: `internal/services/volume/`, `internal/services/networkvolume/`
- Service pattern on client: `client.Volumes.Create()`, `client.NetworkVolumes.Create()`
- Always snake_case attributes with Optional+Computed for server defaults
- Comprehensive plan-time validators
- Per-resource models, API models in `internal/client/`
- Set vs List per-attribute, Vast.ai doc URLs in descriptions
- Dual test strategy: httptest mocks + TF_ACC acceptance tests
- Full acceptance test coverage for all resources

### Claude's Discretion
- Marketplace list/unlist modeling approach
- Offer search data source structure and filter set
- Whether offers are separate or combined data sources
- Timeout defaults for volume operations

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Vast.ai API Reference
- `vast-sdk/vastai/vast.py` — Volume operations: `create__volume`, `delete__volume`, `clone__volume`, `list__volume`, `unlist__volume`, `search__volumes`, `create__network_volume`, `list__network_volume`, `unlist__network_volume`, `search__network_volumes`

### Existing Provider Code (Phase 2 patterns to replicate)
- `internal/services/offer/data_source_gpu_offers.go` — GPU offer search pattern (filter structure, most_affordable, validators)
- `internal/services/template/resource_template.go` — Resource with full CRUD, import, timeouts pattern
- `internal/services/instance/resource_instance.go` — Complex resource pattern (lifecycle, preemption)
- `internal/client/offers.go` — Offer search service pattern
- `internal/client/instances.go` — Service with typed methods pattern

### Phase 2 Context
- `.planning/phases/02-core-compute/02-CONTEXT.md` — Schema conventions (D-14 through D-23)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `VastAIClient` with service sub-objects — add `Volumes` and `NetworkVolumes` services
- GPU offers data source pattern — replicate for volume offer search
- Template resource pattern — replicate for volume resources (CRUD + import + timeouts)
- `internal/acctest/helpers.go` — ProtoV6ProviderFactories for acceptance tests

### Established Patterns
- Schema: snake_case, Optional+Computed, validators, descriptions, plan modifiers
- Client: typed service structs, httptest mocks, error handling
- Testing: unit tests per-service + TF_ACC acceptance tests

### Integration Points
- New service directories: `internal/services/volume/`, `internal/services/networkvolume/`
- New client services: `internal/client/volumes.go`, `internal/client/network_volumes.go`
- Register in `internal/provider/provider.go` Resources() and DataSources()

</code_context>

<specifics>
## Specific Ideas

- Clone is modeled as a creation-time attribute (clone_from_id), not a separate operation — more Terraform-idiomatic
- Volumes and network volumes are separate resources because they have different API paths and capabilities
- Phase 2 patterns should be replicated closely — this is a "follow the template" phase

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-storage*
*Context gathered: 2026-03-27*
