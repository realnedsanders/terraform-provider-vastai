---
phase: 03-storage
verified: 2026-03-27T12:00:00Z
status: passed
score: 10/10 must-haves verified
re_verification: false
---

# Phase 3: Storage Verification Report

**Phase Goal:** Users can provision and manage persistent volumes and network volumes through Terraform, with offer search for finding available storage and clone support for local volumes
**Verified:** 2026-03-27T12:00:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #  | Truth                                                                                | Status     | Evidence                                                                                                                                                      |
|----|--------------------------------------------------------------------------------------|------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1  | User can create a volume from an offer, clone it via clone_from_id, and destroy it   | VERIFIED   | `resource_volume.go` implements createFromOffer and createViaClone paths; Delete calls `client.Volumes.Delete`; tests pass                                   |
| 2  | User can create and manage network volumes with full CRUD (create, read, delete)     | VERIFIED   | `resource_network_volume.go` implements Create, Read (via list), Delete, ImportState; tests pass                                                             |
| 3  | User can search volume and network volume offers with filter attributes              | VERIFIED   | `data_source_volume_offers.go` and `data_source_network_volume_offers.go` both present with full filter mapping; `SearchOffers` wired in both                |
| 4  | VolumeService provides Create, Clone, List, Delete, SearchOffers with correct paths  | VERIFIED   | `volumes.go` implements all five methods; pitfalls respected (query param delete, allocated_storage in search, read-via-list)                                |
| 5  | NetworkVolumeService provides Create, List, Delete, SearchOffers with correct paths  | VERIFIED   | `network_volumes.go` implements all four methods; uses `/volumes?type=network_volume` for List; DELETE uses same query param pattern                          |
| 6  | VastAIClient initializes Volumes and NetworkVolumes service sub-objects              | VERIFIED   | `client.go` lines 57–58: `c.Volumes = &VolumeService{client: c}` and `c.NetworkVolumes = &NetworkVolumeService{client: c}`                                  |
| 7  | User can import an existing volume by ID                                             | VERIFIED   | `resource_volume.go` implements `ImportState` via `resource.ImportStatePassthroughID`                                                                        |
| 8  | User can import an existing network volume by ID                                     | VERIFIED   | `resource_network_volume.go` implements `ImportState` via `resource.ImportStatePassthroughID`                                                                |
| 9  | Provider registers all Phase 3 resources and data sources alongside Phase 2          | VERIFIED   | `provider.go` registers 5 resources and 7 data sources including `volume.NewVolumeResource`, `networkvolume.NewNetworkVolumeResource`, both offers data sources |
| 10 | All 49+ unit tests pass                                                              | VERIFIED   | `go test ./...` exits 0; 7 client volume tests, 4 client network volume tests, 18 volume service tests, 20 network volume service tests                      |

**Score:** 10/10 truths verified

### Required Artifacts

| Artifact                                                                   | Expected                                            | Status   | Details                                                                          |
|----------------------------------------------------------------------------|-----------------------------------------------------|----------|----------------------------------------------------------------------------------|
| `internal/client/volumes.go`                                               | VolumeService with Create, Clone, List, Delete, SearchOffers | VERIFIED | 240 lines; all five methods present; pitfalls 2, 3, 5 handled               |
| `internal/client/network_volumes.go`                                       | NetworkVolumeService with Create, List, Delete, SearchOffers | VERIFIED | 188 lines; all four methods present; network-specific fields included        |
| `internal/client/volumes_test.go`                                          | Unit tests for VolumeService                        | VERIFIED | 7 test functions; covers Create, Clone, List, Delete, SearchOffers, RawQuery, Defaults |
| `internal/client/network_volumes_test.go`                                  | Unit tests for NetworkVolumeService                 | VERIFIED | 4 test functions; covers Create, List, Delete, SearchOffers                     |
| `internal/services/volume/models.go`                                       | VolumeResourceModel, VolumeOfferModel, VolumeOffersDataSourceModel | VERIFIED | All three models present with correct field types                        |
| `internal/services/volume/resource_volume.go`                              | vastai_volume resource with Create/Clone/Read/Delete/Import | VERIFIED | Full implementation; dual create path (offer vs clone); read-via-list            |
| `internal/services/volume/data_source_volume_offers.go`                    | vastai_volume_offers data source with filters       | VERIFIED | All 9 filter attributes; most_affordable convenience attribute; SearchOffers wired |
| `internal/services/volume/resource_volume_test.go`                         | Unit tests for volume resource schema               | VERIFIED | 11 test functions; schema, validators, plan modifiers, interface compliance      |
| `internal/services/volume/data_source_volume_offers_test.go`               | Unit tests for volume offers data source schema     | VERIFIED | 7 test functions; filter attributes, result attributes, validators, metadata     |
| `internal/services/networkvolume/models.go`                                | NetworkVolumeResourceModel, NetworkVolumeOfferModel, NetworkVolumeOffersDataSourceModel | VERIFIED | All three models present |
| `internal/services/networkvolume/resource_network_volume.go`               | vastai_network_volume resource with Create/Read/Delete/Import | VERIFIED | No clone (local-volume-only); read-via-list; full schema                        |
| `internal/services/networkvolume/data_source_network_volume_offers.go`     | vastai_network_volume_offers data source            | VERIFIED | Network-specific fields (nw_disk_min_bw, nw_disk_max_bw, nw_disk_avg_bw, cluster_id); SearchOffers wired |
| `internal/services/networkvolume/resource_network_volume_test.go`          | Unit tests for network volume resource schema       | VERIFIED | 11 test functions including TestNetworkVolumeResource_Schema_NoCloneFromID      |
| `internal/services/networkvolume/data_source_network_volume_offers_test.go`| Unit tests for network volume offers data source    | VERIFIED | 9 test functions; network-specific field verification                            |
| `internal/provider/provider.go`                                            | Updated provider with 5 resources and 7 data sources | VERIFIED | Exactly 5 resources, 7 data sources; both volume and networkvolume packages imported |

### Key Link Verification

| From                                          | To                              | Via                                                | Status   | Details                                                          |
|-----------------------------------------------|---------------------------------|----------------------------------------------------|----------|------------------------------------------------------------------|
| `internal/client/client.go`                   | `internal/client/volumes.go`    | `c.Volumes = &VolumeService{client: c}`            | WIRED    | Line 57 in `NewVastAIClient`                                     |
| `internal/client/client.go`                   | `internal/client/network_volumes.go` | `c.NetworkVolumes = &NetworkVolumeService{client: c}` | WIRED | Line 58 in `NewVastAIClient`                                    |
| `internal/services/volume/resource_volume.go` | `internal/client/volumes.go`    | `r.client.Volumes.Create/Clone/List/Delete`        | WIRED    | Lines 237, 277, 288, 360, 437 verified                           |
| `internal/services/volume/data_source_volume_offers.go` | `internal/client/volumes.go` | `d.client.Volumes.SearchOffers`                | WIRED    | Line 342 verified                                                |
| `internal/services/networkvolume/resource_network_volume.go` | `internal/client/network_volumes.go` | `r.client.NetworkVolumes.Create/List/Delete` | WIRED | Lines 211, 267, 344 verified                                 |
| `internal/services/networkvolume/data_source_network_volume_offers.go` | `internal/client/network_volumes.go` | `d.client.NetworkVolumes.SearchOffers` | WIRED | Line 334 verified                                     |
| `internal/provider/provider.go`               | `internal/services/volume/`     | `volume.NewVolumeResource`, `volume.NewVolumeOffersDataSource` | WIRED | Both present in Resources() and DataSources() slices |
| `internal/provider/provider.go`               | `internal/services/networkvolume/` | `networkvolume.NewNetworkVolumeResource`, `networkvolume.NewNetworkVolumeOffersDataSource` | WIRED | Both present; imports verified |

### Data-Flow Trace (Level 4)

This phase produces Terraform provider resources and data sources — there is no independent data store to trace; all data flows from the Vast.ai API through the client to the service layer to Terraform state. The chain is:

| Artifact                          | Data Variable | Source                          | Produces Real Data | Status   |
|-----------------------------------|---------------|---------------------------------|--------------------|----------|
| `resource_volume.go` (Create)     | `vol *Volume` | `client.Volumes.Create` → PUT /volumes/ | API call with immediate list-back | FLOWING |
| `resource_volume.go` (Clone)      | `newest *Volume` | `client.Volumes.Clone` + `List` | API call then full list-back | FLOWING |
| `resource_volume.go` (Read)       | `found *Volume` | `client.Volumes.List("local_volume")` | Real API list, filtered by ID | FLOWING |
| `data_source_volume_offers.go`    | `offers []VolumeOffer` | `client.Volumes.SearchOffers` → POST /volumes/search/ | Real API search | FLOWING |
| `resource_network_volume.go` (Create) | `vol *Volume` | `client.NetworkVolumes.Create` → PUT /network_volumes/ | API call with list-back | FLOWING |
| `resource_network_volume.go` (Read)   | `found *Volume` | `client.NetworkVolumes.List` → GET /volumes?type=network_volume | Real API list | FLOWING |
| `data_source_network_volume_offers.go` | `offers []NetworkVolumeOffer` | `client.NetworkVolumes.SearchOffers` → POST /network_volumes/search/ | Real API search | FLOWING |

### Behavioral Spot-Checks

| Behavior                              | Command                                                                | Result          | Status |
|---------------------------------------|------------------------------------------------------------------------|-----------------|--------|
| Full project compiles                 | `go build ./...`                                                       | exits 0         | PASS   |
| Volume client tests pass (7 tests)    | `go test ./internal/client/... -run "TestVolume"`                      | 7 PASS          | PASS   |
| Network volume client tests pass (4 tests) | `go test ./internal/client/... -run "TestNetworkVolume"`          | 4 PASS          | PASS   |
| Volume service tests pass (18 tests)  | `go test ./internal/services/volume/... -v`                            | 18 PASS         | PASS   |
| Network volume service tests pass (20 tests) | `go test ./internal/services/networkvolume/... -v`              | 20 PASS         | PASS   |
| Full test suite passes                | `go test ./... -count=1`                                               | all packages ok | PASS   |

### Requirements Coverage

| Requirement | Source Plan | Description                                                          | Status    | Evidence                                                                                  |
|-------------|-------------|----------------------------------------------------------------------|-----------|-------------------------------------------------------------------------------------------|
| STOR-01     | 03-01, 03-02 | `vastai_volume` resource with CRUD (create from offer, delete)      | SATISFIED | `resource_volume.go` implements Create, Read, Delete, Import. List/unlist omitted per documented scope adjustment (host-only operations). |
| STOR-02     | 03-01, 03-02 | `vastai_volume` supports clone operation                             | SATISFIED | `clone_from_id` attribute present; `createViaClone` method calls `client.Volumes.Clone` |
| STOR-03     | 03-01, 03-03 | `vastai_network_volume` resource with CRUD                           | SATISFIED | `resource_network_volume.go` implements Create, Read, Delete, Import. List/unlist omitted per documented scope adjustment. |
| DATA-05     | 03-01, 03-02 | `vastai_volume_offers` data source with filter support               | SATISFIED | `data_source_volume_offers.go` with 9 filter attributes, most_affordable, SearchOffers wired |
| DATA-06     | 03-01, 03-03 | `vastai_network_volume_offers` data source with filter support       | SATISFIED | `data_source_network_volume_offers.go` with 9 filter attributes plus network-specific offer fields |

**Scope adjustment note:** STOR-01 and STOR-03 reference "list/unlist for marketplace" in REQUIREMENTS.md. Research confirmed these are host-only operations not applicable to tenants. The omission is documented in `03-02-PLAN.md` Task 1 and `03-03-PLAN.md` Task 1 under "NOTE on list/unlist" and is captured in the REQUIREMENTS.md Out of Scope table ("Host-side machine operations"). This is not a gap.

**Orphaned requirements check:** REQUIREMENTS.md Traceability table maps STOR-01, STOR-02, STOR-03, DATA-05, DATA-06 to Phase 3. All five are claimed by plans 03-01, 03-02, and 03-03. No orphaned requirements.

### Anti-Patterns Found

None. Scan of all Phase 3 files found:
- Zero TODO/FIXME/HACK/PLACEHOLDER comments
- No stub implementations (return null, empty handlers, unimplemented methods)
- All return paths carry real data from API responses
- No hardcoded empty values passed as props to downstream consumers

### Human Verification Required

#### 1. End-to-end Terraform workflow: volume create from offer

**Test:** With a real Vast.ai API key, write a Terraform config using `vastai_volume_offers` data source and `vastai_volume` resource. Run `terraform apply`, observe a volume is created. Run `terraform destroy`, observe it is deleted.
**Expected:** Volume appears in Vast.ai account, state file populated with computed fields (status, disk_space, machine_id, etc.), and is removed on destroy.
**Why human:** Requires live Vast.ai API key and real marketplace offer IDs.

#### 2. Volume clone workflow

**Test:** Create a volume, then create a second volume using `clone_from_id = <first_id>`. Verify both volumes exist.
**Expected:** Clone completes, second volume has a new ID distinct from the source, state is populated correctly.
**Why human:** Clone behavior (`createViaClone` takes the highest-ID volume as the result) relies on the assumption that the most recently created volume has the highest ID. This is a heuristic that should be validated against real API behavior.

#### 3. Network volume create and import

**Test:** Create a `vastai_network_volume` resource, then run `terraform import vastai_network_volume.existing <id>` for an existing network volume.
**Expected:** Create succeeds; import populates state with all computed fields; `terraform plan` after import shows no diff for stable attributes.
**Why human:** Requires live API key and real network volume offer.

#### 4. Offer search filter accuracy

**Test:** Search `vastai_volume_offers` with `reliability = 0.99` and `geolocation = "US"`. Verify all returned offers satisfy both constraints.
**Expected:** Returned offers match filter criteria.
**Why human:** Requires live API call to verify the structured query body is interpreted correctly by the Vast.ai API.

### Gaps Summary

No gaps found. All must-haves from all three plans are verified at all four levels (exists, substantive, wired, data flowing). The full test suite (49 unit tests across 6 test files) passes. The project compiles cleanly.

The marketplace list/unlist scope adjustment (host-only operations omitted) is correctly documented across plans, the REQUIREMENTS.md Out of Scope table, and carried forward as a known constraint. It does not constitute a gap.

---

_Verified: 2026-03-27T12:00:00Z_
_Verifier: Claude (gsd-verifier)_
