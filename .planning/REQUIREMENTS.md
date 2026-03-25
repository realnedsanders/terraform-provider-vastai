# Requirements: terraform-provider-vastai

**Defined:** 2026-03-25
**Core Value:** Full, reliable IaC control over Vast.ai infrastructure — every API resource manageable through Terraform with the same quality bar as first-party providers.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Foundation

- [ ] **FOUND-01**: Provider authenticates via `VASTAI_API_KEY` env var or `api_key` provider attribute (marked sensitive)
- [ ] **FOUND-02**: Provider supports configurable API endpoint URL via `VASTAI_API_URL` env var or `api_url` provider attribute
- [ ] **FOUND-03**: Go HTTP client with exponential backoff retry on 429/5xx, configurable max retries
- [ ] **FOUND-04**: User-Agent header includes provider version (`terraform-provider-vastai/vX.Y.Z`)
- [ ] **FOUND-05**: API key authentication uses Bearer header (not query parameter) to prevent credential leaks in logs
- [ ] **FOUND-06**: Structured error diagnostics with summary and detail on all API failures

### Compute Resources

- [ ] **COMP-01**: `vastai_instance` resource with full CRUD (create from offer ID, read, update, destroy)
- [ ] **COMP-02**: `vastai_instance` supports start/stop via `status` attribute without destroy/recreate
- [ ] **COMP-03**: `vastai_instance` supports label, bid price change, and template update
- [ ] **COMP-04**: `vastai_instance` handles spot/interruptible instance preemption gracefully (remove from state, not error)
- [ ] **COMP-05**: `vastai_instance` creation polls until instance reaches running state (async create with waiter)
- [ ] **COMP-06**: `vastai_template` resource with full CRUD (image, env vars, onstart_cmd, SSH/Jupyter flags)
- [ ] **COMP-07**: `vastai_ssh_key` resource with full CRUD
- [ ] **COMP-08**: `vastai_ssh_key` supports attach/detach to instances

### Storage Resources

- [ ] **STOR-01**: `vastai_volume` resource with CRUD (create from offer, delete, list/unlist for marketplace)
- [ ] **STOR-02**: `vastai_volume` supports clone operation
- [ ] **STOR-03**: `vastai_network_volume` resource with CRUD (create, delete, list/unlist for marketplace)

### Serverless Resources

- [ ] **SRVL-01**: `vastai_endpoint` resource with CRUD and autoscaling parameters (min_load, target_util, cold_mult, cold_workers, max_workers)
- [ ] **SRVL-02**: `vastai_worker_group` resource with CRUD (bound to endpoint, template, search params, autoscaling config)
- [ ] **SRVL-03**: `vastai_autogroup` resource with CRUD for autoscaling groups

### Networking Resources

- [ ] **NETW-01**: `vastai_cluster` resource with CRUD
- [ ] **NETW-02**: `vastai_overlay` resource with CRUD (bound to cluster)
- [ ] **NETW-03**: Cluster membership management (join/remove machines)
- [ ] **NETW-04**: Overlay membership management (join instances to overlay)

### Account Resources

- [ ] **ACCT-01**: `vastai_api_key` resource with CRUD and permission management (key value marked sensitive)
- [ ] **ACCT-02**: `vastai_environment_variable` resource with CRUD (value marked sensitive)
- [ ] **ACCT-03**: `vastai_team` resource with CRUD
- [ ] **ACCT-04**: `vastai_team_role` resource with CRUD and permission configuration
- [ ] **ACCT-05**: `vastai_team_member` resource for invite/remove management
- [ ] **ACCT-06**: `vastai_subaccount` resource with CRUD

### Data Sources

- [ ] **DATA-01**: `vastai_gpu_offers` data source with structured filter attributes (gpu_name, num_gpus, gpu_ram, price, region, datacenter_only, etc.)
- [ ] **DATA-02**: `vastai_instance` data source (singular by ID)
- [ ] **DATA-03**: `vastai_instances` data source (list with optional filtering)
- [ ] **DATA-04**: `vastai_templates` data source (search by query)
- [ ] **DATA-05**: `vastai_volume_offers` data source with filter support
- [ ] **DATA-06**: `vastai_network_volume_offers` data source with filter support
- [ ] **DATA-07**: `vastai_user` data source (current account profile)
- [ ] **DATA-08**: `vastai_ssh_keys` data source (list all keys)
- [ ] **DATA-09**: `vastai_endpoints` data source (list serverless endpoints)
- [ ] **DATA-10**: `vastai_invoices` data source (billing history, read-only)
- [ ] **DATA-11**: `vastai_audit_logs` data source (account activity, read-only)

### Schema Quality

- [ ] **SCHM-01**: Attribute validators on all constrained fields (int ranges, string lengths, enum values)
- [ ] **SCHM-02**: `Sensitive` flag on all secret attributes (API keys, env var values)
- [ ] **SCHM-03**: Correct Required/Optional/Computed classification matching API behavior
- [ ] **SCHM-04**: Meaningful description on every attribute (rendered in registry docs)
- [ ] **SCHM-05**: Plan modifiers: `UseStateForUnknown` for stable computed fields, `RequiresReplace` for immutable fields
- [ ] **SCHM-06**: Configurable timeouts per resource via `terraform-plugin-framework-timeouts`

### Import Support

- [ ] **IMPT-01**: `terraform import` support for all managed resources via resource ID
- [ ] **IMPT-02**: Import documentation with example commands for each resource

### Testing

- [ ] **TEST-01**: Acceptance tests for all resources (create, read, update, import, destroy)
- [ ] **TEST-02**: Unit tests for validators, plan modifiers, and API client logic
- [ ] **TEST-03**: Resource sweepers to clean up leaked test resources
- [ ] **TEST-04**: CI pipeline running tests on PR (unit tests always, acceptance tests on main)

### Documentation

- [ ] **DOCS-01**: Generated documentation via tfplugindocs for all resources and data sources
- [ ] **DOCS-02**: Provider configuration documentation (auth, endpoint, retry)
- [ ] **DOCS-03**: Working examples in `examples/` directory for common workflows
- [ ] **DOCS-04**: Per-resource example `.tf` files

### Release & Registry

- [ ] **RLSE-01**: GoReleaser configuration with GPG-signed releases
- [ ] **RLSE-02**: GitHub Actions CI/CD pipeline for automated releases on tag push
- [ ] **RLSE-03**: `terraform-registry-manifest.json` declaring protocol version 6.0
- [ ] **RLSE-04**: Semantic versioning with `v` prefix
- [ ] **RLSE-05**: SHA256SUMS and .sig files generated with each release

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced Features

- **ADV-01**: Ephemeral resource for API keys (Terraform 1.10+, prevents state persistence of secrets)
- **ADV-02**: Identity-based import by name/attributes instead of opaque IDs (Terraform 1.12+)
- **ADV-03**: Terraform actions for instance operations like reboot/recycle (Terraform 1.14+)
- **ADV-04**: Guide documentation (training workflow, serverless setup, cost optimization)
- **ADV-05**: State upgrade/migration infrastructure for breaking schema changes
- **ADV-06**: Drift detection improvements (detect out-of-band changes beyond basic refresh)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Instance command execution | Imperative operation — use `remote-exec` provisioner or external tooling |
| File transfer operations (vm_copy, cloud_copy) | Non-idempotent imperative action — use provisioners or Vast.ai CLI |
| Bulk instance creation resource | Violates per-resource model — use `count`/`for_each` meta-arguments |
| 2FA management | Security-critical interactive flow — manage through Vast.ai console |
| Host-side machine operations (list, unlist, maintenance, defjob) | Wrong audience — this provider targets tenants, not GPU hosts |
| Snapshot resource | Point-in-time imperative action — document as CLI workflow |
| Scheduled job resource | Scheduling belongs in external tooling (cron, Kubernetes CronJobs) |
| Billing write operations (transfer_credit) | Financial operations don't belong in IaC |
| Automatic spot instance replacement | Creates infinite loops and unexpected costs — document handling in guides |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| FOUND-01 | TBD | Pending |
| FOUND-02 | TBD | Pending |
| FOUND-03 | TBD | Pending |
| FOUND-04 | TBD | Pending |
| FOUND-05 | TBD | Pending |
| FOUND-06 | TBD | Pending |
| COMP-01 | TBD | Pending |
| COMP-02 | TBD | Pending |
| COMP-03 | TBD | Pending |
| COMP-04 | TBD | Pending |
| COMP-05 | TBD | Pending |
| COMP-06 | TBD | Pending |
| COMP-07 | TBD | Pending |
| COMP-08 | TBD | Pending |
| STOR-01 | TBD | Pending |
| STOR-02 | TBD | Pending |
| STOR-03 | TBD | Pending |
| SRVL-01 | TBD | Pending |
| SRVL-02 | TBD | Pending |
| SRVL-03 | TBD | Pending |
| NETW-01 | TBD | Pending |
| NETW-02 | TBD | Pending |
| NETW-03 | TBD | Pending |
| NETW-04 | TBD | Pending |
| ACCT-01 | TBD | Pending |
| ACCT-02 | TBD | Pending |
| ACCT-03 | TBD | Pending |
| ACCT-04 | TBD | Pending |
| ACCT-05 | TBD | Pending |
| ACCT-06 | TBD | Pending |
| DATA-01 | TBD | Pending |
| DATA-02 | TBD | Pending |
| DATA-03 | TBD | Pending |
| DATA-04 | TBD | Pending |
| DATA-05 | TBD | Pending |
| DATA-06 | TBD | Pending |
| DATA-07 | TBD | Pending |
| DATA-08 | TBD | Pending |
| DATA-09 | TBD | Pending |
| DATA-10 | TBD | Pending |
| DATA-11 | TBD | Pending |
| SCHM-01 | TBD | Pending |
| SCHM-02 | TBD | Pending |
| SCHM-03 | TBD | Pending |
| SCHM-04 | TBD | Pending |
| SCHM-05 | TBD | Pending |
| SCHM-06 | TBD | Pending |
| IMPT-01 | TBD | Pending |
| IMPT-02 | TBD | Pending |
| TEST-01 | TBD | Pending |
| TEST-02 | TBD | Pending |
| TEST-03 | TBD | Pending |
| TEST-04 | TBD | Pending |
| DOCS-01 | TBD | Pending |
| DOCS-02 | TBD | Pending |
| DOCS-03 | TBD | Pending |
| DOCS-04 | TBD | Pending |
| RLSE-01 | TBD | Pending |
| RLSE-02 | TBD | Pending |
| RLSE-03 | TBD | Pending |
| RLSE-04 | TBD | Pending |
| RLSE-05 | TBD | Pending |

**Coverage:**
- v1 requirements: 55 total
- Mapped to phases: 0
- Unmapped: 55

---
*Requirements defined: 2026-03-25*
*Last updated: 2026-03-25 after initial definition*
