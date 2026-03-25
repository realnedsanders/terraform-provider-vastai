# Project Research Summary

**Project:** terraform-provider-vastai
**Domain:** Terraform/OpenTofu Provider for GPU Cloud Marketplace (Vast.ai)
**Researched:** 2026-03-25
**Confidence:** HIGH

## Executive Summary

Building a Terraform provider for Vast.ai is a well-understood problem class: a greenfield Go plugin using HashiCorp's Terraform Plugin Framework (v1.19.0, protocol v6) that wraps a REST API into declarative CRUD resources. The key challenge is not the Terraform plumbing — that is fully documented with official scaffolding templates — but rather the mismatch between Vast.ai's marketplace model and Terraform's declarative model. Vast.ai works like an eBay for GPU instances: users search for ephemeral offers, then create instances from a specific offer ID. Every API resource, data source, and workflow must be designed around this two-step pattern (offer search data source -> instance resource), and the architecture must handle the inevitable reality that spot instances vanish outside Terraform's control.

The recommended approach is a three-layer architecture: a thin provider registration shell, service-per-directory resource/data source implementations, and a clean API client that abstracts all HTTP concerns (auth, retry, rate limiting, error mapping). The project targets ~15 resources and ~10 data sources, which puts it firmly in the tier requiring the service-per-directory pattern from day one — not the flat monolith that the scaffolding template starts with. The most critical architectural decision is to separate `internal/client/` from `internal/services/`, making the Vast.ai API client independently testable and decoupled from Terraform schema logic.

The top risks are: (1) credential leakage via query-parameter auth (must use Authorization header instead), (2) schema misconfiguration causing "inconsistent result after apply" errors (requires careful Optional/Computed/Computed flag design for every attribute), and (3) acceptance test cost accumulation (GPU instances bill continuously; sweepers and cost gates are mandatory). An existing provider (`aalekhpatel07/vastai` v0.1.3) is abandoned and covers only ~2 resources, leaving the entire feature space open. The MVP target (v0.1.0) should be registry-publishable with instance lifecycle, template, SSH key management, and the GPU offer search data source.

## Key Findings

### Recommended Stack

The stack is well-defined with no meaningful alternatives for any component. Go 1.25.x with Terraform Plugin Framework v1.19.0 is mandatory — SDKv2 is legacy and explicitly not recommended for new providers. The `go-retryablehttp` library provides the HTTP transport layer with built-in retry and backoff, eliminating the need to hand-roll exponential backoff. GoReleaser v2 handles the exact release artifacts the Terraform Registry expects (cross-compiled binaries, GPG-signed checksums, registry manifest). See `.planning/research/STACK.md` for full version matrix and rationale.

**Core technologies:**
- Go 1.25.x — language runtime required by Plugin Framework v1.19.0 and testing framework
- Terraform Plugin Framework v1.19.0 — only correct choice for new providers; GA status, protocol v6, type-safe schema
- terraform-plugin-testing v1.15.0 — official acceptance test framework; runs real Terraform operations
- hashicorp/go-retryablehttp v0.7.8 — HTTP client with exponential backoff, 429/503 handling; HashiCorp ecosystem standard
- terraform-plugin-framework-validators v0.19.0 — pre-built validators for string, numeric, and collection attributes
- GoReleaser v2.14.x — cross-compilation, GPG signing, registry artifact generation
- golangci-lint v2.11.x — linting meta-tool; v2 config format required

### Expected Features

The existing provider is abandoned at v0.1.3 (Feb 2024) with ~2 resources. Full coverage of all 25+ Vast.ai resource types is the primary differentiator. See `.planning/research/FEATURES.md` for complete prioritized feature matrix.

**Must have (table stakes — v0.1.0):**
- Provider authentication via `VASTAI_API_KEY` env var + provider block attribute
- GPU offer search data source — entry point for ALL compute workflows; cannot create instances without it
- Instance resource (full CRUD + start/stop + import) — core compute primitive
- Template resource (CRUD + import) — required dependency for instances
- SSH Key resource (CRUD + import) — required for instance access
- HTTP retry with exponential backoff — APIs have rate limits; without retry, `terraform apply` fails unpredictably
- Import support for all resources — `ImportState` on every resource; HashiCorp design principle
- Attribute validators on constrained fields — catch errors at validate time, not API call time
- Configurable timeouts per resource — GPU instance creation/deletion is long-running
- GoReleaser + GitHub Actions release pipeline + GPG signing — required for registry publication
- Generated documentation via tfplugindocs — registry renders docs from `docs/` directory

**Should have (competitive — v0.2-v0.5):**
- Volume and Network Volume resources — persistent storage workflows
- Endpoint + Worker Group + Autoscaler resources — serverless inference at scale
- API Key and Environment Variable resources — account management primitives
- Plan modifiers (UseStateForUnknown, RequiresReplace) — reduces noisy plan output
- Bid price management on instances — marketplace cost optimization
- Instance list data source and user profile data source
- `most_affordable` computed attribute on offer data source — common user need

**Defer (v1.0+):**
- Cluster and Overlay resources — advanced networking, small audience
- Team, Team Role, Subaccount resources — organizational RBAC, complex design
- Ephemeral resource for API keys — requires Terraform 1.10+
- Identity-based import — requires Terraform 1.12+
- Terraform actions for instance operations (reboot, recycle) — requires Terraform 1.14+
- Guide documentation (training workflows, serverless setup)

**Anti-features to avoid:**
- Instance command execution resource — imperative, violates declarative model; use provisioners
- Bulk instance creation resource — use `count`/`for_each` instead
- Financial write operations (credit transfer) — never expose as Terraform resources
- Machine host-side operations — wrong audience; build a separate provider if needed

### Architecture Approach

The provider follows a strict three-layer design: Provider Shell (Terraform Plugin Framework integration, auth, client creation, resource registration) -> Service Layer (`internal/services/<name>/` directories, each self-contained with resource/data source implementations and Terraform model structs) -> API Client (`internal/client/`, pure Go HTTP wrapper with service objects, zero Terraform dependencies). This mirrors the patterns used by terraform-provider-aws (263 service directories), terraform-provider-cloudflare (242 service directories), and terraform-provider-digitalocean. The client-services separation is critical: it enables independent unit testing of the API client with `httptest`, keeps Terraform schema logic isolated from HTTP concerns, and allows the client to evolve when Vast.ai changes their API. See `.planning/research/ARCHITECTURE.md` for full data flow diagrams, interface signatures, and code examples.

**Major components:**
1. `main.go` + `internal/provider/` — binary entry point, gRPC server, auth config, resource/data source registration; thin layer with no business logic
2. `internal/client/` — VastAIClient with service sub-objects (InstanceService, TemplateService, etc.), auth via Authorization header, exponential backoff, error normalization; zero Terraform dependencies
3. `internal/services/<name>/` — one directory per resource type; resource/data source CRUD implementations, Terraform model structs (tfsdk tags), acceptance tests; registers via constructor functions
4. `internal/acctest/` — shared acceptance test helpers, provider factories, test sweepers for cost control
5. `internal/validators/` + `internal/planmodifiers/` — custom validation and plan modification logic

### Critical Pitfalls

1. **API key in query parameters** — Vast.ai API uses `?api_key=` by default; Terraform logs URLs at TRACE level; use `Authorization: Bearer` header exclusively. Address in Phase 1 before any resource is built. Mark `api_key` attribute `Sensitive: true`.

2. **"Provider produced inconsistent result after apply" from schema misconfiguration** — Every attribute the API can modify must be `Computed: true`. Server-generated fields (ID, status, IP) must be `Computed: true` only. After Create, always call Read to populate state from API. Apply `UseStateForUnknown()` to stable computed fields. Address in Phase 2 with a schema design checklist enforced project-wide.

3. **No waiter/poller logic for eventually-consistent instance lifecycle** — Instance creation is asynchronous; the API returns immediately but the instance is not yet running. Without a poller, `terraform apply` completes with incorrect state. Implement a waiter (poll `actual_status` until `running` or timeout). Same pattern required for Delete. Address in Phase 1-2 (client primitives + instance resource).

4. **Spot instance preemption breaks state** — Marketplace instances vanish without Terraform's knowledge. Read() must call `resp.State.RemoveResource(ctx)` on 404 or terminal status, not error. This is expected behavior, not a bug. Document clearly.

5. **Offer ID vs. machine spec confusion** — Unlike AWS/GCP, Vast.ai requires selecting a specific offer ID. Design instance resource to take `offer_id` as required attribute. Mark with `RequiresReplace()`. Handle race condition when offer expires between plan and apply with clear error message. Design the GPU offer data source and instance resource as a pair.

6. **Acceptance test cost accumulation** — GPU instances bill continuously. Implement sweepers for every resource type, use cheapest possible offers in tests, gate acceptance tests behind `TF_ACC=1`, run acceptance tests only on merge to main. Add `CheckDestroy` to every test.

7. **Registry publication failure from wrong GPG key type** — Modern systems default to ECC keys; the Terraform Registry only accepts RSA/DSA. Generate RSA GPG key explicitly. Set up CI/CD and publish a `v0.1.0-alpha` release early to validate the entire pipeline before writing provider code.

## Implications for Roadmap

Based on the dependency graph in FEATURES.md and the build order in ARCHITECTURE.md, a 6-phase structure is recommended:

### Phase 1: Foundation — Project Scaffold + API Client
**Rationale:** Everything else depends on this. Registry publication pipeline must be validated before provider code is written (Pitfall 10). API client security (auth headers, no credential leaks) must be correct before any resource is built. Retry logic and rate limiting belong in the transport layer, not per-resource.
**Delivers:** Compilable provider skeleton with no resources; working CI/CD; `v0.1.0-alpha` release that installs via `terraform init`; Vast.ai API client with auth, retry, error handling, and rate limiting.
**Addresses:** Provider auth, configurable API endpoint, User-Agent header, HTTP retry with exponential backoff, GPG-signed release pipeline.
**Avoids:** API key credential leakage (Pitfall 1), registry publication failure (Pitfall 10), Python SDK as API contract trap (Pitfall 8), rate limiting cascades (Pitfall 9).

### Phase 2: Core Compute — GPU Offers + Instance + Template + SSH Key
**Rationale:** The GPU offer data source is the entry point for ALL compute workflows; it must exist before the instance resource can be tested. Instance resource is the highest-value, highest-risk resource and validates the full stack end-to-end. Template and SSH Key are required dependencies for realistic instance configurations. This phase proves the architecture.
**Delivers:** Complete instance lifecycle workflow: search offers -> create instance from offer -> manage instance (start/stop) -> destroy. Full acceptance test suite for these resources. Import support for all three resources.
**Addresses:** GPU offer search data source, instance resource (CRUD + start/stop), template resource, SSH key resource, configurable timeouts, attribute validators, import support.
**Avoids:** Offer ID vs. spec confusion (Pitfall 5), schema misconfiguration (Pitfall 2), waiter/poller gaps (Pitfall 3), spot instance preemption (Pitfall 4), import incomplete state (Pitfall 7).
**Research flag:** Needs deep validation of Vast.ai instance API response shapes — `GET /api/v0/instances/{id}/` response contract must be verified with curl against live API to identify which creation-time attributes are returned.

### Phase 3: Storage Resources — Volumes + Network Volumes
**Rationale:** Storage is a natural second workflow after compute. Volumes follow the same offer-then-create pattern as instances, so the patterns from Phase 2 carry forward. Volume and network volume APIs are independent of each other and of Phase 4 serverless.
**Delivers:** Volume and network volume resources with CRUD and import; volume offer search data sources.
**Addresses:** Volume resource, Network Volume resource, volume offer data source, network volume offer data source.
**Avoids:** Same eventual consistency pitfalls as instances; volume provisioning is async.

### Phase 4: Serverless Workflows — Endpoints + Worker Groups + Autoscalers
**Rationale:** Serverless is a distinct, high-value Vast.ai product. Endpoint -> Worker Group -> Autoscaler is a strict dependency chain (each requires the previous). This phase requires template resource from Phase 2 as a dependency.
**Delivers:** Complete serverless inference endpoint setup: create endpoint -> configure worker group -> attach autoscaler. Full coverage of Vast.ai's serverless product.
**Addresses:** Endpoint resource, Worker Group resource, Autoscaler resource.
**Research flag:** Worker Group and Autoscaler APIs have complex configuration (min_load, target_util, cold_mult) — likely needs `research-phase` during planning to map Python SDK parameters to Go structs.

### Phase 5: Account Management — API Keys + Env Vars + Additional Data Sources
**Rationale:** Account management resources (API keys, environment variables) are low-complexity and standalone (no dependencies on other resources). Additional data sources (user profile, instance list, template search) complete the coverage for common workflows.
**Delivers:** Full account configuration management via Terraform. Comprehensive data source coverage.
**Addresses:** API Key resource, Environment Variable resource, User profile data source, Instance list data source, Template search data source, plan modifiers (UseStateForUnknown, RequiresReplace), bid price management.

### Phase 6: Polish + Advanced Features
**Rationale:** Cluster/Overlay networking is a small audience with complex lifecycle. Team/RBAC management requires careful design. Documentation guides require all features to be stable. This phase targets v1.0.
**Delivers:** Full API coverage including advanced networking and organizational management. Guide documentation for common workflows. State upgrade infrastructure for breaking schema changes.
**Addresses:** Cluster resource, Overlay resource, Team/Role/Subaccount resources, guide documentation, ephemeral API key resource (Terraform 1.10+), identity-based import (Terraform 1.12+).
**Research flag:** Team and RBAC APIs need research — permissions model and role scope are not well-documented in the Python SDK.

### Phase Ordering Rationale

- Phase 1 before everything because the release pipeline and API client are shared foundations; a broken auth design poisons all resources built on top.
- Phase 2 before Phase 3/4 because the offer-then-create pattern, schema design conventions, and waiter patterns established here must be consistent across all resources. Getting these wrong in Phase 2 means fixing them in all subsequent resources.
- Phase 3 and 4 can proceed in parallel after Phase 2 completes — volumes and serverless have no dependency on each other.
- Phase 5 is low-risk and can overlap with Phase 4; account management resources are simple enough to implement alongside more complex work.
- Phase 6 is deferred because cluster/overlay/team have small audiences, high complexity, and require stable foundations.

### Research Flags

Phases likely needing deeper research (`/gsd:research-phase`) during planning:
- **Phase 2 (Instance API):** The Vast.ai instance API response contract must be verified with live API calls. Which creation-time attributes are returned by `GET /api/v0/instances/{id}/` directly impacts schema design (Computed vs. Optional+Computed) and import completeness.
- **Phase 4 (Serverless APIs):** Worker Group and Autoscaler configuration parameters are not fully documented; Python SDK is the only reference and mixes business logic with HTTP calls. Needs extraction of the actual HTTP contracts.
- **Phase 6 (Team/RBAC APIs):** Permissions model is not clearly documented in the Python SDK. The scope of team roles and how they interact with subaccounts needs research before schema design.

Phases with standard, well-documented patterns (skip `research-phase`):
- **Phase 1:** Terraform provider scaffolding, GoReleaser, GitHub Actions release pipelines are extensively documented with official templates.
- **Phase 3 (Volumes):** Follows same offer-then-create pattern as instances; patterns established in Phase 2 apply directly.
- **Phase 5 (Account Management):** API Key and Environment Variable are simple CRUD resources; standard patterns apply.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All versions verified via GitHub releases, pkg.go.dev, and HashiCorp scaffolding repo. No meaningful alternatives exist for any component. |
| Features | HIGH | Based on HashiCorp design principles, existing provider gap analysis, and Vast.ai Python SDK as feature reference. Anti-features and deferral decisions are well-reasoned. |
| Architecture | HIGH | Three-layer pattern verified against terraform-provider-aws, cloudflare, and digitalocean production architectures. Interface signatures from official Plugin Framework docs. |
| Pitfalls | HIGH | Verified against HashiCorp official docs, AWS provider contributor guides, and Vast.ai Python SDK source. Most pitfalls are Terraform-ecosystem patterns, not Vast.ai-specific. |

**Overall confidence:** HIGH

### Gaps to Address

- **Vast.ai API response contracts:** No OpenAPI spec exists. The Python SDK is the only reference but is a CLI tool, not an API library. Every endpoint's actual request/response JSON shape must be verified with `curl` against the live API during Phase 1-2. This affects schema design for every resource.
- **Vast.ai rate limits:** The API's rate limit thresholds are undocumented. The Python SDK retries on 429 with a 150ms starting backoff, but the actual limits are unknown. The provider should implement conservative rate limiting (shared `golang.org/x/time/rate` limiter) and make retry count configurable. Real limits will be discovered through acceptance testing.
- **Vast.ai offer expiry timing:** The time window between `terraform plan` (offer discovered) and `terraform apply` (instance created) during which an offer expires is unknown. This affects UX guidance. Handle with clear error messages and documented `terraform plan -refresh=true` workflow.
- **Instance field returnability:** Whether creation-time fields like `onstart_cmd`, `template_hash`, and original `bid_price` are returned by the instance GET endpoint is unknown without live API testing. This directly determines which attributes can support `ImportStateVerify: true` in acceptance tests.

## Sources

### Primary (HIGH confidence)
- [Terraform Plugin Framework GitHub Releases](https://github.com/hashicorp/terraform-plugin-framework/releases) — version history, v1.19.0 confirmed
- [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) — go.mod dependencies, .goreleaser.yml, project structure
- [HashiCorp Developer - Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) — official documentation, best practices, interface definitions
- [HashiCorp Provider Design Principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles) — resource modeling, CRUD behavior
- [Terraform Registry Publishing Requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing) — GPG key types, GoReleaser config, manifest requirements
- [AWS Provider Contributor Guide](https://hashicorp.github.io/terraform-provider-aws/) — retries, waiters, error handling patterns
- [terraform-provider-aws internal structure](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal) — service-per-directory pattern reference
- [terraform-provider-cloudflare internal structure](https://github.com/cloudflare/terraform-provider-cloudflare/tree/main/internal/services) — service-per-directory pattern at scale
- [Vast.ai Python SDK](vast-sdk/vastai/vast.py) — HTTP endpoint contracts, auth pattern, retry logic

### Secondary (MEDIUM confidence)
- [go-retryablehttp on pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/go-retryablehttp) — v0.7.8 confirmed; no formal releases page but stable and widely used
- [CoreWeave Terraform Provider](https://registry.terraform.io/providers/coreweave/coreweave/latest/docs) — comparable GPU cloud provider as feature reference
- [Existing vastai provider (abandoned)](https://registry.terraform.io/providers/aalekhpatel07/vastai/latest/docs) — confirmed v0.1.3, last update Feb 2024; gap analysis baseline
- [OpenTofu Compatibility Promises](https://opentofu.org/docs/language/v1-compatibility-promises/) — provider protocol compatibility

### Tertiary (LOW confidence)
- Vast.ai API rate limits — undocumented; discovered through SDK source code comments and community experience; needs validation through testing
- Instance field returnability on GET — not explicitly documented; must be verified empirically during Phase 2

---
*Research completed: 2026-03-25*
*Ready for roadmap: yes*
