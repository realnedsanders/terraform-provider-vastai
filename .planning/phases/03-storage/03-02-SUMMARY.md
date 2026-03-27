---
phase: 03-storage
plan: 02
subsystem: terraform-resources
tags: [go, terraform, volumes, storage, data-source, crud, import]

# Dependency graph
requires:
  - phase: 03-storage
    plan: 01
    provides: "VolumeService with Create, Clone, List, Delete, SearchOffers and typed structs"
  - phase: 02-core-compute
    provides: "Template resource pattern, GPU offers data source pattern, schema conventions"
provides:
  - "vastai_volume resource with Create (from offer + clone), Read (via list), Delete, Import"
  - "vastai_volume_offers data source with structured filters and most_affordable convenience"
  - "VolumeResourceModel, VolumeOfferModel, VolumeOffersDataSourceModel"
  - "18 unit tests covering resource and data source schema validation"
affects: [03-03-PLAN, 06-01-PLAN]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Immutable resource pattern: Update returns error, all creation-time attrs ForceNew"
    - "Clone as creation-time attribute: clone_from_id triggers POST /volumes/copy/ path"
    - "Read-via-list for volumes: List(local_volume) then filter by ID"

key-files:
  created:
    - internal/services/volume/models.go
    - internal/services/volume/resource_volume.go
    - internal/services/volume/data_source_volume_offers.go
    - internal/services/volume/resource_volume_test.go
    - internal/services/volume/data_source_volume_offers_test.go
  modified: []

key-decisions:
  - "Volumes treated as immutable: Update returns error, all creation-time attrs use RequiresReplace"
  - "No list/unlist marketplace operations: HOST-only per research (Pitfall 1), omitted from tenant provider"
  - "Clone returns highest-ID volume from list since clone API does not return new volume ID"
  - "Timeouts block has Create/Read/Delete only (no Update since volumes are immutable)"

patterns-established:
  - "Immutable resource: Update returns error diagnostic, creation attrs are ForceNew"
  - "Clone-as-attribute: clone_from_id optional attr triggers alternate Create path"
  - "Volume offer search: structured filters map to VolumeOfferSearchParams pointer fields"

requirements-completed: [STOR-01, STOR-02, DATA-05]

# Metrics
duration: 5min
completed: 2026-03-27
---

# Phase 3 Plan 02: Volume Resource and Offers Data Source Summary

**vastai_volume resource with CRUD/clone/import and vastai_volume_offers data source with 16 filter attributes, most_affordable convenience, and 18 unit tests**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-27T14:59:52Z
- **Completed:** 2026-03-27T15:05:26Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- vastai_volume resource with Create from offer, Create via clone (D-01), Read via list (Pitfall 3), Delete, and Import
- vastai_volume_offers data source with 13 filter attributes, offers list, and most_affordable convenience attribute
- All creation-time attributes (offer_id, size, clone_from_id, disable_compression) marked ForceNew
- 18 unit tests covering schema structure, validators, plan modifiers, descriptions, and interface compliance

## Task Commits

Each task was committed atomically:

1. **Task 1: Volume resource with CRUD, clone, import, and models** - `6b207f4` (feat)
2. **Task 2: Volume offers data source and unit tests for all volume service types** - `c5e09cd` (feat)

## Files Created/Modified
- `internal/services/volume/models.go` - VolumeResourceModel, VolumeOffersDataSourceModel, VolumeOfferModel
- `internal/services/volume/resource_volume.go` - vastai_volume resource with CRUD, clone, import, timeouts
- `internal/services/volume/data_source_volume_offers.go` - vastai_volume_offers data source with filters and most_affordable
- `internal/services/volume/resource_volume_test.go` - 11 unit tests for volume resource schema
- `internal/services/volume/data_source_volume_offers_test.go` - 7 unit tests for volume offers data source schema

## Decisions Made
- Volumes treated as immutable: Update method returns error diagnostic since all meaningful attributes force replacement
- Marketplace list/unlist operations omitted entirely: per research these are HOST-only operations (Pitfall 1), not relevant for tenant provider scope
- Clone creation path finds new volume by selecting highest ID from list, since clone API does not return the new volume ID
- Timeouts block configured for Create/Read/Delete only (no Update needed for immutable resources)
- allocated_storage filter attribute included per Pitfall 5 with description explaining its impact on pricing calculations

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Known Stubs

None - all resource methods are fully wired to the client API.

## Next Phase Readiness
- Volume resource and offers data source ready for provider registration
- Network volume resource (Plan 03-03) follows the same patterns established here
- Both vastai_volume and vastai_volume_offers need to be registered in provider.go (will happen in 03-03 or a registration task)

## Self-Check: PASSED

All files verified present. All commits verified in git log.

---
*Phase: 03-storage*
*Completed: 2026-03-27*
