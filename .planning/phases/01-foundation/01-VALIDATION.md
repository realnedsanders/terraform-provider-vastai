---
phase: 1
slug: foundation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (standard library) |
| **Config file** | none — Wave 0 creates go.mod and test files |
| **Quick run command** | `go test ./internal/client/... -count=1 -short` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/client/... -count=1 -short`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | FOUND-01,02 | unit | `go build ./...` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 1 | FOUND-03,05 | unit | `go test ./internal/client/... -count=1` | ❌ W0 | ⬜ pending |
| 01-02-02 | 02 | 1 | FOUND-04 | unit | `go test ./internal/client/... -run TestUserAgent -count=1` | ❌ W0 | ⬜ pending |
| 01-02-03 | 02 | 1 | FOUND-06 | unit | `go test ./internal/client/... -run TestError -count=1` | ❌ W0 | ⬜ pending |
| 01-03-01 | 03 | 2 | RLSE-01,02 | manual | verify goreleaser config with `goreleaser check` | ❌ W0 | ⬜ pending |
| 01-03-02 | 03 | 2 | RLSE-03,04,05 | manual | verify release artifacts in `.goreleaser.yml` | ❌ W0 | ⬜ pending |
| 01-03-03 | 03 | 2 | TEST-04 | manual | verify CI workflow files exist with correct triggers | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` — Go module initialization with correct module path
- [ ] `internal/client/client_test.go` — unit tests for HTTP client, auth, retry, error handling
- [ ] `main.go` — provider entry point that compiles

*Test infrastructure is established as part of the phase itself.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `terraform init` installs provider | FOUND-01 | Requires Terraform binary and registry | Install locally, run `terraform init` with provider source |
| GitHub tag triggers release | RLSE-01,02 | Requires GitHub Actions and GPG secrets | Push test tag, verify workflow runs and artifacts published |
| GoReleaser cross-compiles | RLSE-05 | Requires goreleaser binary | Run `goreleaser check` locally to validate config |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
