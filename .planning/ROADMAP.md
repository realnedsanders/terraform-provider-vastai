# Roadmap: terraform-provider-vastai

## Overview

This roadmap delivers a production-grade Terraform provider for Vast.ai, progressing from a compilable provider skeleton with a working release pipeline through complete resource coverage across compute, storage, serverless, account, and networking domains, finishing with registry-ready documentation and examples. Each phase delivers a coherent, verifiable capability that builds on the previous phase's patterns and infrastructure.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation** - Provider scaffold, Go API client, and release pipeline that installs via `terraform init`
- [ ] **Phase 2: Core Compute** - Instance lifecycle with GPU offer search, templates, SSH keys, and schema quality patterns
- [x] **Phase 3: Storage** - Volume and network volume resources with offer search data sources (completed 2026-03-27)
- [x] **Phase 4: Serverless** - Endpoint, worker group, and autoscaler resources for inference workflows (completed 2026-03-27)
- [ ] **Phase 5: Account & Networking** - API keys, teams, clusters, overlays, and remaining data sources
- [ ] **Phase 6: Documentation & Release** - Generated docs, working examples, test sweepers, and registry publication

## Phase Details

### Phase 1: Foundation
**Goal**: A compilable, installable Terraform provider with zero resources but a working API client, CI/CD pipeline, and a published alpha release that installs via `terraform init`
**Depends on**: Nothing (first phase)
**Requirements**: FOUND-01, FOUND-02, FOUND-03, FOUND-04, FOUND-05, FOUND-06, RLSE-01, RLSE-02, RLSE-03, RLSE-04, RLSE-05, TEST-04
**Success Criteria** (what must be TRUE):
  1. Running `terraform init` with the provider source successfully downloads and installs the provider binary
  2. Provider authenticates to Vast.ai API using `VASTAI_API_KEY` env var or provider attribute, with credentials sent via Authorization header (never in URL query parameters)
  3. API client retries failed requests with exponential backoff on 429/5xx responses and surfaces structured error diagnostics on failure
  4. Tagging a release in GitHub triggers automated cross-compilation, GPG signing, and artifact publication
  5. `terraform plan` with an empty configuration and valid API key completes without error
**Plans:** 3 plans

Plans:
- [x] 01-01-PLAN.md -- Provider scaffold: Go module, provider shell with schema, build tooling, registry manifest
- [ ] 01-02-PLAN.md -- API client: Bearer auth, retry with backoff, structured errors, provider wiring
- [ ] 01-03-PLAN.md -- CI/CD: GoReleaser config, release workflow, test workflow, .gitignore

### Phase 2: Core Compute
**Goal**: Users can search GPU offers, create instances from offers, manage instance lifecycle (start/stop/update), and configure templates and SSH keys -- the complete compute workflow end-to-end
**Depends on**: Phase 1
**Requirements**: COMP-01, COMP-02, COMP-03, COMP-04, COMP-05, COMP-06, COMP-07, COMP-08, DATA-01, DATA-02, DATA-03, DATA-04, DATA-08, SCHM-01, SCHM-02, SCHM-03, SCHM-04, SCHM-05, SCHM-06, IMPT-01, IMPT-02, TEST-01, TEST-02
**Success Criteria** (what must be TRUE):
  1. User can write a Terraform config that searches GPU offers by filters (gpu_name, num_gpus, price, region) and creates an instance from the best matching offer
  2. User can start, stop, update labels, change bid price, and destroy instances without recreating -- all via `terraform apply`
  3. User can create and manage templates (image, env vars, onstart_cmd) and SSH keys as Terraform resources, and attach SSH keys to instances
  4. Running `terraform import` for any managed resource populates state correctly, and `terraform plan` after import shows no diff for stable attributes
  5. All resources have attribute validators on constrained fields, sensitive flags on secrets, correct Required/Optional/Computed classification, and meaningful descriptions
**Plans:** 2/6 plans executed

Plans:
- [x] 02-01-PLAN.md -- API client services: InstanceService, OfferService, TemplateService, SSHKeyService with typed structs, unit tests, new dependencies
- [ ] 02-02-PLAN.md -- GPU offers data source and template resource/data-source with schema quality patterns
- [x] 02-03-PLAN.md -- SSH key resource with CRUD, import, sensitive flags, and SSH keys data source
- [ ] 02-04-PLAN.md -- Instance resource with full lifecycle, preemption handling, SSH attachment, and import
- [ ] 02-05-PLAN.md -- Instance data sources, provider registration, and full unit test suite integration
- [ ] 02-06-PLAN.md -- Acceptance tests for all resources and data sources (TF_ACC-gated create/read/update/import/destroy)

### Phase 3: Storage
**Goal**: Users can provision and manage persistent volumes and network volumes through Terraform, with offer search for finding available storage and clone support for local volumes
**Depends on**: Phase 2
**Requirements**: STOR-01, STOR-02, STOR-03, DATA-05, DATA-06
**Success Criteria** (what must be TRUE):
  1. User can create a volume from an offer, clone it via clone_from_id, and destroy it via Terraform
  2. User can create and manage network volumes with full CRUD (create from offer, read, delete)
  3. User can search volume and network volume offers with filter attributes and use results to provision storage resources
**Plans:** 3/3 plans complete

Plans:
- [x] 03-01-PLAN.md -- API client services: VolumeService and NetworkVolumeService with CRUD, clone, offer search, and unit tests
- [x] 03-02-PLAN.md -- Volume resource with CRUD/clone/import and volume offers data source with structured filters
- [x] 03-03-PLAN.md -- Network volume resource with CRUD/import, network volume offers data source, and provider registration

### Phase 4: Serverless
**Goal**: Users can set up complete serverless inference endpoints with worker groups and autoscaling configuration through Terraform
**Depends on**: Phase 2
**Requirements**: SRVL-01, SRVL-02, SRVL-03, DATA-09
**Success Criteria** (what must be TRUE):
  1. User can create a serverless endpoint with autoscaling parameters (min_load, target_util, cold_mult, cold_workers, max_workers) and manage it via Terraform
  2. User can create worker groups bound to endpoints with template and search parameter configuration
  3. User can configure autoscaling groups and query endpoint status via data source
**Plans:** 3/3 plans complete

Plans:
- [x] 04-01-PLAN.md -- API client services: EndpointService and WorkerGroupService with CRUD, typed structs, and unit tests
- [x] 04-02-PLAN.md -- Endpoint resource with CRUD/import/autoscaling config, endpoints data source, and provider registration
- [x] 04-03-PLAN.md -- Worker group resource with CRUD/import, endpoint binding, template config, and provider registration

### Phase 5: Account & Networking
**Goal**: Users can manage their Vast.ai account configuration (API keys, teams, environment variables) and advanced networking (clusters, overlays) entirely through Terraform
**Depends on**: Phase 2
**Requirements**: ACCT-01, ACCT-02, ACCT-03, ACCT-04, ACCT-05, ACCT-06, NETW-01, NETW-02, NETW-03, NETW-04, DATA-07, DATA-10, DATA-11
**Success Criteria** (what must be TRUE):
  1. User can create and manage API keys with permission scoping, environment variables, and subaccounts as Terraform resources (all secret values marked sensitive)
  2. User can create teams, define roles with granular permissions, and invite/remove team members via Terraform
  3. User can create clusters and overlays, manage cluster membership (join/remove machines), and join instances to overlays
  4. User can query their account profile, billing invoices, and audit logs via read-only data sources
**Plans**: TBD

Plans:
- [ ] 05-01: TBD
- [ ] 05-02: TBD
- [ ] 05-03: TBD
- [ ] 05-04: TBD

### Phase 6: Documentation & Release
**Goal**: Provider is registry-ready with generated documentation for every resource and data source, working example configurations, and test sweepers for safe CI operation
**Depends on**: Phase 3, Phase 4, Phase 5
**Requirements**: DOCS-01, DOCS-02, DOCS-03, DOCS-04, TEST-03
**Success Criteria** (what must be TRUE):
  1. Running `tfplugindocs generate` produces complete documentation for all resources and data sources, and the output renders correctly in Terraform Registry format
  2. Every resource and data source has a working example `.tf` file in the `examples/` directory that can be used as-is
  3. Provider configuration documentation covers authentication, endpoint configuration, and retry behavior with examples
  4. Resource sweepers exist for all resource types that create cloud resources, and CI runs sweepers to clean up leaked test resources
**Plans**: TBD

Plans:
- [ ] 06-01: TBD
- [ ] 06-02: TBD
- [ ] 06-03: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6
Note: Phases 3, 4, and 5 depend only on Phase 2 (not on each other) but execute sequentially for focus.

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 0/3 | Not started | - |
| 2. Core Compute | 6/6 | Complete | 2026-03-25 |
| 3. Storage | 3/3 | Complete   | 2026-03-27 |
| 4. Serverless | 3/3 | Complete   | 2026-03-27 |
| 5. Account & Networking | 0/4 | Not started | - |
| 6. Documentation & Release | 0/3 | Not started | - |
