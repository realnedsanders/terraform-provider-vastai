# Phase 5: Account & Networking - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Account configuration resources (API keys, environment variables, teams, team roles, team members, subaccounts) and advanced networking resources (clusters, overlays, membership management) plus read-only data sources (user profile, invoices, audit logs). Reuses all established schema quality patterns.

</domain>

<decisions>
## Implementation Decisions

### Team RBAC
- **D-01:** Claude's Discretion on team member invite modeling — create=invite vs separate invite+member based on API behavior
- **D-02:** Team role permissions as a string set: `permissions = ["create_instance", "delete_instance", "view_billing"]` — simple, matches API

### Cluster/Overlay Networking
- **D-03:** Claude's Discretion on cluster membership modeling (separate `vastai_cluster_member` resource vs inline `machine_ids` attribute) — based on API operations and Terraform membership patterns
- **D-04:** Claude's Discretion on overlay instance joining (separate resource vs inline attribute) — consistent with cluster membership decision

### Data Sources
- **D-05:** Claude's Discretion on user profile depth — based on what API returns
- **D-06:** Claude's Discretion on invoice filtering (date range vs return all) — based on API capabilities
- **D-07:** Claude's Discretion on audit log filtering — based on API capabilities

### Inherited from Prior Phases (DO NOT re-decide)
- All schema conventions: snake_case, Optional+Computed, validators, per-resource models, API models in client
- Service-per-directory, service pattern on client
- Sensitive flag on all secret values (API keys, env var values)
- Dual test strategy: httptest mocks + TF_ACC tests

### Claude's Discretion
- Team member invite flow modeling
- Cluster/overlay membership resource pattern
- Data source field selection and filtering
- Timeout defaults

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Vast.ai API Reference
- `vast-sdk/vastai/vast.py` — Account operations: `create__api_key`, `delete__api_key`, `show__api_keys`, `set__api_key`, `create__env_var`, `delete__env_var`, `update__env_var`, `show__env_vars`, `create__team`, `destroy__team`, `show__members`, `invite__member`, `remove__member`, `create__team_role`, `delete__team_role`, `show__team_roles`, `update__team_role`, `create__subaccount`, `show__subaccounts`
- `vast-sdk/vastai/vast.py` — Networking operations: `create__cluster`, `delete__cluster`, `show__clusters`, `join__cluster`, `remove__machine_from_cluster`, `create__overlay`, `delete__overlay`, `show__overlays`, `join__overlay`
- `vast-sdk/vastai/vast.py` — Data source operations: `show__user`, `show__invoices`, `show__invoices_v1`, `show__audit_logs`

### Existing Provider Code
- `internal/services/sshkey/resource_ssh_key.go` — Simple CRUD resource pattern
- `internal/services/template/resource_template.go` — Resource with CRUD + import
- `internal/services/endpoint/resource_endpoint.go` — Resource with complex schema

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- VastAIClient service pattern — add ApiKeys, EnvVars, Teams, TeamRoles, Clusters, Overlays, etc.
- Simple CRUD resource patterns from SSH key, template
- Data source patterns from offers, endpoints

### Integration Points
- New service directories for each resource type
- New client services in `internal/client/`
- Register all in provider.go (currently 7 resources + 8 data sources)

</code_context>

<specifics>
## Specific Ideas

- Team role permissions as string set keeps the API surface simple and avoids over-abstracting permission categories that may change
- Cluster/overlay membership patterns should follow Terraform conventions for association resources (like aws_security_group_rule)
- API keys and env var values must be marked Sensitive — established pattern from SSH keys

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-account-networking*
*Context gathered: 2026-03-27*
