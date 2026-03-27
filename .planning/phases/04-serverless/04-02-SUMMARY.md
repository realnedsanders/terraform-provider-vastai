---
phase: 04-serverless
plan: 02
subsystem: serverless
tags: [terraform, endpoint, autoscaling, data-source, crud, import]

# Dependency graph
requires:
  - phase: 04-serverless/04-01
    provides: "EndpointService client with Create/List/Update/Delete methods"
provides:
  - "vastai_endpoint resource with full CRUD, import, autoscaling validators"
  - "vastai_endpoints data source listing all endpoints"
  - "EndpointResourceModel, EndpointsDataSourceModel, EndpointModel types"
affects: [04-serverless/04-03, 06-documentation]

# Tech tracking
tech-stack:
  added: []
  patterns: [endpoint-resource-crud, data-source-list-pattern, autoscaling-validators]

key-files:
  created:
    - internal/services/endpoint/models.go
    - internal/services/endpoint/resource_endpoint.go
    - internal/services/endpoint/data_source_endpoints.go
    - internal/services/endpoint/resource_endpoint_test.go
    - internal/services/endpoint/data_source_endpoints_test.go
  modified:
    - internal/provider/provider.go

key-decisions:
  - "SRVL-03 satisfied via endpoint autoscaling config -- no separate autogroup resource needed"
  - "Endpoint state change after creation handled via post-create Update call (Pitfall 6)"
  - "Read timeout set to 2m (vs 5m for CUD) since listing is lightweight"

patterns-established:
  - "Endpoint resource CRUD: create-then-read via List, read-after-write on Update"
  - "Data source list pattern: single computed ListNestedAttribute with type conversion helpers"

requirements-completed: [SRVL-01, SRVL-03, DATA-09]

# Metrics
duration: 4min
completed: 2026-03-27
---

# Phase 4 Plan 2: Endpoint Resource and Data Source Summary

**vastai_endpoint resource with CRUD, import, autoscaling validators (target_util Between(0,1), cold_mult AtLeast(1.0)) and vastai_endpoints data source, all with 15 passing schema unit tests**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-27T20:24:48Z
- **Completed:** 2026-03-27T20:28:51Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Endpoint resource with full CRUD lifecycle, import support, and autoscaling parameter validators per D-04/D-05
- Endpoints data source listing all user endpoints with complete metadata
- Both registered in provider.go and all 15 schema unit tests passing
- SRVL-03 satisfied: autoscaling configuration lives on the endpoint resource (no separate autogroup resource)

## Task Commits

Each task was committed atomically:

1. **Task 1: Endpoint resource, data source, models, and provider registration** - `596832c` (feat)
2. **Task 2: Unit tests for endpoint resource and data source schemas** - `9480129` (test)

## Files Created/Modified
- `internal/services/endpoint/models.go` - EndpointResourceModel, EndpointsDataSourceModel, EndpointModel types
- `internal/services/endpoint/resource_endpoint.go` - vastai_endpoint resource with CRUD, import, autoscaling validators
- `internal/services/endpoint/data_source_endpoints.go` - vastai_endpoints data source with List API call and type conversion
- `internal/services/endpoint/resource_endpoint_test.go` - 10 schema unit tests for endpoint resource
- `internal/services/endpoint/data_source_endpoints_test.go` - 6 schema unit tests for endpoints data source (including nested attribute validation)
- `internal/provider/provider.go` - Added endpoint resource and data source registrations

## Decisions Made
- SRVL-03 (autogroup) is satisfied by endpoint autoscaling config on the resource itself -- no separate resource needed
- Post-create Update call for endpoint_state when user requests non-active state (per Pitfall 6: new endpoints always start active)
- Read timeout set to 2 minutes (lighter than 5-minute CUD timeouts) since List is a single GET

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None - all data paths are fully wired to the API client.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Endpoint resource and data source are complete and registered
- Ready for Plan 04-03 (worker group resource) which depends on endpoint_id from this resource
- All existing tests continue to pass across all packages

## Self-Check: PASSED

All 5 created files verified on disk. Both task commits (596832c, 9480129) verified in git log.

---
*Phase: 04-serverless*
*Completed: 2026-03-27*
