---
phase: 02-core-compute
plan: 03
subsystem: api
tags: [terraform, ssh-key, resource, datasource, crud, import, validators, sensitive]

# Dependency graph
requires:
  - phase: 02-core-compute-01
    provides: SSHKeyService client CRUD methods in internal/client/ssh_keys.go
provides:
  - vastai_ssh_key resource with full CRUD and import
  - vastai_ssh_keys data source listing all keys
  - SSHKeyResourceModel and SSHKeysDataSourceModel
  - Schema quality patterns applied (validators, plan modifiers, sensitive, timeouts, descriptions)
affects: [02-core-compute-04, 06-documentation]

# Tech tracking
tech-stack:
  added: [terraform-plugin-framework-validators v0.19.0, terraform-plugin-framework-timeouts v0.5.0]
  patterns: [resource-with-import, sensitive-flag, ssh-key-format-validator, UseStateForUnknown, configurable-timeouts, list-then-filter-read]

key-files:
  created:
    - internal/services/sshkey/resource_ssh_key.go
    - internal/services/sshkey/data_source_ssh_keys.go
    - internal/services/sshkey/models.go
    - internal/services/sshkey/resource_ssh_key_test.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "Read uses List + filter by ID since Vast.ai has no single-get SSH key endpoint"
  - "SSH key format validator accepts ssh-rsa, ssh-ed25519, ssh-ecdsa, ssh-dsa prefixes per D-16"
  - "Timeouts default to 5 minutes for all operations (create/read/update/delete)"

patterns-established:
  - "Service directory pattern: internal/services/{resource}/ with resource, data_source, models, tests"
  - "Schema quality: Sensitive flag on secrets, UseStateForUnknown on stable computed fields, validators on constrained inputs"
  - "Read-via-list: When API lacks single-get endpoint, list all and filter by ID, remove from state if not found"

requirements-completed: [COMP-07, COMP-08, DATA-08]

# Metrics
duration: 3min
completed: 2026-03-25
---

# Phase 2 Plan 3: SSH Key Resource and Data Source Summary

**SSH key resource with CRUD, import, sensitive flag, format validator, and data source listing all keys with 10 unit tests**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-25T22:00:38Z
- **Completed:** 2026-03-25T22:04:26Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- SSH key resource with full CRUD lifecycle (create, read, update, delete) and import support
- SSH keys data source listing all keys for the authenticated user
- Schema quality patterns: sensitive flag on ssh_key, SSH format validator, UseStateForUnknown on id/created_at, configurable timeouts, descriptions on all attributes
- 10 unit tests covering schema structure, sensitive flags, descriptions, plan modifiers, validators, and metadata

## Task Commits

Each task was committed atomically:

1. **Task 1: SSH key resource with CRUD, import, sensitive flag, and validators** - `d92d370` (feat)
2. **Task 2: SSH keys data source and unit tests** - `ff0fc69` (feat)

## Files Created/Modified
- `internal/services/sshkey/models.go` - SSHKeyResourceModel, SSHKeysDataSourceModel, SSHKeyModel structs with tfsdk tags
- `internal/services/sshkey/resource_ssh_key.go` - vastai_ssh_key resource with full CRUD, import, validators, timeouts
- `internal/services/sshkey/data_source_ssh_keys.go` - vastai_ssh_keys data source listing all keys
- `internal/services/sshkey/resource_ssh_key_test.go` - 10 unit tests for schema quality verification
- `go.mod` - Added terraform-plugin-framework-validators v0.19.0, terraform-plugin-framework-timeouts v0.5.0
- `go.sum` - Updated checksums

## Decisions Made
- Read uses List + filter by ID since Vast.ai has no single-get SSH key endpoint -- matches the Python SDK pattern
- SSH key format validator regex `^ssh-(rsa|ed25519|ecdsa|dsa) ` covers all standard key types per D-16
- All timeout defaults set to 5 minutes (SSH key operations are fast, but consistent with resource pattern)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- SSH key resource and data source ready for provider registration (will be wired in provider.go when all Phase 2 resources are complete)
- SSHKeyService.AttachToInstance/DetachFromInstance methods available in client for Plan 04 (instance resource) to use
- Schema quality patterns (validators, timeouts, sensitive, plan modifiers) established for reuse across template and instance resources

## Self-Check: PASSED

- All 4 created files verified present on disk
- Commit d92d370 (Task 1) verified in git log
- Commit ff0fc69 (Task 2) verified in git log
- go build, go test (10/10 PASS), go vet all exit 0

---
*Phase: 02-core-compute*
*Completed: 2026-03-25*
