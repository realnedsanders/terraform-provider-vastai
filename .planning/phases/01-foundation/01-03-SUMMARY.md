---
phase: 01-foundation
plan: 03
subsystem: infra
tags: [goreleaser, github-actions, ci-cd, gpg-signing, golangci-lint]

# Dependency graph
requires:
  - phase: 01-foundation/01
    provides: "Go module with go.mod (go-version-file reference for CI)"
provides:
  - "GoReleaser v2 config with cross-compilation and GPG-signed checksums"
  - "GitHub Actions release workflow triggered by v* tags"
  - "GitHub Actions test workflow with build/lint/unit/acceptance jobs"
  - "Comprehensive .gitignore for Go Terraform provider project"
affects: [all-phases]

# Tech tracking
tech-stack:
  added: [goreleaser-v2, golangci-lint-v2.11, github-actions]
  patterns: [tag-triggered-release, acceptance-test-gating, gpg-signed-checksums]

key-files:
  created:
    - .goreleaser.yml
    - .github/workflows/release.yml
    - .github/workflows/test.yml
  modified:
    - .gitignore

key-decisions:
  - "GoReleaser v2 format with version: 2 header"
  - "GPG signing on checksum file only (artifacts: checksum) with --batch flag for CI non-interactive mode"
  - "Acceptance tests gated by github.ref == refs/heads/main to avoid API costs on PRs"
  - "golangci-lint v2.11 via golangci-lint-action@v8"

patterns-established:
  - "CI workflow pattern: build job gates test and acceptance jobs via needs dependency"
  - "Release pipeline: v* tag -> GPG import -> GoReleaser -> signed checksums + manifest"
  - "Acceptance test gating: TF_ACC=1 + VASTAI_API_KEY on main branch only"

requirements-completed: [RLSE-01, RLSE-02, RLSE-05, TEST-04]

# Metrics
duration: 2min
completed: 2026-03-25
---

# Phase 1 Plan 3: CI/CD Pipeline Summary

**GitHub Actions CI/CD with GoReleaser v2 cross-compilation, GPG-signed SHA256SUMS, and acceptance test gating on main branch**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-25T19:21:04Z
- **Completed:** 2026-03-25T19:23:14Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- GoReleaser v2 config producing cross-compiled binaries for linux/darwin/windows on amd64/arm64 with GPG-signed SHA256SUMS
- Release workflow triggered by v* tags with GPG key import via crazy-max/ghaction-import-gpg@v7 and goreleaser-action@v7
- Test workflow with build/lint/unit tests on PRs and acceptance tests gated to main branch only
- Comprehensive .gitignore covering Go binaries, Terraform working dirs, GoReleaser dist, IDE files, and env files

## Task Commits

Each task was committed atomically:

1. **Task 1: Create GoReleaser config and release workflow** - `b7d0fb7` (feat)
2. **Task 2: Create test workflow and .gitignore** - `f3eff6f` (feat)

## Files Created/Modified
- `.goreleaser.yml` - GoReleaser v2 config with cross-compilation, SHA256SUMS, GPG signing, and manifest inclusion
- `.github/workflows/release.yml` - Release workflow: v* tag trigger, GPG import, GoReleaser execution
- `.github/workflows/test.yml` - CI workflow: build+lint on PRs, unit tests, acceptance tests on main
- `.gitignore` - Comprehensive ignore patterns for Go Terraform provider project

## Decisions Made
- GoReleaser v2 format (version: 2) used as the current standard
- GPG signing targets checksum file only (artifacts: checksum) with --batch flag to prevent TTY hang in CI
- Acceptance tests gated by `github.ref == 'refs/heads/main'` to avoid spending real API credits on PRs
- golangci-lint v2.11 via golangci-lint-action@v8 for Go linting
- terraform-registry-manifest.json included in both checksum.extra_files and release.extra_files per Terraform Registry requirements

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required

**External services require manual configuration** before CI/CD workflows will function:

### GitHub Actions Secrets
1. **GPG_PRIVATE_KEY** - RSA 4096-bit GPG private key (NOT ECC) for release signing
   - Generate: `gpg --full-generate-key` (select RSA, 4096 bits)
   - Export: `gpg --armor --export-secret-keys KEY_ID`
   - Add at: GitHub repo -> Settings -> Secrets and variables -> Actions -> New repository secret

2. **GPG_PASSPHRASE** - Passphrase used when generating the GPG key
   - Add at: GitHub repo -> Settings -> Secrets and variables -> Actions -> New repository secret

3. **VASTAI_API_KEY** - Vast.ai API key for acceptance tests on main branch
   - Obtain from: Vast.ai console -> Account -> API Keys
   - Add at: GitHub repo -> Settings -> Secrets and variables -> Actions -> New repository secret

### Terraform Registry
4. Add the GPG **public** key to Terraform Registry
   - Export: `gpg --armor --export KEY_ID`
   - Add at: Terraform Registry -> Publishing -> GPG Keys -> Add Key

## Known Stubs

None - all files are complete configurations, no placeholder values.

## Next Phase Readiness
- CI/CD infrastructure complete and ready for all subsequent phases
- Test workflow will activate once Go source code exists (build/lint/test)
- Release workflow will activate on first v* tag push after GPG secrets are configured
- All workflows reference go-version-file: 'go.mod' so Go version is centrally managed

## Self-Check: PASSED

All created files verified on disk. All task commits verified in git history.

---
*Phase: 01-foundation*
*Completed: 2026-03-25*
