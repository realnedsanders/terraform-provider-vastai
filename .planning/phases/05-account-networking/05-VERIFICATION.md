---
phase: 05-account-networking
verified: 2026-03-27T21:00:00Z
status: passed
score: 13/13 must-haves verified
re_verification: null
gaps: []
human_verification:
  - test: "Apply a vastai_api_key resource and confirm the key value is stored sensitively and preserved across plans"
    expected: "The key attribute is set on creation and unchanged on subsequent terraform plan (UseStateForUnknown); plan shows no diff for key"
    why_human: "Cannot exercise Terraform state management in static analysis; UseStateForUnknown plan-modifier behavior requires a real apply cycle"
  - test: "Apply a vastai_subaccount resource and run terraform destroy; confirm no API DELETE call is made"
    expected: "State is removed; warning diagnostic displayed; Vast.ai account still shows the subaccount"
    why_human: "No-op destroy with warning requires a live API key and Terraform run to confirm the API is not contacted"
  - test: "Apply a vastai_overlay_member resource and run terraform destroy; confirm no API call removes the instance from the overlay"
    expected: "State removed; warning diagnostic displayed; instance still in overlay on Vast.ai console"
    why_human: "Same as above — no-op destroy requires live environment verification"
  - test: "Apply vastai_team_role with a nested JSON permissions object and verify Terraform stores and plans it correctly"
    expected: "Permissions stored as a JSON string; subsequent plans show no spurious diff due to JSON normalization"
    why_human: "JSON string equivalence and plan-diff stability requires a real apply cycle"
---

# Phase 5: Account & Networking Verification Report

**Phase Goal:** Users can manage their Vast.ai account configuration (API keys, teams, environment variables) and advanced networking (clusters, overlays) entirely through Terraform
**Verified:** 2026-03-27T21:00:00Z
**Status:** PASSED
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can create and manage API keys with permission scoping (all secret values marked sensitive) | VERIFIED | `resource_api_key.go:95` Sensitive:true on key; `RequiresReplace` plan modifier on name/permissions; `resource_api_key.go:172` calls `client.ApiKeys.Create` |
| 2 | User can manage environment variables (value marked sensitive) with full CRUD | VERIFIED | `resource_env_var.go:68` Sensitive:true on value; wired to `client.EnvVars.Create/List/Update/Delete`; `env_vars.go:61` uses `DeleteWithBody` |
| 3 | User can manage subaccounts (create-only, destroy is no-op with warning) | VERIFIED | `resource_subaccount.go:245` `AddWarning` on Delete with no API call; wired to `client.Subaccounts.Create/List` |
| 4 | User can create teams, define roles with granular permissions, and invite/remove team members | VERIFIED | team/teamrole/teammember resources all substantive; wired to `client.Teams.*`; asymmetric API handled (read/delete by name, update by ID) |
| 5 | User can create clusters and overlays with membership management | VERIFIED | cluster/overlay resources use create-then-read; cluster/overlay member resources exist with composite IDs; `DeleteWithBody` used |
| 6 | Cluster member join/remove works (NETW-03) | VERIFIED | `resource_cluster_member.go:159` calls `JoinMachine`; `resource_cluster_member.go:357` calls `RemoveMachine` |
| 7 | Overlay member join works (NETW-04); destroy is no-op with warning | VERIFIED | `resource_overlay_member.go:141` calls `JoinInstance`; `resource_overlay_member.go:303` AddWarning with no API call |
| 8 | User can query account profile, billing invoices, and audit logs via read-only data sources | VERIFIED | `data_source_user.go:110`, `data_source_invoices.go:134`, `data_source_audit_logs.go:93` all wired to respective client services |
| 9 | InvoiceService uses GetFullPath for v1 API (not v0) | VERIFIED | `invoices.go:60` path="/api/v1/invoices/"; `invoices.go:66` calls `s.client.GetFullPath` |
| 10 | All 10 new resources and 3 new data sources registered in provider | VERIFIED | `provider.go` Resources() has 17 entries; DataSources() has 11 entries; all Phase 5 names confirmed at lines 152-168, 183-185 |
| 11 | All client services compile and unit tests pass | VERIFIED | `go test ./internal/client/...` → ok (13.050s); `go test ./...` → all 24 packages pass |
| 12 | Full project compiles | VERIFIED | `go build ./...` exits 0 with no output |
| 13 | VastAIClient has all 9 Phase 5 service sub-objects initialized | VERIFIED | `client.go:34-42` declares ApiKeys, EnvVars, Teams, Subaccounts, Clusters, Overlays, Users, Invoices, AuditLogs; all initialized in NewVastAIClient at lines 72-80 |

**Score:** 13/13 truths verified

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/client/api_keys.go` | ApiKeyService with Create, List, Delete | VERIFIED | 57 lines; Create/List/Delete all substantive |
| `internal/client/env_vars.go` | EnvVarService with Create, List, Update, Delete | VERIFIED | 65 lines; DeleteWithBody used correctly |
| `internal/client/teams.go` | TeamService with team CRUD, role CRUD, member invite/list/remove | VERIFIED | 159 lines; all 9 methods present; InviteMember uses query params |
| `internal/client/subaccounts.go` | SubaccountService with Create, List | VERIFIED | 59 lines; no Delete (correct per Pitfall 8) |
| `internal/client/clusters.go` | ClusterService with Create, List, Delete, JoinMachine, RemoveMachine | VERIFIED | 124 lines; create-then-read pattern; DeleteWithBody for cluster/remove |
| `internal/client/overlays.go` | OverlayService with Create, List, Delete, JoinInstance | VERIFIED | 84 lines; create-then-read; DeleteWithBody |
| `internal/client/users.go` | UserService with GetCurrent | VERIFIED | 37 lines; GET /users/current?owner=me |
| `internal/client/invoices.go` | InvoiceService with List (GetFullPath for v1 API) | VERIFIED | 70 lines; GetFullPath used; query param building for filters |
| `internal/client/audit_logs.go` | AuditLogService with List | VERIFIED | 31 lines |
| `internal/client/client.go` | VastAIClient with all 9 new service sub-objects + GetFullPath | VERIFIED | All 9 sub-objects at lines 34-42; GetFullPath at line 251 |
| `internal/client/auth.go` | newRequestFullPath method | VERIFIED | Lines 52-83; constructs URL as baseURL + fullPath (no /api/v0) |
| `internal/services/apikey/resource_api_key.go` | vastai_api_key resource | VERIFIED | 9.6k; Sensitive+UseStateForUnknown on key; RequiresReplace on name/permissions |
| `internal/services/envvar/resource_env_var.go` | vastai_environment_variable resource | VERIFIED | 8.2k; full CRUD; value is Sensitive |
| `internal/services/subaccount/resource_subaccount.go` | vastai_subaccount resource (create-only, no-op destroy) | VERIFIED | 8.3k; AddWarning on Delete, no API call |
| `internal/services/team/resource_team.go` | vastai_team resource | VERIFIED | 7.2k; Create/Read/Delete wired |
| `internal/services/teamrole/resource_team_role.go` | vastai_team_role resource | VERIFIED | 11k; asymmetric API: GetRole(name), UpdateRole(id), DeleteRole(name) |
| `internal/services/teammember/resource_team_member.go` | vastai_team_member resource | VERIFIED | 9.4k; invite=create pattern; ListMembers for ID resolution |
| `internal/services/cluster/resource_cluster.go` | vastai_cluster resource | VERIFIED | 7.4k; create-then-read; DeleteWithBody |
| `internal/services/clustermember/resource_cluster_member.go` | vastai_cluster_member resource (composite ID) | VERIFIED | 13k; composite ID "cluster_id/machine_id"; Split for import |
| `internal/services/overlay/resource_overlay.go` | vastai_overlay resource | VERIFIED | 8.4k; create-then-read; DeleteWithBody |
| `internal/services/overlaymember/resource_overlay_member.go` | vastai_overlay_member resource (no-op destroy) | VERIFIED | 11k; AddWarning on Delete, no API call |
| `internal/services/user/data_source_user.go` | vastai_user data source | VERIFIED | 4.6k; GetCurrent wired |
| `internal/services/invoice/data_source_invoices.go` | vastai_invoices data source | VERIFIED | 5.1k; InvoiceListParams built from model; client.Invoices.List called |
| `internal/services/auditlog/data_source_audit_logs.go` | vastai_audit_logs data source | VERIFIED | 3.8k; AuditLogs.List wired |
| `internal/provider/provider.go` | All Phase 5 resources and data sources registered | VERIFIED | 17 resources, 11 data sources; all 10 new resources + 3 data sources present |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `resource_api_key.go` | `client/api_keys.go` | `client.ApiKeys.Create/List/Delete` | WIRED | Lines 172, 231, 308 |
| `resource_env_var.go` | `client/env_vars.go` | `client.EnvVars.Create/List/Update/Delete` | WIRED | Lines 126, 170, 220, 260 |
| `resource_subaccount.go` | `client/subaccounts.go` | `client.Subaccounts.Create/List` | WIRED | Lines 150, 202 |
| `resource_team.go` | `client/teams.go` | `client.Teams.CreateTeam/DestroyTeam` | WIRED | Lines 119, 215 |
| `resource_team_role.go` | `client/teams.go` | `client.Teams.CreateRole/GetRole/UpdateRole/DeleteRole` | WIRED | Lines 154, 248, 307, 358 |
| `resource_team_member.go` | `client/teams.go` | `client.Teams.InviteMember/ListMembers/RemoveMember` | WIRED | Lines 133, 142, 204, 294 |
| `resource_cluster.go` | `client/clusters.go` | `client.Clusters.Create/List/Delete` | WIRED | Lines 130, 175, 239 |
| `resource_cluster_member.go` | `client/clusters.go` | `client.Clusters.JoinMachine/RemoveMachine/List` | WIRED | Lines 159, 171, 251, 357 |
| `resource_overlay.go` | `client/overlays.go` | `client.Overlays.Create/List/Delete` | WIRED | Lines 143, 196, 269 |
| `resource_overlay_member.go` | `client/overlays.go` | `client.Overlays.JoinInstance/List` | WIRED | Lines 141, 150, 230 |
| `data_source_invoices.go` | `client/invoices.go` | `client.Invoices.List` | WIRED | Line 134 |
| `client/invoices.go` | `client/auth.go` | `GetFullPath -> newRequestFullPath` | WIRED | `invoices.go:66` calls GetFullPath; `client.go:252` calls newRequestFullPath |
| `data_source_user.go` | `client/users.go` | `client.Users.GetCurrent` | WIRED | Line 110 |
| `data_source_audit_logs.go` | `client/audit_logs.go` | `client.AuditLogs.List` | WIRED | Line 93 |
| `provider.go` | all service packages | import + register in Resources()/DataSources() | WIRED | 13 Phase 5 packages imported; all entries in Resources() and DataSources() slices |

---

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `data_source_user.go` | user (*User) | `client.Users.GetCurrent()` → GET /users/current?owner=me | Yes — live API query | FLOWING |
| `data_source_invoices.go` | result (*InvoiceListResponse) | `client.Invoices.List(params)` → GET /api/v1/invoices/ via GetFullPath | Yes — live API query with optional filters | FLOWING |
| `data_source_audit_logs.go` | entries ([]AuditLogEntry) | `client.AuditLogs.List()` → GET /audit_logs/ | Yes — live API query | FLOWING |
| `resource_api_key.go` | key (create response) | `client.ApiKeys.Create()` → POST /auth/apikeys/ | Yes — API returns key value on create only | FLOWING |
| `resource_cluster.go` | cluster (*Cluster) | `client.Clusters.Create()` → create-then-read via List | Yes — create-then-read resolves real ID | FLOWING |
| `resource_overlay.go` | overlay (*Overlay) | `client.Overlays.Create()` → create-then-read via List | Yes — create-then-read resolves real ID | FLOWING |

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| All client tests pass | `go test ./internal/client/... -count=1 -timeout 120s` | ok (13.050s) | PASS |
| Account service tests pass | `go test ./internal/services/apikey/... ./internal/services/envvar/... ./internal/services/subaccount/...` | all ok | PASS |
| Team service tests pass | `go test ./internal/services/team/... ./internal/services/teamrole/... ./internal/services/teammember/...` | all ok | PASS |
| Networking service tests pass | `go test ./internal/services/cluster/... ./internal/services/clustermember/... ./internal/services/overlay/... ./internal/services/overlaymember/...` | all ok | PASS |
| Data source tests pass | `go test ./internal/services/user/... ./internal/services/invoice/... ./internal/services/auditlog/...` | all ok | PASS |
| Full project compiles | `go build ./...` | exit 0, no output | PASS |
| Provider has 17 resources | `grep -c "New.*Resource\b" internal/provider/provider.go` | 17 | PASS |
| Provider has 11 data sources | `grep -c "New.*DataSource\b" internal/provider/provider.go` | 11 | PASS |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| ACCT-01 | 05-01, 05-03 | `vastai_api_key` resource with CRUD and permission management (key marked sensitive) | SATISFIED | `resource_api_key.go` exists; Sensitive:true on key; RequiresReplace on immutable fields; REQUIREMENTS.md now shows Pending but implementation is complete |
| ACCT-02 | 05-01, 05-03 | `vastai_environment_variable` resource with CRUD (value marked sensitive) | SATISFIED | `resource_env_var.go` exists; Sensitive:true on value; full CRUD via client.EnvVars.* |
| ACCT-03 | 05-01, 05-04 | `vastai_team` resource with CRUD | SATISFIED | `resource_team.go` exists; CreateTeam/DestroyTeam wired |
| ACCT-04 | 05-01, 05-04 | `vastai_team_role` resource with CRUD and permission configuration | SATISFIED | `resource_team_role.go` exists; permissions as JSON string (D-02 revised); asymmetric API handled |
| ACCT-05 | 05-01, 05-04 | `vastai_team_member` resource for invite/remove management | SATISFIED | `resource_team_member.go` exists; InviteMember/RemoveMember wired; invite uses query params |
| ACCT-06 | 05-01, 05-03 | `vastai_subaccount` resource with CRUD (no delete endpoint — no-op destroy) | SATISFIED | `resource_subaccount.go` exists; AddWarning on Delete; no API call |
| NETW-01 | 05-02, 05-05 | `vastai_cluster` resource with CRUD | SATISFIED | `resource_cluster.go` exists; create-then-read; DeleteWithBody |
| NETW-02 | 05-02, 05-05 | `vastai_overlay` resource with CRUD (bound to cluster) | SATISFIED | `resource_overlay.go` exists; create-then-read; DeleteWithBody |
| NETW-03 | 05-02, 05-05 | Cluster membership management (join/remove machines) | SATISFIED | `resource_cluster_member.go` exists; composite ID; JoinMachine/RemoveMachine wired |
| NETW-04 | 05-02, 05-05 | Overlay membership management (join instances; no remove endpoint — no-op destroy) | SATISFIED | `resource_overlay_member.go` exists; JoinInstance wired; AddWarning on Delete |
| DATA-07 | 05-02, 05-06 | `vastai_user` data source (current account profile) | SATISFIED | `data_source_user.go` exists; GetCurrent wired |
| DATA-10 | 05-02, 05-06 | `vastai_invoices` data source (billing history, read-only) | SATISFIED | `data_source_invoices.go` exists; Invoices.List wired; GetFullPath for v1 API |
| DATA-11 | 05-02, 05-06 | `vastai_audit_logs` data source (account activity, read-only) | SATISFIED | `data_source_audit_logs.go` exists; AuditLogs.List wired |

**Note on REQUIREMENTS.md traceability table:** The traceability table in REQUIREMENTS.md marks ACCT-01, ACCT-02, and ACCT-06 as Pending while ACCT-03/04/05 and NETW-01/02/03/04 and DATA-07/10/11 show Complete. The Pending status is a documentation artifact — the implementation for all 13 requirements is complete and tests pass. REQUIREMENTS.md was not updated to reflect ACCT-01/02/06 completion after Phase 5 execution.

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None | — | — | — | — |

No TODO/FIXME/placeholder comments found in any Phase 5 service files. No stub implementations detected. All resources have substantive CRUD implementations.

---

### Human Verification Required

#### 1. API Key Sensitive Preservation Across Plans

**Test:** Create a `vastai_api_key` resource with `terraform apply`. Then run `terraform plan` again.
**Expected:** The `key` attribute shows no diff; plan is empty for the key field despite the API not returning the key value on re-reads.
**Why human:** UseStateForUnknown plan-modifier behavior requires a real Terraform plan/apply cycle with live state. Cannot be verified by static analysis.

#### 2. Subaccount No-Op Destroy

**Test:** Create a `vastai_subaccount` resource, then run `terraform destroy`.
**Expected:** Terraform shows a warning diagnostic "Subaccount Not Deleted"; the account is removed from Terraform state; the subaccount still exists in the Vast.ai console.
**Why human:** Requires live API key, creates a real subaccount, and verifies no delete request was made to the API.

#### 3. Overlay Member No-Op Destroy

**Test:** Join an instance to an overlay via `vastai_overlay_member`, then run `terraform destroy`.
**Expected:** Terraform shows a warning diagnostic "Overlay Member Not Removed"; state is cleared; instance still appears in the overlay on Vast.ai.
**Why human:** Same as above — requires live environment to confirm the API is not contacted on destroy.

#### 4. Team Role JSON Permissions Stability

**Test:** Create a `vastai_team_role` with nested JSON permissions (e.g., `{"api":{"instance_read":{}}}`). Run `terraform plan` again without changes.
**Expected:** Plan shows no diff for the permissions attribute despite JSON string comparison.
**Why human:** JSON normalization and plan stability for string-typed JSON attributes requires a real apply cycle to confirm no spurious diff.

---

### Gaps Summary

No gaps. All 13 observable truths verified. All 25 required artifacts exist, are substantive, and are wired. All key links confirmed. 24 test packages pass. Full project compiles. The phase goal — managing Vast.ai account configuration and advanced networking entirely through Terraform — is fully achieved.

The only open items are 4 human verification points that require live Terraform runs with a real API key, all relating to Terraform state management behaviors (UseStateForUnknown, no-op destroy warnings) that cannot be exercised through static code analysis.

**REQUIREMENTS.md note:** The traceability table still shows ACCT-01, ACCT-02, ACCT-06 as "Pending" — this is a documentation inconsistency, not an implementation gap. The implementation is complete and all tests pass.

---

_Verified: 2026-03-27T21:00:00Z_
_Verifier: Claude (gsd-verifier)_
