# terraform-provider-vastai

## What This Is

A production-grade Terraform/OpenTofu provider for Vast.ai written in Go. It enables infrastructure-as-code management of Vast.ai GPU compute, storage, networking, serverless endpoints, and account resources — bringing Vast.ai into existing IaC workflows alongside other cloud providers. Published to the Terraform Registry for community use.

## Core Value

Full, reliable IaC control over Vast.ai infrastructure — every API resource manageable through Terraform with the same quality bar as first-party providers.

## Requirements

### Validated

- Provider authenticates via VASTAI_API_KEY environment variable — Phase 1
- Go API client with Bearer auth, exponential backoff retry, structured errors — Phase 1
- CI/CD pipeline with GitHub Actions and goreleaser for signed releases — Phase 1
- Built on Terraform Plugin Framework (not SDKv2) — Phase 1

### Active

- [ ] Go API client covers all Vast.ai REST endpoints (ported from Python SDK)
- [ ] Go API client covers all Vast.ai REST endpoints (ported from Python SDK)
- [ ] Instance resource with full CRUD, start/stop, reboot, label, bid management
- [ ] Template resource with full CRUD
- [ ] SSH Key resource with full CRUD and instance attach/detach
- [ ] Volume and Network Volume resources with CRUD, clone, list/unlist
- [ ] Serverless Endpoint resource with CRUD and autoscaling configuration
- [ ] Worker Group resource with CRUD
- [ ] Cluster and Overlay resources with CRUD and membership management
- [ ] API Key resource with CRUD and permission management
- [ ] Team, Team Role, and Subaccount resources with CRUD
- [ ] Environment Variable resource with CRUD
- [ ] Data source for GPU offer search with full filter support
- [ ] Data source for machine listing
- [ ] Data source for template search
- [ ] Data source for volume/network volume offers
- [ ] Data source for user profile, invoices, and audit logs
- [ ] Data source for endpoint status and worker health
- [ ] Acceptance tests against real Vast.ai API for all resources
- [ ] Generated documentation via tfplugindocs
- [ ] CI/CD pipeline with GitHub Actions and goreleaser for signed releases
- [ ] Examples directory with working configurations for common workflows
- [ ] Import support for all managed resources
- [ ] Built on Terraform Plugin Framework (not SDKv2)

### Out of Scope

- Serverless request routing / inference proxying — runtime concern, not IaC
- 2FA management — security-sensitive, better handled interactively
- File transfer operations (vm_copy, cloud_copy) — imperative operations, not declarative resources
- Billing/credit transfer operations — financial operations, not infrastructure
- Instance command execution — imperative, use provisioners or external tooling
- Scheduled job management — better suited to external schedulers
- Maintenance scheduling — provider-side (host) operations, not tenant IaC

## Context

- **API Surface:** Vast.ai exposes ~126 operations across 25+ resource types via REST at `https://console.vast.ai/api/v0/`
- **Reference Implementation:** Python SDK cloned at `vast-sdk/` — authoritative source for API behavior, endpoint paths, query parameters, and response shapes
- **Authentication:** API key passed as query parameter `?api_key=<key>` on all requests
- **API Style:** REST with GET (list/show), PUT (create/update), DELETE (destroy), POST (actions)
- **No official Go SDK exists** — we build a Go API client from scratch using the Python SDK as reference
- **Terraform Plugin Framework** is the modern approach (vs legacy SDKv2) — required for new providers per HashiCorp guidance
- **Provider naming:** `terraform-provider-vastai` following registry conventions
- **Registry namespace:** `realnedsanders/vastai` (personal account)
- **GitHub repo:** `github.com/realnedsanders/terraform-provider-vastai`

## Constraints

- **Language**: Go — required by Terraform provider ecosystem
- **Framework**: Terraform Plugin Framework — HashiCorp's current recommendation for new providers
- **API Reference**: Python SDK at `vast-sdk/` is the source of truth for API behavior
- **Testing**: Acceptance tests require a real Vast.ai account and API key (no mock API available)
- **Auth**: `VASTAI_API_KEY` environment variable — matches existing SDK convention
- **Registry**: Must follow HashiCorp's publishing requirements (signed binaries, documentation format, naming conventions)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Terraform Plugin Framework over SDKv2 | HashiCorp's current guidance for new providers; better type safety, testing support | — Pending |
| Hand-written Go API client over code generation | No OpenAPI spec available; Python SDK is the reference | — Pending |
| Full API coverage in v1 | User requirement — bring all Vast.ai resources under IaC | — Pending |
| Exclude imperative operations (file transfer, exec) | Not declarative resources — don't fit Terraform's model | — Pending |
| VASTAI_API_KEY env var for auth | Matches existing SDK convention, standard Terraform pattern | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-25 after Phase 1 completion*
