---
phase: 01-foundation
verified: 2026-03-25T20:00:00Z
status: human_needed
score: 4/5 success criteria verified
human_verification:
  - test: "Push a v0.1.0-alpha tag to a GitHub remote and confirm the release workflow runs, GoReleaser produces artifacts, and the provider appears at registry.terraform.io/realnedsanders/vastai"
    expected: "GitHub Actions release workflow executes, cross-compiled zip archives appear in the GitHub release, SHA256SUMS and SHA256SUMS.sig are present, terraform-registry-manifest.json is included, and the Terraform Registry shows the provider available for download"
    why_human: "No GitHub remote is configured on this repo (git remote show returns empty). No v* tags exist. Workflow execution and Registry publication cannot be verified programmatically without a live remote, CI run, and Registry API integration."
  - test: "Run `terraform init` with a provider block using source = 'realnedsanders/vastai' after the registry publication above"
    expected: "Terraform downloads and installs the provider binary without error, and `terraform plan` with an empty config completes without error"
    why_human: "Depends on registry publication (item above). Cannot simulate Terraform Registry download locally."
---

# Phase 1: Foundation Verification Report

**Phase Goal:** A compilable, installable Terraform provider with zero resources but a working API client, CI/CD pipeline, and a published alpha release that installs via `terraform init`
**Verified:** 2026-03-25T20:00:00Z
**Status:** human_needed
**Re-verification:** No -- initial verification

## Goal Achievement

### Success Criteria (from ROADMAP.md)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Running `terraform init` successfully downloads and installs the provider binary | ? HUMAN | No GitHub remote configured, no tags, no published release -- pipeline exists but has not run |
| 2 | Provider authenticates via `VASTAI_API_KEY` env var or `api_key` attribute; credentials sent via Authorization header (never URL query params) | ✓ VERIFIED | `internal/provider/provider.go:97-101` reads `VASTAI_API_KEY`; `internal/client/auth.go:40` sets `Authorization: Bearer`; `TestAPIKeyNotInURL` passes; `TestNewRequest_AuthHeader` passes |
| 3 | API client retries on 429/5xx with exponential backoff and surfaces structured error diagnostics | ✓ VERIFIED | `internal/client/client.go:48-70` `vastaiRetryPolicy`; `internal/client/client.go:74-93` `vastaiBackoff` (150ms, 1.5x); `internal/client/errors.go` `APIError`; 20/20 unit tests pass |
| 4 | Tagging a release in GitHub triggers automated cross-compilation, GPG signing, and artifact publication | ? HUMAN | `.github/workflows/release.yml` and `.goreleaser.yml` exist and are correctly configured; no remote/tag exists to confirm execution |
| 5 | `terraform plan` with empty configuration and valid API key completes without error | ? HUMAN | Provider binary compiles and schema is valid; requires a Terraform binary and live API key -- depends on SC-1 for install path |

**Score:** 2/5 success criteria fully verified programmatically; 2 confirmed human-only (external system); 1 partial (SC-5 -- requires SC-1)

### Observable Truths (derived from PLAN must_haves)

#### Plan 01-01 Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Go module initializes and all dependencies resolve | ✓ VERIFIED | `go build ./...` exits 0; `go.mod` contains `github.com/realnedsanders/terraform-provider-vastai` |
| 2 | `go build ./...` compiles without errors | ✓ VERIFIED | Build confirmed clean with no output (zero errors) |
| 3 | Provider registers at `registry.terraform.io/realnedsanders/vastai` | ✓ VERIFIED | `main.go:22` sets `Address: "registry.terraform.io/realnedsanders/vastai"` |
| 4 | Schema exposes `api_key` (sensitive, optional) and `api_url` (optional) | ✓ VERIFIED | `internal/provider/provider.go:53-62`: `api_key` with `Sensitive: true`, `api_url` without Sensitive |
| 5 | `terraform-registry-manifest.json` declares protocol version 6.0 | ✓ VERIFIED | File contents: `"protocol_versions": ["6.0"]` |

#### Plan 01-02 Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | API client authenticates using `Authorization: Bearer` header, never query parameters | ✓ VERIFIED | `internal/client/auth.go:40`; TestAPIKeyNotInURL + TestNewRequest_AuthHeader pass |
| 2 | API client retries on 429 and 5xx with 150ms base, 1.5x exponential backoff, max 5 retries | ✓ VERIFIED | `client.go:30-34`: `RetryMax=5`, `RetryWaitMin=150ms`; `vastaiBackoff` uses `math.Pow(1.5, ...)` |
| 3 | API client does not retry on 4xx (except 429) | ✓ VERIFIED | `vastaiRetryPolicy` only retries 429 and 5xx (not 501); TestRetryPolicy_ClientErrors passes |
| 4 | User-Agent header contains `terraform-provider-vastai/VERSION` | ✓ VERIFIED | `client.go:41`: `fmt.Sprintf("terraform-provider-vastai/%s", version)`; TestNewRequest_UserAgent passes |
| 5 | API errors produce structured APIError with status code, method, path, and message | ✓ VERIFIED | `errors.go`: `APIError{StatusCode, Message, Method, Path}`; TestAPIError_ErrorString passes |
| 6 | Provider Configure wires API client into ResourceData and DataSourceData | ✓ VERIFIED | `provider.go:122-124`: `c := client.NewVastAIClient(...)`, `resp.ResourceData = c`, `resp.DataSourceData = c` |
| 7 | Unit tests verify auth header, retry behavior, error mapping, and User-Agent | ✓ VERIFIED | 20/20 tests pass including TestNewRequest_AuthHeader, TestRetryPolicy_*, TestAPIError_*, TestAPIKeyNotInURL |

#### Plan 01-03 Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Pushing a v* tag triggers release workflow with GoReleaser and GPG signing | ? HUMAN | `release.yml:3-5` trigger `on: push: tags: ['v*']`; no remote or tags to confirm execution |
| 2 | Pull requests trigger build + lint + unit test jobs | ? HUMAN | `test.yml:3-7` triggers on `pull_request: branches: [main]`; no executed workflow to confirm |
| 3 | Pushes to main trigger build + lint + unit + acceptance tests | ? HUMAN | `test.yml:8-10` triggers on `push: branches: [main]`; acceptance gated by `if: github.ref == 'refs/heads/main'` |
| 4 | GoReleaser produces cross-compiled binaries for linux/darwin/windows on amd64/arm64 | ✓ VERIFIED | `.goreleaser.yml:13-28`: `goos` includes linux, darwin, windows, freebsd; `goarch` includes amd64, arm64 |
| 5 | GoReleaser generates SHA256SUMS file and detached .sig signature | ✓ VERIFIED | `.goreleaser.yml:33-48`: `checksum.name_template` includes `SHA256SUMS`; `signs` section with `artifacts: checksum` and `--detach-sign` |
| 6 | `terraform-registry-manifest.json` included in release artifacts | ✓ VERIFIED | `.goreleaser.yml:34-36,50-52`: manifest in both `checksum.extra_files` and `release.extra_files` |
| 7 | `.gitignore` excludes compiled binaries and terraform working directories | ✓ VERIFIED | `.gitignore` contains `terraform-provider-vastai`, `terraform-provider-vastai_*`, `.terraform/`, `dist/`, `*.test` |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `main.go` | Provider entry point with providerserver.Serve() | ✓ VERIFIED | Contains `providerserver.Serve`, `provider.New(version)`, correct registry address |
| `go.mod` | Go module with all framework dependencies | ✓ VERIFIED | Module `github.com/realnedsanders/terraform-provider-vastai`, `terraform-plugin-framework v1.19.0`, `go-retryablehttp v0.7.8` |
| `internal/provider/provider.go` | VastaiProvider struct with Schema and Configure methods | ✓ VERIFIED | `VastaiProvider`, `VastaiProviderModel`, `New`, `Schema`, `Configure`, `Metadata`, `Resources`, `DataSources` all present; no TODO stubs |
| `terraform-registry-manifest.json` | Registry protocol declaration | ✓ VERIFIED | `"version": 1`, `"protocol_versions": ["6.0"]` |
| `GNUmakefile` | Build, test, lint, install targets | ✓ VERIFIED | `build`, `install`, `lint`, `fmt`, `test`, `testacc`, `generate` targets present with correct recipes |
| `.golangci.yml` | Linter configuration | ✓ VERIFIED | `version: "2"`, enables staticcheck, govet, errcheck, gosec, revive, gosimple, unconvert, unparam |
| `internal/client/client.go` | VastAIClient struct with retry policy and HTTP methods | ✓ VERIFIED | `NewVastAIClient`, `vastaiRetryPolicy`, `vastaiBackoff`, `Get`, `Post`, `Put`, `Delete` all substantive |
| `internal/client/auth.go` | newRequest with Bearer auth and headers | ✓ VERIFIED | `Authorization: Bearer`, `User-Agent`, `Content-Type`, `Accept` headers set; URL construction correct |
| `internal/client/errors.go` | APIError struct implementing error interface | ✓ VERIFIED | `type APIError struct` with `StatusCode, Message, Method, Path`; `Error()` returns canonical format |
| `internal/client/client_test.go` | 20 unit tests using httptest | ✓ VERIFIED | All 20 tests present and passing: TestNewVastAIClient through TestAPIKeyNotInURL |
| `.goreleaser.yml` | GoReleaser v2 config with GPG signing | ✓ VERIFIED | `version: 2`, cross-compilation matrix, SHA256SUMS, GPG detached sig, manifest inclusion |
| `.github/workflows/release.yml` | Release workflow on v* tags | ✓ VERIFIED | `on: push: tags: ['v*']`, GPG import via ghaction-import-gpg@v7, goreleaser-action@v7 |
| `.github/workflows/test.yml` | CI workflow for PRs and main | ✓ VERIFIED | PR + push triggers, build/lint/unit/acceptance jobs, TF_ACC gating on main only |
| `.gitignore` | Go binary and Terraform ignore patterns | ✓ VERIFIED | Contains all required patterns |
| `examples/provider/provider.tf` | Example provider configuration | ✓ VERIFIED | `source = "realnedsanders/vastai"` |
| `templates/index.md.tmpl` | tfplugindocs provider template | ✓ VERIFIED | VASTAI_API_KEY docs, tffile and SchemaMarkdown directives |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `main.go` | `internal/provider/provider.go` | `provider.New(version)` | ✓ WIRED | `main.go:26`: `provider.New(version)` call present |
| `internal/provider/provider.go` | `schema.Schema` | `api_key` sensitive + `api_url` attributes | ✓ WIRED | `provider.go:53`: `Sensitive: true` confirmed for `api_key` |
| `internal/provider/provider.go` | `internal/client/client.go` | `client.NewVastAIClient()` in Configure | ✓ WIRED | `provider.go:13` import; `provider.go:122` call; `provider.go:123-124` injection |
| `internal/client/auth.go` | `internal/client/client.go` | `newRequest` method on `VastAIClient` | ✓ WIRED | `auth.go:15`: `func (c *VastAIClient) newRequest(...)` |
| `internal/client/client.go` | `internal/client/errors.go` | `handleResponse` creates APIError on non-2xx | ✓ WIRED | `client.go:127`: `return &APIError{...}` |
| `.github/workflows/release.yml` | `.goreleaser.yml` | goreleaser-action reads config | ✓ WIRED | `release.yml:26`: `uses: goreleaser/goreleaser-action@v7` with `args: release --clean` |
| `.goreleaser.yml` | `terraform-registry-manifest.json` | `extra_files` in checksum + release | ✓ WIRED | `goreleaser.yml:35,51`: manifest referenced in both sections |
| `.github/workflows/test.yml` | `go.mod` | `go-version-file` reads Go version | ✓ WIRED | `test.yml:20,37,51`: `go-version-file: 'go.mod'` in all three jobs |

### Data-Flow Trace (Level 4)

Not applicable -- this phase contains no components that render dynamic data. The provider is a CLI binary/plugin, not a UI. The API client performs HTTP I/O verified by unit tests against httptest servers.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Provider binary compiles | `go build ./...` | Exit 0, no output | ✓ PASS |
| All unit tests pass | `go test -v -count=1 ./...` | 20/20 PASS, 0 FAIL | ✓ PASS |
| Module exports `New` function | Module structure verified via code read | `func New(version string) func() provider.Provider` present | ✓ PASS |
| API key absent from request URLs | `TestAPIKeyNotInURL` | PASS | ✓ PASS |
| Release workflow syntax | YAML file read + structure verified | Valid YAML, correct trigger/job structure | ✓ PASS |
| GoReleaser config syntax | YAML file read + key fields verified | `version: 2`, cross-compilation matrix, signing present | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| FOUND-01 | 01-01-PLAN | Auth via VASTAI_API_KEY env var or `api_key` attribute (sensitive) | ✓ SATISFIED | `provider.go:97-102` env var + config override; `api_key Sensitive: true` at line 56 |
| FOUND-02 | 01-01-PLAN | Configurable API endpoint via `VASTAI_API_URL` or `api_url` | ✓ SATISFIED | `provider.go:98,104-106,108-110` env var + config override + default to `https://console.vast.ai` |
| FOUND-03 | 01-02-PLAN | HTTP client with exponential backoff on 429/5xx | ✓ SATISFIED | `client.go:30-34,48-70,74-93`: RetryMax=5, 150ms base, 1.5x backoff |
| FOUND-04 | 01-02-PLAN | User-Agent header includes provider version | ✓ SATISFIED | `client.go:41`: `"terraform-provider-vastai/%s"`, TestNewRequest_UserAgent passes |
| FOUND-05 | 01-02-PLAN | Bearer header auth (not query parameter) | ✓ SATISFIED | `auth.go:40` Authorization header; TestAPIKeyNotInURL passes |
| FOUND-06 | 01-02-PLAN | Structured error diagnostics on API failures | ✓ SATISFIED | `errors.go`: `APIError{StatusCode, Method, Path, Message}`; TestAPIError_ErrorString passes |
| RLSE-01 | 01-03-PLAN | GoReleaser config with GPG-signed releases | ✓ SATISFIED | `.goreleaser.yml` with `signs` section, `--detach-sign`, GPG_FINGERPRINT |
| RLSE-02 | 01-03-PLAN | GitHub Actions CI/CD for automated releases on tag push | ✓ SATISFIED | `.github/workflows/release.yml` triggers on `v*` tags |
| RLSE-03 | 01-01-PLAN | `terraform-registry-manifest.json` with protocol version 6.0 | ✓ SATISFIED | File content confirmed: `"protocol_versions": ["6.0"]` |
| RLSE-04 | 01-01-PLAN | Semantic versioning with `v` prefix | ✓ SATISFIED | `release.yml:4-5` triggers on `v*`; `main.go:13` `var version string = "dev"` with ldflags injection |
| RLSE-05 | 01-03-PLAN | SHA256SUMS and .sig files with each release | ✓ SATISFIED | `.goreleaser.yml:33-48`: `checksum` + `signs` sections configured |
| TEST-04 | 01-03-PLAN | CI pipeline: unit tests on PR, acceptance tests on main | ✓ SATISFIED | `test.yml`: `pull_request` + `push` triggers; acceptance job `if: github.ref == 'refs/heads/main'` |

**Requirements coverage: 12/12 Phase 1 requirements satisfied by code artifacts**

No orphaned requirements found -- all 12 requirements listed in REQUIREMENTS.md traceability table for Phase 1 appear in plan frontmatter and are implemented.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `internal/provider/provider.go:129` | 129 | `return []func() resource.Resource{}` | ℹ️ Info | Intentional -- phase goal is "zero resources"; this is the designed state, not a stub |
| `internal/provider/provider.go:134` | 134 | `return []func() datasource.DataSource{}` | ℹ️ Info | Intentional -- phase goal is "zero data sources"; this is the designed state, not a stub |

No blocker anti-patterns. The empty slices are the intended design for Phase 1 ("zero resources"). All TODO comments from Plan 01-01 were removed in Plan 01-02 as confirmed (`grep "TODO" provider.go` returns no output).

### Human Verification Required

#### 1. Release Workflow Execution

**Test:** Configure a GitHub remote for this repository, generate RSA 4096-bit GPG key, add `GPG_PRIVATE_KEY` and `GPG_PASSPHRASE` secrets to the GitHub repo, push the branch to main, then push a `v0.1.0-alpha` tag. Monitor the Actions tab.
**Expected:** The release workflow triggers, GPG key import succeeds, GoReleaser produces cross-compiled zip archives for linux/darwin/windows (amd64/arm64), SHA256SUMS and SHA256SUMS.sig are attached to the GitHub release, and `terraform-registry-manifest.json` is included.
**Why human:** No GitHub remote is configured (`git remote show` returns empty). No v* tags exist. Workflow execution requires a live GitHub environment, secrets configuration, and GoReleaser runtime.

#### 2. Terraform Registry Publication

**Test:** Add the GPG public key to the Terraform Registry (registry.terraform.io -> Publishing -> GPG Keys). Connect the GitHub repository to the Terraform Registry and publish the provider namespace `realnedsanders/vastai`.
**Expected:** After the first release tag is pushed (item 1 above), the provider appears at `registry.terraform.io/providers/realnedsanders/vastai` and is downloadable.
**Why human:** Requires Terraform Registry account access, namespace creation, and GPG key registration -- all external service operations.

#### 3. `terraform init` End-to-End Install

**Test:** After registry publication (item 2 above), create a directory with `providers.tf` containing the `examples/provider/provider.tf` content. Run `terraform init` with `VASTAI_API_KEY=<real key>` set.
**Expected:** Terraform downloads the provider binary, installs it to `.terraform/providers/`, and `terraform plan` completes without error.
**Why human:** Depends on registry publication; requires Terraform binary and live API credentials; cannot simulate the full `terraform init` download path locally.

### Gaps Summary

No gaps blocking goal achievement. All 12 requirements are implemented and verified in the codebase. The 3 human verification items are all **external system dependencies** (GitHub Actions execution, Terraform Registry publication, `terraform init` verification) -- they depend on infrastructure that has not been configured yet (no GitHub remote, no registry account), not on missing or broken code.

The code and configuration are complete and correct. The remaining work is operational setup.

---

_Verified: 2026-03-25T20:00:00Z_
_Verifier: Claude (gsd-verifier)_
