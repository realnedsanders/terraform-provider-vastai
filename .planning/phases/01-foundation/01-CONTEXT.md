# Phase 1: Foundation - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Compilable, installable Terraform provider with zero resources but a working API client, CI/CD pipeline, and a published alpha release that installs via `terraform init`. This phase delivers the infrastructure everything else builds on — Go module, Plugin Framework wiring, REST API client, GoReleaser config, GitHub Actions CI, and GPG-signed release pipeline.

</domain>

<decisions>
## Implementation Decisions

### GitHub Namespace & Module Path
- **D-01:** GitHub repo is `github.com/realnedsanders/terraform-provider-vastai`
- **D-02:** Go module path is `github.com/realnedsanders/terraform-provider-vastai`
- **D-03:** Terraform Registry namespace will be `realnedsanders/vastai`

### API Client Architecture
- **D-04:** Service pattern — `client.Instances.Create()`, `client.Offers.List()` — sub-services mirror the API surface, scales to 25+ resource types
- **D-05:** Standalone-capable design in `internal/client/` — clean interfaces, no Terraform dependencies, could be extracted to a public Go SDK later
- **D-06:** Built-in waiter methods (e.g., `WaitForInstanceReady()`) with configurable timeout and polling interval — AWS provider pattern
- **D-07:** Rate limiting matches Python SDK: 150ms base delay, 1.5x exponential backoff, max 5 retries
- **D-08:** Full tflog integration — request method/URL at DEBUG, response status at DEBUG, response body at TRACE. Errors and retries always logged.
- **D-09:** Authentication via Bearer header (never query parameter) — prevents credential leaks in logs

### Project Layout
- **D-10:** Service-per-directory pattern: `internal/services/instance/`, `internal/services/template/` — each directory contains resource.go, data_source.go, models.go, and tests
- **D-11:** API client in `internal/client/` — private to module, cleanly separated from provider/resource code

### CI/CD & Release
- **D-12:** Tag-based releases — push git tag (v0.1.0) triggers GitHub Actions → GoReleaser → signed binaries published to registry
- **D-13:** GPG private key + passphrase stored as GitHub Actions secrets
- **D-14:** CI strategy: unit tests + lint on PRs (fast, free); acceptance tests on main branch only (uses real API credits)
- **D-15:** GoReleaser generates SHA256SUMS, .sig files, and terraform-registry-manifest.json automatically

### Claude's Discretion
- Go version selection (latest stable that satisfies Plugin Framework requirements)
- Specific GoReleaser configuration details
- golangci-lint configuration
- Exact directory structure for templates/, examples/, docs/ scaffolding
- Makefile targets and developer workflow

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Vast.ai API Reference
- `vast-sdk/vastai/vast.py` — Primary API client implementation, all endpoint paths, query parameters, response shapes
- `vast-sdk/vastai/vastai_sdk.py` — SDK wrapper with method signatures

### Project Research
- `.planning/research/STACK.md` — Technology stack recommendations with versions
- `.planning/research/ARCHITECTURE.md` — Provider architecture patterns and component boundaries
- `.planning/research/PITFALLS.md` — Critical mistakes to avoid (especially: Bearer auth, schema design, async creation)
- `.planning/research/SUMMARY.md` — Synthesized research findings

### Project Planning
- `.planning/PROJECT.md` — Project context, constraints, and key decisions
- `.planning/REQUIREMENTS.md` — Phase 1 requirements: FOUND-01..06, RLSE-01..05, TEST-04

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- No existing Go code — this is a greenfield project
- Python SDK at `vast-sdk/` serves as API behavior reference only (not code to port directly)

### Established Patterns
- None yet — Phase 1 establishes all patterns

### Integration Points
- Provider binary must be installable via `terraform init` with the registry source `realnedsanders/vastai`
- API client must authenticate against `https://console.vast.ai/api/v0/` using Bearer token from `VASTAI_API_KEY`

</code_context>

<specifics>
## Specific Ideas

- API client should match Python SDK's retry behavior (150ms base, 1.5x backoff) since that's battle-tested against Vast.ai's actual rate limits
- Service pattern chosen specifically because the API has 25+ resource types — flat client with 100+ methods would be unwieldy
- Standalone-capable client design was chosen so it could become a community Go SDK if there's demand

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-foundation*
*Context gathered: 2026-03-25*
