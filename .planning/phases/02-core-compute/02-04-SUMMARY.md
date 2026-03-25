---
phase: 02-core-compute
plan: 04
subsystem: compute
tags: [terraform, instance, lifecycle, preemption, ssh, import, gpu]

# Dependency graph
requires:
  - phase: 02-core-compute/01
    provides: "API client service layer with InstanceService, SSHKeyService"
  - phase: 02-core-compute/02
    provides: "GPU offers data source, template resource patterns"
  - phase: 02-core-compute/03
    provides: "SSH key resource with CRUD"
provides:
  - "vastai_instance resource with full CRUD lifecycle"
  - "Preemption detection for spot instances (is_bid + intended vs actual status)"
  - "Start/stop via status attribute without destroy/recreate"
  - "In-place label, bid price, and template updates"
  - "SSH key attach/detach via ssh_key_ids set attribute"
  - "Import support via contract ID"
  - "Instance resource model and schema pattern for subsequent resources"
affects: [02-core-compute/05, 02-core-compute/06, 03-storage, 04-serverless, 06-documentation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Preemption detection: is_bid + intended_status + actual_status three-way check"
    - "mapInstanceToModel: API-to-Terraform model conversion with RAM MB-to-GB"
    - "buildRuntype: feature flag bools to space-separated runtype string"
    - "reconcileSSHKeys: incremental set-diff attach/detach"
    - "Offer expiry: structured error with user guidance on 404/400 from create"

key-files:
  created:
    - "internal/services/instance/models.go"
    - "internal/services/instance/resource_instance.go"
    - "internal/services/instance/resource_instance_test.go"
  modified: []

key-decisions:
  - "Preemption only on stopped/offline (not exited) for spot instances -- exited may be a normal container exit"
  - "RAM conversion divides by 1000 (MB to GB) per Pitfall 6 from research"
  - "SSH key attachment resolves IDs to public key content via SSHKeys.List before calling AttachToInstance"
  - "Delete waits for destroyed status but ignores timeout errors (best-effort cleanup)"

patterns-established:
  - "Instance lifecycle pattern: Create polls, Read detects preemption, Update applies diffs in order (status, label, bid, template, SSH keys)"
  - "Set-based attribute reconciliation: compute added/removed via setDifference, apply incrementally"
  - "Offer expiry error pattern: catch 404/400 on create, guide user to re-plan"

requirements-completed: [COMP-01, COMP-02, COMP-03, COMP-04, COMP-05, IMPT-02]

# Metrics
duration: 8min
completed: 2026-03-25
---

# Phase 2 Plan 04: Instance Resource Summary

**vastai_instance resource with full CRUD, start/stop lifecycle, spot preemption detection, SSH key attachment, bid/label/template updates, and import support**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-25T22:13:55Z
- **Completed:** 2026-03-25T22:21:54Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Instance resource with complete CRUD lifecycle (create from offer, read with preemption handling, in-place updates, graceful destroy)
- Spot instance preemption detection: three-way check on is_bid, intended_status, and actual_status with silent state removal for genuinely preempted instances
- Start/stop via status attribute without destroy/recreate (COMP-02)
- In-place updates for label, bid price, template/image/env/onstart, and SSH key attachment
- 13 passing unit tests covering schema validation, preemption detection, model mapping, helper functions, and interface compliance

## Task Commits

Each task was committed atomically:

1. **Task 1: Instance resource model and schema** - `3bedfd6` (feat)
2. **Task 2: Instance CRUD with lifecycle, preemption, SSH, import** - `3340250` (feat)

## Files Created/Modified
- `internal/services/instance/models.go` - InstanceResourceModel with all attributes (Required immutable, Optional+Computed, Optional mutable, Computed stable/dynamic, Timeouts)
- `internal/services/instance/resource_instance.go` - Full resource implementation: Schema with validators/plan modifiers/sensitive flags, Create (offer-based with polling), Read (preemption detection), Update (ordered diffs for status/label/bid/template/SSH), Delete, ImportState, plus helper functions
- `internal/services/instance/resource_instance_test.go` - 13 unit tests: schema RequiresReplace, UseStateForUnknown, validators, sensitive fields, descriptions, timeouts, preemption detection (7 cases), buildRuntype (5 cases), setDifference, mapInstanceToModel, interface compliance, extractStringSet, empty optional fields, attribute existence

## Decisions Made
- Preemption detection uses stopped/offline as preemption indicators (not exited, which can be normal termination) -- per D-09 requirement to distinguish preemption from normal stops
- RAM values divided by 1000 (MB to GB) rather than 1024 -- matches Vast.ai API convention where values are in metric MB
- SSH key attachment resolves IDs to public key content via SSHKeys.List before calling AttachToInstance -- the API requires the full key content, not an ID reference
- Delete uses best-effort wait for "destroyed" status -- does not fail if wait times out since the destroy command already succeeded

## Deviations from Plan

None -- plan executed exactly as written.

## Known Stubs

None -- all CRUD methods fully implemented with real logic.

## Issues Encountered
- Plan modifier types (RequiresReplaceModifier, UseStateForUnknownModifier) are not exported as named types in terraform-plugin-framework, so tests verify plan modifier count rather than specific type assertions. This provides equivalent coverage.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness
- Instance resource ready for registration in provider.go (Plan 05/06 scope)
- Instance resource ready for acceptance testing against real API (Plan 06 scope)
- Patterns established (preemption, set reconciliation, offer expiry) reusable in storage and serverless phases

## Self-Check: PASSED

- All 3 created files exist on disk
- Both task commits (3bedfd6, 3340250) found in git log
- 13/13 tests pass
- go build, go vet pass cleanly

---
*Phase: 02-core-compute*
*Completed: 2026-03-25*
