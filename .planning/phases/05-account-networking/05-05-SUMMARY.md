---
phase: 05-account-networking
plan: 05
subsystem: networking
tags: [terraform, cluster, overlay, composite-id, create-then-read, no-op-destroy]

# Dependency graph
requires:
  - phase: 05-account-networking/02
    provides: "Cluster and overlay client services (clusters.go, overlays.go)"
provides:
  - "vastai_cluster resource with create-then-read and DeleteWithBody"
  - "vastai_cluster_member resource with composite ID and JoinMachine/RemoveMachine"
  - "vastai_overlay resource with create-then-read and DeleteWithBody"
  - "vastai_overlay_member resource with join-only and no-op destroy warning"
affects: [06-documentation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Composite ID pattern for association resources (cluster_id/machine_id, overlay_id/instance_id)"
    - "No-op destroy with AddWarning diagnostic for API-limited resources"
    - "Create-then-read pattern for resources that return only messages on create"

key-files:
  created:
    - internal/services/cluster/models.go
    - internal/services/cluster/resource_cluster.go
    - internal/services/cluster/resource_cluster_test.go
    - internal/services/clustermember/models.go
    - internal/services/clustermember/resource_cluster_member.go
    - internal/services/clustermember/resource_cluster_member_test.go
    - internal/services/overlay/models.go
    - internal/services/overlay/resource_overlay.go
    - internal/services/overlay/resource_overlay_test.go
    - internal/services/overlaymember/models.go
    - internal/services/overlaymember/resource_overlay_member.go
    - internal/services/overlaymember/resource_overlay_member_test.go
  modified:
    - internal/provider/provider.go

key-decisions:
  - "Composite IDs for membership resources enable import via cluster_id/machine_id and overlay_id/instance_id"
  - "Overlay member destroy is a no-op with AddWarning because the API has no remove-instance-from-overlay endpoint"

patterns-established:
  - "Composite ID pattern: strings.Split on '/' for association resources with multi-part identifiers"
  - "No-op destroy: AddWarning diagnostic for API limitations, state-only removal"

requirements-completed: [NETW-01, NETW-02, NETW-03, NETW-04]

# Metrics
duration: 6min
completed: 2026-03-28
---

# Phase 5 Plan 05: Networking Resources Summary

**Cluster, overlay, and membership resources with composite IDs, create-then-read pattern, and no-op destroy for API-limited overlay membership**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-28T00:36:53Z
- **Completed:** 2026-03-28T00:42:52Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments
- 4 new Terraform resources: vastai_cluster, vastai_cluster_member, vastai_overlay, vastai_overlay_member
- Composite ID pattern for membership resources enabling proper import support
- No-op destroy with warning diagnostic for overlay members (API limitation documented)
- All resources registered in provider with full schema unit test coverage

## Task Commits

Each task was committed atomically:

1. **Task 1: Cluster and cluster member resources** - `7e01127` (feat)
2. **Task 2: Overlay and overlay member resources** - `f1f0807` (feat)

**Provider registration (deviation):** `9b788a3` (feat)
**Plan metadata:** pending (docs: complete plan)

## Files Created/Modified
- `internal/services/cluster/models.go` - ClusterResourceModel with ID, Subnet, ManagerID
- `internal/services/cluster/resource_cluster.go` - vastai_cluster with create-then-read, DeleteWithBody, import
- `internal/services/cluster/resource_cluster_test.go` - Schema unit tests for cluster resource
- `internal/services/clustermember/models.go` - ClusterMemberResourceModel with composite ID
- `internal/services/clustermember/resource_cluster_member.go` - vastai_cluster_member with JoinMachine/RemoveMachine, composite ID import
- `internal/services/clustermember/resource_cluster_member_test.go` - Schema unit tests for cluster member resource
- `internal/services/overlay/models.go` - OverlayResourceModel with ID, Name, ClusterID, InternalSubnet
- `internal/services/overlay/resource_overlay.go` - vastai_overlay with create-then-read, DeleteWithBody, name validator, import
- `internal/services/overlay/resource_overlay_test.go` - Schema unit tests for overlay resource
- `internal/services/overlaymember/models.go` - OverlayMemberResourceModel with composite ID
- `internal/services/overlaymember/resource_overlay_member.go` - vastai_overlay_member with join-only, no-op destroy warning
- `internal/services/overlaymember/resource_overlay_member_test.go` - Schema unit tests for overlay member resource
- `internal/provider/provider.go` - Added 4 new resource registrations and imports

## Decisions Made
- Composite IDs (cluster_id/machine_id, overlay_id/instance_id) for membership resources enable `terraform import` with a single string
- Overlay member destroy uses AddWarning diagnostic (not error) because the API limitation is expected behavior, not a failure
- All key fields use RequiresReplace (ForceNew) since the API does not support in-place updates for networking resources

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Registered new resources in provider**
- **Found during:** After Task 2 completion
- **Issue:** The 4 new resources (cluster, clustermember, overlay, overlaymember) were not registered in the provider's Resources() function, making them invisible to Terraform
- **Fix:** Added imports and resource registrations to internal/provider/provider.go
- **Files modified:** internal/provider/provider.go
- **Verification:** `go build ./...` compiles successfully
- **Committed in:** 9b788a3

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** Essential for correctness -- resources must be registered to be usable. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All networking resources (NETW-01 through NETW-04) are complete
- Remaining Phase 5 plans (03, 04, 06) handle account resources and data sources
- Phase 6 can generate documentation for these resources once all phases complete

## Self-Check: PASSED

All 12 created files verified present. All 3 commits (7e01127, f1f0807, 9b788a3) verified in git log.

---
*Phase: 05-account-networking*
*Completed: 2026-03-28*
