# Feature Research

**Domain:** Terraform/OpenTofu Provider for GPU Cloud Marketplace (Vast.ai)
**Researched:** 2026-03-25
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete or unpublishable.

#### Registry Publishing Requirements

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| GPG-signed releases via GoReleaser | Required by Terraform Registry for publication | MEDIUM | RSA/DSA keys only (no ECC). GitHub Actions workflow + `.goreleaser.yml` from scaffolding repo. Secrets: `GPG_PRIVATE_KEY`, `PASSPHRASE` |
| Semantic versioning with `v` prefix | Registry rejects non-semver tags | LOW | `v0.1.0`, `v1.0.0` etc. Never modify released versions (causes checksum errors) |
| `terraform-registry-manifest.json` | Required release artifact declaring protocol version | LOW | `{"version": 1, "metadata": {"protocol_versions": ["6.0"]}}` for Plugin Framework |
| SHA256SUMS + .sig files | Registry verification of binary integrity | LOW | GoReleaser generates these automatically |
| Repository named `terraform-provider-vastai` | Registry naming convention, public GitHub repo required | LOW | Lowercase only |
| Generated documentation via tfplugindocs | Registry renders docs from `docs/` directory | MEDIUM | `docs/index.md` + `docs/resources/*.md` + `docs/data-sources/*.md`. Templates in `templates/`, examples in `examples/` |
| Provider index documentation | Registry requires overview page | LOW | `docs/index.md` with auth configuration, provider block example |

#### Core Provider Functionality

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| API key authentication via `VASTAI_API_KEY` env var | Standard Terraform pattern; every cloud provider does this | LOW | Also support `api_key` in provider block. Mark as sensitive. Follow `VASTAI_API_KEY` convention from Python SDK |
| Configurable API endpoint URL | Needed for testing, staging environments | LOW | Default `https://console.vast.ai`, overridable via provider config or `VASTAI_API_URL` env var |
| User-Agent header with provider version | API operators expect this for debugging and analytics | LOW | `terraform-provider-vastai/v0.1.0` format |

#### Resource Lifecycle (CRUD)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Instance resource (create/read/update/delete) | Core compute primitive -- the most important resource | HIGH | Create from offer ID, read status, update label/template, destroy. Must handle spot vs on-demand. `start`/`stop` as state attribute, not separate resources |
| Template resource (CRUD) | Templates define instance configurations; foundational resource | MEDIUM | Create with image, env, onstart_cmd, ssh/jupyter flags. Update mutable fields. Delete by hash or ID |
| SSH Key resource (CRUD) | Standard access management; every cloud provider has this | LOW | Create, read, update, delete. Simple key-value resource |
| API Key resource (CRUD) | Account management primitive | MEDIUM | Create with name and permissions. Sensitive value handling for the key itself |
| Environment Variable resource (CRUD) | Account-level config management | LOW | Simple name/value CRUD. Sensitive flag on value |
| Endpoint resource (CRUD) | Serverless is a core Vast.ai product | MEDIUM | Create with autoscaling params (min_load, target_util, cold_mult, cold_workers, max_workers). Update all params |
| Worker Group resource (CRUD) | Manages serverless worker scaling | HIGH | Create with template, search params, autoscaling config. Complex dependency on endpoints and templates |
| Volume resource (CRUD) | Storage is fundamental infrastructure | MEDIUM | Create from offer ID with size. Delete. List/unlist for marketplace |
| Network Volume resource (CRUD) | Networked storage variant | MEDIUM | Similar to volume but different API paths and listing behavior |
| Team resource (CRUD) | Multi-user account management | LOW | Create team, destroy team |
| Team Role resource (CRUD) | RBAC for teams | MEDIUM | Create with permissions, update, delete |
| Subaccount resource (CRUD) | Organizational account hierarchy | LOW | Create with email, username, type |

#### Data Sources

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| GPU offer search data source | THE key data source -- users need to find GPU instances to rent | HIGH | Must support Vast.ai query language (key operator value). Filter by GPU type, RAM, price, region, etc. Return structured offer list |
| Instance data source (singular) | Look up existing instance by ID | LOW | Read-only view of instance attributes |
| Instances data source (plural/list) | List all account instances | LOW | With optional filtering |
| Template search data source | Find templates by query | MEDIUM | Mirror SDK `search_templates` |
| Volume search data source | Find volume offers | MEDIUM | Mirror SDK `search_volumes` and `search_network_volumes` |
| User profile data source | Reference account info in configs | LOW | Read-only, from `show_user` |
| SSH Keys data source | List existing SSH keys | LOW | Read-only list |

#### Import Support

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `terraform import` for all managed resources | HashiCorp design principle: resources MUST support import for brownfield adoption | MEDIUM | Implement `ImportState` on every resource. Import by resource ID. Critical for users with existing Vast.ai infrastructure |
| Import documentation with examples | Registry shows import commands in resource docs | LOW | `examples/resources/<name>/import.sh` for each resource |

#### Schema Quality

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Attribute validators on all constrained fields | Catch config errors at `terraform validate` time, not API call time | MEDIUM | Use `terraform-plugin-framework-validators`: int ranges, string length, regex patterns, enum values (e.g., volume type, instance status). Provider-level validation for API key format |
| `Sensitive` flag on secrets | API keys, passwords must not appear in plan/state output | LOW | Mark `api_key`, env var values, SSH private keys |
| Required vs Optional vs Computed correctly set | Incorrect schema causes confusing plan diffs | MEDIUM | Match API behavior exactly. `Computed` for server-set fields (ID, status, IP). `Required` for create-time mandatory fields. `Optional + Computed` for server-defaulted fields |
| Meaningful attribute descriptions | Rendered in registry docs; undescribed attributes are useless | MEDIUM | Every attribute needs a description. Copy from API docs where available |

#### Error Handling

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Structured diagnostics with summary + detail | Framework standard; raw errors confuse users | MEDIUM | `diags.AddError("Failed to create instance", fmt.Sprintf("API returned %d: %s", status, body))`. Never expose raw HTTP errors without context |
| HTTP retry with exponential backoff | APIs have rate limits and transient failures; without retry, `terraform apply` fails unpredictably | MEDIUM | Retry on 429 (rate limit) and 5xx. Configurable max retries (default 3, matching Python SDK). Exponential backoff with jitter |

#### Testing

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Acceptance tests for all resources | HashiCorp expectation for registry providers; validates real API behavior | HIGH | Use `terraform-plugin-testing`. Basic create/read, update, import, and destroy tests for each resource. `resource.ParallelTest()` where safe |
| Unit tests for validators and plan modifiers | Fast feedback without API calls | LOW | Pure Go tests for custom logic |
| CI pipeline running tests on PR | Standard quality gate | MEDIUM | GitHub Actions matrix for Go versions. Acceptance tests on main branch (costs real API credits) |

#### Documentation

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Per-resource documentation with examples | Registry renders these; users expect usage examples | MEDIUM | `templates/resources/<name>.md.tmpl` with `{{.SchemaMarkdown}}` and example `.tf` files |
| Per-data-source documentation with examples | Same as resources | MEDIUM | `templates/data-sources/<name>.md.tmpl` |
| Provider configuration documentation | First thing users read | LOW | Auth methods, endpoint config, retry config |
| Working examples in `examples/` directory | Users copy-paste from examples | MEDIUM | At minimum: basic instance creation, template + instance workflow, serverless endpoint setup |

### Differentiators (Competitive Advantage)

Features that set this provider apart. The existing `aalekhpatel07/vastai` provider is v0.1.3 (Feb 2024, apparently abandoned), with only basic resource/data source support. Nearly everything below differentiates.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Full API coverage (all 25+ resource types) | Existing provider covers ~2 resources. Full coverage = only provider users need | HIGH | This is the primary differentiator. Every Vast.ai resource manageable through Terraform |
| GPU offer search with rich filtering | Vast.ai's query language is powerful but complex; exposing it as structured Terraform attributes makes it accessible | HIGH | Translate Vast.ai query DSL into typed Terraform attributes: `gpu_name`, `num_gpus`, `gpu_ram_min`, `price_max`, `region`, `datacenter_only`. Much better DX than raw query strings |
| Configurable timeouts per resource | Long-running GPU instance creation/deletion needs configurable waits | MEDIUM | Use `terraform-plugin-framework-timeouts`. Sensible defaults (create: 10m for instances, 5m for volumes). User-overridable |
| Plan modifiers for computed attributes | Reduces noisy `(known after apply)` in plans | MEDIUM | `UseStateForUnknown` for stable computed fields (ID, SSH host, created_at). `RequiresReplace` for immutable fields (offer_id on instances, image on templates) |
| Instance lifecycle management (start/stop/reboot) | Users want to stop instances without destroying them (cost savings) | MEDIUM | Model as `status` attribute (`running`, `stopped`). Changing status triggers start/stop API. Alternative: separate `vastai_instance_state` resource |
| Bid price management | Vast.ai is a marketplace; bid management is core to cost optimization | MEDIUM | `price` attribute on instance resource. `change_bid` on update. Plan modifier to detect price drift |
| Cluster and overlay networking | Advanced networking for multi-GPU workflows; no existing provider supports this | HIGH | `vastai_cluster`, `vastai_overlay` resources with membership management. Join/remove machines from clusters |
| Autoscaler resource with endpoint binding | Serverless autoscaling is a premium Vast.ai feature | HIGH | `vastai_autoscaler` resource bound to endpoints. Configure test_workers, gpu_ram, template, search params |
| Cross-resource references | Terraform's strength is connecting resources; this makes workflows declarative | MEDIUM | Template ID referenced by instance. Endpoint ID referenced by worker group. SSH key ID referenced by instance attachment |
| Ephemeral resource for API key | API keys are sensitive; ephemeral resources (Terraform 1.10+) prevent state persistence | MEDIUM | `ephemeral "vastai_api_key" "temp"` for short-lived API keys. Forward-looking feature few community providers implement yet |
| Identity-based import (Terraform 1.12+) | Import by name/attributes instead of opaque IDs | LOW | For resources with unique names (templates by hash, env vars by name). Implement `Identity` schema on resources |
| Drift detection and refresh accuracy | Detect when resources change outside Terraform (e.g., instance terminated by marketplace) | MEDIUM | Read function accurately refreshes all state. Handle "resource gone" gracefully (remove from state, not error) |
| `most_affordable` computed attribute on offer data source | Users almost always want the cheapest matching offer | LOW | Return offers sorted by price; expose `most_affordable` as convenience attribute |
| Guide documentation | Beyond reference docs: "How to deploy a GPU training job", "Serverless inference endpoint setup" | MEDIUM | `docs/guides/` directory with workflow-oriented docs. Sets apart from bare-minimum providers |
| Terraform Cloud / HCP Terraform compatibility | Enterprise users expect remote execution support | LOW | Mostly automatic with Plugin Framework, but needs testing. Document any env var requirements |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems or violate Terraform's declarative model.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Instance command execution resource | Users want to run setup commands on GPU instances | Imperative operation; violates Terraform's declarative model. Creates state management nightmares (what's the "state" of a command?). Terraform has provisioners for this | Use `remote-exec` provisioner, `null_resource` with provisioner, or external tooling (Ansible, SSH scripts). Document this pattern in guides |
| File transfer resources (copy, cloud_copy, sync) | Users want to upload training data/models to instances | Imperative, non-idempotent operations. No clear "state" to manage. Copy is not a resource -- it's an action | Use provisioners (`file` provisioner), external tooling, or Vast.ai's native cloud copy CLI. Document in guides |
| Bulk instance creation resource | Vast.ai SDK has `create_instances` for multiple at once | Terraform's resource model is per-instance. Bulk creation breaks count/for_each, makes individual management impossible | Use `count` or `for_each` meta-arguments on the standard instance resource. Terraform parallelizes automatically |
| Invoice/billing data sources with write operations | Users want billing visibility in Terraform | Billing is read-only and not infrastructure. Invoices don't represent managed state. Including financial operations (transfer_credit) in IaC is dangerous | Data source for invoices is fine (read-only). Never expose financial write operations as Terraform resources |
| 2FA management resources | Might seem like full API coverage | Security-critical interactive flow. TOTP codes are time-sensitive. Breaking 2FA via automation = security incident | Keep out of scope. 2FA is managed through Vast.ai console interactively |
| Machine host-side operations (list_machine, unlist_machine, schedule_maint, set_defjob) | Host operators might want IaC for their machines | These are provider-side (host) operations, not tenant (consumer) operations. Wrong audience for this provider | Build a separate `terraform-provider-vastai-host` if host-side IaC is needed. Keep tenant provider focused |
| Snapshot resource | Taking container snapshots seems IaC-worthy | Snapshots are point-in-time imperative actions, not declarative resources. "Desired state = snapshot exists" doesn't map well | Expose as a data source (read existing snapshots) or document as CLI workflow. Consider Terraform actions (1.14+) in the future |
| Raw query string pass-through in data sources | Power users want the full Vast.ai query DSL | Bypasses Terraform's type system and validation. Users get no plan-time feedback on typos. Makes provider docs useless for understanding filters | Expose structured attributes for common filters AND a `raw_query` escape hatch clearly marked as advanced/unsupported |
| Automatic instance replacement on spot eviction | Marketplace instances can be reclaimed; auto-replace sounds useful | Creates infinite loops and unexpected costs. Terraform's model is "desired state" not "keep retrying". Race conditions with plan/apply cycles | Document spot eviction handling in guides. Use external monitoring + `terraform apply` trigger. Consider `lifecycle { create_before_destroy = true }` pattern |
| Scheduled job resource | SDK has `show_scheduled_jobs` and `delete_scheduled_job` | Cron-like scheduling is better handled by external schedulers (CloudWatch Events, cron, Kubernetes CronJobs). Terraform isn't a scheduler | Data source to read scheduled jobs is acceptable. Actual scheduling belongs in external tooling |

## Feature Dependencies

```
[Provider Auth & HTTP Client]
    |
    +---> [Instance Resource]
    |         |
    |         +--requires--> [GPU Offer Search Data Source] (to find offer_id)
    |         +--requires--> [Template Resource] (optional, for template_hash)
    |         +--requires--> [SSH Key Resource] (optional, for attach_ssh)
    |         +--requires--> [Volume Resource] (optional, for link_volume)
    |
    +---> [Template Resource]
    |         (standalone, no hard dependencies)
    |
    +---> [SSH Key Resource]
    |         (standalone)
    |
    +---> [Endpoint Resource]
    |         |
    |         +---> [Worker Group Resource] (requires endpoint_id)
    |         |         +--requires--> [Template Resource] (template_hash)
    |         |         +--requires--> [GPU Offer Search Data Source] (search_params)
    |         |
    |         +---> [Autoscaler Resource] (requires endpoint_id)
    |                   +--requires--> [Template Resource]
    |
    +---> [Volume Resource]
    |         +--requires--> [Volume Offer Search Data Source] (to find offer)
    |
    +---> [Network Volume Resource]
    |         +--requires--> [Network Volume Offer Search Data Source]
    |
    +---> [Cluster Resource]
    |         +---> [Overlay Resource] (requires cluster_id)
    |
    +---> [Team Resource]
    |         +---> [Team Role Resource] (requires team)
    |         +---> [Team Member (invite)] (requires team + role)
    |         +---> [Subaccount Resource] (team context)
    |
    +---> [API Key Resource]
    |         (standalone)
    |
    +---> [Environment Variable Resource]
              (standalone)
```

### Dependency Notes

- **Instance requires GPU Offer Search:** You cannot create an instance without an offer ID. The offer data source is the entry point for compute provisioning.
- **Worker Group requires Endpoint:** Worker groups are bound to serverless endpoints. Endpoint must exist first.
- **Autoscaler requires Endpoint + Template:** Autoscaler needs both a target endpoint and a template defining what to launch.
- **Overlay requires Cluster:** Overlay networks are created on top of physical clusters.
- **Team Role requires Team:** Roles are scoped to teams.
- **Volume requires Volume Offer Search:** Similar to instances -- you need an offer to create a volume.
- **Import support depends on Read:** Every resource's import relies on a correct Read implementation. Get Read right first.

## MVP Definition

### Launch With (v0.1.0)

Minimum viable product for registry publication -- validates the provider works and gets real users.

- [ ] Provider configuration (API key auth, endpoint URL, User-Agent) -- foundation for everything
- [ ] Go HTTP client with retry logic -- required by all resources
- [ ] GPU offer search data source -- entry point for ALL compute workflows
- [ ] Instance resource (full CRUD + start/stop + import) -- core compute primitive
- [ ] Template resource (full CRUD + import) -- needed by instances
- [ ] SSH Key resource (full CRUD + import) -- needed for instance access
- [ ] Instance data source (singular by ID) -- basic lookup
- [ ] Attribute validators for required fields -- catch errors early
- [ ] Configurable timeouts on instance resource -- long-running operations
- [ ] Acceptance tests for all v0.1 resources -- quality gate
- [ ] Generated documentation with examples -- registry requirement
- [ ] GoReleaser + GitHub Actions release pipeline -- registry publication
- [ ] GPG signing -- registry requirement

### Add After Validation (v0.2.0 - v0.5.0)

Features to add once core is working and users validate the approach.

- [ ] Volume resource (CRUD + import) -- triggered by user demand for persistent storage
- [ ] Network Volume resource (CRUD + import) -- networked storage variant
- [ ] Endpoint resource (CRUD + import) -- unlocks serverless workflows
- [ ] Worker Group resource (CRUD + import) -- serverless scaling
- [ ] Autoscaler resource (CRUD + import) -- advanced serverless
- [ ] API Key resource (CRUD + import, sensitive handling) -- account management
- [ ] Environment Variable resource (CRUD + import) -- account config
- [ ] Volume/network volume offer search data sources -- mirror GPU offer search for storage
- [ ] Template search data source -- find existing templates
- [ ] Plan modifiers (UseStateForUnknown, RequiresReplace) -- improve plan output quality
- [ ] Bid price management on instances -- marketplace cost optimization
- [ ] Instance list data source -- bulk lookups
- [ ] User profile data source -- reference account info

### Future Consideration (v1.0+)

Features to defer until the provider is stable and has real users.

- [ ] Cluster resource -- advanced networking, complex lifecycle
- [ ] Overlay resource -- depends on cluster, limited audience
- [ ] Team resource + Team Role + member management -- organization features, needs careful RBAC design
- [ ] Subaccount resource -- organizational hierarchy
- [ ] Ephemeral resource for API keys -- requires Terraform 1.10+, cutting-edge
- [ ] Identity-based import -- requires Terraform 1.12+, cutting-edge
- [ ] Terraform actions for instance operations (reboot, recycle) -- requires Terraform 1.14+
- [ ] Audit log data source -- nice-to-have observability
- [ ] Invoice data source -- billing visibility
- [ ] Guide documentation (training workflows, serverless setup) -- user education
- [ ] State upgrade/migration infrastructure -- needed for breaking schema changes

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Provider auth + HTTP client | HIGH | MEDIUM | P1 |
| GPU offer search data source | HIGH | HIGH | P1 |
| Instance resource (full CRUD) | HIGH | HIGH | P1 |
| Template resource (CRUD) | HIGH | MEDIUM | P1 |
| SSH Key resource (CRUD) | HIGH | LOW | P1 |
| Acceptance tests for core resources | HIGH | HIGH | P1 |
| Generated docs + examples | HIGH | MEDIUM | P1 |
| Release pipeline (GoReleaser + Actions) | HIGH | MEDIUM | P1 |
| Import support for all resources | HIGH | MEDIUM | P1 |
| Attribute validators | MEDIUM | MEDIUM | P1 |
| Configurable timeouts | MEDIUM | LOW | P1 |
| Retry with exponential backoff | HIGH | MEDIUM | P1 |
| Volume resource | HIGH | MEDIUM | P2 |
| Network Volume resource | MEDIUM | MEDIUM | P2 |
| Endpoint resource | HIGH | MEDIUM | P2 |
| Worker Group resource | MEDIUM | HIGH | P2 |
| Autoscaler resource | MEDIUM | HIGH | P2 |
| API Key resource | MEDIUM | LOW | P2 |
| Environment Variable resource | MEDIUM | LOW | P2 |
| Plan modifiers | MEDIUM | MEDIUM | P2 |
| Bid price management | MEDIUM | LOW | P2 |
| Data sources (templates, volumes, user) | MEDIUM | LOW | P2 |
| Cluster resource | LOW | HIGH | P3 |
| Overlay resource | LOW | HIGH | P3 |
| Team/Role/Subaccount resources | LOW | MEDIUM | P3 |
| Ephemeral API key resource | LOW | MEDIUM | P3 |
| Identity-based import | LOW | LOW | P3 |
| Guide documentation | MEDIUM | MEDIUM | P3 |

**Priority key:**
- P1: Must have for launch (v0.1.0)
- P2: Should have, add in subsequent releases (v0.2-v0.5)
- P3: Nice to have, future consideration (v1.0+)

## Competitor Feature Analysis

| Feature | aalekhpatel07/vastai (v0.1.3) | vast-data/vastdata (different product) | Our Approach |
|---------|-------------------------------|----------------------------------------|--------------|
| Resources | ~1-2 basic resources | 50+ resources (VAST Data storage) | All 25+ Vast.ai resource types |
| Data sources | ~1 basic data source | 50+ data sources | Rich offer search + all list endpoints |
| Import support | Unknown/unlikely | Yes | Yes, all resources from day one |
| Documentation | Minimal, auto-generated | Comprehensive | Generated + guides + examples |
| Testing | Unknown | Yes | Full acceptance test suite |
| Maintenance | Abandoned (last update Feb 2024) | Actively maintained | Actively maintained |
| Plugin Framework | Unknown (may be SDKv2) | Plugin Framework | Plugin Framework (protocol 6.0) |
| Timeouts | No | Yes | Yes, configurable per resource |
| Validators | No | Yes | Yes, comprehensive |
| Plan modifiers | No | Yes | Yes (UseStateForUnknown, RequiresReplace) |
| Retry logic | No | Unknown | Yes, configurable exponential backoff |
| Sensitive handling | Unknown | Yes | Yes, ephemeral resource support planned |

## Sources

- [Terraform Registry Publishing Requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing) -- HIGH confidence
- [HashiCorp Provider Design Principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles) -- HIGH confidence
- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework) -- HIGH confidence
- [Terraform Plugin Framework Validation](https://developer.hashicorp.com/terraform/plugin/framework/validation) -- HIGH confidence
- [Terraform Plugin Framework Timeouts](https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts) -- HIGH confidence
- [Terraform Plugin Framework Plan Modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification) -- HIGH confidence
- [Terraform Plugin Framework State Upgrade](https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade) -- HIGH confidence
- [tfplugindocs Documentation Generator](https://github.com/hashicorp/terraform-plugin-docs) -- HIGH confidence
- [Terraform Testing Patterns](https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns) -- HIGH confidence
- [AWS Provider Retries and Waiters](https://hashicorp.github.io/terraform-provider-aws/retries-and-waiters/) -- HIGH confidence
- [AWS Provider Resource Filtering](https://hashicorp.github.io/terraform-provider-aws/resource-filtering/) -- HIGH confidence
- [AWS Provider Error Handling](https://hashicorp.github.io/terraform-provider-aws/error-handling/) -- HIGH confidence
- [Terraform Sensitive State Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices/sensitive-state) -- HIGH confidence
- [Terraform Ephemeral Resources](https://developer.hashicorp.com/terraform/plugin/framework/ephemeral-resources) -- HIGH confidence
- [Terraform Actions](https://developer.hashicorp.com/terraform/language/block/action) -- MEDIUM confidence (very new, Terraform 1.14)
- [Terraform Identity-Based Import](https://developer.hashicorp.com/terraform/language/block/import) -- MEDIUM confidence (Terraform 1.12)
- [HashiCorp Release Publishing Workflow](https://github.com/hashicorp/ghaction-terraform-provider-release) -- HIGH confidence
- [CoreWeave Terraform Provider](https://registry.terraform.io/providers/coreweave/coreweave/latest/docs) -- MEDIUM confidence (comparable GPU cloud provider)
- [Existing vastai provider (abandoned)](https://registry.terraform.io/providers/aalekhpatel07/vastai/latest/docs) -- HIGH confidence (confirmed v0.1.3, last update Feb 2024)
- [Vast.ai Python SDK](vast-sdk/) -- HIGH confidence (local reference implementation)

---
*Feature research for: Terraform/OpenTofu Provider for Vast.ai GPU Cloud Marketplace*
*Researched: 2026-03-25*
