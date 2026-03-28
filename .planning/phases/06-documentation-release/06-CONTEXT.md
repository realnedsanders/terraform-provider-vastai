# Phase 6: Documentation & Release - Context

**Gathered:** 2026-03-28
**Status:** Ready for planning

<domain>
## Phase Boundary

Final phase: generated documentation via tfplugindocs for all 17 resources and 11 data sources, working example configurations (per-resource basics + workflow examples), provider configuration docs, resource sweepers for CI cleanup, README update with full resource table, CHANGELOG.md, LICENSE file, and CONTRIBUTING.md.

</domain>

<decisions>
## Implementation Decisions

### Example Configurations
- **D-01:** Per-resource basic examples — one .tf per resource/data source (minimum for registry docs)
- **D-02:** GPU instance workflow example — end-to-end: search offers -> create template -> add SSH key -> launch instance
- **D-03:** Serverless workflow example — create endpoint -> add worker group -> query status
- **D-04:** Team management example — create team -> define roles -> invite members
- **D-05:** Claude's Discretion on placeholder vs realistic values in examples

### Documentation
- **D-06:** README.md updated with comprehensive resource/data source table listing everything the provider supports
- **D-07:** CHANGELOG.md added — standard changelog tracking releases
- **D-08:** LICENSE file — MPL-2.0 (Mozilla Public License 2.0, standard for Terraform providers)
- **D-09:** .github/CONTRIBUTING.md — standard contributing guide with dev setup, testing, PR process

### Sweepers
- **D-10:** Sweepers for ALL mutable resources — instances, volumes, network volumes, endpoints, worker groups, SSH keys, templates, teams, API keys, env vars, clusters, overlays, subaccounts
- **D-11:** Claude's Discretion on sweeper identification strategy (prefix convention vs label-based, based on API filtering capabilities)

### Claude's Discretion
- Example placeholder vs realistic values
- Sweeper identification strategy
- tfplugindocs template customization beyond defaults

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Existing Documentation
- `docs/index.md` — Already generated provider docs (from Phase 1 fix)
- `templates/index.md.tmpl` — Provider doc template
- `examples/provider/provider.tf` — Existing provider example
- `README.md` — Current README (needs update with resource table)

### tfplugindocs Conventions
- `templates/resources/<name>.md.tmpl` — Resource doc templates with `{{.SchemaMarkdown}}`
- `templates/data-sources/<name>.md.tmpl` — Data source doc templates
- `examples/resources/<name>/resource.tf` — Per-resource example files
- `examples/data-sources/<name>/data-source.tf` — Per-data-source example files

### Existing Provider Code
- `internal/provider/provider.go` — 17 resources + 11 data sources registered
- `internal/services/*/` — All resource and data source implementations

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `docs/index.md` already generated from templates/index.md.tmpl
- `examples/provider/provider.tf` already exists
- `GNUmakefile` has `generate` target for `go generate ./...`
- `internal/acctest/helpers.go` for test infrastructure

### Integration Points
- tfplugindocs reads from `templates/` and `examples/`, writes to `docs/`
- Sweepers register via `resource.AddTestSweepers()` in `*_test.go` files
- CI workflow at `.github/workflows/test.yml` needs sweeper step

</code_context>

<specifics>
## Specific Ideas

- The GPU instance workflow example should be the "hero" example — the most common use case users will copy
- Sweepers should be conservative — only delete resources with identifiable test prefixes to avoid deleting user resources
- README resource table should match what the registry shows for quick reference

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 06-documentation-release*
*Context gathered: 2026-03-28*
