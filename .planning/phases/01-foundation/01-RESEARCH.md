# Phase 1: Foundation - Research

**Researched:** 2026-03-25
**Domain:** Terraform Provider Foundation (Go module, Plugin Framework wiring, REST API client, CI/CD, release pipeline)
**Confidence:** HIGH

## Summary

Phase 1 delivers the complete infrastructure that every subsequent phase builds on: a compilable Go binary that registers as a Terraform provider with zero resources, a REST API client for Vast.ai with authentication, retry, and structured error handling, CI/CD pipelines for testing and releasing, and a GPG-signed release pipeline that publishes to the Terraform Registry. The end-state is a `v0.1.0-alpha` tag that installs successfully via `terraform init`.

The Terraform provider scaffolding is fully templated by HashiCorp -- the scaffolding repository provides the exact `main.go`, `.goreleaser.yml`, `release.yml`, `test.yml`, and `terraform-registry-manifest.json` files needed. The primary engineering work is the Vast.ai API client: wrapping `go-retryablehttp` with Bearer token authentication, custom retry policy for 429/5xx, configurable backoff (150ms base, 1.5x multiplier matching the Python SDK), User-Agent header injection, and structured error mapping to Terraform diagnostics. The provider shell is thin -- it reads `VASTAI_API_KEY` from config or environment, creates the API client in `Configure()`, and passes it to resources via `ResourceData`/`DataSourceData`.

The highest-risk items are: (1) GPG signing in CI -- the registry rejects ECC keys and GoReleaser fails silently without `--batch` flag, so the pipeline must be validated with an actual tag push before any provider code is written; (2) the API client's authentication must use `Authorization: Bearer` header exclusively (never query parameters) to prevent credential leaks in Terraform logs; and (3) the retry configuration must be correct from day one because every resource depends on the client's HTTP transport layer.

**Primary recommendation:** Start with the release pipeline (GoReleaser + GitHub Actions + GPG), validate it works end-to-end with an empty provider binary, then build the API client, then wire the provider shell. This order catches the hardest-to-debug failures first.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** GitHub repo is `github.com/realnedsanders/terraform-provider-vastai`
- **D-02:** Go module path is `github.com/realnedsanders/terraform-provider-vastai`
- **D-03:** Terraform Registry namespace will be `realnedsanders/vastai`
- **D-04:** Service pattern -- `client.Instances.Create()`, `client.Offers.List()` -- sub-services mirror the API surface, scales to 25+ resource types
- **D-05:** Standalone-capable design in `internal/client/` -- clean interfaces, no Terraform dependencies, could be extracted to a public Go SDK later
- **D-06:** Built-in waiter methods (e.g., `WaitForInstanceReady()`) with configurable timeout and polling interval -- AWS provider pattern
- **D-07:** Rate limiting matches Python SDK: 150ms base delay, 1.5x exponential backoff, max 5 retries
- **D-08:** Full tflog integration -- request method/URL at DEBUG, response status at DEBUG, response body at TRACE. Errors and retries always logged.
- **D-09:** Authentication via Bearer header (never query parameter) -- prevents credential leaks in logs
- **D-10:** Service-per-directory pattern: `internal/services/instance/`, `internal/services/template/` -- each directory contains resource.go, data_source.go, models.go, and tests
- **D-11:** API client in `internal/client/` -- private to module, cleanly separated from provider/resource code
- **D-12:** Tag-based releases -- push git tag (v0.1.0) triggers GitHub Actions -> GoReleaser -> signed binaries published to registry
- **D-13:** GPG private key + passphrase stored as GitHub Actions secrets
- **D-14:** CI strategy: unit tests + lint on PRs (fast, free); acceptance tests on main branch only (uses real API credits)
- **D-15:** GoReleaser generates SHA256SUMS, .sig files, and terraform-registry-manifest.json automatically

### Claude's Discretion
- Go version selection (latest stable that satisfies Plugin Framework requirements)
- Specific GoReleaser configuration details
- golangci-lint configuration
- Exact directory structure for templates/, examples/, docs/ scaffolding
- Makefile targets and developer workflow

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| FOUND-01 | Provider authenticates via `VASTAI_API_KEY` env var or `api_key` provider attribute (marked sensitive) | Provider Configure pattern with env var fallback documented in Architecture Patterns; `Sensitive: true` schema flag prevents log leaks |
| FOUND-02 | Provider supports configurable API endpoint URL via `VASTAI_API_URL` env var or `api_url` provider attribute | Same Configure pattern; Python SDK uses `VAST_URL` env var defaulting to `https://console.vast.ai`; our provider uses `VASTAI_API_URL` |
| FOUND-03 | Go HTTP client with exponential backoff retry on 429/5xx, configurable max retries | go-retryablehttp v0.7.8 provides RetryMax, RetryWaitMin, RetryWaitMax, custom CheckRetry and Backoff functions; Python SDK: 150ms base, 1.5x backoff, default 3 retries |
| FOUND-04 | User-Agent header includes provider version (`terraform-provider-vastai/vX.Y.Z`) | go-retryablehttp supports RequestLogHook for header injection; version injected via ldflags at build time |
| FOUND-05 | API key authentication uses Bearer header (not query parameter) to prevent credential leaks in logs | Python SDK supports `Authorization: Bearer` at line 628-629 of vast.py; locked decision D-09 |
| FOUND-06 | Structured error diagnostics with summary and detail on all API failures | Plugin Framework Diagnostics API: AddError(summary, detail) with contextual information |
| RLSE-01 | GoReleaser configuration with GPG-signed releases | Scaffolding repo `.goreleaser.yml` with `signs.artifacts: checksum` and `--batch` flag; version 2 format |
| RLSE-02 | GitHub Actions CI/CD pipeline for automated releases on tag push | Scaffolding repo `release.yml` with `goreleaser/goreleaser-action@v7`, GPG import, tag trigger |
| RLSE-03 | `terraform-registry-manifest.json` declaring protocol version 6.0 | `{"version": 1, "metadata": {"protocol_versions": ["6.0"]}}` -- included in release via GoReleaser extra_files |
| RLSE-04 | Semantic versioning with `v` prefix | GoReleaser extracts version from git tag; tag format `v0.1.0-alpha.1` |
| RLSE-05 | SHA256SUMS and .sig files generated with each release | GoReleaser checksum section generates SHA256SUMS; signs section creates detached .sig |
| TEST-04 | CI pipeline running tests on PR (unit tests always, acceptance tests on main) | Scaffolding repo `test.yml` pattern; `TF_ACC=1` gating; separate jobs for build/lint and test |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.25.x (1.25.7 installed locally, 1.25.8 latest) | Language runtime | Required by terraform-plugin-framework v1.19.0; local machine has 1.25.7 which satisfies the requirement |
| terraform-plugin-framework | v1.19.0 | Provider SDK (protocol v6) | Only correct choice for new providers; GA, type-safe schemas, compile-time checking |
| terraform-plugin-go | v0.31.0 | Low-level plugin protocol | Transitive dependency of framework; provides providerserver.Serve() |
| terraform-plugin-log (tflog) | v0.10.0 | Structured logging | Required for all Terraform provider logging; never use fmt/log stdlib |
| hashicorp/go-retryablehttp | v0.7.8 | HTTP client with retry | Foundation for API client; exponential backoff, 429/503 handling, Retry-After parsing |
| GoReleaser | v2.14.x | Cross-compilation + release | Standard for Terraform providers; generates exact registry artifacts |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| terraform-plugin-testing | v1.15.0 | Acceptance test framework | Phase 1 needs it as dependency in go.mod even though no acceptance tests yet; will use for provider config tests |
| terraform-plugin-framework-validators | v0.19.0 | Pre-built validators | Phase 1 needs it as dependency; used later for attribute validation |
| golangci-lint | v2.11.x (v2.11.3 installed) | Linting meta-tool | CI and local development; v2 config format |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| go-retryablehttp | net/http + custom retry | Lose built-in backoff, Retry-After parsing, 429 handling; no benefit |
| go-retryablehttp | go-resty | Not HashiCorp ecosystem; adds unnecessary dependency outside the standard stack |
| GoReleaser v2 | Manual cross-compilation | Lose checksums, signing, manifest injection; no reason to hand-roll |

**Installation:**
```bash
# Initialize Go module
go mod init github.com/realnedsanders/terraform-provider-vastai

# Add dependencies
go get github.com/hashicorp/terraform-plugin-framework@v1.19.0
go get github.com/hashicorp/terraform-plugin-go@v0.31.0
go get github.com/hashicorp/terraform-plugin-log@v0.10.0
go get github.com/hashicorp/terraform-plugin-testing@v1.15.0
go get github.com/hashicorp/terraform-plugin-framework-validators@v0.19.0
go get github.com/hashicorp/go-retryablehttp@v0.7.8
go mod tidy
```

**Version verification:** Go 1.25.7 is installed locally (satisfies 1.25+ requirement). golangci-lint v2.11.3 is installed locally. All package versions are verified against the HashiCorp scaffolding repo go.mod (which pins framework v1.19.0, testing v1.15.0) and GitHub releases pages as documented in `.planning/research/STACK.md`.

## Architecture Patterns

### Recommended Project Structure (Phase 1 scope)
```
terraform-provider-vastai/
  main.go                              # Entry point: providerserver.Serve()
  go.mod / go.sum                      # Go module files
  .goreleaser.yml                      # GoReleaser v2 config (cross-platform + signing)
  .golangci.yml                        # golangci-lint v2 config
  GNUmakefile                          # build, test, lint, generate, install targets
  terraform-registry-manifest.json     # {"version":1,"metadata":{"protocol_versions":["6.0"]}}
  .github/
    workflows/
      test.yml                         # CI: build, lint, unit tests (PRs); acceptance tests (main)
      release.yml                      # CD: GoReleaser on v* tag push
  internal/
    provider/
      provider.go                      # VastaiProvider struct, Schema, Configure, Resources, DataSources
      provider_test.go                 # Provider configuration unit tests
    client/
      client.go                        # VastAIClient struct, NewClient(), base HTTP methods (Get, Post, Put, Delete)
      client_test.go                   # Client unit tests (httptest server)
      errors.go                        # APIError type, error mapping from HTTP status codes
      auth.go                          # Bearer token authentication, header injection
  examples/
    provider/
      provider.tf                      # Provider configuration example
  templates/                           # tfplugindocs templates (scaffolding only in Phase 1)
    index.md.tmpl                      # Provider overview template
```

### Pattern 1: Provider Configure -> Client Injection
**What:** Create API client once in provider Configure(), inject into all resources/data sources via ResourceData/DataSourceData.
**When to use:** Always -- this is the standard pattern for all Terraform providers.
**Example:**
```go
// Source: HashiCorp Plugin Framework docs + CONTEXT.md decisions
// internal/provider/provider.go

type VastaiProvider struct {
    version string
}

type VastaiProviderModel struct {
    APIKey types.String `tfsdk:"api_key"`
    APIURL types.String `tfsdk:"api_url"`
}

func (p *VastaiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var config VastaiProviderModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Handle unknown values during plan
    if config.APIKey.IsUnknown() {
        resp.Diagnostics.AddWarning(
            "Unknown Vast.ai API Key",
            "The provider cannot be configured during plan when the API key is unknown. "+
                "Resources and data sources will return errors until the key is known.",
        )
        return
    }
    if config.APIURL.IsUnknown() {
        resp.Diagnostics.AddWarning(
            "Unknown Vast.ai API URL",
            "The provider cannot be configured during plan when the API URL is unknown.",
        )
        return
    }

    // API key from config or environment variable
    apiKey := os.Getenv("VASTAI_API_KEY")
    if !config.APIKey.IsNull() {
        apiKey = config.APIKey.ValueString()
    }
    if apiKey == "" {
        resp.Diagnostics.AddError(
            "Missing API Key",
            "Set the VASTAI_API_KEY environment variable or configure api_key in the provider block.",
        )
        return
    }

    // API URL from config or environment variable
    apiURL := os.Getenv("VASTAI_API_URL")
    if apiURL == "" {
        apiURL = "https://console.vast.ai"
    }
    if !config.APIURL.IsNull() {
        apiURL = config.APIURL.ValueString()
    }

    c := client.NewVastAIClient(apiKey, apiURL, p.version)
    resp.DataSourceData = c
    resp.ResourceData = c
}
```

### Pattern 2: API Client with Service Objects (Phase 1 skeleton)
**What:** Root client struct with base HTTP methods; service sub-objects added in later phases.
**When to use:** Always for APIs with 10+ resource types (D-04).
**Example:**
```go
// Source: CONTEXT.md D-04, D-05, D-07, D-08, D-09 + go-retryablehttp docs
// internal/client/client.go

type VastAIClient struct {
    httpClient  *retryablehttp.Client
    baseURL     string
    apiKey      string
    userAgent   string

    // Service objects added in later phases:
    // Instances *InstanceService
    // Templates *TemplateService
}

func NewVastAIClient(apiKey, baseURL, version string) *VastAIClient {
    retryClient := retryablehttp.NewClient()
    retryClient.RetryMax = 5                              // D-07: max 5 retries
    retryClient.RetryWaitMin = 150 * time.Millisecond     // D-07: 150ms base delay
    retryClient.RetryWaitMax = 30 * time.Second
    retryClient.CheckRetry = vastaiRetryPolicy             // Custom: retry on 429 + 5xx
    retryClient.Logger = nil                               // Silence default logger; use tflog instead

    c := &VastAIClient{
        httpClient: retryClient,
        baseURL:    strings.TrimRight(baseURL, "/"),
        apiKey:     apiKey,
        userAgent:  fmt.Sprintf("terraform-provider-vastai/%s", version),
    }
    return c
}
```

### Pattern 3: Bearer Authentication (Never Query Parameters)
**What:** All API requests use `Authorization: Bearer <key>` header. API key never appears in URL.
**When to use:** Always -- locked decision D-09.
**Example:**
```go
// Source: CONTEXT.md D-09, vast.py line 628-629
// internal/client/auth.go

func (c *VastAIClient) newRequest(ctx context.Context, method, path string, body interface{}) (*retryablehttp.Request, error) {
    url := c.baseURL + "/api/v0" + path

    var bodyReader io.Reader
    if body != nil {
        jsonBytes, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("marshaling request body: %w", err)
        }
        bodyReader = bytes.NewReader(jsonBytes)
    }

    req, err := retryablehttp.NewRequestWithContext(ctx, method, url, bodyReader)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    req.Header.Set("User-Agent", c.userAgent)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    return req, nil
}
```

### Pattern 4: Custom Retry Policy (429 + 5xx)
**What:** Retry on 429 (rate limited) and 5xx (server errors), skip 4xx client errors.
**When to use:** Always -- D-07 mandates matching Python SDK behavior.
**Example:**
```go
// Source: vast.py line 357-361 (retry on 429), go-retryablehttp DefaultRetryPolicy
// internal/client/client.go

func vastaiRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
    // Do not retry on context cancellation
    if ctx.Err() != nil {
        return false, ctx.Err()
    }
    // Retry on connection errors
    if err != nil {
        return true, nil
    }
    // Retry on 429 (rate limited) and 5xx (server errors, except 501)
    if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode != 501) {
        return true, nil
    }
    return false, nil
}
```

### Pattern 5: Structured Error Diagnostics
**What:** Map HTTP errors to Terraform diagnostics with summary + detail.
**When to use:** Every API call result that surfaces to Terraform.
**Example:**
```go
// Source: Plugin Framework Diagnostics docs, FOUND-06
// internal/client/errors.go

type APIError struct {
    StatusCode int
    Message    string
    Method     string
    Path       string
}

func (e *APIError) Error() string {
    return fmt.Sprintf("Vast.ai API error: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Message)
}

// Usage in resource code:
// resp.Diagnostics.AddError(
//     "Error Reading Instance",
//     fmt.Sprintf("Could not read instance %s: %s", id, err.Error()),
// )
```

### Pattern 6: Compile-Time Interface Verification
**What:** Static assertions that the provider implements required interfaces.
**When to use:** Always -- in provider.go.
**Example:**
```go
// Source: Plugin Framework docs
var _ provider.Provider = &VastaiProvider{}
```

### Pattern 7: main.go Entry Point
**What:** Minimal entry point that serves the provider via gRPC.
**When to use:** Always -- single canonical pattern.
**Example:**
```go
// Source: terraform-provider-scaffolding-framework main.go
package main

import (
    "context"
    "flag"
    "log"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/realnedsanders/terraform-provider-vastai/internal/provider"
)

var version string = "dev"

func main() {
    var debug bool
    flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
    flag.Parse()

    opts := providerserver.ServeOpts{
        Address: "registry.terraform.io/realnedsanders/vastai",
        Debug:   debug,
    }

    err := providerserver.Serve(context.Background(), provider.New(version), opts)
    if err != nil {
        log.Fatal(err.Error())
    }
}
```

### Anti-Patterns to Avoid
- **API key in query parameters:** The Python SDK uses `?api_key=` by default. Never replicate this. Terraform logs URLs at TRACE level, exposing credentials. Use Bearer header exclusively (D-09).
- **Standard library `log` or `fmt.Println`:** Terraform requires structured logging via `tflog`. Raw output breaks log filtering and level management.
- **Monolithic provider package:** Do not put client code, resource code, and provider code in the same package. The three-layer separation (`provider/`, `client/`, `services/`) is mandatory from day one.
- **Hardcoded API base URL:** Always make the base URL configurable via `VASTAI_API_URL` env var and `api_url` provider attribute. Required for testing against staging environments.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTTP retry with backoff | Custom retry loop with time.Sleep | `hashicorp/go-retryablehttp` v0.7.8 | Handles 429, Retry-After headers, connection errors, backoff with jitter; battle-tested across HashiCorp ecosystem |
| Cross-platform compilation | Shell scripts with `GOOS`/`GOARCH` | GoReleaser v2 | Generates 10+ binaries, checksums, signatures, manifest in one command; exact registry format |
| GPG signing in CI | Manual `gpg --sign` in workflow | `crazy-max/ghaction-import-gpg@v7` + GoReleaser signs section | Handles key import, passphrase, TTY-less signing in CI; `--batch` flag critical |
| Terraform structured logging | `log.Printf` or custom logger | `terraform-plugin-log/tflog` v0.10.0 | Integrates with Terraform's log levels, subsystems, and filtering; required by the ecosystem |
| Provider server gRPC setup | Custom gRPC server | `providerserver.Serve()` from terraform-plugin-go | Handles protocol negotiation, versioning, health checks; zero configuration needed |

**Key insight:** Phase 1 is almost entirely "wiring together existing tools correctly" rather than building novel functionality. The engineering value is in configuration accuracy (GoReleaser, GitHub Actions, GPG signing) and API client design quality (auth, retry, errors), not in custom logic.

## Common Pitfalls

### Pitfall 1: GPG Key Type Rejection
**What goes wrong:** The Terraform Registry rejects releases signed with ECC (ed25519) GPG keys, which are the modern default on most systems.
**Why it happens:** Registry only accepts RSA and DSA keys. Modern `gpg --full-generate-key` defaults to ECC.
**How to avoid:** Generate an RSA 4096-bit key specifically for release signing. Store as `GPG_PRIVATE_KEY` secret in GitHub Actions. Always include `--batch` flag in GoReleaser signing config.
**Warning signs:** GoReleaser signing step produces empty `.sig` file. Registry shows "signature verification failed." `terraform init` fails with "checksum mismatch."

### Pitfall 2: GoReleaser Passphrase Prompt in CI
**What goes wrong:** GoReleaser's GPG signing step hangs or fails silently in CI because no TTY is available for passphrase entry.
**Why it happens:** CI runners have no interactive terminal. Without `--batch` flag and `PASSPHRASE` environment variable, GPG waits for user input indefinitely.
**How to avoid:** Include `--batch` in the GoReleaser signs.args config (the scaffolding template already does this). Store the passphrase as `GPG_PASSPHRASE` GitHub Actions secret. Validate by running a test release.
**Warning signs:** Release workflow hangs at the signing step. No `.sig` file appears in release artifacts.

### Pitfall 3: API Key Leaked via Terraform Logging
**What goes wrong:** If the API key appears in a URL (query parameter) or error message, Terraform's TRACE-level logging exposes it in CI logs, Terraform Cloud runs, and developer terminals.
**Why it happens:** The Python SDK sends `?api_key=<key>` on every request. Developers copy this pattern. Error messages include the full request URL.
**How to avoid:** Use `Authorization: Bearer` header exclusively (D-09). Never include the API key in URLs. Sanitize URLs in error messages before creating Terraform diagnostics.
**Warning signs:** `TF_LOG=TRACE` output shows the API key. Error messages contain `?api_key=`.

### Pitfall 4: Unknown Values in Provider Configure
**What goes wrong:** When a practitioner interpolates a resource output into the provider config (e.g., `api_key = some_resource.key`), the value is "unknown" during plan. If Configure treats unknown as empty and errors, `terraform plan` breaks entirely.
**Why it happens:** Terraform evaluates provider configuration before all resource values are known. The Plugin Framework passes unknown values as `types.String` with `IsUnknown() == true`.
**How to avoid:** Check `IsUnknown()` before `IsNull()` in Configure. Return a warning (not error) when values are unknown, allowing plan to proceed. Only error on empty values when the value is known.
**Warning signs:** `terraform plan` fails with "Missing API Key" even though the key is configured via a data source or resource output.

### Pitfall 5: Missing terraform-registry-manifest.json
**What goes wrong:** The Terraform Registry cannot determine the protocol version and rejects the release, or `terraform init` fails to install the provider.
**Why it happens:** The manifest file must exist in the repo root AND be included in release artifacts via GoReleaser's `extra_files` config. Missing either one breaks publishing.
**How to avoid:** Create `terraform-registry-manifest.json` with `{"version": 1, "metadata": {"protocol_versions": ["6.0"]}}` in the repo root. Verify it appears in GoReleaser checksum and release extra_files sections.
**Warning signs:** `terraform init` shows "protocol version mismatch" or "provider does not support the requested protocol."

### Pitfall 6: go-retryablehttp Default Logger Noise
**What goes wrong:** go-retryablehttp logs retry attempts to stderr by default, which interleaves with Terraform's structured logging and produces confusing output.
**Why it happens:** The library has a built-in logger that is non-nil by default.
**How to avoid:** Set `retryClient.Logger = nil` when creating the client. Log retries explicitly via tflog at DEBUG level in the custom retry policy or RequestLogHook.
**Warning signs:** Duplicate log lines for retries. Non-structured log output mixed with Terraform's JSON logging.

## Code Examples

Verified patterns from official sources:

### GoReleaser Configuration (.goreleaser.yml)
```yaml
# Source: terraform-provider-scaffolding-framework .goreleaser.yml (verified 2026-03-25)
version: 2
before:
  hooks:
    - go mod tidy
builds:
- env:
    - CGO_ENABLED=0
  mod_timestamp: '{{ .CommitTimestamp }}'
  flags:
    - -trimpath
  ldflags:
    - '-s -w -X main.version={{.Version}} -X main.commit={{.Commit}}'
  goos:
    - freebsd
    - windows
    - linux
    - darwin
  goarch:
    - amd64
    - '386'
    - arm
    - arm64
  ignore:
    - goos: darwin
      goarch: '386'
    - goos: windows
      goarch: arm
  binary: '{{ .ProjectName }}_v{{ .Version }}'
archives:
- formats:
  - zip
  name_template: '{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}'
checksum:
  extra_files:
    - glob: 'terraform-registry-manifest.json'
      name_template: '{{ .ProjectName }}_{{ .Version }}_manifest.json'
  name_template: '{{ .ProjectName }}_{{ .Version }}_SHA256SUMS'
  algorithm: sha256
signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"
release:
  extra_files:
    - glob: 'terraform-registry-manifest.json'
      name_template: '{{ .ProjectName }}_{{ .Version }}_manifest.json'
changelog:
  disable: true
```

### GitHub Actions Release Workflow (.github/workflows/release.yml)
```yaml
# Source: terraform-provider-scaffolding-framework release.yml (verified 2026-03-25)
name: Release
on:
  push:
    tags:
      - 'v*'
permissions:
  contents: write
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v6
        with:
          go-version-file: 'go.mod'
          cache: true
      - name: Import GPG key
        uses: crazy-max/ghaction-import-gpg@v7
        id: import_gpg
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.GPG_PASSPHRASE }}
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v7
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
```

### GitHub Actions Test Workflow (.github/workflows/test.yml)
```yaml
# Source: terraform-provider-scaffolding-framework test.yml (adapted per D-14)
name: Tests
on:
  pull_request:
    branches: [main]
    paths-ignore:
      - 'README.md'
  push:
    branches: [main]
    paths-ignore:
      - 'README.md'
jobs:
  build:
    name: Build
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version-file: 'go.mod'
          cache: true
      - run: go mod download
      - run: go build -v ./...
      - name: Run linters
        uses: golangci/golangci-lint-action@v8
        with:
          version: v2.11
  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    timeout-minutes: 15
    needs: build
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version-file: 'go.mod'
          cache: true
      - run: go mod download
      - run: go test -v -count=1 -parallel=4 -timeout 120s ./...
  # Acceptance tests only on main branch (D-14)
  acceptance:
    name: Acceptance Tests
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    timeout-minutes: 120
    needs: build
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version-file: 'go.mod'
          cache: true
      - run: go mod download
      - run: go test -v -count=1 -parallel=4 -timeout 120m ./...
        env:
          TF_ACC: "1"
          VASTAI_API_KEY: ${{ secrets.VASTAI_API_KEY }}
```

### terraform-registry-manifest.json
```json
{
    "version": 1,
    "metadata": {
        "protocol_versions": ["6.0"]
    }
}
```

### GNUmakefile
```makefile
# Source: terraform-provider-scaffolding-framework GNUmakefile (adapted)
default: fmt lint install generate

.PHONY: build install lint fmt test testacc generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w .

test:
	go test -v -count=1 -parallel=4 -timeout 120s ./...

testacc:
	TF_ACC=1 go test -v -count=1 -parallel=4 -timeout 120m ./...

generate:
	go generate ./...
```

### Provider Schema (provider.go skeleton)
```go
// Source: Plugin Framework docs
func (p *VastaiProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Interact with Vast.ai GPU compute infrastructure.",
        Attributes: map[string]schema.Attribute{
            "api_key": schema.StringAttribute{
                Description: "Vast.ai API key. Can also be set via VASTAI_API_KEY environment variable.",
                Optional:    true,
                Sensitive:   true,
            },
            "api_url": schema.StringAttribute{
                Description: "Vast.ai API base URL. Can also be set via VASTAI_API_URL environment variable. Defaults to https://console.vast.ai.",
                Optional:    true,
            },
        },
    }
}
```

### go-retryablehttp Custom Backoff (matching D-07)
```go
// Source: go-retryablehttp docs + D-07 (150ms base, 1.5x multiplier)
func vastaiBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
    // Respect Retry-After header if present
    if resp != nil && resp.StatusCode == 429 {
        if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
            if seconds, err := strconv.ParseFloat(retryAfter, 64); err == nil {
                return time.Duration(seconds * float64(time.Second))
            }
        }
    }

    // D-07: 150ms base, 1.5x exponential backoff
    wait := float64(min) * math.Pow(1.5, float64(attemptNum))
    if time.Duration(wait) > max {
        return max
    }
    return time.Duration(wait)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Terraform Plugin SDKv2 | Plugin Framework v1.19.0 | 2022+ (GA) | New providers MUST use Plugin Framework; SDKv2 is legacy/maintenance-only |
| golangci-lint v1 config | golangci-lint v2 config | 2025 | v2 uses `linters.default` syntax; `enable-all`/`disable-all` deprecated |
| GoReleaser v1 config | GoReleaser v2 config | 2024 | Requires `version: 2` in config; different YAML structure |
| `goreleaser/goreleaser-action@v5` | `goreleaser/goreleaser-action@v7` | 2025 | Current scaffolding uses v7 |
| `crazy-max/ghaction-import-gpg@v6` | `crazy-max/ghaction-import-gpg@v7` | 2025 | Current scaffolding uses v7 |
| Query parameter auth | Bearer header auth | Always best practice | Prevents credential leaks in logs |

**Deprecated/outdated:**
- Plugin SDKv2: Legacy, do not use for new providers
- GoReleaser v1 config format: Will not work with current GoReleaser binary
- golangci-lint v1 config format: Deprecated `enable-all` syntax

## Open Questions

1. **Terraform binary availability for local testing**
   - What we know: Terraform is not installed on the local machine. The provider compiles and publishes via CI, but local `terraform init` / `terraform plan` testing requires the Terraform binary.
   - What's unclear: Whether the planner should include a Terraform installation task or assume it is available.
   - Recommendation: Include a dev environment setup task that installs Terraform via `go install github.com/hashicorp/terraform@latest` or documents the requirement. Not blocking for Phase 1 since CI handles all Terraform operations.

2. **GitHub repository creation and secrets setup**
   - What we know: The repo must exist at `github.com/realnedsanders/terraform-provider-vastai` (D-01). GPG secrets must be configured (D-13).
   - What's unclear: Whether the repo already exists or needs creation. Whether GPG key generation and secret configuration is in scope for the plan or a prerequisite.
   - Recommendation: Plan should include tasks for repo creation (if needed) and document GPG key + secrets setup as a prerequisite that the developer must complete manually before the release pipeline can work. GPG key generation and GitHub secrets are manual operations.

3. **Terraform Registry namespace registration**
   - What we know: Namespace is `realnedsanders/vastai` (D-03). The registry requires the repo to follow naming conventions.
   - What's unclear: Whether the developer has registered the namespace on the Terraform Registry.
   - Recommendation: Document as a prerequisite. Registry namespace registration is a manual one-time operation via the Terraform Registry web UI.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Build, all code | Yes | 1.25.7 | -- |
| golangci-lint | Linting (CI + local) | Yes | v2.11.3 | -- |
| GPG | Release signing (local key generation) | Yes | 2.4.9 | -- |
| GitHub CLI (gh) | Repo management | Yes | v2.88.1 | -- |
| make (GNU Make) | Developer workflow | Yes | 4.4.1 | -- |
| Terraform | Local provider testing | No | -- | CI runs acceptance tests; local testing can use `go test` with TF_ACC=1 if Terraform is installed later |
| GoReleaser | Release builds | No (local) | -- | Only needed in CI (GitHub Actions installs it via goreleaser-action); not needed locally |

**Missing dependencies with no fallback:**
- None -- all missing tools are CI-only (GoReleaser) or not strictly required for Phase 1 implementation (Terraform binary can be installed if needed for local testing)

**Missing dependencies with fallback:**
- Terraform binary: Not installed locally, but CI handles all acceptance testing. For local validation, `go build ./...` and unit tests (`go test ./...`) cover Phase 1 scope.
- GoReleaser: Not installed locally, but GitHub Actions handles releases. Local builds use `go build`.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + terraform-plugin-testing v1.15.0 |
| Config file | None yet -- Wave 0 creates test infrastructure |
| Quick run command | `go test -v -count=1 -parallel=4 -timeout 120s ./...` |
| Full suite command | `TF_ACC=1 go test -v -count=1 -parallel=4 -timeout 120m ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FOUND-01 | Provider reads API key from env var and config attribute | unit | `go test -v -run TestProvider_Configure ./internal/provider/ -count=1` | Wave 0 |
| FOUND-02 | Provider reads API URL from env var and config attribute | unit | `go test -v -run TestProvider_ConfigureURL ./internal/provider/ -count=1` | Wave 0 |
| FOUND-03 | HTTP client retries on 429/5xx with exponential backoff | unit | `go test -v -run TestClient_Retry ./internal/client/ -count=1` | Wave 0 |
| FOUND-04 | User-Agent header includes provider version | unit | `go test -v -run TestClient_UserAgent ./internal/client/ -count=1` | Wave 0 |
| FOUND-05 | API key sent via Bearer header, never in URL | unit | `go test -v -run TestClient_BearerAuth ./internal/client/ -count=1` | Wave 0 |
| FOUND-06 | Structured error diagnostics on API failure | unit | `go test -v -run TestClient_ErrorDiagnostics ./internal/client/ -count=1` | Wave 0 |
| RLSE-01 | GoReleaser config with GPG signing | config-only | Manual verification via test tag push | N/A |
| RLSE-02 | GitHub Actions release workflow on tag push | config-only | Manual verification via test tag push | N/A |
| RLSE-03 | Registry manifest declares protocol 6.0 | unit | `go test -v -run TestRegistryManifest ./... -count=1` | Wave 0 (optional) |
| RLSE-04 | Semantic versioning with v prefix | config-only | GoReleaser enforces via tag pattern | N/A |
| RLSE-05 | SHA256SUMS and .sig files in release | config-only | Manual verification via test release | N/A |
| TEST-04 | CI runs unit tests on PR, acceptance on main | config-only | Verify via GitHub Actions run after first push | N/A |

### Sampling Rate
- **Per task commit:** `go test -v -count=1 -parallel=4 -timeout 120s ./...`
- **Per wave merge:** `go test -v -count=1 -parallel=4 -timeout 120s ./...` (full unit suite)
- **Phase gate:** Full unit suite green + successful `go build ./...` + test tag push validates release pipeline

### Wave 0 Gaps
- [ ] `internal/provider/provider_test.go` -- covers FOUND-01, FOUND-02
- [ ] `internal/client/client_test.go` -- covers FOUND-03, FOUND-04, FOUND-05, FOUND-06
- [ ] Go module initialization (`go mod init`, `go mod tidy`)
- [ ] All source files in `internal/provider/` and `internal/client/` (greenfield -- nothing exists yet)

## Sources

### Primary (HIGH confidence)
- [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) -- go.mod (v1.19.0), .goreleaser.yml (v2 format), release.yml, test.yml, main.go, terraform-registry-manifest.json. All files verified via raw GitHub content 2026-03-25.
- [HashiCorp Developer - Plugin Framework Providers](https://developer.hashicorp.com/terraform/plugin/framework/providers) -- Provider interface, Configure pattern, Unknown value handling
- [HashiCorp Developer - Plugin Framework Diagnostics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics) -- AddError, AddWarning, AddAttributeError patterns
- [go-retryablehttp on pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/go-retryablehttp) -- Client struct, NewClient defaults, CheckRetry, Backoff signatures, retry policies
- [HashiCorp Developer - Publish Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing) -- GPG key requirements (RSA/DSA only), manifest format, registry publishing flow
- `.planning/research/STACK.md` -- version matrix, compatibility, confidence assessment
- `.planning/research/ARCHITECTURE.md` -- three-layer architecture, data flow, interface signatures
- `.planning/research/PITFALLS.md` -- credential leaks, GPG key types, CI signing, rate limiting

### Secondary (MEDIUM confidence)
- [vast-sdk/vastai/vast.py](vast-sdk/vastai/vast.py) -- `http_request()` retry logic (line 333-362), `apiurl()` query param auth (line 575-619), `apiheaders()` Bearer auth (line 621-630), default retry=3 (line 8723), base URL `https://console.vast.ai` (line 72)

### Tertiary (LOW confidence)
- Vast.ai API rate limits -- undocumented; Python SDK defaults to 3 retries with 150ms base. Decision D-07 sets 5 retries. Actual limits unknown; must be discovered through testing.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- All versions verified against scaffolding repo, pkg.go.dev, and GitHub releases
- Architecture: HIGH -- Patterns from scaffolding template, AWS/Cloudflare provider references, official Plugin Framework docs
- Pitfalls: HIGH -- Verified against HashiCorp docs, Python SDK source, and community experience
- Release pipeline: HIGH -- Exact files from scaffolding repo verified and documented

**Research date:** 2026-03-25
**Valid until:** 2026-04-25 (stable ecosystem, 30-day validity)
