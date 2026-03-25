# Phase 1: Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-25
**Phase:** 01-foundation
**Areas discussed:** GitHub namespace, API client design, Project layout, CI/CD & release

---

## GitHub Namespace

| Option | Description | Selected |
|--------|-------------|----------|
| Personal account | github.com/your-username/terraform-provider-vastai | |
| New org for this | Create a dedicated GitHub org | |
| Existing org | Use an existing org | |

**User's choice:** Personal account
**Notes:** Username is `realnedsanders`. Module path: `github.com/realnedsanders/terraform-provider-vastai`. Registry namespace: `realnedsanders/vastai`.

---

## API Client Design

### Client Structure

| Option | Description | Selected |
|--------|-------------|----------|
| Single client + methods | One Client struct with all methods | |
| Service pattern | Client with sub-services: client.Instances.Create() | ✓ |
| You decide | Claude picks | |

**User's choice:** Service pattern
**Notes:** Mirrors API surface, scales to 25+ resource types.

### Async Handling

| Option | Description | Selected |
|--------|-------------|----------|
| Built-in waiters | WaitForInstanceReady() with configurable timeout | ✓ |
| Return immediately | Client returns API response, resource handles polling | |
| You decide | Claude picks | |

**User's choice:** Built-in waiters

### Client Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Standalone-capable | internal/client/ with clean interfaces, extractable later | ✓ |
| Provider-coupled | Direct HTTP calls within resource code | |
| You decide | Claude picks | |

**User's choice:** Standalone-capable

### Rate Limiting

| Option | Description | Selected |
|--------|-------------|----------|
| Match Python SDK | 150ms base, 1.5x backoff, max 5 retries | ✓ |
| Use go-retryablehttp | HashiCorp's standard, 1s base, 30s max, 4 retries | |
| You decide | Claude picks | |

**User's choice:** Match Python SDK

### Logging

| Option | Description | Selected |
|--------|-------------|----------|
| Full tflog integration | Request/response at DEBUG/TRACE, errors always | ✓ |
| Minimal | Only errors and retries | |
| You decide | Claude picks | |

**User's choice:** Full tflog integration

---

## Project Layout

### Resource Organization

| Option | Description | Selected |
|--------|-------------|----------|
| Service directories | internal/services/<resource>/ with resource, data source, models, tests | ✓ |
| Flat internal/provider | All in internal/provider/ with naming convention | |
| You decide | Claude picks | |

**User's choice:** Service directories

### Client Location

| Option | Description | Selected |
|--------|-------------|----------|
| internal/client/ | Standard internal package, cleanly separated | ✓ |
| internal/api/ | Alternative naming | |
| You decide | Claude picks | |

**User's choice:** internal/client/

---

## CI/CD & Release

### Release Trigger

| Option | Description | Selected |
|--------|-------------|----------|
| Tag-based | Push git tag → GoReleaser → signed binaries | ✓ |
| Manual dispatch | Manually trigger from GitHub Actions UI | |
| You decide | Claude picks | |

**User's choice:** Tag-based

### GPG Signing

| Option | Description | Selected |
|--------|-------------|----------|
| GitHub Secrets | GPG private key + passphrase as Actions secrets | ✓ |
| You decide | Claude picks | |

**User's choice:** GitHub Secrets

### CI Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Unit on PR, acc on main | PRs: unit + lint. Main: acceptance tests | ✓ |
| Everything on PR | All tests on every PR | |
| You decide | Claude picks | |

**User's choice:** Unit on PR, acc on main

---

## Claude's Discretion

- Go version selection
- GoReleaser configuration details
- golangci-lint configuration
- Directory structure for templates/, examples/, docs/
- Makefile targets

## Deferred Ideas

None
