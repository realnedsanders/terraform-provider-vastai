---
phase: 3
slug: storage
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (standard library) + terraform-plugin-testing |
| **Config file** | go.mod (existing) |
| **Quick run command** | `go test ./internal/client/... ./internal/services/volume/... ./internal/services/networkvolume/... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~10 seconds (unit) |

---

## Sampling Rate

- **After every task commit:** Run quick command
- **After every plan wave:** Run full suite
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Wave 0 Requirements

- [ ] `internal/client/volumes.go` — Volume service API methods
- [ ] `internal/client/network_volumes.go` — Network volume service API methods
- [ ] `internal/services/volume/` — Volume resource and data source
- [ ] `internal/services/networkvolume/` — Network volume resource and data source

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Volume clone creates copy | STOR-02 | Requires existing volume to clone from | Create volume, then create clone, verify independent lifecycle |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity maintained
- [ ] Wave 0 covers all MISSING references
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
