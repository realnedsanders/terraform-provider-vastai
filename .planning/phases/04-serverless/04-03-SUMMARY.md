---
phase: 04-serverless
plan: 03
subsystem: serverless
tags: [terraform, worker-group, serverless, vastai, gpu-compute]

# Dependency graph
requires:
  - phase: 04-01
    provides: Worker group client service with Create/List/Update/Delete API methods
  - phase: 04-02
    provides: Endpoint resource pattern and provider registration structure
provides:
  - vastai_worker_group resource with full CRUD, import, and endpoint binding
  - Worker group schema unit tests with negative autoscaling params test
affects: [06-documentation, acceptance-tests]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "AtLeastOneOf cross-validation for mutually optional template fields"
    - "Omitting unused API fields from schema (Pitfall 3: autoscaling at endpoint level)"

key-files:
  created:
    - internal/services/workergroup/models.go
    - internal/services/workergroup/resource_worker_group.go
    - internal/services/workergroup/resource_worker_group_test.go
  modified:
    - internal/provider/provider.go

key-decisions:
  - "Omit min_load/target_util/cold_mult from worker group schema (Pitfall 3: not used at workergroup level, autoscaling driven by endpoint)"
  - "Send sensible autoscaling defaults (MinLoad=0, TargetUtil=0.9, ColdMult=2.0) in Create since API requires them"
  - "AtLeastOneOf cross-validation on template_hash/template_id (at least one must be provided)"

patterns-established:
  - "AtLeastOneOf: Use stringvalidator.AtLeastOneOf and int64validator.AtLeastOneOf for cross-field validation between string and int64 attributes"
  - "Schema omission: When API accepts fields that have no effect at a resource level, omit from schema and document why"

requirements-completed: [SRVL-02]

# Metrics
duration: 4min
completed: 2026-03-27
---

# Phase 4 Plan 3: Worker Group Resource Summary

**vastai_worker_group resource with CRUD, import, endpoint binding (ForceNew), and template AtLeastOneOf validation -- autoscaling params correctly omitted per Pitfall 3**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-27T20:32:01Z
- **Completed:** 2026-03-27T20:36:03Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Worker group resource with full CRUD, import support, and endpoint binding via RequiresReplace
- AtLeastOneOf cross-validation ensures at least one of template_hash or template_id is provided
- Autoscaling params (min_load, target_util, cold_mult) intentionally omitted from schema -- documented that autoscaling is endpoint-level
- 12 schema unit tests including negative test confirming no autoscaling params in schema
- Provider now registers 7 resources and 8 data sources for complete Phase 4 coverage

## Task Commits

Each task was committed atomically:

1. **Task 1: Worker group resource, models, and provider registration** - `336b356` (feat)
2. **Task 2: Unit tests for worker group resource schema** - `4189ab2` (test)

## Files Created/Modified
- `internal/services/workergroup/models.go` - WorkerGroupResourceModel with endpoint binding, template config, search params
- `internal/services/workergroup/resource_worker_group.go` - vastai_worker_group resource with CRUD, import, RequiresReplace, AtLeastOneOf
- `internal/services/workergroup/resource_worker_group_test.go` - 12 schema unit tests including negative autoscaling test
- `internal/provider/provider.go` - Added workergroup.NewWorkerGroupResource to Resources()

## Decisions Made
- Omitted min_load/target_util/cold_mult from worker group schema per Pitfall 3 (not used at workergroup level; autoscaling driven by endpoint)
- Send sensible autoscaling defaults in Create (MinLoad=0, TargetUtil=0.9, ColdMult=2.0) since API requires these fields in the request body
- Used AtLeastOneOf cross-validation on template_hash/template_id to enforce at least one is provided

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Known Stubs
None - all data flows are wired to the API client.

## Next Phase Readiness
- Phase 4 (Serverless) is complete: endpoint resource (04-02), worker group resource (04-03), and client services (04-01)
- Provider now registers 7 resources (instance, template, sshkey, volume, networkvolume, endpoint, worker_group) and 8 data sources
- Ready for Phase 5 (Account & Networking) or Phase 6 (Documentation & Release)

## Self-Check: PASSED

All files exist. All commits verified.

---
*Phase: 04-serverless*
*Completed: 2026-03-27*
