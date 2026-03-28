---
phase: 05-account-networking
plan: 04
subsystem: account
tags: [terraform, team, rbac, roles, members, invite]

# Dependency graph
requires:
  - phase: 05-01
    provides: "Team client services (CreateTeam, DestroyTeam, CreateRole, GetRole, UpdateRole, DeleteRole, InviteMember, ListMembers, RemoveMember)"
provides:
  - "vastai_team resource with create/destroy"
  - "vastai_team_role resource with asymmetric API handling (read/delete by name, update by ID)"
  - "vastai_team_member resource with invite-as-create pattern"
affects: [06-documentation-release]

# Tech tracking
tech-stack:
  added: []
  patterns: [asymmetric-identifier-resource, invite-as-create, parameterless-delete, json-string-permissions]

key-files:
  created:
    - internal/services/team/models.go
    - internal/services/team/resource_team.go
    - internal/services/team/resource_team_test.go
    - internal/services/teamrole/models.go
    - internal/services/teamrole/resource_team_role.go
    - internal/services/teamrole/resource_team_role_test.go
    - internal/services/teammember/models.go
    - internal/services/teammember/resource_team_member.go
    - internal/services/teammember/resource_team_member_test.go
  modified:
    - internal/provider/provider.go

key-decisions:
  - "Team role permissions as JSON string attribute with validity validator (D-02 revised -- nested JSON objects, not flat string set)"
  - "Team member role is ForceNew (no update-member endpoint exists; role change requires remove and re-invite)"
  - "Team Read verifies existence via ListRoles (no single-team GET endpoint)"
  - "Team role import by numeric ID resolves name via ListRoles for subsequent GetRole calls"

patterns-established:
  - "Asymmetric identifier resource: read/delete by name, update by ID (team role)"
  - "Invite-as-create: Create operation sends invitation, ListMembers resolves ID"
  - "Parameterless delete: DestroyTeam takes no parameters (API key context determines team)"

requirements-completed: [ACCT-03, ACCT-04, ACCT-05]

# Metrics
duration: 5min
completed: 2026-03-28
---

# Phase 5 Plan 4: Team RBAC Resources Summary

**Team, team role, and team member Terraform resources with asymmetric API handling and invite-as-create pattern**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-28T00:37:08Z
- **Completed:** 2026-03-28T00:42:00Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- Team resource with create/destroy (parameterless DELETE per API design)
- Team role resource handles the asymmetric API correctly (read/delete by name, update by ID -- Pitfall 3)
- Team member resource uses invite as create (D-01), with email matching for ID resolution from ListMembers
- All three resources have ImportState support and passing unit tests
- Permissions stored as JSON string with custom validity validator (D-02 revised)

## Task Commits

Each task was committed atomically:

1. **Task 1: Team and team role resources** - `07600e0` (feat)
2. **Task 2: Team member resource** - `638ac4f` (feat)
3. **Provider registration** - `11c3539` (feat)

## Files Created/Modified
- `internal/services/team/models.go` - TeamResourceModel with ID, TeamName, Timeouts
- `internal/services/team/resource_team.go` - vastai_team resource with create/destroy, ListRoles-based read
- `internal/services/team/resource_team_test.go` - Schema unit tests for team resource
- `internal/services/teamrole/models.go` - TeamRoleResourceModel with ID, Name, Permissions, Timeouts
- `internal/services/teamrole/resource_team_role.go` - vastai_team_role resource with asymmetric CRUD
- `internal/services/teamrole/resource_team_role_test.go` - Schema unit tests for team role resource
- `internal/services/teammember/models.go` - TeamMemberResourceModel with ID, Email, Role, Timeouts
- `internal/services/teammember/resource_team_member.go` - vastai_team_member resource with invite-as-create
- `internal/services/teammember/resource_team_member_test.go` - Schema unit tests for team member resource
- `internal/provider/provider.go` - Registered all 3 new resources (10 total)

## Decisions Made
- Team role permissions stored as JSON string attribute with custom JSON validity validator, matching API's nested JSON object format (D-02 revised)
- Team member role marked ForceNew since no update-member API endpoint exists; role change requires remove + re-invite
- Team Read verifies existence via ListRoles call (no dedicated single-team GET endpoint)
- Team role import by numeric ID resolves the role name via ListRoles before using GetRole for subsequent reads
- Team member Read searches by both ID and email for robustness (handles pending invites)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Provider registration for team resources**
- **Found during:** After Task 2
- **Issue:** Plan did not include provider.go registration of the 3 new resources
- **Fix:** Added imports and resource registrations for team, teamrole, teammember in provider.go
- **Files modified:** internal/provider/provider.go
- **Verification:** `go build ./...` succeeds
- **Committed in:** 11c3539

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** Essential for resources to be usable. No scope creep.

## Issues Encountered
None

## Known Stubs
None -- all resources are fully wired to client services created in Plan 05-01.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Team RBAC resources complete, ready for data sources (Plan 05-05/05-06)
- All three resources follow established patterns and are registered in the provider

---
*Phase: 05-account-networking*
*Completed: 2026-03-28*
