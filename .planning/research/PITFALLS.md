# Pitfalls Research

**Domain:** Terraform/OpenTofu provider for GPU marketplace (Vast.ai)
**Researched:** 2026-03-25
**Confidence:** HIGH (verified against HashiCorp official docs, AWS provider contributor guides, Vast.ai API/SDK source)

## Critical Pitfalls

### Pitfall 1: API Key Leaked in Query Parameters and Terraform Logs

**What goes wrong:**
The Vast.ai API authenticates via `?api_key=<key>` as a URL query parameter. Terraform logs URLs at TRACE level during HTTP operations. Any error message that includes the request URL will expose the API key in plaintext to logs, CI output, and Terraform Cloud run logs. The Python SDK also uses this pattern (`apiurl()` function at line 575-619 of `vast.py`). Additionally, Terraform state files may store the provider configuration including the API key if it is set as a provider attribute rather than read from an environment variable.

**Why it happens:**
The Vast.ai API was designed for CLI usage where query-parameter auth is convenient. Terraform providers, however, operate in automated pipelines where logs are captured and shared. Developers copy the auth pattern directly from the Python SDK without considering the Terraform logging context.

**How to avoid:**
- Always send the API key as an `Authorization: Bearer <key>` header instead of a query parameter. The Python SDK already supports this pattern (line 628-629 of `vast.py`: `result["Authorization"] = "Bearer " + args.api_key`). Use the header-based auth exclusively in the Go client.
- Mark the `api_key` provider attribute as `Sensitive: true` in the schema so Terraform redacts it from plan/apply output.
- Prefer reading the API key from `VASTAI_API_KEY` environment variable rather than the provider block, to avoid storing it in state at all.
- Never include URLs with embedded credentials in error messages returned to Terraform diagnostics.

**Warning signs:**
- API key appears in `TF_LOG=TRACE` output during development.
- API key visible in CI/CD logs for acceptance tests.
- Error messages from the Go HTTP client contain the full request URL.

**Phase to address:**
Phase 1 (Go API Client) -- this must be designed correctly from the start, as every resource depends on the client.

---

### Pitfall 2: "Provider Produced Inconsistent Result After Apply" From Schema Misconfiguration

**What goes wrong:**
Terraform compares the planned state with the actual state after apply. If they differ, it raises the dreaded "Provider produced inconsistent result after apply" error. This happens constantly in new providers because developers misconfigure the `Optional`/`Computed` attribute flags. Common triggers:
- An attribute is `Optional` but the API returns a normalized or default value different from what was sent (e.g., the API lowercases a string, or fills in a default). Without `Computed: true`, Terraform flags the mismatch.
- An attribute is `Computed` only but appears in the config block -- Terraform rejects this during validation.
- The Create method sets state values that don't match what the Read method subsequently returns (e.g., because of eventual consistency or API normalization).

**Why it happens:**
The Terraform Plugin Framework is strict about data consistency. Any attribute whose value the API might modify must be marked `Computed: true`. Any attribute that accepts user input must be `Optional: true` or `Required: true`. The Vast.ai API returns many server-computed fields (timestamps, status, actual pricing, IDs) and normalizes user inputs (e.g., disk sizes, image names). Without careful schema design, mismatches are inevitable.

**How to avoid:**
- For every attribute, ask: "Can the API return a value different from what the user configured?" If yes, mark it `Optional: true, Computed: true`.
- For server-generated fields (id, created_at, actual_status, ssh_host, ssh_port), mark them `Computed: true` only.
- After Create, always call Read to populate state from the API response rather than trusting the request payload.
- Use `UseStateForUnknown()` plan modifier for computed attributes that do not change after creation (e.g., id, created_at).
- Use `RequiresReplace()` plan modifier for immutable fields (e.g., GPU type, machine_id on an instance).
- Test with `TF_LOG=TRACE` to catch data consistency warnings early -- the framework logs warnings before they become hard errors.

**Warning signs:**
- Acceptance tests fail intermittently with "inconsistent result after apply."
- `terraform plan` shows changes on attributes the user did not modify.
- Repeated `(known after apply)` in plan output for values that should be predictable.

**Phase to address:**
Phase 2 (First Resource + Data Source) -- establish the schema design patterns here and enforce them project-wide. Create a schema design checklist before writing any resource.

---

### Pitfall 3: No Retry/Waiter Logic for Eventually-Consistent Instance Lifecycle

**What goes wrong:**
When you create a Vast.ai instance via `PUT /api/v0/asks/{id}/`, the API returns `{'success': True, 'new_contract': 7835610}` immediately. But the instance is not yet running -- it must be provisioned, which takes seconds to minutes. If the provider immediately tries to read the instance (which it must, to populate state), the API may return incomplete data, a different status than expected, or even a 404 if the instance has not propagated. Terraform then either errors or stores incorrect state.

Similarly, after destroying an instance, the API may still return the instance record briefly. If Terraform's next plan refresh sees the "ghost" instance, it will believe it still exists.

**Why it happens:**
The Vast.ai API is a marketplace -- instance creation is an asynchronous process involving bid matching, host provisioning, and container startup. The Python SDK just prints "Started." and leaves the user to poll. Terraform providers must handle this gap between API acknowledgment and resource readiness.

**How to avoid:**
- Implement a waiter/poller pattern (modeled on the AWS provider's `retry.StateChangeConf`) that polls the instance status after creation until it reaches a target state (e.g., "running") or times out.
- In the Create method: call the API, then poll `GET /api/v0/instances/{id}/` until `actual_status` reaches the expected state.
- In the Delete method: call the API, then poll until the instance is gone or in a terminal state.
- Use exponential backoff with jitter in the poller to avoid hammering the API.
- Set reasonable timeouts (2-5 minutes for creation, 1-2 minutes for deletion) and make them configurable via `timeouts` block.
- For Read, if the instance has been preempted or destroyed externally, remove it from state (`resp.State.RemoveResource(ctx)`) rather than erroring.

**Warning signs:**
- Acceptance tests for instance creation are flaky, passing sometimes and failing others.
- `terraform apply` succeeds but `terraform plan` immediately shows drift.
- Error messages reference attributes being null or empty right after creation.

**Phase to address:**
Phase 1 (Go API Client) for the retry/backoff primitives. Phase 2 (Instance Resource) for the waiter implementation.

---

### Pitfall 4: Interruptible (Spot) Instances Vanish and Break State

**What goes wrong:**
Vast.ai interruptible instances can be stopped at any time when outbid. When this happens outside of Terraform's knowledge, the next `terraform plan` or `terraform apply` discovers the instance is gone or stopped. If the provider's Read function does not handle this gracefully, it returns an error instead of removing the resource from state. The user then faces a broken state that requires manual `terraform state rm`.

Even worse: if the provider treats a stopped interruptible instance the same as a running on-demand instance, it may try to "update" a stopped instance (which the API may reject), or show confusing plan diffs.

**Why it happens:**
Terraform assumes resources are durable -- you create them, they exist until you destroy them. GPU marketplace spot instances violate this assumption fundamentally. The Vast.ai API returns `actual_status` values like "running", "stopped", "exited", "loading" that represent states Terraform does not natively understand.

**How to avoid:**
- In the Read function, check `actual_status`. If the instance no longer exists (404) or is in a terminal state (destroyed), call `resp.State.RemoveResource(ctx)` to cleanly remove it from state. Terraform will then show it as needing recreation on the next plan.
- Document clearly that interruptible instances may be preempted and will appear as "destroyed" in the next plan. This is expected behavior, not a bug.
- Consider exposing `actual_status` as a computed attribute so users can see the current state.
- Consider adding a `lifecycle_policy` attribute or documenting the use of Terraform's `create_before_destroy` lifecycle for spot instances.
- Do NOT attempt to automatically recreate preempted instances in the provider -- that is orchestration logic, not provider logic.

**Warning signs:**
- Users report "resource not found" errors during plan/apply on interruptible instances.
- State file contains references to instances that no longer exist.
- Users open issues asking "why does Terraform want to recreate my instance every apply?"

**Phase to address:**
Phase 2 (Instance Resource) -- this must be handled from the first instance resource implementation.

---

### Pitfall 5: Creating Instances Requires an Offer ID, Not a Machine Spec

**What goes wrong:**
Developers model the instance resource like a traditional cloud VM: you specify GPU type, memory, disk, region, and the provider creates it. But Vast.ai works differently -- you must first search for available offers (ephemeral, time-limited listings from hosts), then create an instance from a specific offer ID. The offer may expire between search and creation. This fundamental mismatch between Terraform's declarative model ("I want a machine with these specs") and Vast.ai's marketplace model ("I want to rent this specific offer") causes the resource design to fail.

**Why it happens:**
Developers coming from AWS/GCP/Azure providers expect `instance_type = "p3.2xlarge"` to be sufficient. Vast.ai's model is more like bidding on eBay -- you select a specific listing. The Python SDK's `create_instance` function takes an offer `id` (line 2386: `def create_instance(id, args)`), not a specification.

**How to avoid:**
- Design the instance resource to take an `offer_id` (or `machine_id`) as a required attribute. This is the ID returned from the offers data source.
- Create a `vastai_offers` data source that searches for available offers with filters (GPU type, RAM, disk, price, etc.) and returns matching offer IDs.
- Document the two-step workflow: (1) use data source to find offers, (2) use resource to create instance from offer.
- Mark `offer_id` as `RequiresReplace()` -- you cannot change the underlying machine after creation.
- Handle the race condition where an offer expires between plan and apply by returning a clear error message ("Offer no longer available, re-run plan to find new offers").
- Consider caching offer data in the data source with a short TTL, or documenting that users should use `-refresh=true` before apply.

**Warning signs:**
- Users confused about why they cannot just specify "give me an A100."
- Frequent "offer not found" errors during apply after a plan showed valid offers.
- Resource design attempts to abstract away the offer concept, creating a leaky abstraction.

**Phase to address:**
Phase 2 (First Resource + Data Source) -- the offers data source and instance resource must be designed together as a pair.

---

### Pitfall 6: Acceptance Tests Create Real GPU Instances and Accumulate Costs

**What goes wrong:**
Terraform acceptance tests create real infrastructure. GPU instances on Vast.ai cost $0.10-$10+/hour depending on GPU type. A single test run that creates an instance, fails midway, and does not clean up can cost significant money. A CI pipeline running tests on every PR can quickly burn through credits. Leaked resources from failed test cleanup compound the problem.

**Why it happens:**
Vast.ai has no sandbox or mock API. Every acceptance test hits the real API and provisions real GPU instances. The Python SDK has no local testing mode. Test failures, timeouts, or CI runner crashes leave instances running and billing.

**How to avoid:**
- Implement test sweepers (using `resource.AddTestSweepers()`) that destroy all instances with a test prefix (e.g., `tf-acc-test-`) before and after test runs.
- Use the cheapest possible offers in acceptance tests -- filter for the lowest-cost GPU type, smallest disk, and use interruptible instances for cost savings.
- Set aggressive test timeouts to prevent runaway tests.
- Add a CI budget guard: track Vast.ai credit balance before and after test runs, fail the pipeline if costs exceed a threshold.
- Use `CheckDestroy` functions in every test to verify resource cleanup.
- Separate unit tests (testing schema logic, plan modifiers, validators -- no API calls) from acceptance tests (gated behind `TF_ACC=1`).
- Consider running acceptance tests only on merge to main, not on every PR commit.
- Add a scheduled CI job that runs sweepers to clean up any leaked resources.

**Warning signs:**
- Vast.ai credit balance dropping faster than expected.
- `show instances` reveals instances with `tf-acc-test` prefixes that should have been destroyed.
- Acceptance tests timing out because instance provisioning is slow.

**Phase to address:**
Phase 2 (Testing Infrastructure) -- set up sweepers and cost controls before writing the first acceptance test. Phase 3+ (Each Resource) -- add sweepers for each new resource type.

---

### Pitfall 7: Import Support That Reconstructs Incomplete State

**What goes wrong:**
Terraform import reads a resource by its ID and populates state. If the Import function only fetches the subset of attributes available from the GET API endpoint (which is common), the resulting state will be missing attributes that were set during creation but are not returned by the API (e.g., the original `offer_id`, `onstart_cmd`, `bid_price`, user-specified `label`). The next `terraform plan` then shows a diff for every missing attribute, forcing the user to either add all attributes to their config or accept a noisy plan.

**Why it happens:**
Many APIs do not return all creation parameters on subsequent GETs. The Vast.ai API's `show instance` endpoint returns instance runtime state (status, IP, ports) but may not return all original creation parameters (template_hash, onstart script contents, original bid price). Developers implement import as just calling Read, without accounting for the data gap.

**How to avoid:**
- Audit the Vast.ai `GET /api/v0/instances/{id}/` response to determine exactly which creation-time attributes are returned and which are lost.
- For attributes not returned by the API, mark them as `Optional: true, Computed: true` so Terraform does not require them in config after import.
- Document which attributes are populated by import and which must be manually added to the configuration after import.
- Use Terraform 1.5+ import blocks (declarative import) which handle config generation better than `terraform import` CLI.
- Test import explicitly in acceptance tests using `ImportState` test steps with `ImportStateVerify: true` to catch missing attributes early.

**Warning signs:**
- `terraform plan` shows many changes immediately after `terraform import`.
- Users report that imported resources want to be "replaced" on next apply.
- `ImportStateVerify` test steps fail because state does not match config.

**Phase to address:**
Phase 2-3 (Each Resource) -- implement and test import for every resource, not as an afterthought.

---

### Pitfall 8: Treating the Python SDK as a Reliable API Contract

**What goes wrong:**
The Go API client is ported from the Python SDK, which is the CLI tool -- not a formal API specification. The Python SDK has hardcoded URL paths, undocumented query parameters, inconsistent error handling (sometimes `raise_for_status()`, sometimes manual status checks), and implicit behaviors (like automatically retrying on 429 with a 150ms starting backoff). Developers copy these patterns into Go without understanding they are implementation details, not API guarantees. When Vast.ai updates their API, the Python SDK may change in ways that break assumptions baked into the Go client.

**Why it happens:**
No OpenAPI spec exists. The Python SDK at `vast-sdk/` is the "source of truth," but it is a CLI application with print statements, global state, and ad-hoc error handling -- not an API client library. Functions like `create_instance()` (line 2386) mix business logic (file reading, runtype detection) with HTTP calls. The actual API contract is buried under CLI concerns.

**How to avoid:**
- Extract only the HTTP-level information from the Python SDK: endpoint URLs, HTTP methods, request/response JSON shapes, and authentication. Ignore CLI-specific logic (argument parsing, output formatting, file I/O).
- For every endpoint, manually test with `curl` to verify the actual request/response format. Do not trust the Python SDK's string formatting of URLs.
- Build the Go client as a clean, independent HTTP client -- not a line-by-line port of the Python SDK.
- Document each API endpoint with its actual request/response contract as discovered through testing.
- Implement integration tests that validate the Go client's assumptions against the live API.
- Subscribe to Vast.ai changelog/updates to catch API changes early.

**Warning signs:**
- Go client functions mirror Python SDK function signatures instead of being idiomatic Go.
- Error handling in the Go client is inconsistent (some functions check status codes, others do not).
- API calls work in the Python SDK but fail in the Go client for the same parameters.

**Phase to address:**
Phase 1 (Go API Client) -- this is the foundation. A poorly-abstracted API client will poison every resource built on top of it.

---

### Pitfall 9: Rate Limiting Causes Cascading Failures in Parallel Operations

**What goes wrong:**
Terraform defaults to 10 parallel operations. When managing multiple instances, templates, and volumes, the provider fires many concurrent API calls. The Vast.ai API returns HTTP 429 when rate-limited (the Python SDK handles this at line 357-361 with a retry loop). Without proper backoff in the Go client, Terraform operations fail noisily, and users see partial applies -- some resources created, others failed with 429 errors. Terraform's state then reflects a partially-applied configuration.

**Why it happens:**
Developers test with single resources and never hit rate limits. In production, users manage tens or hundreds of GPU instances. The Vast.ai API's rate limits are undocumented (no official rate limit documentation was found), and the Python SDK's retry logic (starting at 150ms, multiplying by 1.5x) is minimal.

**How to avoid:**
- Implement exponential backoff with jitter in the Go HTTP client for all API calls, not just specific endpoints. Start at 200ms, cap at 30 seconds, with full jitter.
- Respect `Retry-After` headers if the API returns them.
- Make retry count configurable via provider configuration (default: 10 retries).
- Use a shared rate limiter (e.g., `golang.org/x/time/rate`) across all API calls from the same client instance to proactively throttle before hitting 429s.
- Log retries at DEBUG level so users can diagnose slowness.
- Test with `terraform apply -parallelism=20` to stress-test rate limiting.

**Warning signs:**
- Intermittent 429 errors in CI acceptance tests.
- `terraform apply` takes much longer than expected due to retry storms.
- Partial applies where some resources are created and others fail.

**Phase to address:**
Phase 1 (Go API Client) -- rate limiting must be built into the HTTP transport layer from day one.

---

### Pitfall 10: Registry Publication Fails Due to GPG Key Type or Goreleaser Misconfiguration

**What goes wrong:**
The Terraform Registry rejects provider releases signed with ECC GPG keys (the modern default). GoReleaser fails silently when the GPG key requires a passphrase in CI. The `terraform-registry-manifest.json` file specifies the wrong protocol version. Binary names do not match the expected pattern. Any of these causes the release to fail after all development work is complete, blocking users from installing the provider.

**Why it happens:**
Registry publishing has strict but non-obvious requirements. Developers focus on code quality and leave publishing for last. GPG key generation defaults to ECC (ed25519) on modern systems, but the registry only accepts RSA and DSA. GoReleaser's signing step silently fails in CI because no TTY is available for passphrase entry. The manifest file's `protocol_versions` must match the framework version used.

**How to avoid:**
- Generate an RSA GPG key specifically for release signing: `gpg --full-generate-key` and select RSA.
- Store the private key and passphrase as GitHub Secrets (`GPG_PRIVATE_KEY`, `PASSPHRASE`).
- Copy the GoReleaser config and GitHub Actions workflow from `terraform-provider-scaffolding-framework` -- do not write from scratch.
- Include `terraform-registry-manifest.json` with `"protocol_versions": ["6.0"]` (for Plugin Framework).
- Ensure the `--batch` flag is set in GoReleaser signing config.
- Set up CI/CD and publish a `v0.1.0-alpha` release early to validate the entire pipeline before investing months in code.
- Never modify an already-released version -- publish a new patch instead.
- Verify: no branch has the same name as any version tag.

**Warning signs:**
- GoReleaser signing step produces no output or an empty `.sig` file.
- Terraform Registry shows "signature verification failed" after publish.
- `terraform init` fails with "checksum mismatch" for users trying to install.

**Phase to address:**
Phase 1 (Project Scaffold) -- set up CI/CD, GPG signing, and publish a dummy release to validate the pipeline before writing any provider code.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Skip waiter/poller logic, just sleep for N seconds | Faster to implement | Flaky tests, slow operations, race conditions in production | Never -- implement proper waiters from the start |
| Hardcode API base URL | Simpler config | Cannot test against staging, cannot support self-hosted Vast.ai | Never -- make it configurable via provider attribute and `VASTAI_URL` env var |
| Skip import support initially | Ship resources faster | Users cannot adopt provider for existing infrastructure; adding import later requires understanding all state mappings retroactively | Only for MVP/alpha, add before v1.0 |
| Copy Python SDK patterns directly into Go | Faster porting | Non-idiomatic Go, inherited bugs, brittle to API changes | Never -- extract API contracts, implement idiomatically |
| Use SDKv2 instead of Plugin Framework | More examples available | SDKv2 is legacy; no new features; migration later is painful for complex resources | Never -- HashiCorp mandates Plugin Framework for new providers |
| Test only happy paths | Faster test suite | Fails in production when resources are preempted, rate-limited, or API returns unexpected data | Only for initial development; add error path tests before beta |
| Single retry on 429, no backoff | Simpler code | Fails under load, cascading failures with parallel operations | Never -- implement proper exponential backoff from the start |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Vast.ai Instance API | Treating `PUT /asks/{id}/` as idempotent (it creates a new contract each time) | Track the returned `new_contract` ID as the resource ID, not the offer ID. The offer ID is consumed on creation. |
| Vast.ai Instance API | Assuming create returns the full instance object | Create returns `{"success": true, "new_contract": 12345}`. You must then GET the instance by the contract ID to populate state. |
| Vast.ai Auth | Sending API key in both query parameter and Authorization header | Use only the Authorization header. The query parameter approach leaks credentials in logs. |
| Vast.ai Volume API | Assuming volumes are immediately available after creation | Volumes may take time to provision. Implement a waiter for volume status. |
| Vast.ai Search/Offers API | Caching offer results across plan and apply | Offers are ephemeral. A data source read during plan may return stale offers by apply time. Document this limitation. |
| Terraform Registry | Using `protocol_versions: ["5.0"]` with Plugin Framework | Plugin Framework uses protocol version 6. Set `"protocol_versions": ["6.0"]` in manifest. |
| GoReleaser | Using default `.goreleaser.yml` from an older provider | Copy from `terraform-provider-scaffolding-framework` repository to get the current recommended configuration. |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| No pagination in list operations | Slow data source reads, timeouts for users with many instances | Implement pagination in the Go client for all list endpoints | >50 resources of one type |
| Reading all instances to find one by ID | O(n) instance lookups | Use the single-instance GET endpoint (`/instances/{id}/`), not the list endpoint with client-side filter | >20 concurrent instances |
| No caching in the offers data source | Every `terraform plan` re-queries the full offers list | Accept staleness: cache offers for the duration of a single Terraform operation (provider lifetime) | >10 data source references in config |
| Polling too frequently in waiters | 429 rate limits during instance creation | Use exponential backoff starting at 2 seconds, not sub-second polling | Creating >5 instances simultaneously |
| Serializing all API calls | `terraform apply` takes minutes for multi-resource configs | Ensure the HTTP client is safe for concurrent use (no shared mutable state) | >10 resources in a single apply |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| API key in query parameters | Key leaked in logs, browser history, proxy logs, Terraform state | Use Authorization header exclusively; mark provider `api_key` attribute as `Sensitive: true` |
| API key stored in Terraform state | Credential exposure from state file access | Read from `VASTAI_API_KEY` env var only; do not store in provider config if possible |
| SSH private keys in resource attributes | Keys stored in plaintext in state | Use `WriteOnly` attributes (Terraform 1.11+) or reference keys by ID rather than value |
| Error messages containing full API responses | Responses may include tokens, internal IDs, or other sensitive data | Sanitize API responses before including in Terraform diagnostics |
| Acceptance test API key in CI logs | CI runners may log environment variables | Use GitHub Actions masked secrets; never echo `VASTAI_API_KEY` in CI scripts |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Opaque error messages ("API returned 400") | Users cannot diagnose issues | Parse Vast.ai error responses and surface the `msg` field in diagnostics |
| No documentation of the offer-then-create workflow | Users try to create instances without first searching for offers | Provide examples showing the complete workflow: data source search + resource creation |
| Computed attributes without `UseStateForUnknown` | Every `terraform plan` shows `(known after apply)` for stable values | Apply `UseStateForUnknown()` to all computed attributes that do not change after creation |
| Missing `timeouts` block support | Users cannot extend timeouts for slow provisioning | Implement configurable timeouts on all resources that wait for async operations |
| Instance resource recreated when offer expires | Users see "destroy + create" on every plan if the offer data source changes | Decouple offer selection from instance lifecycle; instance references `machine_id`, not `offer_id` after creation |
| No examples directory | Users must reverse-engineer from docs | Ship a `examples/` directory with complete, working configurations for common workflows |

## "Looks Done But Isn't" Checklist

- [ ] **Instance Resource:** Often missing waiter logic for async provisioning -- verify that `terraform apply` waits for the instance to reach `running` state before completing.
- [ ] **Instance Resource:** Often missing preemption handling -- verify that Read gracefully removes preempted/destroyed instances from state.
- [ ] **All Resources:** Often missing `ImportState` implementation -- verify that `terraform import <resource> <id>` works and produces a clean subsequent plan.
- [ ] **All Resources:** Often missing `CheckDestroy` in acceptance tests -- verify that destroyed resources are actually gone from the API.
- [ ] **Data Sources:** Often missing error handling for empty results -- verify that an offers search with no matches returns a useful error, not a panic.
- [ ] **Provider:** Often missing configurable base URL -- verify that `VASTAI_URL` env var can override the default `https://console.vast.ai` for testing.
- [ ] **CI/CD:** Often missing test sweepers -- verify that a scheduled job cleans up leaked test resources.
- [ ] **Registry:** Often missing `terraform-registry-manifest.json` -- verify that it exists with correct protocol version.
- [ ] **Documentation:** Often missing examples for interruptible instances -- verify that bid_price and preemption behavior are documented.
- [ ] **Schema:** Often missing plan modifiers on computed attributes -- verify that `terraform plan` is clean after `terraform apply` with no config changes.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| API key leaked in logs | MEDIUM | Rotate API key immediately via Vast.ai console. Audit CI logs for exposure. Update `VASTAI_API_KEY` in all environments. |
| "Inconsistent result after apply" errors | LOW-MEDIUM | Fix the schema (add `Computed: true` or plan modifiers). Run `terraform refresh` to sync state. No data loss. |
| Preempted instance breaks state | LOW | `terraform plan` will show the resource needs recreation. Apply to create a new instance. Or `terraform state rm` to remove manually. |
| Leaked test instances accumulating cost | MEDIUM | Run test sweepers (`go test -sweep=all`). Check Vast.ai console for instances with test prefixes. Destroy manually if needed. |
| Registry publication failure | LOW | Fix GPG key type or GoReleaser config. Publish a new patch version. Cannot fix an existing release -- must increment version. |
| Offers expire between plan and apply | LOW | Re-run `terraform plan` to refresh offers. Add `-refresh=true` flag. Document this as expected behavior. |
| Go client rate-limited under load | MEDIUM | Add/tune exponential backoff. Reduce Terraform parallelism (`-parallelism=5`). Add a shared rate limiter to the client. |
| Import produces noisy plan | LOW | Add missing attributes to config manually. Run `terraform plan` to identify diffs. Or re-import with corrected config. |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| API key in query params | Phase 1: Go API Client | `grep -r "api_key" *.go` finds zero query-parameter usages; all auth is via headers |
| Schema inconsistency errors | Phase 2: First Resource | Acceptance tests pass with `TF_LOG=TRACE` showing zero data consistency warnings |
| No waiter logic | Phase 1-2: Client + Instance | `terraform apply` for instance creation waits until `actual_status == "running"` |
| Spot instance state breakage | Phase 2: Instance Resource | Test scenario: create interruptible instance, externally stop it, verify `terraform plan` shows recreation (not error) |
| Offer ID vs spec confusion | Phase 2: Data Source + Instance | Documentation and examples show the two-step workflow |
| Test cost accumulation | Phase 2: Test Infrastructure | Sweepers exist for every resource type; CI has cost threshold checks |
| Import incomplete state | Phase 2-3: Each Resource | `ImportStateVerify: true` test step passes for every resource |
| Python SDK as API contract | Phase 1: Go API Client | Each endpoint has a documented request/response contract verified with curl |
| Rate limiting cascade | Phase 1: Go API Client | Tests pass with `-parallelism=20`; client has exponential backoff with jitter |
| Registry publication failure | Phase 1: Project Scaffold | A `v0.1.0-alpha` release successfully installs via `terraform init` |

## Sources

- [HashiCorp Provider Design Principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles) -- resource modeling, CRUD behavior, schema alignment
- [Terraform Plugin Framework Data Consistency Errors](https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/data-consistency-errors) -- schema misconfiguration causes and fixes
- [AWS Provider Retries and Waiters](https://hashicorp.github.io/terraform-provider-aws/retries-and-waiters/) -- retry patterns, StateChangeConf, eventual consistency handling
- [AWS Provider Error Handling](https://hashicorp.github.io/terraform-provider-aws/error-handling/) -- error wrapping, not-found handling, diagnostic patterns
- [Terraform Registry Publishing Requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing) -- GPG key types, GoReleaser config, manifest file
- [Terraform Plugin Framework Acceptance Testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests) -- test setup, provider server configuration, missing id attribute
- [Terraform Plugin Framework Plan Modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification) -- UseStateForUnknown, RequiresReplace gotchas
- [Terraform Test Sweepers](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/sweepers) -- cleanup leaked resources from failed tests
- [Vast.ai Pricing Documentation](https://docs.vast.ai/documentation/instances/pricing) -- on-demand, reserved, interruptible instance types
- [Vast.ai Rental Types FAQ](https://docs.vast.ai/documentation/reference/faq/rental-types) -- bid mechanics, preemption behavior
- [Vast.ai Python SDK](vast-sdk/vastai/vast.py) -- `http_request()` retry logic (line 333), `apiurl()` auth pattern (line 575), `create_instance()` contract (line 2386), `destroy_instance()` behavior (line 3078)
- [HashiCorp Discuss: Terraform provider acceptance testing and eventual consistency](https://discuss.hashicorp.com/t/terraform-provider-acceptance-testing-and-eventually-consistentancy/52901) -- community strategies for flaky tests
- [How to Handle API Rate Limiting in Custom Providers](https://oneuptime.com/blog/post/2026-02-23-how-to-handle-api-rate-limiting-in-custom-providers/view) -- rate limiting patterns for custom providers
- [Terraform Provider Authentication Best Practices](https://oneuptime.com/blog/post/2026-02-23-terraform-provider-authentication/view) -- credential security in providers

---
*Pitfalls research for: Terraform/OpenTofu provider for Vast.ai GPU marketplace*
*Researched: 2026-03-25*
