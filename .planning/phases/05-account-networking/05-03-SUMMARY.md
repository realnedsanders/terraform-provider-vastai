---
phase: 05-account-networking
plan: 03
subsystem: api
tags: [terraform, api-key, environment-variable, subaccount, sensitive, crud]

# Dependency graph
requires:
  - phase: 05-account-networking-01
    provides: Client services for API keys, env vars, and subaccounts
provides:
  - vastai_api_key resource with create/read/delete and sensitive key preservation
  - vastai_environment_variable resource with full CRUD and name-based ID
  - vastai_subaccount resource with create-only and no-op destroy
affects: [06-documentation-release]

# Tech tracking
tech-stack:
  added: [terraform-plugin-framework-timeouts, terraform-plugin-framework-validators, booldefault]
  patterns: [immutable-resource-with-RequiresReplace, create-only-resource-with-noop-destroy, name-keyed-resource-id, sensitive-write-only-attribute, custom-json-validator]

key-files:
  created:
    - internal/services/apikey/resource_api_key.go
    - internal/services/apikey/models.go
    - internal/services/apikey/resource_api_key_test.go
    - internal/services/envvar/resource_env_var.go
    - internal/services/envvar/models.go
    - internal/services/envvar/resource_env_var_test.go
    - internal/services/subaccount/resource_subaccount.go
    - internal/services/subaccount/models.go
    - internal/services/subaccount/resource_subaccount_test.go
    - internal/client/api_keys.go
    - internal/client/env_vars.go
    - internal/client/subaccounts.go
  modified:
    - internal/client/client.go
    - internal/provider/provider.go
    - go.mod
    - go.sum

key-decisions:
  - "Custom JSON validator for API key permissions using json.Valid"
  - "Environment variable ID is the key name (name-keyed resource)"
  - "Subaccount destroy is a no-op with AddWarning diagnostic (no API delete endpoint)"
  - "Password on subaccount uses UseStateForUnknown to preserve write-only value"

patterns-established:
  - "Immutable resource pattern: RequiresReplace on all mutable fields, Update returns error"
  - "Create-only resource pattern: no-op Delete with AddWarning diagnostic"
  - "Name-keyed ID pattern: import sets both id and key attributes"
  - "Write-only sensitive attribute: UseStateForUnknown preserves value never returned by API"

requirements-completed: [ACCT-01, ACCT-02, ACCT-06]

# Metrics
duration: 6min
completed: 2026-03-28
---

# Phase 5 Plan 3: Account Resources Summary

**API key, environment variable, and subaccount Terraform resources with immutable/create-only patterns, sensitive value handling, and 32 passing unit tests**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-28T00:37:24Z
- **Completed:** 2026-03-28T00:43:30Z
- **Tasks:** 2
- **Files modified:** 15

## Accomplishments
- vastai_api_key resource: Sensitive key value stored on create via UseStateForUnknown, immutable name/permissions with RequiresReplace, custom JSON validator for permissions
- vastai_environment_variable resource: Full CRUD with name-based ID, Sensitive value, key ForceNew, import by key name
- vastai_subaccount resource: Create-only with no-op destroy warning, write-only password, all fields immutable
- Client services (api_keys.go, env_vars.go, subaccounts.go) with DeleteWithBody for env vars
- All 3 resources registered in provider.go with import support

## Task Commits

Each task was committed atomically:

1. **Task 1: API key and environment variable resources** - `04b0300` (feat)
2. **Task 2: Subaccount resource with no-op destroy** - `2483d01` (feat)

## Files Created/Modified
- `internal/services/apikey/resource_api_key.go` - vastai_api_key resource with create/read/delete, Sensitive key, immutable
- `internal/services/apikey/models.go` - ApiKeyResourceModel with ID, Name, Key, Permissions, CreatedAt
- `internal/services/apikey/resource_api_key_test.go` - 10 unit tests for schema, sensitivity, plan modifiers, validators
- `internal/services/envvar/resource_env_var.go` - vastai_environment_variable resource with full CRUD, name-keyed ID
- `internal/services/envvar/models.go` - EnvVarResourceModel with ID, Key, Value
- `internal/services/envvar/resource_env_var_test.go` - 9 unit tests for schema, sensitivity, validators
- `internal/services/subaccount/resource_subaccount.go` - vastai_subaccount resource with create-only, no-op destroy
- `internal/services/subaccount/models.go` - SubaccountResourceModel with ID, Email, Username, Password, HostOnly
- `internal/services/subaccount/resource_subaccount_test.go` - 13 unit tests including no-op delete verification
- `internal/client/api_keys.go` - ApiKeyService with Create, List, Delete
- `internal/client/env_vars.go` - EnvVarService with Create, List, Update, Delete (via DeleteWithBody)
- `internal/client/subaccounts.go` - SubaccountService with Create, List
- `internal/client/client.go` - Added service fields and DeleteWithBody method
- `internal/provider/provider.go` - Registered all 3 account resources
- `go.mod` / `go.sum` - Added timeouts and validators dependencies

## Decisions Made
- Custom JSON validator for API key permissions field uses `json.Valid` for validation
- Environment variable uses key name as resource ID (name-keyed pattern); import sets both id and key attributes
- Subaccount destroy emits AddWarning diagnostic (no API delete endpoint exists)
- Password attribute on subaccount uses RequiresReplace + UseStateForUnknown (write-only, preserved in state)
- host_only uses booldefault.StaticBool(false) for computed default

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Created client services and updated client.go**
- **Found during:** Task 1 (API key and env var resources)
- **Issue:** This worktree only had Phase 1 code; client service files and service fields on VastAIClient did not exist
- **Fix:** Created api_keys.go, env_vars.go, subaccounts.go client services; added service fields and DeleteWithBody to client.go
- **Files modified:** internal/client/client.go, internal/client/api_keys.go, internal/client/env_vars.go, internal/client/subaccounts.go
- **Verification:** go build ./... succeeds
- **Committed in:** 04b0300 (Task 1 commit)

**2. [Rule 3 - Blocking] Created subaccount resource in Task 1 commit**
- **Found during:** Task 1 (provider.go registration)
- **Issue:** provider.go registers all 3 resources; subaccount package needed to exist for go mod tidy to succeed
- **Fix:** Created subaccount resource and models alongside Task 1 work
- **Files modified:** internal/services/subaccount/resource_subaccount.go, internal/services/subaccount/models.go
- **Verification:** go build ./... succeeds
- **Committed in:** 04b0300 (Task 1 commit)

**3. [Rule 1 - Bug] Fixed diagnostic severity constant in subaccount test**
- **Found during:** Task 2 (subaccount test)
- **Issue:** Test checked d.Severity() == 1 for warnings, but SeverityWarning is 2 (SeverityError is 1)
- **Fix:** Changed severity check from 1 to 2
- **Files modified:** internal/services/subaccount/resource_subaccount_test.go
- **Verification:** All 13 subaccount tests pass
- **Committed in:** 2483d01 (Task 2 commit)

---

**Total deviations:** 3 auto-fixed (1 bug, 2 blocking)
**Impact on plan:** All auto-fixes necessary for compilation and correct test behavior. No scope creep.

## Issues Encountered
None beyond the documented deviations.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All 3 account resources (API key, env var, subaccount) ready for documentation generation in Phase 6
- Patterns established: immutable resource (RequiresReplace), create-only resource (no-op destroy), name-keyed ID

## Known Stubs
None - all resources are fully wired to client services with complete CRUD implementations.

## Self-Check: PASSED

All 12 created files verified present. Both task commits (04b0300, 2483d01) found in git log.

---
*Phase: 05-account-networking*
*Completed: 2026-03-28*
