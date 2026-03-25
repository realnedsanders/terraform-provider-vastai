---
phase: 02-core-compute
plan: 06
subsystem: testing
tags: [acceptance-tests, terraform-plugin-testing, TF_ACC, gpu-offers, instance, template, ssh-key]

# Dependency graph
requires:
  - phase: 02-core-compute/02-05
    provides: "All Phase 2 resources and data sources registered in provider"
provides:
  - "TF_ACC-gated acceptance tests for all 3 resources (instance, template, ssh_key)"
  - "TF_ACC-gated acceptance tests for all 5 data sources (gpu_offers, instance, instances, templates, ssh_keys)"
  - "CheckDestroy verification for instance cleanup"
  - "Cheapest-offer cost strategy for instance tests (D-21)"
affects: [06-documentation-release]

# Tech tracking
tech-stack:
  added: [terraform-plugin-testing v1.15.0]
  patterns: [acceptance-test-pattern, cheapest-offer-strategy, TF_ACC-gating]

key-files:
  created:
    - internal/services/sshkey/resource_ssh_key_acc_test.go
    - internal/services/template/resource_template_acc_test.go
    - internal/services/offer/data_source_gpu_offers_acc_test.go
    - internal/services/instance/resource_instance_acc_test.go
    - internal/services/instance/data_source_instance_acc_test.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "terraform-plugin-testing v1.15.0 added as test dependency for acceptance test framework"
  - "Instance tests use max $0.50/hr offer cap to minimize API costs per D-21"
  - "ImportStateVerifyIgnore lists exclude creation-only and timeout attributes"

patterns-established:
  - "Acceptance test naming: TestAcc{Resource}_{scenario} convention"
  - "Cost strategy: testInstanceBaseConfig constant with cheapest GPU offer data source"
  - "Data source tests: create resource first, then verify data source reads it"

requirements-completed: [TEST-01]

# Metrics
duration: 3min
completed: 2026-03-25
---

# Phase 2 Plan 6: Acceptance Tests Summary

**TF_ACC-gated acceptance tests covering full CRUD/import lifecycle for all Phase 2 resources (instance, template, SSH key) and read verification for all 5 data sources, using terraform-plugin-testing v1.15.0**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-25T22:37:48Z
- **Completed:** 2026-03-25T22:41:00Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Full acceptance test coverage for SSH key, template, and GPU offers (Task 1: low-cost, fast resources)
- Full acceptance test coverage for instance resource and data sources (Task 2: higher-cost lifecycle tests)
- All 15 test functions properly gated behind TF_ACC=1 environment variable
- Instance tests use cheapest available offer ($0.50/hr cap) per D-21 cost minimization strategy

## Task Commits

Each task was committed atomically:

1. **Task 1: Acceptance tests for SSH key, template, and GPU offers** - `3730d86` (test)
2. **Task 2: Acceptance tests for instance resource and data sources** - `8e28f81` (test)

## Files Created/Modified
- `internal/services/sshkey/resource_ssh_key_acc_test.go` - SSH key CRUD, update, import, and data source acceptance tests (4 test functions)
- `internal/services/template/resource_template_acc_test.go` - Template CRUD, update, import, and data source search acceptance tests (4 test functions)
- `internal/services/offer/data_source_gpu_offers_acc_test.go` - GPU offers basic and filtered search acceptance tests (2 test functions)
- `internal/services/instance/resource_instance_acc_test.go` - Instance create/destroy, label update, import acceptance tests with CheckDestroy (3 test functions)
- `internal/services/instance/data_source_instance_acc_test.go` - Instance lookup by ID and instances list with label filter acceptance tests (2 test functions)
- `go.mod` - Added terraform-plugin-testing v1.15.0 dependency
- `go.sum` - Updated dependency checksums

## Decisions Made
- Added terraform-plugin-testing v1.15.0 as the acceptance test framework -- this is HashiCorp's official testing module, required for resource.Test with ProtoV6ProviderFactories
- Instance acceptance tests cap offer price at $0.50/hr and use 15m create / 5m delete timeouts to balance cost vs reliability
- ImportStateVerifyIgnore lists are specific per resource: timeouts for SSH key/template, plus offer_id/disk_gb/ssh_key_ids/image_login/cancel_unavail/env/use_ssh/use_jupyter_lab for instance (attributes not returned from API read or only set at creation)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added missing terraform-plugin-testing dependency**
- **Found during:** Task 1 (SSH key acceptance tests)
- **Issue:** terraform-plugin-testing v1.15.0 not in go.mod; acceptance tests use helper/resource package
- **Fix:** Ran `go get github.com/hashicorp/terraform-plugin-testing/helper/resource` and `go mod tidy`
- **Files modified:** go.mod, go.sum
- **Verification:** All test files compile, unit tests pass
- **Committed in:** 3730d86 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Essential dependency addition. No scope creep.

## Issues Encountered
None -- plan executed straightforwardly after resolving the missing dependency.

## User Setup Required
None -- no external service configuration required. Acceptance tests require only TF_ACC=1 and VASTAI_API_KEY environment variables.

## Next Phase Readiness
- Phase 2 (Core Compute) is now complete with all 6 plans executed
- All resources have unit tests (schema validation) and acceptance tests (API integration)
- Patterns established: acceptance test naming, cost strategy, data source test approach
- Ready for Phase 3 (Storage) or Phases 4/5 (Serverless, Account & Networking)

## Self-Check: PASSED

All 5 created files verified present. Both commit hashes (3730d86, 8e28f81) confirmed in git log.

---
*Phase: 02-core-compute*
*Completed: 2026-03-25*
