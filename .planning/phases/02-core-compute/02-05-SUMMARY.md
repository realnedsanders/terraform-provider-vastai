---
phase: 02-core-compute
plan: 05
subsystem: compute
tags: [terraform, instance, data-source, provider-registration, gpu, schema-quality]

# Dependency graph
requires:
  - phase: 02-core-compute/01
    provides: "API client with InstanceService Get/List methods"
  - phase: 02-core-compute/02
    provides: "GPU offers data source, template resource and data source"
  - phase: 02-core-compute/03
    provides: "SSH key resource and data source"
  - phase: 02-core-compute/04
    provides: "Instance resource with full CRUD"
provides:
  - "vastai_instance data source for looking up a single instance by ID"
  - "vastai_instances data source for listing all instances with optional label filter"
  - "Provider registration of all Phase 2 resources (3) and data sources (5)"
  - "Complete Phase 2 provider wiring -- terraform plan can discover all compute resources"
affects: [02-core-compute/06, 03-storage, 04-serverless, 05-account-networking, 06-documentation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Instance data source: singular by-ID lookup pattern with Computed read-only attributes"
    - "Instances data source: list-all with optional client-side substring filter"
    - "Provider registration: package import + factory function slice pattern"
    - "apiInstanceToAttrValues: API-to-attr.Value conversion for ListNestedAttribute elements"

key-files:
  created:
    - "internal/services/instance/data_source_instance.go"
    - "internal/services/instance/data_source_instances.go"
    - "internal/services/instance/data_source_instance_test.go"
  modified:
    - "internal/services/instance/models.go"
    - "internal/provider/provider.go"

key-decisions:
  - "Instance data source uses string ID (Required) to match resource ID type -- consistent with import pattern"
  - "Instances data source filters by label substring match client-side (API does not support server-side label filtering)"
  - "RAM conversion divides by 1000 (consistent with resource and offers patterns)"

patterns-established:
  - "Data source pair pattern: singular (by-ID Required) + plural (list-all with optional filter) per resource domain"
  - "Provider registration pattern: import service packages, add factory functions to Resources() and DataSources() slices"

requirements-completed: [DATA-02, DATA-03, TEST-02, IMPT-02]

# Metrics
duration: 4min
completed: 2026-03-25
---

# Phase 2 Plan 05: Instance Data Sources and Provider Registration Summary

**Instance data sources (singular by-ID, plural list with label filter) plus full provider registration of 3 resources and 5 data sources**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-25T22:27:25Z
- **Completed:** 2026-03-25T22:31:25Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- vastai_instance data source reads single instance by ID with all hardware, network, and status attributes
- vastai_instances data source lists all instances with optional label substring filter
- Provider registers all Phase 2 resources (vastai_instance, vastai_template, vastai_ssh_key) and data sources (vastai_gpu_offers, vastai_instance, vastai_instances, vastai_templates, vastai_ssh_keys)
- Full unit test suite passes: 75+ tests across all packages (client, instance, offer, sshkey, template)
- Every schema attribute has non-empty Description per SCHM-04

## Task Commits

Each task was committed atomically:

1. **Task 1: Instance data sources with schema quality** - `4fac6b8` (feat)
2. **Task 2: Register all resources and data sources in provider** - `1563cbd` (feat)

## Files Created/Modified
- `internal/services/instance/data_source_instance.go` - Singular instance data source (Required id, all Computed attributes, API-to-model mapping)
- `internal/services/instance/data_source_instances.go` - Plural instances data source (Optional label filter, ListNestedAttribute, client-side substring matching)
- `internal/services/instance/data_source_instance_test.go` - 6 unit tests: schema structure, Required/Computed flags, descriptions, metadata
- `internal/services/instance/models.go` - Added InstanceDataSourceModel, InstancesDataSourceModel, InstanceDataModel, instanceDataModelAttrTypes()
- `internal/provider/provider.go` - Added service package imports, registered 3 resources and 5 data sources

## Decisions Made
- Instance data source uses string ID (Required) to be consistent with the resource's string ID and import pattern
- Instances data source uses client-side substring filtering on label (the Vast.ai API List endpoint does not support server-side label filtering)
- RAM conversion uses /1000 (MB to GB) consistent with the established pattern from offers and instance resource

## Deviations from Plan

None -- plan executed exactly as written.

## Known Stubs

None -- all data sources and provider registration fully implemented with real logic.

## Issues Encountered

None.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness
- All Phase 2 resources and data sources are registered and discoverable by Terraform
- Ready for acceptance testing against real Vast.ai API (Plan 06 scope)
- Provider binary compiles and all unit tests pass
- Patterns established (data source pair, provider registration) reusable in Phases 3-5

## Self-Check: PASSED

- All 5 created/modified files exist on disk
- Both task commits (4fac6b8, 1563cbd) found in git log
- 75+ tests pass across all packages
- go build ./... and go vet ./... clean

---
*Phase: 02-core-compute*
*Completed: 2026-03-25*
