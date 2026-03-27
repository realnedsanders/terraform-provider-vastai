---
phase: 04-serverless
plan: 01
subsystem: api
tags: [go, rest-client, serverless, endpoints, worker-groups, httptest]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: "VastAIClient with HTTP methods (Get, Post, Put, Delete, DeleteWithBody), Bearer auth, retry"
  - phase: 02-core-compute
    provides: "Service sub-object pattern (InstanceService, OfferService, TemplateService, SSHKeyService)"
  - phase: 03-storage
    provides: "VolumeService and NetworkVolumeService extending VastAIClient"
provides:
  - "EndpointService with Create, List, Update, Delete for /endptjobs/ API"
  - "WorkerGroupService with Create, List, Update, Delete for /autojobs/ API"
  - "VastAIClient.Endpoints and VastAIClient.WorkerGroups service fields"
  - "Typed request/response structs for serverless API operations"
affects: [04-02, 04-03, 06-documentation]

# Tech tracking
tech-stack:
  added: []
  patterns: [create-then-read-by-name, create-then-read-by-highest-id, delete-with-json-body, autoscaler-instance-injection]

key-files:
  created:
    - internal/client/endpoints.go
    - internal/client/worker_groups.go
    - internal/client/endpoints_test.go
    - internal/client/worker_groups_test.go
  modified:
    - internal/client/client.go

key-decisions:
  - "Endpoint create-then-read searches by name (backwards iteration for most recent match)"
  - "Worker group create-then-read finds by highest ID (most recently created)"
  - "autoscaler_instance always set to 'prod' internally, not exposed to users"
  - "Delete operations use DeleteWithBody pattern matching template delete convention"

patterns-established:
  - "Serverless service pattern: same VastAIClient service sub-object structure as Phase 2/3"
  - "Endpoint list response uses {success, results, msg} envelope (different from volume {volumes} pattern)"

requirements-completed: [SRVL-01, SRVL-02, SRVL-03, DATA-09]

# Metrics
duration: 4min
completed: 2026-03-27
---

# Phase 4 Plan 1: Serverless Client Services Summary

**Go API client services for serverless endpoints (/endptjobs/) and worker groups (/autojobs/) with typed CRUD methods and 8 httptest unit tests**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-27T20:16:27Z
- **Completed:** 2026-03-27T20:20:51Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- EndpointService with Create (POST + list readback by name), List, Update, Delete for /endptjobs/ API
- WorkerGroupService with Create (POST + list readback by highest ID), List, Update, Delete for /autojobs/ API
- VastAIClient extended with Endpoints and WorkerGroups service fields initialized in constructor
- 8 new unit tests covering all CRUD operations with httptest mock servers verifying auth headers, HTTP methods, paths, and request bodies

## Task Commits

Each task was committed atomically:

1. **Task 1: Endpoint and worker group client services with all API methods** - `09e580b` (feat)
2. **Task 2: Unit tests for endpoint and worker group client services** - `20e75ef` (test)

## Files Created/Modified
- `internal/client/endpoints.go` - EndpointService with Create/List/Update/Delete for serverless endpoints
- `internal/client/worker_groups.go` - WorkerGroupService with Create/List/Update/Delete for worker groups
- `internal/client/client.go` - Added Endpoints and WorkerGroups fields to VastAIClient struct and constructor
- `internal/client/endpoints_test.go` - 4 unit tests for EndpointService CRUD operations
- `internal/client/worker_groups_test.go` - 4 unit tests for WorkerGroupService CRUD operations

## Decisions Made
- Endpoint Create uses create-then-read with backwards name matching (most recent match wins), consistent with the list-based read pattern since no single-GET endpoint exists
- Worker group Create uses create-then-read with highest-ID selection (newest group), since worker group names are not unique identifiers
- autoscaler_instance is always injected as "prod" in Create and Update methods -- internal API detail never exposed to Terraform users (Pitfall 5)
- Delete operations use DeleteWithBody (not Delete) since the API requires JSON body with identifiers (Pitfall 2)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Client services ready for Plan 04-02 (endpoint Terraform resource and data source)
- Client services ready for Plan 04-03 (worker group Terraform resource)
- All existing tests pass (59 total including 8 new)

## Self-Check: PASSED

All 5 created files verified on disk. Both task commits (09e580b, 20e75ef) found in git log.

---
*Phase: 04-serverless*
*Completed: 2026-03-27*
