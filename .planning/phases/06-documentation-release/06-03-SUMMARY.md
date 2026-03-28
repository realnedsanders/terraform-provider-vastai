---
phase: 06-documentation-release
plan: 03
subsystem: testing-infrastructure
tags: [sweepers, test-cleanup, ci-integration, resource-management]
dependency_graph:
  requires: [client-services, ci-workflow]
  provides: [test-sweepers, sweep-make-target, ci-sweep-job]
  affects: [GNUmakefile, .github/workflows/test.yml, internal/services/*, internal/sweep/]
tech_stack:
  added: [terraform-plugin-testing/helper/resource (sweeper framework)]
  patterns: [init()-based sweeper registration, tfacc- prefix convention, create-then-read sweep pattern]
key_files:
  created:
    - internal/sweep/client.go
    - internal/services/sweep_test.go
    - internal/services/instance/sweep_test.go
    - internal/services/template/sweep_test.go
    - internal/services/sshkey/sweep_test.go
    - internal/services/volume/sweep_test.go
    - internal/services/networkvolume/sweep_test.go
    - internal/services/endpoint/sweep_test.go
    - internal/services/workergroup/sweep_test.go
    - internal/services/apikey/sweep_test.go
    - internal/services/envvar/sweep_test.go
    - internal/services/team/sweep_test.go
    - internal/services/cluster/sweep_test.go
    - internal/services/overlay/sweep_test.go
    - internal/services/subaccount/sweep_test.go
  modified:
    - GNUmakefile
    - .github/workflows/test.yml
decisions:
  - "Created internal/sweep package instead of putting SharedSweepClient in acctest to avoid import cycle (acctest -> provider -> services -> acctest)"
  - "Skipped sweeper for SSH keys (no name/label field), clusters (no name field), and subaccounts (no delete API)"
  - "Team sweeper targets roles only (teams are account-scoped with parameterless delete, unsafe for sweeper)"
  - "Worker group sweeper filters on endpoint name prefix since worker groups lack their own name field"
metrics:
  duration: 5min
  completed: "2026-03-28T07:47:00Z"
  tasks_completed: 2
  tasks_total: 2
  files_created: 15
  files_modified: 2
---

# Phase 06 Plan 03: Resource Sweepers Summary

Resource sweepers for test cleanup with tfacc- prefix convention, internal/sweep package to avoid import cycles, and CI integration for automatic post-test cleanup.

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Create sweeper entry point and per-resource sweepers | 9e46144 | internal/sweep/client.go, internal/services/sweep_test.go, 13 per-resource sweep_test.go files |
| 2 | Add sweep make target and CI workflow sweeper step | 91bf97a | GNUmakefile, .github/workflows/test.yml |

## Decisions Made

1. **internal/sweep package over acctest for SharedClient** -- The acctest package imports provider, which imports all service packages. Service-level sweep_test.go files importing acctest would create an import cycle. Created a dedicated internal/sweep package that only depends on internal/client.

2. **Skipped 3 sweepers due to API limitations** -- SSH keys have no name/label field (only key content), clusters have no name field (only subnet), and subaccounts have no delete API. These resource types cannot be safely identified or cleaned up by sweepers.

3. **Team sweeper limited to roles** -- The team API uses a parameterless DELETE (destroys the entire team). Since there's no listing or name-based filtering, sweeping teams would risk destroying production teams. Roles have names and support safe filtering.

4. **Worker groups filter on endpoint name** -- Worker groups don't have their own name field but carry their parent endpoint's name. Filtering by endpoint_name with tfacc- prefix correctly identifies test worker groups.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Import cycle between acctest and service packages**
- **Found during:** Task 1
- **Issue:** Plan directed adding SharedSweepClient to internal/acctest/helpers.go. However, acctest imports provider, which imports all service packages. Service-level sweep_test.go files importing acctest created a circular dependency.
- **Fix:** Created new internal/sweep package containing only SharedClient(), depending only on internal/client. Updated all sweeper files to import internal/sweep instead of internal/acctest.
- **Files created:** internal/sweep/client.go
- **Commit:** 9e46144

**2. [Rule 2 - Missing functionality] Resources without identifiable name fields**
- **Found during:** Task 1
- **Issue:** SSH keys, clusters, and subaccounts lack name/label fields needed for tfacc- prefix matching. Subaccounts additionally lack a delete API.
- **Fix:** Created documentation-only sweep_test.go files explaining why sweepers are omitted. Removed these 3 packages from the sweep_test.go blank imports. Adjusted sweeper count from planned 13 to actual 10.
- **Files modified:** internal/services/sshkey/sweep_test.go, internal/services/cluster/sweep_test.go, internal/services/subaccount/sweep_test.go
- **Commit:** 9e46144

## Sweeper Coverage

| Resource | Sweeper | Name Field | Delete Method |
|----------|---------|------------|---------------|
| vastai_instance | Yes | Label | Destroy(ctx, id) |
| vastai_template | Yes | Name | Delete(ctx, hashID) |
| vastai_ssh_key | No (no name field) | - | - |
| vastai_volume | Yes | Label | Delete(ctx, id) |
| vastai_network_volume | Yes | Label | Delete(ctx, id) |
| vastai_endpoint | Yes | EndpointName | Delete(ctx, id) |
| vastai_worker_group | Yes | EndpointName (parent) | Delete(ctx, id) |
| vastai_api_key | Yes | Name | Delete(ctx, id) |
| vastai_environment_variable | Yes | Key | Delete(ctx, key) |
| vastai_team_role | Yes | Name | DeleteRole(ctx, name) |
| vastai_cluster | No (no name field) | - | - |
| vastai_overlay | Yes | Name | Delete(ctx, overlayID) |
| vastai_subaccount | No (no delete API) | - | - |

## Known Stubs

None -- all sweeper files are fully implemented with working list/filter/delete logic.

## Self-Check: PASSED

All 15 created files verified present. Both commit hashes (9e46144, 91bf97a) verified in git log.
