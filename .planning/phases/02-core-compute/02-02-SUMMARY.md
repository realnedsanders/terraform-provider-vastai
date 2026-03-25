---
phase: 02-core-compute
plan: 02
subsystem: api
tags: [terraform, data-source, resource, gpu-offers, templates, validators, plan-modifiers, timeouts]

# Dependency graph
requires:
  - phase: 02-core-compute/01
    provides: "API client service layer (OfferService, TemplateService) with typed structs"
provides:
  - "vastai_gpu_offers data source with structured filters and most_affordable"
  - "vastai_template resource with full CRUD and import"
  - "vastai_templates data source with query search"
  - "Schema quality patterns: validators, plan modifiers, timeouts, sensitive flags"
affects: [02-core-compute/03, 02-core-compute/04, 02-core-compute/05, 03-storage, 04-serverless, 05-account-networking]

# Tech tracking
tech-stack:
  added: [terraform-plugin-framework-validators v0.19.0, terraform-plugin-framework-timeouts v0.5.0]
  patterns: [data-source-with-nested-attributes, resource-with-crud-import-timeouts, model-conversion-layer, mb-to-gb-unit-conversion]

key-files:
  created:
    - internal/services/offer/data_source_gpu_offers.go
    - internal/services/offer/models.go
    - internal/services/offer/data_source_gpu_offers_test.go
    - internal/services/template/resource_template.go
    - internal/services/template/data_source_templates.go
    - internal/services/template/models.go
    - internal/services/template/resource_template_test.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "MB-to-GB conversion in offer model layer: API returns RAM in MB, Terraform state stores in GB for user friendliness"
  - "hash_id as primary template identifier (not numeric ID) for import and CRUD operations"
  - "Optional+Computed pattern for server-defaulted boolean fields (ssh_direct, jup_direct, etc.)"
  - "booldefault.StaticBool(false) for private field to match expected API behavior"

patterns-established:
  - "Data source pattern: offerModelAttrTypes() + offerNestedAttributes() shared between list and single nested attributes"
  - "Resource CRUD pattern: modelToCreateRequest() converter + apiTemplateToModel() reverse mapper"
  - "Schema quality pattern: validators on constrained fields, Sensitive on secrets, UseStateForUnknown on stable computed, timeouts block"
  - "Unit test pattern: schema inspection tests verifying attribute types, validators, descriptions, sensitive flags, plan modifiers"

requirements-completed: [COMP-06, DATA-01, DATA-04, SCHM-01, SCHM-02, SCHM-03, SCHM-04, SCHM-05, SCHM-06, IMPT-01]

# Metrics
duration: 6min
completed: 2026-03-25
---

# Phase 2 Plan 2: GPU Offers Data Source and Template Resource Summary

**GPU offers data source with structured filters/most_affordable and template resource with full CRUD, import, timeouts, and comprehensive schema quality patterns**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-25T22:00:07Z
- **Completed:** 2026-03-25T22:07:03Z
- **Tasks:** 2/2
- **Files modified:** 9

## Accomplishments
- GPU offers data source with 10 filter attributes (gpu_name, num_gpus, gpu_ram_gb, max_price_per_hour, datacenter_only, region, offer_type, order_by, limit, raw_query) and computed most_affordable convenience attribute
- Template resource with full CRUD (create/read/update/delete), import via hash_id, configurable timeouts, and Docker image regex validation
- Templates data source for searching by query string
- Comprehensive schema quality: validators on all constrained fields, Sensitive on docker_login_repo, correct Required/Optional/Computed classification, descriptions on every attribute, UseStateForUnknown plan modifiers, and configurable timeouts
- 16 unit tests covering schema validation, model conversion, and attribute verification

## Task Commits

Each task was committed atomically:

1. **Task 1: GPU offers data source** - `c6b9696` (feat)
2. **Task 2: Template resource and data source** - `c21b612` (feat)

## Files Created/Modified
- `internal/services/offer/models.go` - GpuOffersDataSourceModel and OfferModel with GB-unit fields
- `internal/services/offer/data_source_gpu_offers.go` - vastai_gpu_offers data source with structured filters, validators, and most_affordable
- `internal/services/offer/data_source_gpu_offers_test.go` - Schema and model conversion unit tests (5 tests)
- `internal/services/template/models.go` - TemplateResourceModel, TemplatesDataSourceModel, TemplateDataSourceModel
- `internal/services/template/resource_template.go` - vastai_template resource with CRUD, import, timeouts
- `internal/services/template/data_source_templates.go` - vastai_templates data source with query search
- `internal/services/template/resource_template_test.go` - Schema, validator, sensitive, plan modifier, and model tests (11 tests)
- `go.mod` - Added framework-validators v0.19.0 and framework-timeouts v0.5.0
- `go.sum` - Updated checksums

## Decisions Made
- MB-to-GB conversion in the offer model layer: the Vast.ai API returns GPU/CPU RAM in MB, but the Terraform state exposes values in GB for user-friendliness (divide by 1000)
- hash_id as primary template identifier: templates use hash_id (not numeric ID) as the Terraform resource ID, matching the API's CRUD operations
- Optional+Computed for server-defaulted fields: boolean fields like ssh_direct, jup_direct, use_ssh use Optional+Computed so users can set them or accept server defaults without noisy diffs
- Static false default for private: the private field uses booldefault.StaticBool(false) to match expected API behavior

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None - all data sources and resources are fully wired to the client service layer.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Offer and template services are ready for the instance resource (Plan 03) to reference
- Schema quality patterns (validators, plan modifiers, timeouts, sensitive) are established and can be replicated to instance, SSH key, and all subsequent resources
- The data source pattern (attrTypes + nestedAttributes + model conversion) can be directly applied to instance and SSH key data sources

## Self-Check: PASSED

All 8 created files verified on disk. Both task commits (c6b9696, c21b612) verified in git log.

---
*Phase: 02-core-compute*
*Completed: 2026-03-25*
