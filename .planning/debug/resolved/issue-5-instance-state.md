# Debug Session: issue-5-instance-state

**Status:** resolved
**Issue:** https://github.com/realnedsanders/terraform-provider-vastai/issues/5
**Branch:** `fix/issue-5-instance-state`

## Symptom

After `vastai_instance` Create, OpenTofu reports that omitted Optional+Computed values (`env`, `image_login`, `ssh_key_ids`, `use_jupyter_lab`) remain unknown. The instance itself is created successfully.

## Evidence / root-cause trace

1. Terraform Plugin Framework represents omitted Optional+Computed plan values as unknown during Create.
2. `Create` reads the plan into `InstanceResourceModel` and later calls `mapInstanceToModel`.
3. The observed create response supplies only `new_contract`; no resource state can be hydrated from it. The response model retains its existing `success` compatibility field, but Create does not rely on it.
4. The follow-up instance GET exposes `extra_env` and `image_runtype`, but `mapInstanceToModel` currently maps neither.
5. `mapInstanceToModel` leaves every other user field unchanged. Therefore omitted Optional+Computed values remain unknown when `resp.State.Set` runs.
6. `GET /instances/{id}/ssh/` exposes current SSH attachments. Its `ssh_keys` field is itself a JSON-encoded array; the client has no method for it.
7. `use_jupyter_lab`, `image_login`, and `cancel_unavail` are not readable from the instance object. They require explicit create-only state semantics or safe known normalization. Explicit config/secrets must be preserved.

## Working hypothesis

Map authoritative readable fields from the post-create GET and SSH endpoint, preserve explicit config, and normalize only genuinely unreadable omitted residual unknowns. Model create-only inputs so the API request does not invent false values when omitted. This will make every Create result known without replacing explicit values or creating template-merge drift.

## Plan / success criteria

1. Add focused tests for client decoding, mapping/inference, explicit preservation, unknown normalization, and schema/request semantics; confirm RED.
2. Implement the smallest spec-aligned client and resource changes; confirm GREEN.
3. Run `gofmt`, focused tests, `go test ./...`, and `go vet ./...`.
4. Record red/green and final verification evidence below.

## Evidence log

- Baseline issue output: omitted `env`, `image_login`, `ssh_key_ids`, and `use_jupyter_lab` remain unknown after apply.

- RED (`go test ./internal/client ./internal/services/instance`): failed as expected. Missing `Instance.ImageRuntype`, `InstanceService.GetSSHKeys`, `inferUseSSH`, SSH mapping/normalization helpers; request optional booleans are still non-pointer `bool` fields.

- Design decision: `ssh_key_ids` remains Optional+Computed with standard authoritative remote-set semantics. Omitted config adopts the GET result. Explicit config is exact desired state; Create reconciles missing keys and server extras before the final read. No private state is used.
- GREEN (`go test ./internal/client ./internal/services/instance`): both focused packages passed after implementation.
- Additional disk RED: the optional-disk serialization test did not compile because `CreateInstanceRequest.Disk` was still a non-pointer `float64`; omitted requests therefore serialized `disk: 0`.
- Additional disk GREEN: `Disk` is now `*float64,omitempty`; Create omits unconfigured disk, and the follow-up GET hydrates `disk_gb` from positive `disk_space` while preserving explicit config.
- Documentation generation: passed via `tfplugindocs v0.24.0` using Terraform 1.11.4 from `PATH` (the tool's automatic latest-download verification first failed because its OpenPGP key was expired).
- Full verification: `go test ./...` passed; `go vet ./...` passed; `git diff --check` passed.

**Status:** resolved locally; no commit or push performed.
