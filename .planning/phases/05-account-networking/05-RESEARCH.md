# Phase 5: Account & Networking - Research

**Researched:** 2026-03-27
**Domain:** Vast.ai Account Management, Team RBAC, Networking (Clusters/Overlays), Read-Only Data Sources
**Confidence:** HIGH

## Summary

Phase 5 covers the largest batch of new resources and data sources in the provider: 6 account resources (API key, environment variable, team, team role, team member, subaccount), 2 networking resources (cluster, overlay), membership management for both networking types, and 3 read-only data sources (user profile, invoices, audit logs). All API contracts have been extracted from the Python SDK (`vast-sdk/vastai/vast.py`) and documented below.

The API patterns are straightforward CRUD operations consistent with prior phases, with a few notable exceptions: (1) environment variables use `key` (name) as their identifier rather than a numeric ID, (2) team role delete/read uses NAME as path parameter while update uses numeric ID, (3) the cluster delete and overlay delete use `DeleteWithBody` pattern (JSON body, not path parameter), (4) the invoices v1 endpoint uses a different API version prefix (`/api/v1/invoices/`) requiring a small client enhancement, and (5) the permissions model for API keys and team roles is a nested JSON object that should be stored as a JSON string in Terraform state.

**Primary recommendation:** Implement account resources first (simpler CRUD, well-understood patterns), then networking resources with separate membership resources following the `aws_security_group_rule` pattern, and finally read-only data sources. The client needs a `GetFullPath` method (or similar) to support the `/api/v1/` invoices endpoint.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- D-02: Team role permissions as a string set: `permissions = ["create_instance", "delete_instance", "view_billing"]` -- simple, matches API
- All schema conventions: snake_case, Optional+Computed, validators, per-resource models, API models in client
- Service-per-directory, service pattern on client
- Sensitive flag on all secret values (API keys, env var values)
- Dual test strategy: httptest mocks + TF_ACC tests

### Claude's Discretion
- D-01: Team member invite modeling -- create=invite vs separate invite+member based on API behavior
- D-03: Cluster membership modeling (separate `vastai_cluster_member` resource vs inline `machine_ids` attribute)
- D-04: Overlay instance joining (separate resource vs inline attribute)
- D-05: User profile depth -- based on what API returns
- D-06: Invoice filtering (date range vs return all) -- based on API capabilities
- D-07: Audit log filtering -- based on API capabilities
- Timeout defaults

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ACCT-01 | `vastai_api_key` resource with CRUD and permission management (key value marked sensitive) | API contracts extracted: POST/GET/DELETE `/auth/apikeys/`. Permissions are nested JSON object. Key value returned on create only. |
| ACCT-02 | `vastai_environment_variable` resource with CRUD (value marked sensitive) | API contracts extracted: POST/GET/PUT/DELETE `/secrets/`. Keyed by name, not numeric ID. Response is `{"secrets": {"key": "value"}}` map. |
| ACCT-03 | `vastai_team` resource with CRUD | API contracts extracted: POST/DELETE/GET `/team/`. Create sends `team_name`, destroy is parameterless DELETE. |
| ACCT-04 | `vastai_team_role` resource with CRUD and permission configuration | API contracts extracted: POST `/team/roles/`, GET `/team/roles-full/`, GET/PUT/DELETE `/team/roles/{identifier}/`. Delete uses NAME, update uses integer ID. |
| ACCT-05 | `vastai_team_member` resource for invite/remove management | API contracts extracted: POST `/team/invite/?email=X&role=Y`, GET `/team/members/`, DELETE `/team/members/{id}/`. Invite uses query params, not JSON body. |
| ACCT-06 | `vastai_subaccount` resource with CRUD | API contracts extracted: POST `/users/` with email/username/password/host_only. GET `/subaccounts?owner=me`. No individual delete endpoint found. |
| NETW-01 | `vastai_cluster` resource with CRUD | API contracts extracted: POST/DELETE `/cluster/`, GET `/clusters/`. Create sends subnet+manager_id. Delete uses DeleteWithBody with cluster_id. |
| NETW-02 | `vastai_overlay` resource with CRUD (bound to cluster) | API contracts extracted: POST/DELETE `/overlay/`, GET `/overlay/`. Create sends cluster_id+name. Delete uses DeleteWithBody with overlay_id or overlay_name. |
| NETW-03 | Cluster membership management (join/remove machines) | API contracts extracted: PUT `/cluster/` for join (cluster_id+machine_ids), DELETE `/cluster/remove_machine/` for remove (cluster_id+machine_id+optional new_manager_id). |
| NETW-04 | Overlay membership management (join instances to overlay) | API contracts extracted: PUT `/overlay/` for join (name+instance_id). No explicit leave/remove endpoint found in SDK. |
| DATA-07 | `vastai_user` data source (current account profile) | API contracts extracted: GET `/users/current?owner=me`. Returns rich user object with 20+ fields (balance, email, billing, etc.). |
| DATA-10 | `vastai_invoices` data source (billing history, read-only) | API contracts extracted: GET `/api/v1/invoices/` with pagination, date filtering, type filtering. Uses v1 API prefix. |
| DATA-11 | `vastai_audit_logs` data source (account activity, read-only) | API contracts extracted: GET `/audit_logs/`. Returns list of {ip_address, api_key_id, created_at, api_route, args}. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| terraform-plugin-framework | v1.19.0 | Provider SDK | Already in go.mod, established pattern |
| terraform-plugin-framework-timeouts | v0.5.0 | Configurable timeouts | Already in go.mod, used on all resources |
| terraform-plugin-framework-validators | v0.19.0 | Attribute validators | Already in go.mod, string/int/float validators |
| terraform-plugin-log (tflog) | v0.10.0 | Structured logging | Already in go.mod, required for all provider logging |
| hashicorp/go-retryablehttp | v0.7.8 | HTTP client with retries | Already in go.mod, foundation of API client |

No new dependencies required. All libraries are already in go.mod from prior phases.

**Installation:** No new `go get` commands needed.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  client/
    api_keys.go              # ApiKeyService + API models
    api_keys_test.go
    clusters.go              # ClusterService + API models
    clusters_test.go
    env_vars.go              # EnvVarService + API models
    env_vars_test.go
    invoices.go              # InvoiceService + API models
    invoices_test.go
    overlays.go              # OverlayService + API models
    overlays_test.go
    teams.go                 # TeamService + API models (team, members, roles)
    teams_test.go
    subaccounts.go           # SubaccountService + API models
    subaccounts_test.go
    users.go                 # UserService + API models
    users_test.go
    audit_logs.go            # AuditLogService + API models
    audit_logs_test.go
    client.go                # Add new service sub-objects + GetFullPath method
  services/
    apikey/
      models.go
      resource_api_key.go
      resource_api_key_test.go
    envvar/
      models.go
      resource_env_var.go
      resource_env_var_test.go
    team/
      models.go
      resource_team.go
      resource_team_test.go
    teamrole/
      models.go
      resource_team_role.go
      resource_team_role_test.go
    teammember/
      models.go
      resource_team_member.go
      resource_team_member_test.go
    subaccount/
      models.go
      resource_subaccount.go
      resource_subaccount_test.go
    cluster/
      models.go
      resource_cluster.go
      resource_cluster_test.go
    clustermember/
      models.go
      resource_cluster_member.go
      resource_cluster_member_test.go
    overlay/
      models.go
      resource_overlay.go
      resource_overlay_test.go
    overlaymember/
      models.go
      resource_overlay_member.go
      resource_overlay_member_test.go
    user/
      models.go
      data_source_user.go
      data_source_user_test.go
    invoice/
      models.go
      data_source_invoices.go
      data_source_invoices_test.go
    auditlog/
      models.go
      data_source_audit_logs.go
      data_source_audit_logs_test.go
  provider/
    provider.go              # Register all new resources + data sources
```

### Pattern 1: Standard CRUD Resource (API Key, Team, Cluster, Overlay)
**What:** Same resource pattern used in SSH key, template, endpoint -- struct with `client *client.VastAIClient`, interface assertions, Schema/Metadata/Configure/Create/Read/Update/Delete methods, ImportState.
**When to use:** Every managed resource.
**Example:** Follow `internal/services/sshkey/resource_ssh_key.go` exactly.

### Pattern 2: Name-Keyed Resource (Environment Variable)
**What:** Resource where the unique identifier is a string name (key) rather than a numeric ID. The `id` attribute in Terraform state stores the name string.
**When to use:** `vastai_environment_variable` -- the API uses key name for CRUD, not a numeric ID.
**Example:**
```go
// Environment variables are keyed by name, not numeric ID
// Terraform ID = env var name (the "key" field)
// Create: POST /secrets/ with {"key": name, "value": value}
// Read: GET /secrets/ returns {"secrets": {"name1": "val1", "name2": "val2"}}
// Update: PUT /secrets/ with {"key": name, "value": value}
// Delete: DELETE /secrets/ with {"key": name} (DeleteWithBody)
```

### Pattern 3: Association/Membership Resource (Cluster Member, Overlay Member)
**What:** Separate resource representing the join between a parent (cluster/overlay) and a child (machine/instance). Uses composite ID (`cluster_id/machine_id`) for Terraform state.
**When to use:** Cluster membership (NETW-03) and overlay membership (NETW-04).
**Example:**
```go
// Composite ID pattern for association resources
// ID format: "cluster_id/machine_id" (parsed on import)
// Create: PUT /cluster/ with {"cluster_id": X, "machine_ids": [Y]}
// Read: GET /clusters/ and check if machine_id is in cluster's nodes
// Delete: DELETE /cluster/remove_machine/ with {"cluster_id": X, "machine_id": Y}
```

### Pattern 4: Invite-as-Create Resource (Team Member)
**What:** Resource where "create" means "invite" -- the team member invite endpoint functions as the create operation. The member's user ID becomes the resource ID after invite acceptance.
**When to use:** `vastai_team_member` (ACCT-05).
**Recommendation for D-01:** Model as a single `vastai_team_member` resource where Create calls the invite endpoint. This is simpler and more idiomatic for Terraform -- users declare "this member should exist" and Terraform ensures they're invited.

### Pattern 5: Read-Only Data Source (User, Invoices, Audit Logs)
**What:** Data source with no CRUD -- only Read. No timeouts block needed (data sources are read-only).
**When to use:** DATA-07, DATA-10, DATA-11.
**Example:** Follow `internal/services/sshkey/data_source_ssh_keys.go`.

### Anti-Patterns to Avoid
- **Inline membership lists on parent resources:** Do NOT put `machine_ids` as an attribute on `vastai_cluster` or `instance_ids` on `vastai_overlay`. This forces full resource replacement when membership changes and cannot be managed independently. Use separate membership resources instead.
- **Hardcoded API version prefix for v1 endpoints:** The invoices v1 endpoint uses `/api/v1/invoices/`. Do NOT work around this by string manipulation in service code. Add a proper `GetFullPath` client method.
- **Treating permissions as individual boolean attributes:** The permissions JSON is nested and evolving. Store as a JSON string attribute, not as 10+ boolean fields that would break when Vast.ai adds new permission categories.

## Discretion Recommendations

### D-01: Team Member Invite Modeling
**Recommendation:** Single `vastai_team_member` resource where Create = invite.
**Rationale:** The API has `POST /team/invite/?email=X&role=Y` for invite and `DELETE /team/members/{id}/` for remove. There is no separate "accept invite" or "member exists" state that Terraform needs to manage. The invite IS the membership creation. A separate `vastai_team_invite` + `vastai_team_member` pair would add complexity with no benefit -- the invite IS the member creation from the API's perspective.

### D-03: Cluster Membership Modeling
**Recommendation:** Separate `vastai_cluster_member` resource.
**Rationale:** The API has distinct endpoints for join (`PUT /cluster/` with machine_ids) and remove (`DELETE /cluster/remove_machine/` with machine_id). These are independent operations that should be independently manageable in Terraform. The `aws_security_group_rule` pattern is the established Terraform convention for this. Inline `machine_ids` on the cluster resource would cause issues: (1) can't manage individual machines, (2) ordering issues with lists, (3) force-new on any change.

### D-04: Overlay Instance Joining
**Recommendation:** Separate `vastai_overlay_member` resource, consistent with D-03.
**Rationale:** Same reasoning as clusters. API has `PUT /overlay/` for join. No explicit leave/remove-instance endpoint found in SDK, but overlay delete removes all instances. For individual instance removal, the overlay may need to be recreated -- document this limitation. Alternatively, the join operation may be one-way (instance stays until overlay is deleted).

### D-05: User Profile Depth
**Recommendation:** Include all non-sensitive read-only fields from the API response. The `user_fields` tuple in the SDK shows 20+ fields: balance, credit, email, username, fullname, billing address fields, has_billing, has_payout, etc. Expose a focused subset: id, username, email, email_verified, fullname, balance, credit, has_billing, ssh_key. Mark `api_key` as excluded (the SDK explicitly pops it).

### D-06: Invoice Filtering
**Recommendation:** Support the v1 API pagination and date filtering. The v1 endpoint accepts `start_date`, `end_date`, `limit`, `after_token`, `latest_first`, and type filters (`invoices` vs `charges`, with sub-type filtering). For the Terraform data source, expose `start_date` and `end_date` as optional string attributes (YYYY-MM-DD format), `limit` as optional int (default 100), and `type` as required string enum ("invoices" or "charges"). Return `results` list and `total` count.

### D-07: Audit Log Filtering
**Recommendation:** No filtering needed -- the API endpoint `GET /audit_logs/` returns all logs without filter parameters. The data source should just expose the full list with the 5 fields: ip_address, api_key_id, created_at, api_route, args.

## API Contract Reference

### API Keys (ACCT-01)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Create | POST | `/auth/apikeys/` | `{"name": str, "permissions": {json}, "key_params": str?}` | `{id, key, name, ...}` (key only on create) |
| List | GET | `/auth/apikeys/` | -- | `[{id, name, permissions, ...}]` |
| Delete | DELETE | `/auth/apikeys/{id}/` | -- | `{success, msg}` |
| (No update) | -- | -- | -- | API keys are immutable after creation |

**Key pitfall:** API key `key` value is only returned on create. Must be stored in state on create and never retrievable again. Mark as `Sensitive: true` and `UseStateForUnknown` on read.

### Environment Variables (ACCT-02)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Create | POST | `/secrets/` | `{"key": str, "value": str}` | `{success, msg}` |
| List | GET | `/secrets/` | -- | `{"secrets": {"key1": "val1", ...}}` |
| Update | PUT | `/secrets/` | `{"key": str, "value": str}` | `{success, msg}` |
| Delete | DELETE | `/secrets/` | `{"key": str}` (body) | `{success, msg}` |

**Key pitfall:** Delete uses `DeleteWithBody` pattern (JSON body, not path parameter). List returns a map, not an array -- must find by key name. No individual GET by key; must list all and filter.

### Teams (ACCT-03)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Create | POST | `/team/` | `{"team_name": str}` | `{id, team_name, ...}` |
| Destroy | DELETE | `/team/` | -- (no params) | `{success, msg}` |
| (No update) | -- | -- | -- | Team name appears immutable |
| (No single read) | -- | -- | -- | No single-team GET; may need to infer from context |

**Key pitfall:** Destroy is parameterless -- it destroys the team associated with the current API key context. There is no team ID in the delete endpoint. This means the provider can only manage one team per API key context. Create likely returns a team ID or switches context.

### Team Roles (ACCT-04)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Create | POST | `/team/roles/` | `{"name": str, "permissions": {json}}` | `{id, name, permissions}` |
| List All | GET | `/team/roles-full/` | -- | `[{id, name, permissions}, ...]` |
| Read One | GET | `/team/roles/{name}/` | -- | `{id, name, permissions}` |
| Update | PUT | `/team/roles/{id}/` | `{"name": str, "permissions": {json}}` | `{id, name, permissions}` |
| Delete | DELETE | `/team/roles/{name}/` | -- | `{success, msg}` |

**Key pitfall:** Read and Delete use role NAME in the path. Update uses role ID in the path. This inconsistency means the provider must track both name and ID. Import should support ID (since that is the stable identifier).

### Team Members (ACCT-05)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Invite | POST | `/team/invite/?email=X&role=Y` | -- (query params) | `{success}` (200 on success) |
| List | GET | `/team/members/` | -- | `[{id, email, role, ...}]` |
| Remove | DELETE | `/team/members/{id}/` | -- | `{success, msg}` |

**Key pitfall:** Invite uses query parameters, NOT JSON body. This is unusual. The Go client's `Post` method only sends JSON bodies, so invite will need a custom method or path-level query parameter encoding (e.g., `Post(ctx, "/team/invite/?email=X&role=Y", nil, &resp)`).

### Subaccounts (ACCT-06)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Create | POST | `/users/` | `{"email": str, "username": str, "password": str, "host_only": bool, "parent_id": "me"}` | `{id, ...}` |
| List | GET | `/subaccounts?owner=me` | -- | `{"users": [{id, email, username, ...}]}` |
| (No delete) | -- | -- | -- | No delete endpoint found in SDK |
| (No update) | -- | -- | -- | No update endpoint found in SDK |

**Key pitfall:** No delete endpoint found. This means subaccounts may need to be create-only (no destroy). Terraform can track the resource but destroy would be a no-op or error. Document this limitation. The `password` field is sensitive and write-only (never returned by API).

### Clusters (NETW-01)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Create | POST | `/cluster/` | `{"subnet": str, "manager_id": int}` | `{msg}` (no ID in response) |
| List | GET | `/clusters/` | -- | `{"clusters": {"id1": {subnet, nodes: [{machine_id, is_cluster_manager, local_ip}]}, ...}}` |
| Delete | DELETE | `/cluster/` | `{"cluster_id": int}` (body) | `{msg}` |

**Key pitfall:** Create returns only `{msg}` -- no cluster ID. Must use create-then-read pattern (list all clusters, find the new one by subnet match). Delete uses `DeleteWithBody`. The cluster list response uses cluster IDs as map keys, not array indices. Nodes contain `machine_id`, `is_cluster_manager`, `local_ip`.

### Overlays (NETW-02)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Create | POST | `/overlay/` | `{"cluster_id": int, "name": str}` | `{msg}` |
| List | GET | `/overlay/` | -- | `[{overlay_id, name, internal_subnet, cluster_id, instances: [...]}]` |
| Delete | DELETE | `/overlay/` | `{"overlay_id": int}` or `{"overlay_name": str}` (body) | `{msg}` |

**Key pitfall:** Create returns only `{msg}` -- no overlay ID. Must use create-then-read pattern (list overlays, find by name). Delete uses `DeleteWithBody` and supports either ID or name. The overlay list returns `instances` array showing which instances are joined.

### Cluster Membership (NETW-03)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Join | PUT | `/cluster/` | `{"cluster_id": int, "machine_ids": [int, ...]}` | `{msg}` |
| Remove | DELETE | `/cluster/remove_machine/` | `{"cluster_id": int, "machine_id": int, "new_manager_id": int?}` | `{msg}` |
| Read | GET | `/clusters/` | -- (check nodes list) | Via cluster list response |

**Key pitfall:** Join accepts multiple machine_ids at once, but Terraform should model as one-machine-per-resource for proper lifecycle management. Remove has an optional `new_manager_id` -- if removing the manager node, must specify which node takes over. This needs special handling.

### Overlay Membership (NETW-04)
| Operation | Method | Endpoint | Request Body | Response |
|-----------|--------|----------|-------------|----------|
| Join | PUT | `/overlay/` | `{"name": str, "instance_id": int}` | `{msg}` |
| Read | GET | `/overlay/` | -- (check instances list) | Via overlay list response |
| (No remove) | -- | -- | -- | No individual instance removal found |

**Key pitfall:** No explicit remove-instance-from-overlay endpoint found in the SDK. The overlay DELETE removes the entire overlay with all instances. Individual instance removal may not be supported by the API. The resource destroy may need to be a no-op (with documentation), or we may need to investigate if there's an unlisted endpoint. This is a **significant limitation** for the `vastai_overlay_member` resource pattern.

### User Profile (DATA-07)
| Operation | Method | Endpoint | Response |
|-----------|--------|----------|----------|
| Read | GET | `/users/current?owner=me` | `{id, username, email, balance, credit, fullname, email_verified, has_billing, has_payout, ssh_key, billaddress_*, ...}` |

Response explicitly excludes `api_key` (popped by SDK). Fields from `user_fields` tuple: balance, balance_threshold, balance_threshold_enabled, billaddress_city, billaddress_country, billaddress_line1, billaddress_line2, billaddress_zip, billed_expected, billed_verified, billing_creditonly, can_pay, credit, email, email_verified, fullname, got_signup_credit, has_billing, has_payout, id, last4, paid_expected, paid_verified, password_resettable, paypal_email, ssh_key, user, username.

### Invoices (DATA-10)
| Operation | Method | Endpoint | Query Params | Response |
|-----------|--------|----------|-------------|----------|
| Read (v1) | GET | `/api/v1/invoices/` | `select_filters={date_col: {gte: ts, lte: ts}}, limit, after_token, latest_first` | `{results: [...], count, total, next_token}` |
| Read (deprecated) | GET | `/users/me/invoices` | `owner=me, sdate, edate, inc_charges` | `{invoices: [...], current: ...}` |

**Key pitfall:** The v1 endpoint uses `/api/v1/` prefix, but the Go client's `newRequest` hardcodes `/api/v0`. Need to add a `GetFullPath` method that takes the complete URL path (including version prefix) without prepending `/api/v0`. The Python SDK handles this with a regex check: `if not re.match(r"^/api/v(\d)+/", subpath): subpath = "/api/v0" + subpath`.

### Audit Logs (DATA-11)
| Operation | Method | Endpoint | Response |
|-----------|--------|----------|----------|
| Read | GET | `/audit_logs/` | `[{ip_address, api_key_id, created_at, api_route, args}]` |

Simple list endpoint with no filtering parameters.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Permissions JSON validation | Custom struct parser for each permission category | Store as `jsontypes.Normalized` or string with JSON validator | Permissions schema is nested, evolving, has constraints -- modeling as a rigid Go struct will break |
| Composite ID parsing | String splitting with manual error handling | Reusable `parseCompositeID(id, parts int) ([]string, error)` helper | Used by cluster_member and overlay_member; DRY |
| API version routing | Conditional prefix logic scattered in each service | Single `GetFullPath` client method | Only invoices v1 needs it now, but clean abstraction for future |
| Query parameter encoding | Manual string concatenation in paths | Use `url.Values` and `req.URL.RawQuery` | Proper URL encoding, no injection bugs |

**Key insight:** The permissions model is the most complex schema element in this phase. The API uses a deeply nested JSON object (`{"api": {"instance_read": {...}, "instance_write": {...}, ...}}`). Rather than modeling every permission as a Terraform attribute (which would be brittle and break when Vast.ai adds new categories), store the entire permissions blob as a JSON string attribute. Use `terraform-plugin-framework`'s `jsontypes.Normalized` type if available, or a plain string with a JSON validity validator.

## Common Pitfalls

### Pitfall 1: API Key Value Only Available on Create
**What goes wrong:** API key's actual key value is only returned in the create response. Subsequent reads (list) do NOT include the key value. If you don't store it on create, it's lost forever.
**Why it happens:** Security design -- keys are write-once, read-never from the API.
**How to avoid:** In the Create method, extract the key value from the create response and set it in state. Mark as `Sensitive: true`, `Computed: true` with `UseStateForUnknown` so reads don't overwrite it.
**Warning signs:** Key shows as empty/null after terraform refresh.

### Pitfall 2: Environment Variable Keyed by Name, Not ID
**What goes wrong:** Env vars use the "key" (name) as their identifier. There is no numeric ID. The list endpoint returns a map `{"secrets": {"name": "value"}}`, not an array.
**Why it happens:** The API was designed as a simple key-value store, not a CRUD resource.
**How to avoid:** Use the env var name as the Terraform resource ID. For read, list all env vars and filter by name. For delete, use `DeleteWithBody` with the key name.
**Warning signs:** Cannot find env var by ID, delete fails with path-based endpoint.

### Pitfall 3: Team Role Name vs ID Inconsistency
**What goes wrong:** Team role read (`GET /team/roles/{name}/`) and delete (`DELETE /team/roles/{name}/`) use the role NAME in the path. But update (`PUT /team/roles/{id}/`) uses the role's numeric ID. Using the wrong identifier for the wrong operation causes 404 errors.
**Why it happens:** Inconsistent API design -- likely different developers built different endpoints.
**How to avoid:** Track BOTH name and ID in state. Use name for read/delete, ID for update. Import should resolve both from the list endpoint.
**Warning signs:** 404 on update (using name instead of ID) or 404 on read (using ID instead of name).

### Pitfall 4: Cluster/Overlay Create Returns No ID
**What goes wrong:** Creating a cluster or overlay returns only `{"msg": "..."}` -- no ID in the response. Cannot populate state directly from create response.
**Why it happens:** API returns a success message, not the created object.
**How to avoid:** Use the create-then-read pattern established in Phase 3 (volumes) and Phase 4 (endpoints). After create, immediately list all clusters/overlays and find the new one by matching subnet (cluster) or name (overlay).
**Warning signs:** Resource created but ID is empty in state.

### Pitfall 5: Invite Uses Query Params, Not JSON Body
**What goes wrong:** Team member invite sends email and role as URL query parameters (`POST /team/invite/?email=X&role=Y`), not as a JSON request body. Using the standard `Post(ctx, path, body, result)` with a JSON body will fail silently.
**Why it happens:** Inconsistent API design -- most endpoints use JSON body, but invite uses query params.
**How to avoid:** Encode the email and role directly into the URL path (e.g., `"/team/invite/?email=X&role=Y"`). The existing client already supports inline query params in paths (see volumes: `"/volumes?owner=me&type=local_volume"`).
**Warning signs:** Invite returns 200 but no member is actually invited (params ignored).

### Pitfall 6: Invoices V1 Uses Different API Version Prefix
**What goes wrong:** The invoices v1 endpoint is at `/api/v1/invoices/`, but the Go client's `newRequest` method hardcodes `/api/v0` prefix on all paths. Passing `/api/v1/invoices/` as the path would result in `/api/v0/api/v1/invoices/`.
**Why it happens:** Client was built for v0 API; invoices v1 is the only known v1 endpoint.
**How to avoid:** Add a new `GetFullPath(ctx, fullPath, result)` method to VastAIClient that constructs the URL as `baseURL + fullPath` without the `/api/v0` prefix. Alternatively, add a `newRequestFullPath` internal method.
**Warning signs:** 404 on invoices endpoint, doubled path prefix in logs.

### Pitfall 7: Cluster Delete Uses DeleteWithBody
**What goes wrong:** Cluster and overlay delete send the ID in the request body, not the URL path. Using the standard `Delete(ctx, path, result)` sends no body. The existing `DeleteWithBody` method handles this.
**Why it happens:** Consistent with template delete pattern (already handled in Phase 2).
**How to avoid:** Use `client.DeleteWithBody(ctx, "/cluster/", body, &resp)` pattern.
**Warning signs:** Delete returns 400 "missing cluster_id" error.

### Pitfall 8: No Subaccount Delete Endpoint
**What goes wrong:** The API has no endpoint for deleting subaccounts. Terraform expects resources to be destroyable.
**Why it happens:** Subaccounts may be a permanent relationship in Vast.ai's model.
**How to avoid:** Implement destroy as a no-op that removes from state but does not call the API. Add a warning diagnostic: "Subaccount was removed from Terraform state but NOT deleted from Vast.ai." Document this in the resource description.
**Warning signs:** Destroy operation fails with 404 or method-not-allowed.

### Pitfall 9: No Overlay Instance Removal Endpoint
**What goes wrong:** There is no SDK function to remove a single instance from an overlay. The only way to disassociate is to delete the entire overlay.
**Why it happens:** The overlay model may not support partial membership changes.
**How to avoid:** If `vastai_overlay_member` resource destroy is a no-op (with warning), or if we can confirm there IS an unlisted removal endpoint. Consider making overlay_member create-only with a lifecycle note. Alternative: skip the separate overlay_member resource and document that overlay membership is managed at instance creation time.
**Warning signs:** Resource destroy appears to succeed but instance is still in overlay.

## Code Examples

### Client Service Pattern (API Keys)
```go
// Source: Pattern from internal/client/ssh_keys.go adapted for API keys

// ApiKeyService handles API key-related API operations.
type ApiKeyService struct {
    client *VastAIClient
}

// ApiKey represents an API key from the Vast.ai API.
type ApiKey struct {
    ID          int             `json:"id"`
    Name        string          `json:"name"`
    Key         string          `json:"key,omitempty"` // Only present on create response
    Permissions json.RawMessage `json:"permissions"`
    KeyParams   string          `json:"key_params,omitempty"`
    CreatedAt   string          `json:"created_at,omitempty"`
}

// Create creates a new API key.
// POST /auth/apikeys/ with {"name": name, "permissions": perms}
func (s *ApiKeyService) Create(ctx context.Context, name string, permissions json.RawMessage, keyParams string) (*ApiKey, error) {
    body := map[string]interface{}{
        "name":        name,
        "permissions": permissions,
    }
    if keyParams != "" {
        body["key_params"] = keyParams
    }
    var resp ApiKey
    if err := s.client.Post(ctx, "/auth/apikeys/", body, &resp); err != nil {
        return nil, fmt.Errorf("creating API key: %w", err)
    }
    return &resp, nil
}

// List retrieves all API keys.
// GET /auth/apikeys/
func (s *ApiKeyService) List(ctx context.Context) ([]ApiKey, error) {
    var resp []ApiKey
    if err := s.client.Get(ctx, "/auth/apikeys/", &resp); err != nil {
        return nil, fmt.Errorf("listing API keys: %w", err)
    }
    return resp, nil
}

// Delete deletes an API key by ID.
// DELETE /auth/apikeys/{id}/
func (s *ApiKeyService) Delete(ctx context.Context, id int) error {
    path := fmt.Sprintf("/auth/apikeys/%d/", id)
    if err := s.client.Delete(ctx, path, nil); err != nil {
        return fmt.Errorf("deleting API key %d: %w", id, err)
    }
    return nil
}
```

### Client Enhancement: GetFullPath Method
```go
// Source: Needed for invoices v1 endpoint (Pitfall 6)

// GetFullPath sends a GET request using the full URL path (no /api/v0 prefix).
// Used for API endpoints that don't follow the /api/v0/ convention (e.g., /api/v1/invoices/).
func (c *VastAIClient) GetFullPath(ctx context.Context, fullPath string, result interface{}) error {
    req, err := c.newRequestFullPath(ctx, http.MethodGet, fullPath, nil)
    if err != nil {
        return err
    }
    return c.do(ctx, req, result)
}

// newRequestFullPath creates a request with the exact path (no /api/v0 prefix).
func (c *VastAIClient) newRequestFullPath(ctx context.Context, method, fullPath string, body interface{}) (*retryablehttp.Request, error) {
    url := c.baseURL + fullPath
    // ... same body encoding and header logic as newRequest
}
```

### Composite ID Helper
```go
// Source: Pattern for association resources

// parseCompositeID splits a composite ID "part1/part2" into components.
func parseCompositeID(id string, expectedParts int) ([]string, error) {
    parts := strings.Split(id, "/")
    if len(parts) != expectedParts {
        return nil, fmt.Errorf("expected %d parts in composite ID %q, got %d", expectedParts, id, len(parts))
    }
    return parts, nil
}

// formatCompositeID creates a composite ID from parts.
func formatCompositeID(parts ...string) string {
    return strings.Join(parts, "/")
}
```

### Environment Variable Resource Pattern (Name-Keyed)
```go
// Source: Pattern adapted from SDK analysis

// EnvVarService handles environment variable operations.
type EnvVarService struct {
    client *VastAIClient
}

// EnvVarMap is the response from GET /secrets/.
type EnvVarMap struct {
    Secrets map[string]string `json:"secrets"`
}

// Create creates a new env var. POST /secrets/ with {"key": name, "value": value}
func (s *EnvVarService) Create(ctx context.Context, key, value string) error {
    body := map[string]string{"key": key, "value": value}
    var resp map[string]interface{}
    return s.client.Post(ctx, "/secrets/", body, &resp)
}

// List retrieves all env vars. GET /secrets/
func (s *EnvVarService) List(ctx context.Context) (map[string]string, error) {
    var resp EnvVarMap
    if err := s.client.Get(ctx, "/secrets/", &resp); err != nil {
        return nil, err
    }
    return resp.Secrets, nil
}

// Update updates an env var. PUT /secrets/ with {"key": name, "value": value}
func (s *EnvVarService) Update(ctx context.Context, key, value string) error {
    body := map[string]string{"key": key, "value": value}
    var resp map[string]interface{}
    return s.client.Put(ctx, "/secrets/", body, &resp)
}

// Delete deletes an env var. DELETE /secrets/ with body {"key": name}
func (s *EnvVarService) Delete(ctx context.Context, key string) error {
    body := map[string]string{"key": key}
    return s.client.DeleteWithBody(ctx, "/secrets/", body, nil)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `show invoices` (v0) | `show invoices-v1` with pagination | 2025 | v1 endpoint adds pagination, type filtering, proper date handling |
| Inline cluster membership | Separate membership resources | Terraform convention | Better lifecycle management, independent changes |

**Deprecated/outdated:**
- `show__invoices` (v0): Marked `(DEPRECATED)` in the SDK. Use `show__invoices_v1` instead. The v0 endpoint at `/users/me/invoices` still works but lacks pagination.

## Open Questions

1. **Subaccount delete behavior**
   - What we know: No delete endpoint found in the Python SDK.
   - What's unclear: Whether the API supports subaccount deletion at all, or if it's permanent.
   - Recommendation: Implement destroy as state-removal-only with warning diagnostic. Document the limitation.

2. **Overlay instance removal**
   - What we know: `join__overlay` adds instances, `delete__overlay` removes the entire overlay. No `remove_instance_from_overlay` function exists in the SDK.
   - What's unclear: Whether the API supports removing a single instance from an overlay without destroying the overlay.
   - Recommendation: Two options: (a) Make `vastai_overlay_member` destroy a no-op with warning, or (b) skip the separate overlay_member resource entirely and document that overlay membership is managed through overlay lifecycle. Recommend option (b) -- skip `vastai_overlay_member` as a separate resource and instead include `instance_ids` as a computed read-only attribute on `vastai_overlay` that shows current membership. Joining is done at instance creation time, not as a separate Terraform resource.

3. **Team context for team operations**
   - What we know: `destroy__team` is a parameterless DELETE to `/team/` -- it destroys the team associated with the current API key.
   - What's unclear: How multiple teams work with a single API key. The create response likely includes a team ID.
   - Recommendation: Track team ID in state from create response. For read, may need to infer from team member or role list responses.

4. **API key create response schema**
   - What we know: `create__api_key` prints `r.json()` -- the full response shape is not documented in the SDK.
   - What's unclear: Exact fields returned (especially whether `id` and `key` are both present).
   - Recommendation: HIGH confidence that `id` and `key` are returned based on the delete endpoint accepting `id` and the purpose of the create call. Verify during implementation with a live API test.

5. **Permissions JSON schema validation**
   - What we know: Permissions are a nested JSON object with categories like `instance_read`, `instance_write`, etc., with optional constraints.
   - What's unclear: Complete list of valid permission keys, whether invalid keys cause errors or are silently ignored.
   - Recommendation: Accept any valid JSON for permissions. Use D-02's approach (permissions as a string set was the initial proposal) but note the actual API uses a nested object, not a flat set. Store as JSON string.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | All code | Yes | 1.25.7 | -- |
| terraform-plugin-framework | All resources | Yes | v1.19.0 (in go.mod) | -- |
| terraform-plugin-testing | Acceptance tests | Yes | v1.15.0 (in go.mod) | -- |

No new external dependencies required.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + terraform-plugin-testing v1.15.0 |
| Config file | go.mod (module dependencies) |
| Quick run command | `go test ./internal/... -count=1 -timeout 60s` |
| Full suite command | `go test ./... -count=1 -timeout 120s` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ACCT-01 | API key CRUD + sensitive key + permissions | unit (schema + client) | `go test ./internal/services/apikey/... ./internal/client/... -run TestApiKey -count=1` | Wave 0 |
| ACCT-02 | Env var CRUD + sensitive value + name-keyed | unit (schema + client) | `go test ./internal/services/envvar/... ./internal/client/... -run TestEnvVar -count=1` | Wave 0 |
| ACCT-03 | Team CRUD | unit (schema + client) | `go test ./internal/services/team/... ./internal/client/... -run TestTeam -count=1` | Wave 0 |
| ACCT-04 | Team role CRUD + permissions + name/ID handling | unit (schema + client) | `go test ./internal/services/teamrole/... ./internal/client/... -run TestTeamRole -count=1` | Wave 0 |
| ACCT-05 | Team member invite/remove + query param invite | unit (schema + client) | `go test ./internal/services/teammember/... ./internal/client/... -run TestTeamMember -count=1` | Wave 0 |
| ACCT-06 | Subaccount create + no-delete + sensitive password | unit (schema + client) | `go test ./internal/services/subaccount/... ./internal/client/... -run TestSubaccount -count=1` | Wave 0 |
| NETW-01 | Cluster CRUD + create-then-read + DeleteWithBody | unit (schema + client) | `go test ./internal/services/cluster/... ./internal/client/... -run TestCluster -count=1` | Wave 0 |
| NETW-02 | Overlay CRUD + create-then-read + DeleteWithBody | unit (schema + client) | `go test ./internal/services/overlay/... ./internal/client/... -run TestOverlay -count=1` | Wave 0 |
| NETW-03 | Cluster membership join/remove + composite ID | unit (schema + client) | `go test ./internal/services/clustermember/... -run TestClusterMember -count=1` | Wave 0 |
| NETW-04 | Overlay membership join + read via list | unit (schema + client) | `go test ./internal/services/overlaymember/... -run TestOverlayMember -count=1` | Wave 0 |
| DATA-07 | User profile data source read | unit (schema + client) | `go test ./internal/services/user/... ./internal/client/... -run TestUser -count=1` | Wave 0 |
| DATA-10 | Invoices data source + v1 endpoint + pagination | unit (schema + client) | `go test ./internal/services/invoice/... ./internal/client/... -run TestInvoice -count=1` | Wave 0 |
| DATA-11 | Audit logs data source read | unit (schema + client) | `go test ./internal/services/auditlog/... ./internal/client/... -run TestAuditLog -count=1` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -count=1 -timeout 60s`
- **Per wave merge:** `go test ./... -count=1 -timeout 120s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- All test files listed above are Wave 0 (none exist yet)
- No new test framework or fixture configuration needed -- existing patterns cover everything
- Existing test count: 193 tests, all passing

## Project Constraints (from CLAUDE.md)

- **Language:** Go (required by Terraform provider ecosystem)
- **Framework:** Terraform Plugin Framework (HashiCorp's current recommendation)
- **API Reference:** Python SDK at `vast-sdk/` is the source of truth for API behavior
- **Auth:** Bearer header only, never query parameter
- **Testing:** Acceptance tests require real Vast.ai account; unit tests use httptest mocks

## Sources

### Primary (HIGH confidence)
- `vast-sdk/vastai/vast.py` lines 2016-2028 -- API key create endpoint
- `vast-sdk/vastai/vast.py` lines 2859-2863 -- API key delete endpoint
- `vast-sdk/vastai/vast.py` lines 4906-4913 -- API key list endpoint
- `vast-sdk/vastai/vast.py` lines 2066-2077 -- Env var create endpoint
- `vast-sdk/vastai/vast.py` lines 2976-2987 -- Env var delete endpoint
- `vast-sdk/vastai/vast.py` lines 5185-5214 -- Env var list endpoint
- `vast-sdk/vastai/vast.py` lines 6731-6742 -- Env var update endpoint
- `vast-sdk/vastai/vast.py` lines 2643-2647 -- Team create endpoint
- `vast-sdk/vastai/vast.py` lines 3132-3136 -- Team destroy endpoint
- `vast-sdk/vastai/vast.py` lines 2659-2664 -- Team role create endpoint
- `vast-sdk/vastai/vast.py` lines 5810-5818 -- Team roles list endpoint
- `vast-sdk/vastai/vast.py` lines 5800-5804 -- Team role read (by name)
- `vast-sdk/vastai/vast.py` lines 6804-6812 -- Team role update (by ID)
- `vast-sdk/vastai/vast.py` lines 3786-3790 -- Team role delete (by name)
- `vast-sdk/vastai/vast.py` lines 3308-3316 -- Team member invite endpoint
- `vast-sdk/vastai/vast.py` lines 5785-5793 -- Team member list endpoint
- `vast-sdk/vastai/vast.py` lines 3775-3779 -- Team member remove endpoint
- `vast-sdk/vastai/vast.py` lines 2578-2612 -- Subaccount create endpoint
- `vast-sdk/vastai/vast.py` lines 5765-5779 -- Subaccount list endpoint
- `vast-sdk/vastai/vast.py` lines 2039-2058 -- Cluster create endpoint
- `vast-sdk/vastai/vast.py` lines 2896-2911 -- Cluster delete endpoint
- `vast-sdk/vastai/vast.py` lines 5701-5727 -- Cluster list endpoint
- `vast-sdk/vastai/vast.py` lines 3328-3344 -- Cluster join endpoint
- `vast-sdk/vastai/vast.py` lines 5885-5903 -- Cluster remove machine endpoint
- `vast-sdk/vastai/vast.py` lines 2836-2852 -- Overlay create endpoint
- `vast-sdk/vastai/vast.py` lines 2994-3016 -- Overlay delete endpoint
- `vast-sdk/vastai/vast.py` lines 5737-5755 -- Overlay list endpoint
- `vast-sdk/vastai/vast.py` lines 3355-3371 -- Overlay join endpoint
- `vast-sdk/vastai/vast.py` lines 5354-5462 -- Invoices v1 endpoint
- `vast-sdk/vastai/vast.py` lines 4920-4934 -- Audit logs endpoint
- `vast-sdk/vastai/vast.py` lines 5828-5844 -- User profile endpoint
- `vast-sdk/vastai/vast.py` lines 575-591 -- apiurl function (v0/v1 prefix logic)
- `internal/client/auth.go` line 16 -- hardcoded `/api/v0` prefix
- `internal/client/client.go` -- VastAIClient service pattern, DeleteWithBody method
- `internal/client/ssh_keys.go` -- canonical client service pattern
- `internal/services/sshkey/resource_ssh_key.go` -- canonical resource pattern
- `internal/services/sshkey/models.go` -- canonical model pattern
- `internal/services/endpoint/models.go` -- canonical data source model pattern

### Secondary (MEDIUM confidence)
- [Vast.ai Permissions Documentation](https://docs.vast.ai/api-reference/permissions-and-authorization) -- permission categories and JSON structure
- [Vast.ai API Permissions](https://docs.vast.ai/api/permissions-and-authorization) -- team role management details

### Tertiary (LOW confidence)
- Subaccount delete behavior -- no evidence found for or against; assumed no-delete based on SDK absence
- Overlay instance removal -- no endpoint found; assumed one-way based on SDK absence
- API key create response schema -- inferred from SDK print statement and delete endpoint signature

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all patterns established in prior phases
- Architecture: HIGH -- follows established service-per-directory and client service patterns exactly
- API contracts: HIGH -- extracted directly from Python SDK source code
- Pitfalls: HIGH -- identified 9 specific pitfalls from SDK analysis and client code review
- Discretion recommendations: MEDIUM -- recommendations based on API analysis but not validated with live API

**Research date:** 2026-03-27
**Valid until:** 2026-04-27 (stable -- patterns are established, API unlikely to change)
