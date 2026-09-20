---
status: complete
completed: 2026-09-20
branch: feat/openapi-codegen
---

# Official OpenAPI code generation summary

Migrated the Vast.ai HTTP request layer to generated operations from the official `vast-ai/docs` OpenAPI specification while keeping the provider's stable domain facade and live-API compatibility behavior.

## Completed

- Vendored the complete official OpenAPI 3.1 document with immutable commit and blob provenance.
- Added an operation-filtered `oapi-codegen` v2.8.0 client. The generator is pinned in a separate Go tools module and reproducible with `make generate-openapi`.
- Routed every exact documented provider method/path through generated requests. Kept the 19 missing or incompatible route shapes as explicit handwritten fallbacks.
- Preserved retry/backoff, dual authentication, User-Agent, Terraform logging, structured `APIError`, HTTP 200 failure envelopes, raw v1 invoice decoding, waiters, and domain response quirks.
- Preserved known live-API contracts where the combined specification differs: canonical trailing slashes, endpoint delete bodies, volume delete query IDs, and team invite query parameters.
- Added adapter tests for generated request metadata and retry behavior. Existing wire-contract tests continue to cover the compatibility overrides.
- Added a two-stage scheduled OpenAPI updater. It validates untrusted upstream input with read-only permissions, runs generation/build/tests/lint, and then opens a PR. Failed updates open an issue.
- Added weekly Dependabot updates for the root Go module, generator tools module, and GitHub Actions.
- Updated vulnerable dependencies to `google.golang.org/grpc v1.83.2`, `golang.org/x/crypto v0.55.0`, `golang.org/x/net v0.58.0`, and `github.com/cloudflare/circl v1.6.3`.
- Raised the Go patch baseline to 1.25.14 and enabled the repository setting that lets GitHub Actions create pull requests.

## Verification

- `GOTOOLCHAIN=go1.25.14 go test -count=1 ./...`
- `go build ./...`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run`
- `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12`
- `GOTOOLCHAIN=go1.26.8 go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...`
- `go mod verify && go -C tools mod verify`
- Regeneration hash check with `make generate-openapi`

## Commits

- `90e4998` — generated client, compatibility adapter, migrations, tests, and documentation
- `c78ee63` — vulnerable Go module and toolchain patch updates
- `222cbda` — Dependabot and scheduled OpenAPI update automation
