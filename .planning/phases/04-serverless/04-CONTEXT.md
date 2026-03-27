# Phase 4: Serverless - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Serverless inference endpoint lifecycle: endpoint CRUD, worker group binding with template/search config, autoscaling groups, and endpoint status data source. Reuses all established schema quality patterns.

</domain>

<decisions>
## Implementation Decisions

### Endpoint/Worker Resource Model
- **D-01:** Claude's Discretion on resource separation (all separate vs combined) — based on API structure and dependency patterns
- **D-02:** Explicit deletion required — worker groups must be deleted before endpoint. Terraform manages dependencies, no cascade.
- **D-03:** Claude's Discretion on endpoint status data source depth — metadata only vs full worker health/load details based on API response

### Autoscaling Parameters
- **D-04:** Some params required (min_load, max_workers — must-think-about), rest optional with sensible defaults
- **D-05:** Strict validation ranges on all autoscaling params — target_util 0-1, cold_mult >= 1, workers >= 0, etc. Catch bad configs at plan time.

### Inherited from Prior Phases (DO NOT re-decide)
- Service-per-directory, service pattern on client
- Always snake_case, Optional+Computed for server defaults
- Comprehensive validators, per-resource models, API models in client
- Dual test strategy: httptest mocks + TF_ACC acceptance tests

### Claude's Discretion
- Resource model separation (endpoint, worker group, autogroup)
- Endpoint status data source depth
- Specific default values for optional autoscaling params
- Timeout defaults for endpoint/worker operations

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Vast.ai API Reference
- `vast-sdk/vastai/vast.py` — Serverless operations: `create__endpoint`, `delete__endpoint`, `show__endpoints`, `update__endpoint`, `create__workergroup`, `delete__workergroup`, `show__workergroups`, `update__workergroup`, `create__autogroup`, `get__wrkgrp_logs`, `get__endpt_logs`

### Existing Provider Code
- `internal/services/instance/resource_instance.go` — Complex resource pattern (lifecycle, updates, import)
- `internal/services/template/resource_template.go` — Simpler CRUD resource pattern
- `internal/client/instances.go` — Client service with typed methods

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- VastAIClient service pattern — add Endpoints, WorkerGroups services
- Resource/data source patterns from Phase 2/3 — replicate for serverless resources
- `internal/acctest/helpers.go` for acceptance tests

### Integration Points
- New service directories: `internal/services/endpoint/`, `internal/services/workergroup/`
- New client services: `internal/client/endpoints.go`, `internal/client/worker_groups.go`
- Register in provider.go

</code_context>

<specifics>
## Specific Ideas

- Worker groups depend on endpoints — Terraform's dependency graph handles this naturally via endpoint_id reference
- Autogroup may be part of endpoint or worker group creation rather than a standalone resource — research needs to clarify
- Endpoint status should include enough info to be useful for monitoring/conditional logic in Terraform configs

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-serverless*
*Context gathered: 2026-03-27*
