---
phase: 02-core-compute
plan: 01
subsystem: api
tags: [go, vastai, api-client, httptest, instances, offers, templates, ssh-keys]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: VastAIClient with HTTP methods (Get/Post/Put/Delete), retry, Bearer auth
provides:
  - InstanceService with full CRUD + lifecycle (start/stop/bid/label) + WaitForStatus poller
  - OfferService with structured search and raw query passthrough
  - TemplateService with CRUD and body-carrying DELETE
  - SSHKeyService with CRUD + instance attach/detach
  - DeleteWithBody method on VastAIClient
  - ProtoV6ProviderFactories in acctest package
  - terraform-plugin-framework-validators, timeouts, testing dependencies
affects: [02-02, 02-03, 02-04, 02-05, 02-06, 03-storage, 04-serverless, 05-account]

# Tech tracking
tech-stack:
  added: [terraform-plugin-framework-validators v0.19.0, terraform-plugin-framework-timeouts v0.5.0, terraform-plugin-testing v1.15.0]
  patterns: [service-sub-objects on VastAIClient, httptest mock servers for client testing, typed request/response structs with JSON tags]

key-files:
  created:
    - internal/client/instances.go
    - internal/client/offers.go
    - internal/client/templates.go
    - internal/client/ssh_keys.go
    - internal/acctest/helpers.go
    - internal/client/instances_test.go
    - internal/client/offers_test.go
    - internal/client/templates_test.go
    - internal/client/ssh_keys_test.go
  modified:
    - internal/client/client.go
    - internal/client/client_test.go
    - go.mod
    - go.sum

key-decisions:
  - "Service sub-objects pattern: VastAIClient.Instances, .Offers, .Templates, .SSHKeys initialized in constructor"
  - "GPU RAM conversion: OfferSearchParams.GPURamGB * 1000 = MB for API (Pitfall 6 from research)"
  - "Template delete uses DeleteWithBody (hash_id in body, not URL path) per Pitfall 5"
  - "WaitForStatus treats 404 as success when target is 'destroyed', checks for terminal 'exited' state"
  - "5-second poll interval for instance status waiter"

patterns-established:
  - "Client service pattern: each domain gets a *Service struct with client *VastAIClient field"
  - "Typed request/response structs with snake_case JSON tags matching Vast.ai API"
  - "httptest mock server pattern: create server, verify method/path/body/auth, return JSON response"
  - "Pointer types for optional filter params (nil = not set)"

requirements-completed: [COMP-05, TEST-02]

# Metrics
duration: 6min
completed: 2026-03-25
---

# Phase 2 Plan 1: API Client Service Layer Summary

**Go API client services for instances, offers, templates, and SSH keys with full CRUD, lifecycle management, status polling, and 40 unit tests using httptest mocks**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-25T21:48:05Z
- **Completed:** 2026-03-25T21:54:31Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments
- Four typed client service files covering all Phase 2 API operations (instances, offers, templates, SSH keys)
- Instance lifecycle: Create from offer, Get/List, Start/Stop, Destroy, SetLabel, ChangeBid, UpdateTemplate, WaitForStatus with polling
- Offer search with structured filters (GPU name, count, RAM GB-to-MB conversion, price, datacenter, region) and raw query passthrough
- 40 unit tests passing against httptest mock servers covering all methods including edge cases (timeout, 404 on destroy, terminal state)

## Task Commits

Each task was committed atomically:

1. **Task 1: Install dependencies and create API client service files** - `48ad9a9` (feat)
2. **Task 2: Unit tests for all client services using httptest mocks** - `2bc2d3a` (test)

## Files Created/Modified
- `internal/client/instances.go` - InstanceService with full CRUD, lifecycle, and WaitForStatus poller
- `internal/client/offers.go` - OfferService with structured search and GPU RAM MB conversion
- `internal/client/templates.go` - TemplateService with CRUD and body-carrying DELETE
- `internal/client/ssh_keys.go` - SSHKeyService with CRUD and instance attach/detach
- `internal/acctest/helpers.go` - ProtoV6ProviderFactories for acceptance tests
- `internal/client/client.go` - Added service sub-objects and DeleteWithBody method
- `internal/client/instances_test.go` - 13 tests for instance operations and waiter logic
- `internal/client/offers_test.go` - 4 tests for offer search with filter verification
- `internal/client/templates_test.go` - 4 tests for template CRUD
- `internal/client/ssh_keys_test.go` - 6 tests for SSH key operations
- `internal/client/client_test.go` - Added TestDeleteWithBody
- `go.mod` / `go.sum` - Added validators, timeouts, testing dependencies

## Decisions Made
- Service sub-objects initialized in NewVastAIClient constructor for ergonomic access (client.Instances.Create)
- GPU RAM conversion: user specifies GB in params, client multiplies by 1000 for API's MB format
- Template delete implemented via DeleteWithBody (hash_id in request body, not URL path) per Vast.ai API contract
- WaitForStatus uses 5-second poll interval, context-based timeout, detects terminal "exited" state to fail fast
- Pointer types for optional OfferSearchParams fields so nil means "omit from query"

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None - all services are fully implemented with typed structs and methods.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- All four client service files are ready for Terraform resource/data-source wiring in plans 02-02 through 02-06
- ProtoV6ProviderFactories available for acceptance tests
- Dependencies installed: validators for plan-time validation, timeouts for resource timeouts, testing for acceptance tests

## Self-Check: PASSED

All 10 created files verified present on disk. Both task commits (48ad9a9, 2bc2d3a) verified in git history.

---
*Phase: 02-core-compute*
*Completed: 2026-03-25*
