# Stack Research

**Domain:** Terraform/OpenTofu Provider Development (Go)
**Researched:** 2026-03-25
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.25.x (currently 1.25.8) | Language runtime | Required by Terraform plugin ecosystem. 1.25 is the minimum required by terraform-plugin-framework v1.19.0 and terraform-plugin-testing v1.15.0 per HashiCorp's "two latest Go releases" policy. Go 1.26 is also supported but 1.25 is the safer baseline for broadest tooling compatibility. |
| Terraform Plugin Framework | v1.19.0 | Provider SDK | The only correct choice for new providers in 2026. GA status, semantic versioning, protocol v6 support. Provides type-safe schema definitions, compile-time error checking, and native support for modern Terraform features (functions, ephemeral resources, actions). |
| Terraform Plugin Testing | v1.15.0 | Acceptance test framework | HashiCorp's official testing framework. Runs real Terraform operations (plan, apply, refresh, destroy) against actual infrastructure. No alternative exists for proper provider acceptance testing. |
| Terraform Plugin Docs | v0.24.0 | Documentation generation (tfplugindocs CLI) | Generates Terraform Registry-compatible documentation from provider schema and example files. Required for registry publishing. |
| GoReleaser | v2.14.x | Build and release automation | Standard tool for cross-platform Go builds with GPG signing. Terraform's scaffolding repo uses it. The v2 config format is current (requires `version: 2` in config). |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| terraform-plugin-go | v0.31.0 | Low-level plugin protocol bindings | Pulled in as a dependency of the framework. You rarely import this directly, but it underpins the framework's RPC communication with Terraform. |
| terraform-plugin-log (tflog) | v0.10.0 | Structured logging | Use for ALL provider logging. Wraps go-hclog with Terraform-specific conventions. Supports subsystems (e.g., separate logger for API client). Never use fmt.Println or log.Printf. |
| terraform-plugin-framework-validators | v0.19.0 | Pre-built attribute validators | Use for common validation patterns (string length, regex, numeric ranges, IP addresses, etc.) instead of writing custom validators. Covers string, int64, float64, bool, list, map, set, and object types. |
| hashicorp/go-retryablehttp | v0.7.8 | HTTP client with automatic retries | Use as the foundation for the Vast.ai API client. Provides exponential backoff, configurable retry policies, and rate limit awareness (429/503 + Retry-After header support). Wraps net/http so the interface is familiar. |

### Development Tools

| Tool | Version | Purpose | Notes |
|------|---------|---------|-------|
| golangci-lint | v2.11.x | Go linting meta-tool | Use v2 config format (`.golangci.yml` with `version: 2` semantics). Enable at minimum: staticcheck, govet, errcheck, gosec, revive, gosimple, unconvert, unparam. Run `golangci-lint migrate` if starting from v1 config. |
| tfplugindocs | v0.24.0 | Registry documentation generator | Installed via `go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.24.0`. Reads schema from compiled provider binary and merges with templates in `templates/` directory. |
| goreleaser | v2.14.x | Cross-compilation and release | Install via GitHub Actions (`goreleaser/goreleaser-action@v6`). Config in `.goreleaser.yml` at repo root. |
| GitHub Actions | N/A | CI/CD | Use `hashicorp/ghaction-terraform-provider-release` v4 reusable workflows, or build custom workflows based on the scaffolding repo patterns. |

### Registry and Release Infrastructure

| Component | Purpose | Notes |
|-----------|---------|-------|
| terraform-registry-manifest.json | Registry protocol declaration | Must exist in repo root. Declares protocol version (6 for Plugin Framework). Included in release checksums via GoReleaser `extra_files`. |
| GPG key (RSA) | Release signing | Registry requires GPG-signed SHA256SUMS. Must be RSA or DSA (not ECC). Store private key as GitHub Actions secret `GPG_PRIVATE_KEY` with fingerprint in `GPG_FINGERPRINT`. |
| GitHub Releases | Distribution mechanism | GoReleaser creates GitHub Releases with cross-compiled binaries, checksums, and signatures. Registry polls GitHub Releases for new versions. |

## Go Module Dependencies (go.mod)

```go
module github.com/[org]/terraform-provider-vastai

go 1.25

require (
    github.com/hashicorp/terraform-plugin-framework v1.19.0
    github.com/hashicorp/terraform-plugin-framework-validators v0.19.0
    github.com/hashicorp/terraform-plugin-go v0.31.0
    github.com/hashicorp/terraform-plugin-log v0.10.0
    github.com/hashicorp/terraform-plugin-testing v1.15.0
    github.com/hashicorp/go-retryablehttp v0.7.8
)
```

## Project Directory Structure

Based on the HashiCorp scaffolding repo and patterns from production providers:

```
terraform-provider-vastai/
  main.go                           # Provider server entrypoint
  go.mod / go.sum                   # Go module files
  .goreleaser.yml                   # GoReleaser v2 config
  .golangci.yml                     # Linter config
  .github/
    workflows/
      test.yml                      # CI: lint, unit tests, acceptance tests
      release.yml                   # CD: GoReleaser on version tags
  terraform-registry-manifest.json  # Registry protocol manifest
  internal/
    provider/
      provider.go                   # Provider definition, schema, Configure()
      provider_test.go              # Provider unit tests
    client/
      client.go                     # Vast.ai API client (wraps go-retryablehttp)
      client_test.go                # Client unit tests
      instances.go                  # Instance API operations
      templates.go                  # Template API operations
      ssh_keys.go                   # SSH Key API operations
      volumes.go                    # Volume API operations
      ...                           # One file per API resource domain
    resources/
      instance_resource.go          # vastai_instance resource
      instance_resource_test.go     # Acceptance tests
      template_resource.go          # vastai_template resource
      ssh_key_resource.go           # vastai_ssh_key resource
      volume_resource.go            # vastai_volume resource
      ...                           # One file per managed resource
    datasources/
      offer_data_source.go          # vastai_offers data source
      machine_data_source.go        # vastai_machines data source
      ...                           # One file per data source
  examples/
    provider/
      main.tf                       # Provider configuration example
    resources/
      vastai_instance/
        resource.tf                 # Instance resource example
      ...
    data-sources/
      vastai_offers/
        data-source.tf              # Offers data source example
      ...
  templates/                        # tfplugindocs templates (optional overrides)
    resources/
      vastai_instance.md.tmpl
    data-sources/
      vastai_offers.md.tmpl
  docs/                             # Generated by tfplugindocs (do not hand-edit)
    resources/
      instance.md
    data-sources/
      offers.md
  tools/
    tools.go                        # Tool dependencies (tfplugindocs)
```

**Why this structure:** Separating `internal/client/`, `internal/resources/`, and `internal/datasources/` keeps the API client cleanly decoupled from Terraform schema logic. The client package is independently testable with unit tests (mocking HTTP responses), while resources/datasources use acceptance tests that exercise the full stack. This separation also makes it easy to evolve the API client independently when Vast.ai changes their API.

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Terraform Plugin Framework v1.19.0 | Terraform Plugin SDKv2 | Never for a new provider. SDKv2 is legacy/maintenance-only. Only relevant if migrating an existing SDKv2 provider incrementally via terraform-plugin-mux. |
| Terraform Plugin Framework v1.19.0 | terraform-plugin-go (low-level) | Only if you need protocol-level control that the framework cannot provide. Extremely unlikely for this provider. |
| Hand-written Go API client | Code generation (terraform-plugin-codegen-framework) | Only if Vast.ai had an OpenAPI spec. They don't. Code generation is in tech preview and designed for OpenAPI-to-Framework workflows. Our Python SDK reference makes hand-written the correct approach. |
| hashicorp/go-retryablehttp | net/http (stdlib) | Never for a cloud API client. You need retry logic, backoff, and rate limit handling. go-retryablehttp wraps net/http so you lose nothing but gain resilience. |
| hashicorp/go-retryablehttp | resty, go-resty, or similar | go-retryablehttp is the HashiCorp ecosystem standard. Using it keeps dependencies minimal and consistent with what Terraform itself uses internally. |
| golangci-lint v2 | Individual linters (go vet, staticcheck separately) | Never. golangci-lint aggregates 50+ linters into one tool with consistent config. No reason to manage them separately. |
| GoReleaser v2 | Manual cross-compilation + scripts | Never. GoReleaser is the standard for Terraform providers and handles the exact artifacts the registry expects (checksums, signing, manifest). |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Terraform Plugin SDKv2 | Legacy framework. HashiCorp explicitly recommends Plugin Framework for all new providers. SDKv2 lacks type safety, has worse testing support, and will not receive new features. | Terraform Plugin Framework v1.19.0 |
| terraform-plugin-mux | Only needed for incremental SDKv2-to-Framework migration. Since this is a greenfield provider, muxing adds complexity for zero benefit. | Build everything on Plugin Framework from the start. |
| terraform-plugin-codegen-framework | Tech preview status. Designed for providers with OpenAPI specs. Vast.ai has no OpenAPI spec; the Python SDK is our reference. Code generation would add a fragile intermediate step. | Hand-write the Go API client using Python SDK as reference. |
| Terraform Plugin SDKv2 testing (resource.Test from SDKv2) | The old testing approach from SDKv2. terraform-plugin-testing v1.15.0 is the current standalone testing module that works with both Framework and SDKv2. | terraform-plugin-testing v1.15.0 |
| Standard library `log` package | Terraform uses structured logging via terraform-plugin-log/tflog. Raw log output breaks Terraform's log filtering and level management. | terraform-plugin-log (tflog package) |
| ECC GPG keys | Terraform Registry only accepts RSA and DSA keys. ECC keys will fail signature verification. | RSA GPG key (4096-bit recommended) |
| golangci-lint v1 config format | v2 is stable and current. v1 config uses deprecated `enable-all`/`disable-all` syntax. | golangci-lint v2 with `linters.default` config |

## OpenTofu Compatibility

Terraform Plugin Framework providers are automatically compatible with OpenTofu. OpenTofu implements the same provider plugin protocol (v5/v6) as Terraform. No additional code or configuration is needed.

Key compatibility notes:
- OpenTofu v1.6+ supports protocol v5 and v6 providers
- All provider binaries built for Terraform work identically with OpenTofu
- The OpenTofu registry at registry.opentofu.org mirrors Terraform Registry providers
- Test with both `terraform` and `tofu` binaries in CI for confidence (terraform-plugin-testing supports configuring the binary path)

## Version Compatibility Matrix

| Package | Compatible Go | Compatible Terraform | Compatible OpenTofu |
|---------|---------------|---------------------|---------------------|
| terraform-plugin-framework v1.19.0 | Go 1.25+ | Terraform >= 1.0 | OpenTofu >= 1.6 |
| terraform-plugin-testing v1.15.0 | Go 1.25+ | Terraform >= 1.0 | OpenTofu >= 1.6 |
| terraform-plugin-framework-validators v0.19.0 | Go 1.24+ | (via framework) | (via framework) |
| terraform-plugin-docs v0.24.0 | Go 1.22+ | Terraform >= 1.0 | OpenTofu >= 1.6 |
| go-retryablehttp v0.7.8 | Go 1.13+ | N/A | N/A |
| golangci-lint v2.11.x | Go 1.23+ | N/A | N/A |

## Confidence Assessment

| Component | Confidence | Basis |
|-----------|------------|-------|
| Go 1.25.x | HIGH | Verified via scaffolding repo go.mod (1.25.5), pkg.go.dev, and Go release history |
| Terraform Plugin Framework v1.19.0 | HIGH | Verified via GitHub releases (Mar 10, 2025), pkg.go.dev, and scaffolding repo go.mod |
| terraform-plugin-testing v1.15.0 | HIGH | Verified via GitHub releases (Mar 10, 2025) and scaffolding repo go.mod |
| terraform-plugin-docs v0.24.0 | HIGH | Verified via GitHub releases (Oct 13, 2024) |
| terraform-plugin-framework-validators v0.19.0 | HIGH | Verified via GitHub releases (Oct 7, 2024) and pkg.go.dev |
| go-retryablehttp v0.7.8 | MEDIUM | Verified via pkg.go.dev (Jun 18, 2025). No formal GitHub releases page, but package is stable and widely used across HashiCorp ecosystem. |
| GoReleaser v2.14.x | HIGH | Verified via GitHub releases and goreleaser.com |
| golangci-lint v2.11.x | HIGH | Verified via GitHub releases (v2.11.4, Mar 22, 2026) |
| Directory structure | HIGH | Based on HashiCorp scaffolding repo and patterns from AWS/Google/Azure providers |
| OpenTofu compatibility | HIGH | Verified via OpenTofu FAQ and compatibility promises documentation |

## Sources

- [Terraform Plugin Framework GitHub Releases](https://github.com/hashicorp/terraform-plugin-framework/releases) -- version history, v1.19.0 confirmed
- [Terraform Plugin Framework on pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework) -- Go version requirements, module info
- [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) -- go.mod dependencies, .goreleaser.yml, project structure
- [HashiCorp Developer - Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) -- official documentation, best practices
- [HashiCorp Developer - Release and Publish](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-release-publish) -- registry publishing requirements
- [HashiCorp Developer - Acceptance Testing](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests) -- testing framework usage
- [terraform-plugin-testing GitHub Releases](https://github.com/hashicorp/terraform-plugin-testing/releases) -- v1.15.0 confirmed
- [terraform-plugin-docs GitHub Releases](https://github.com/hashicorp/terraform-plugin-docs/releases) -- v0.24.0 confirmed
- [terraform-plugin-framework-validators GitHub](https://github.com/hashicorp/terraform-plugin-framework-validators) -- v0.19.0 confirmed
- [go-retryablehttp on pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/go-retryablehttp) -- v0.7.8 confirmed
- [ghaction-terraform-provider-release](https://github.com/hashicorp/ghaction-terraform-provider-release) -- reusable GitHub Actions workflows
- [HashiCorp Developer - Publish Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing) -- GPG signing, manifest requirements
- [OpenTofu Compatibility Promises](https://opentofu.org/docs/language/v1-compatibility-promises/) -- provider protocol compatibility
- [Go Release History](https://go.dev/doc/devel/release) -- Go 1.25.8, 1.26.1 confirmed

---
*Stack research for: Terraform/OpenTofu Provider Development (Go)*
*Researched: 2026-03-25*
