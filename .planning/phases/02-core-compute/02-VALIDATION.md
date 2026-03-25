---
phase: 2
slug: core-compute
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (standard library) + terraform-plugin-testing |
| **Config file** | go.mod (existing) |
| **Quick run command** | `go test ./internal/client/... ./internal/services/... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~10 seconds (unit), ~5 minutes (acceptance) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/client/... ./internal/services/... -count=1 -short`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Wave 0 Requirements

- [ ] `internal/client/instances.go` — Instance service API methods
- [ ] `internal/client/offers.go` — Offer search API methods
- [ ] `internal/client/templates.go` — Template service API methods
- [ ] `internal/client/sshkeys.go` — SSH key service API methods
- [ ] `internal/services/instance/` — Instance resource and data sources
- [ ] `internal/services/template/` — Template resource
- [ ] `internal/services/sshkey/` — SSH key resource and data source

*All test files created within plans.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `terraform import` populates state correctly | IMPT-01 | Requires existing Vast.ai resources | Create resource via API, run terraform import, verify plan shows no diff |
| Instance preemption handling | COMP-04 | Requires actual spot eviction event | Create interruptible instance, wait for preemption or simulate via API |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
