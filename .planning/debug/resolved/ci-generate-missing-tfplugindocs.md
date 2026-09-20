---
status: resolved
trigger: "Tests/generate fails on PRs #6, #7, and #8"
created: 2026-09-20
---

# Debug Session: CI generation tool missing

## Symptom

All three PR `Tests / generate` jobs fail during `make generate`.

## Evidence

- `gh run view --job ... --log-failed` shows `go generate ./...` reaches `main.go:1` and fails with `exec: "tfplugindocs": executable file not found in $PATH`.
- `.github/workflows/test.yml` installs Go and Terraform in the generate job, but never installs `tfplugindocs`.
- `main.go` invokes the external executable through `//go:generate tfplugindocs generate --provider-name vastai`.

## Root cause hypothesis

The generate job lacks the documentation generator required by the repository's `go:generate` directive. Install the project-pinned v0.24.0 binary and add `GOPATH/bin` to subsequent-step PATH before `make generate`.

## Success criteria

1. A clean environment can install `tfplugindocs@v0.24.0` and run `make generate` with Terraform on PATH.
2. Generation produces no uncommitted changes on each PR branch.
3. GitHub `Tests / generate` reruns successfully.

## Resolution

Installed `github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.24.0` in the generate job and exported `$(go env GOPATH)/bin` through `GITHUB_PATH` before `make generate`.

## Verification

- Installed the same pinned tool into an empty temporary `GOBIN`.
- Ran `make generate` with Terraform 1.11.4 on PATH successfully.
- Generation produced no documentation or source diff.
- `git diff --check` passed.
