# terraform-provider-vastai

## What This Is

A production-grade Terraform/OpenTofu provider for Vast.ai written in Go. Manages 17 resource types and 11 data sources across GPU compute, storage, serverless endpoints, networking, and account management. Published to the Terraform Registry at `realnedsanders/vastai`.

## Core Value

Full, reliable IaC control over Vast.ai infrastructure — every API resource manageable through Terraform with the same quality bar as first-party providers.

## Current State

**v1.0 shipped 2026-03-28** — Full API coverage achieved.

- 17 resources, 11 data sources, 25,967 lines of Go
- 100+ unit tests, 15 acceptance tests, 10 sweepers
- Published to Terraform Registry with generated docs and examples
- CI/CD: GoReleaser with GPG signing, GitHub Actions (build/lint/test/sweep)

## Requirements

### Validated

- Provider authenticates via VASTAI_API_KEY environment variable — v1.0
- Go API client with Bearer auth, exponential backoff retry, structured errors — v1.0
- CI/CD pipeline with GitHub Actions and goreleaser for signed releases — v1.0
- Built on Terraform Plugin Framework (not SDKv2) — v1.0
- Instance resource with full CRUD, start/stop, label, bid management — v1.0
- Template resource with full CRUD — v1.0
- SSH Key resource with full CRUD and instance attach/detach — v1.0
- Volume and Network Volume resources with CRUD, clone — v1.0
- Serverless Endpoint resource with CRUD and autoscaling — v1.0
- Worker Group resource with CRUD — v1.0
- Cluster and Overlay resources with CRUD and membership — v1.0
- API Key, Team, Team Role, Team Member, Subaccount, Env Var resources — v1.0
- GPU offer search, instance, template, SSH key, volume, endpoint data sources — v1.0
- User profile, invoices, audit logs data sources — v1.0
- Acceptance tests for all resources — v1.0
- Generated documentation via tfplugindocs — v1.0
- Examples directory with workflow configurations — v1.0
- Import support for all managed resources — v1.0
- Resource sweepers for CI cleanup — v1.0

### Active

(None — v1.0 milestone complete. Next milestone TBD.)

### Out of Scope

- Serverless request routing / inference proxying — runtime concern, not IaC
- 2FA management — security-sensitive, better handled interactively
- File transfer operations (vm_copy, cloud_copy) — imperative, not declarative
- Billing/credit transfer operations — financial operations, not infrastructure
- Instance command execution — imperative, use provisioners
- Scheduled job management — better suited to external schedulers
- Host-side machine operations — wrong audience (provider targets tenants)
- Marketplace volume list/unlist — confirmed as host-only operations in v1.0

## Context

- **Codebase:** 25,967 lines of Go across 23 packages
- **Registry:** `realnedsanders/vastai` with full documentation
- **GitHub:** `github.com/realnedsanders/terraform-provider-vastai`
- **API Surface:** 17 resources covering instances, templates, SSH keys, volumes, network volumes, endpoints, worker groups, clusters, overlays, API keys, env vars, teams, roles, members, subaccounts
- **Data Sources:** 11 covering GPU offers, instances, templates, SSH keys, volume offers, network volume offers, endpoints, user profile, invoices, audit logs

## Constraints

- **Language**: Go — required by Terraform provider ecosystem
- **Framework**: Terraform Plugin Framework v1.19.0
- **Testing**: Acceptance tests require real Vast.ai API key
- **Auth**: Bearer header authentication (never query parameter)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Plugin Framework over SDKv2 | HashiCorp's current guidance; better type safety | Good |
| Hand-written Go API client | No OpenAPI spec; Python SDK as reference | Good |
| Full API coverage in v1 | User requirement — all resources under IaC | Good |
| Exclude imperative operations | Don't fit Terraform's declarative model | Good |
| Service-per-directory layout | Scales to 17 resources without flat file bloat | Good |
| Bearer header auth (not query param) | Prevents credential leaks in logs | Good |
| Permissions as JSON string (not flat set) | API uses nested JSON objects — D-02 revised | Good |
| Marketplace list/unlist omitted | Confirmed as host-only operations via SDK research | Good |
| No standalone autogroup resource | API has no separate autogroup endpoint — SRVL-03 via endpoint | Good |
| Subaccount create-only (no-op destroy) | API has no delete endpoint | Revisit |
| Overlay member join-only (no-op destroy) | API has no remove-instance endpoint | Revisit |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition:**
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone:**
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-28 after v1.0 milestone*
