# Milestones

## v1.0 Full API Coverage (Shipped: 2026-03-28)

**Phases completed:** 6 phases, 24 plans, 44 tasks

**Key accomplishments:**

- Terraform provider Go project with compilable binary, VastaiProvider schema (api_key sensitive, api_url optional), registry manifest protocol 6.0, and build/lint tooling
- REST API client with Bearer auth, 150ms/1.5x exponential backoff, structured APIError, and provider wiring via go-retryablehttp
- GitHub Actions CI/CD with GoReleaser v2 cross-compilation, GPG-signed SHA256SUMS, and acceptance test gating on main branch
- Go API client services for instances, offers, templates, and SSH keys with full CRUD, lifecycle management, status polling, and 40 unit tests using httptest mocks
- GPU offers data source with structured filters/most_affordable and template resource with full CRUD, import, timeouts, and comprehensive schema quality patterns
- SSH key resource with CRUD, import, sensitive flag, format validator, and data source listing all keys with 10 unit tests
- vastai_instance resource with full CRUD, start/stop lifecycle, spot preemption detection, SSH key attachment, bid/label/template updates, and import support
- Instance data sources (singular by-ID, plural list with label filter) plus full provider registration of 3 resources and 5 data sources
- TF_ACC-gated acceptance tests covering full CRUD/import lifecycle for all Phase 2 resources (instance, template, SSH key) and read verification for all 5 data sources, using terraform-plugin-testing v1.15.0
- Go API client services for volumes and network volumes with typed CRUD methods, offer search, and 11 unit tests
- vastai_volume resource with CRUD/clone/import and vastai_volume_offers data source with 16 filter attributes, most_affordable convenience, and 18 unit tests
- vastai_network_volume resource with CRUD/import, vastai_network_volume_offers data source with bandwidth metrics, and full Phase 3 provider registration (5 resources, 7 data sources)
- Go API client services for serverless endpoints (/endptjobs/) and worker groups (/autojobs/) with typed CRUD methods and 8 httptest unit tests
- vastai_endpoint resource with CRUD, import, autoscaling validators (target_util Between(0,1), cold_mult AtLeast(1.0)) and vastai_endpoints data source, all with 15 passing schema unit tests
- vastai_worker_group resource with CRUD, import, endpoint binding (ForceNew), and template AtLeastOneOf validation -- autoscaling params correctly omitted per Pitfall 3
- Go API client services for API keys, env vars, teams (roles/members), and subaccounts with GetFullPath for v1 API support
- Cluster, overlay, user, invoice, and audit log Go API client services with create-then-read and v1 API support
- API key, environment variable, and subaccount Terraform resources with immutable/create-only patterns, sensitive value handling, and 32 passing unit tests
- Team, team role, and team member Terraform resources with asymmetric API handling and invite-as-create pattern
- Cluster, overlay, and membership resources with composite IDs, create-then-read pattern, and no-op destroy for API-limited overlay membership
- Read-only data sources for user profile, billing invoices, and audit logs with full Phase 5 provider registration (17 resources, 11 data sources)
- tfplugindocs templates, per-resource examples, and registry-ready generated docs for all 17 resources and 11 data sources
- Plan:
- 1. [Rule 3 - Blocking] Import cycle between acctest and service packages

---
