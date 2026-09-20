---
status: complete
created: 2026-09-20
mode: full
branch: feat/openapi-codegen
---

# Migrate the Vast.ai client to official OpenAPI code generation

Use Vast.ai's official `vast-ai/docs/api-reference/openapi.yaml` as the generated HTTP contract while preserving the provider's stable service API and custom retry/error behavior. Automate upstream spec update pull requests and clear current dependency vulnerability alerts.

## Success criteria

- The official OpenAPI document is vendored with clear provenance and a reproducible, pinned Go code generator.
- Every provider endpoint represented by the official spec uses generated request paths and methods. Unsupported official-spec gaps remain explicit raw-client fallbacks.
- Existing request/response adapters, retry policy, structured errors, waiters, and Terraform-facing behavior remain covered by unit tests.
- CI rejects stale generated code.
- A scheduled GitHub workflow detects upstream spec changes, regenerates and validates the client, then opens a PR; breaking updates open an issue.
- Dependabot tracks Go modules and GitHub Actions.
- Current GitHub Dependabot alerts are fixed by selecting patched dependency versions.
- `go test ./...`, generation checks, build, and lint pass before the PR is opened.

## Tasks

1. Add the vendored official spec, an operation-filtered `oapi-codegen` config, pinned tool dependency, generated client, provenance docs, and generated-response adapter.
2. Route spec-covered service methods through generated operations. Keep raw calls only for endpoints absent from the official spec, and document that boundary.
3. Update and extend unit tests for official paths, request shapes, authentication, retries, pagination, and structured errors.
4. Add Dependabot configuration and a scheduled spec-update PR workflow with failure issue reporting.
5. Upgrade vulnerable transitive Go modules to patched versions and verify GitHub's open alert set after the branch is pushed.
6. Run formatting, generation, unit tests, build, and lint; write the GSD summary; commit atomically; push and open the PR.

## Constraints and decisions

- Keep the existing `internal/client` service API so Terraform resources do not need a broad rewrite.
- Generate only operation IDs used by the provider to keep generated output reviewable; retain the complete official spec as the source input.
- Use generated low-level request methods and the existing handwritten domain models where the official schemas are incomplete or less stable.
- Dependabot cannot watch an arbitrary upstream file, so use Dependabot for supported dependency ecosystems and a dedicated GitHub Actions updater for the OpenAPI file.
