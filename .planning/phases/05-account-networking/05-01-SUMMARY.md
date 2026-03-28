---
phase: 05-account-networking
plan: 01
subsystem: api
tags: [go, http-client, api-keys, env-vars, teams, subaccounts, json-raw-message]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: VastAIClient base with auth, retry, Get/Post/Put/Delete/DeleteWithBody methods
provides:
  - ApiKeyService with Create, List, Delete for API key management
  - EnvVarService with Create, List, Update, Delete (DeleteWithBody pattern)
  - TeamService with team CRUD, role CRUD, member invite/list/remove
  - SubaccountService with Create, List
  - GetFullPath client method for non-v0 API endpoints (e.g., /api/v1/invoices/)
  - newRequestFullPath in auth.go for full-path URL construction
affects: [05-02, 05-03, 05-04, 05-05, 05-06]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "json.RawMessage for nested permissions JSON (API keys and team roles)"
    - "DeleteWithBody for env var deletion (key in body, not URL)"
    - "Query parameter encoding in URL path for team invite (not JSON body)"
    - "newRequestFullPath for non-v0 API version endpoints"

key-files:
  created:
    - internal/client/api_keys.go
    - internal/client/api_keys_test.go
    - internal/client/env_vars.go
    - internal/client/env_vars_test.go
    - internal/client/teams.go
    - internal/client/teams_test.go
    - internal/client/subaccounts.go
    - internal/client/subaccounts_test.go
  modified:
    - internal/client/client.go
    - internal/client/auth.go

key-decisions:
  - "TeamRole permissions use json.RawMessage for nested JSON objects (D-02 revised from flat string set)"
  - "Team invite uses query params in URL path, not JSON body (Pitfall 5)"
  - "Only 4 account service sub-objects added; remaining 5 deferred to Plan 05-02"

patterns-established:
  - "json.RawMessage for evolving/nested JSON API fields (permissions)"
  - "GetFullPath/newRequestFullPath for non-v0 API endpoints"

requirements-completed: [ACCT-01, ACCT-02, ACCT-03, ACCT-04, ACCT-05, ACCT-06]

# Metrics
duration: 4min
completed: 2026-03-28
---

# Phase 5 Plan 1: Account Client Services Summary

**Go API client services for API keys, env vars, teams (roles/members), and subaccounts with GetFullPath for v1 API support**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-28T00:18:13Z
- **Completed:** 2026-03-28T00:22:44Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- 4 new client service files covering all 6 account requirements (ACCT-01 through ACCT-06)
- Full unit test coverage with httptest mock servers for success and error paths
- GetFullPath client method enabling /api/v1/ invoices endpoint in later plans
- All 4 account service sub-objects wired into VastAIClient constructor

## Task Commits

Each task was committed atomically:

1. **Task 1: Client services for account APIs** - `649ecce` (feat)
2. **Task 2: Auth and client wiring (GetFullPath + service sub-objects)** - `19bafb5` (feat)

## Files Created/Modified
- `internal/client/api_keys.go` - ApiKeyService with Create, List, Delete (keys immutable after creation)
- `internal/client/api_keys_test.go` - Unit tests for API key operations
- `internal/client/env_vars.go` - EnvVarService with Create, List, Update, Delete using DeleteWithBody
- `internal/client/env_vars_test.go` - Unit tests for env var operations including DeleteWithBody verification
- `internal/client/teams.go` - TeamService with team CRUD, role CRUD (ID for update, name for delete), member invite/list/remove
- `internal/client/teams_test.go` - Unit tests for all team, role, and member operations
- `internal/client/subaccounts.go` - SubaccountService with Create, List (no delete endpoint per API)
- `internal/client/subaccounts_test.go` - Unit tests for subaccount operations
- `internal/client/client.go` - Added 4 account service sub-objects and GetFullPath method
- `internal/client/auth.go` - Added newRequestFullPath for non-v0 API endpoints

## Decisions Made
- TeamRole permissions stored as json.RawMessage (D-02 revised) -- API uses nested JSON objects, not flat string sets
- Team invite encodes email/role as URL query parameters, not JSON body (Pitfall 5)
- Only added 4 account service sub-objects to client.go; remaining 5 (Clusters, Overlays, Users, Invoices, AuditLogs) deferred to Plan 05-02 to avoid compile errors from missing types
- Subaccounts have no Delete method since no API endpoint exists (Pitfall 8)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Known Stubs
None - all services are fully implemented with real API contracts.

## Next Phase Readiness
- Account client services ready for Terraform resource layer (Plans 05-03, 05-04, 05-05)
- GetFullPath ready for invoices data source (Plan 05-06)
- Plan 05-02 should add remaining 5 service sub-objects (Clusters, Overlays, Users, Invoices, AuditLogs) to client.go

---
*Phase: 05-account-networking*
*Completed: 2026-03-28*
