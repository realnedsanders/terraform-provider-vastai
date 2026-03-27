# Phase 3: Storage - Research

**Researched:** 2026-03-27
**Domain:** Vast.ai Volume and Network Volume Terraform Resources + Offer Data Sources
**Confidence:** HIGH

## Summary

Phase 3 implements persistent volume and network volume management through Terraform. The Vast.ai API exposes two distinct storage systems: **local volumes** (storage on a specific machine) and **network volumes** (network-attached storage that can be mounted across instances). Both follow a create-from-offer pattern identical to instances: search offers, pick one, create a volume from that offer ID.

The API contracts are well-defined in the Python SDK. Volume creation is a simple PUT with `{size, id}` (offer ID). Volume search uses POST with the same structured query format as GPU offers. The key complexity lies in the marketplace list/unlist operations, which are **host-side** operations (marked `[Host]` in the SDK), not tenant-side operations. This means they are out of scope for this provider (which targets tenants, not GPU hosts -- see REQUIREMENTS.md Out of Scope). The "list/unlist" in the requirements likely refers to Terraform list/read capabilities, not marketplace listing.

**Primary recommendation:** Implement `vastai_volume` and `vastai_network_volume` as separate resources following the template resource CRUD pattern, with `vastai_volume_offers` and `vastai_network_volume_offers` as separate data sources following the GPU offers search pattern. Volume clone uses a dedicated endpoint (`POST /volumes/copy/`). Delete volume uses query parameter `?id=X` on DELETE, which is a minor deviation from other resources.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Volume cloning via `clone_from_id` optional attribute on the volume resource -- when set, creates by cloning instead of from offer. ForceNew on change.
- **D-02:** Claude's Discretion on marketplace list/unlist modeling (boolean attribute vs separate resource) -- based on API behavior fit
- **D-03:** Separate resources for volumes and network volumes: `vastai_volume` and `vastai_network_volume` -- different API paths, different capabilities

### Inherited from Phase 1 & 2 (DO NOT re-decide)
- Service-per-directory: `internal/services/volume/`, `internal/services/networkvolume/`
- Service pattern on client: `client.Volumes.Create()`, `client.NetworkVolumes.Create()`
- Always snake_case attributes with Optional+Computed for server defaults
- Comprehensive plan-time validators
- Per-resource models, API models in `internal/client/`
- Set vs List per-attribute, Vast.ai doc URLs in descriptions
- Dual test strategy: httptest mocks + TF_ACC acceptance tests
- Full acceptance test coverage for all resources

### Claude's Discretion
- Marketplace list/unlist modeling approach
- Offer search data source structure and filter set
- Whether offers are separate or combined data sources
- Timeout defaults for volume operations

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| STOR-01 | `vastai_volume` resource with CRUD (create from offer, delete, list/unlist for marketplace) | API contracts extracted: PUT /volumes/ for create, DELETE /volumes/?id=X for delete, GET /volumes?owner=me&type=local_volume for read. List/unlist are host-only operations -- see Pitfall 1 |
| STOR-02 | `vastai_volume` supports clone operation | API contract: POST /volumes/copy/ with {src_id, dst_id, size?, disable_compression?}. Modeled as clone_from_id ForceNew attribute per D-01 |
| STOR-03 | `vastai_network_volume` resource with CRUD (create, delete, list/unlist for marketplace) | API contracts: PUT /network_volumes/ for create, GET /volumes?owner=me&type=network_volume for read. No separate delete endpoint found -- may share DELETE /volumes/?id=X. List/unlist are host-only |
| DATA-05 | `vastai_volume_offers` data source with filter support | API contract: POST /volumes/search/ with structured query. Filter fields documented. Same query engine as GPU offers |
| DATA-06 | `vastai_network_volume_offers` data source with filter support | API contract: POST /network_volumes/search/ with same structured query format. Separate endpoint, slightly different display fields |
</phase_requirements>

## Standard Stack

### Core
No new libraries needed. All dependencies are already installed from Phase 1 and Phase 2.

| Library | Version | Purpose | Already In Project |
|---------|---------|---------|-------------------|
| terraform-plugin-framework | v1.19.0 | Resource/data source framework | Yes |
| terraform-plugin-framework-timeouts | v0.5.0 | Configurable timeouts | Yes |
| terraform-plugin-framework-validators | v0.19.0 | Attribute validators | Yes |
| terraform-plugin-testing | v1.15.0 | Acceptance tests | Yes |
| terraform-plugin-log/tflog | v0.10.0 | Structured logging | Yes |

**Installation:** None needed -- all dependencies already in go.mod.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  client/
    volumes.go              # VolumeService + VolumeSearchService
    network_volumes.go      # NetworkVolumeService + NetworkVolumeSearchService
  services/
    volume/
      resource_volume.go           # vastai_volume resource
      resource_volume_test.go      # Unit tests (schema, metadata, httptest mocks)
      resource_volume_acc_test.go  # TF_ACC acceptance tests
      data_source_volume_offers.go      # vastai_volume_offers data source
      data_source_volume_offers_test.go
      data_source_volume_offers_acc_test.go
      models.go                    # VolumeResourceModel, VolumeOfferModel, etc.
    networkvolume/
      resource_network_volume.go           # vastai_network_volume resource
      resource_network_volume_test.go
      resource_network_volume_acc_test.go
      data_source_network_volume_offers.go  # vastai_network_volume_offers data source
      data_source_network_volume_offers_test.go
      data_source_network_volume_offers_acc_test.go
      models.go
  provider/
    provider.go            # Register new resources and data sources
```

### Pattern 1: Volume Resource (CRUD following Template pattern)

**What:** Resource with Create (from offer), Read (via show volumes), Update (none -- volumes appear immutable), Delete, Import.

**When to use:** Both `vastai_volume` and `vastai_network_volume`.

**API Contract -- Volume:**
```
Create:      PUT  /api/v0/volumes/          body: {size: int, id: int (offer_id), name?: string}
Read:        GET  /api/v0/volumes?owner=me&type=local_volume   -> {volumes: [...]}
Delete:      DELETE /api/v0/volumes/?id={volume_id}
Clone:       POST /api/v0/volumes/copy/     body: {src_id: int, dst_id: int, size?: float, disable_compression?: bool}
```

**API Contract -- Network Volume:**
```
Create:      PUT  /api/v0/network_volumes/  body: {size: int, id: int (offer_id), name?: string}
Read:        GET  /api/v0/volumes?owner=me&type=network_volume -> {volumes: [...]}
Delete:      DELETE /api/v0/volumes/?id={volume_id}  (assumed same endpoint -- SDK has no separate delete)
```

**Key observations from SDK:**
- `create__volume` takes `id` (offer_id), `size` (default 15 GB), optional `name`
- `create__network_volume` takes same args, different endpoint `/network_volumes/`
- `delete__volume` uses DELETE with query param `?id=X` (not body, not path param)
- `show__volumes` returns all volumes with `type` filter: "local_volume", "network_volume", "all_volume"
- No single-volume GET endpoint exists -- must use list+filter (same as SSH keys pattern)

### Pattern 2: Volume Offer Search (following GPU Offers pattern)

**What:** Data source that searches for available volume offers using structured query filters.

**When to use:** Both `vastai_volume_offers` and `vastai_network_volume_offers`.

**API Contract -- Volume Offer Search:**
```
Search:      POST /api/v0/volumes/search/          body: {query JSON} -> {offers: [...]}
```

**API Contract -- Network Volume Offer Search:**
```
Search:      POST /api/v0/network_volumes/search/  body: {query JSON} -> {offers: [...]}
```

**Query format (identical for both):**
```json
{
  "verified": {"eq": true},
  "external": {"eq": false},
  "disk_space": {"gte": 1},
  "order": [["score", "desc"]],
  "limit": 10,
  "allocated_storage": 1.0
}
```

**Filter fields available (from vol_offers_fields):**
- `disk_space` (float, GB) -- storage space
- `storage_cost` (float, $/GB/month) -- pricing
- `inet_up` / `inet_down` (float, Mb/s) -- network speed
- `reliability` (float, 0-1) -- host reliability score
- `geolocation` (string) -- two-letter country code
- `duration` (float, seconds -> days) -- max rental duration
- `verified` (bool) -- machine verified
- `static_ip` (bool) -- static IP
- `disk_bw` (float, MB/s) -- disk bandwidth
- `machine_id` (int) -- specific machine
- `id` (int) -- offer ID
- `pci_gen`, `pcie_bw` (float) -- PCI specs
- `gpu_arch`, `cpu_arch` (string) -- architecture
- `cuda_vers` (float, alias for cuda_max_good) -- CUDA version
- `total_flops` (float) -- GPU compute power
- `driver_version` (string) -- driver version
- `has_avx` (bool) -- AVX support
- `ubuntu_version` (string) -- OS version

**Network volume search** has additional fields in display: `cluster_id`, `nw_disk_min_bw`, `nw_disk_max_bw`, `nw_disk_avg_bw`.

**Response fields -- Volume Offers (from vol_displayable_fields):**
- `id`, `cuda_max_good`, `cpu_ghz`, `disk_bw`, `disk_space`, `disk_name`
- `storage_cost`, `driver_version`, `inet_up`, `inet_down`
- `reliability`, `duration`, `machine_id`, `verification`, `host_id`, `geolocation`

**Response fields -- Network Volume Offers (from nw_vol_displayable_fields):**
- `id`, `disk_space`, `storage_cost`, `inet_up`, `inet_down`
- `reliability`, `duration`, `verification`, `host_id`
- `cluster_id`, `geolocation`
- `nw_disk_min_bw`, `nw_disk_max_bw`, `nw_disk_avg_bw`

### Pattern 3: Volume Clone (creation-time attribute)

**What:** Volume clone modeled as `clone_from_id` attribute on `vastai_volume` resource. When set, uses POST /volumes/copy/ instead of PUT /volumes/.

**Clone API contract:**
```
POST /api/v0/volumes/copy/
Body: {
  "src_id": <source_volume_contract_id>,
  "dst_id": <destination_offer_id>,
  "size": <optional, GB, defaults to source size>,
  "disable_compression": <optional, bool>
}
```

**Key:** Clone requires BOTH a source volume ID AND a destination offer ID. This means:
- `clone_from_id` = source volume ID (Required when cloning)
- `offer_id` = destination offer ID (Required always -- for both create and clone)
- `disable_compression` = optional bool for clone operations
- All three attributes should be ForceNew (changing any of them requires recreate)

### Pattern 4: Read-via-List (same as SSH keys)

**What:** No single-volume GET endpoint. Must use `GET /volumes?owner=me&type=X` and filter results client-side by ID.

**Response format (from show__volumes):**
```json
{
  "volumes": [
    {
      "id": 123,
      "cluster_id": 456,
      "label": "my-volume",
      "disk_space": 50.0,
      "status": "active",
      "disk_name": "nvme0n1",
      "driver_version": "535.86.05",
      "inet_up": 1000.0,
      "inet_down": 1000.0,
      "reliability2": 0.99,
      "start_date": 1711900000.0,
      "machine_id": 789,
      "verification": "verified",
      "host_id": 101,
      "geolocation": "US",
      "instances": [1001, 1002]
    }
  ]
}
```

### Anti-Patterns to Avoid

- **Implementing list/unlist as tenant operations:** The SDK marks `list__volume`, `list__network_volume`, `unlist__volume`, `unlist__network_volume` as `[Host]` operations. The provider targets tenants, not hosts. These are out of scope per the project's Out of Scope section.
- **Assuming single GET endpoint exists:** Like SSH keys, volumes use read-via-list. Do NOT try to call GET /volumes/{id}/ -- it does not exist in the SDK.
- **Using body on DELETE:** The `delete__volume` SDK uses DELETE with query parameter `?id=X`, NOT a request body. The path passed to `client.Delete()` should include the query parameter: `/volumes/?id=123`.

## Discretion Recommendations

### D-02: Marketplace list/unlist modeling

**Recommendation: OMIT list/unlist entirely from this phase.**

Evidence: The `list__volume`, `list__network_volume`, `unlist__volume`, and `unlist__network_volume` commands in the SDK are all marked as `[Host]` operations:
- `list__volume`: `help="[Host] list disk space for rent as a volume on a machine"`
- `list__network_volume`: `help="[Host] list disk space for rent as a network volume"`
- `unlist__volume`: `help="[Host] unlist volume offer"`
- `unlist__network_volume`: `help="[Host] Unlists network volume offer"`

These are operations performed by GPU **hosts** (people who own physical machines and want to rent storage), not by **tenants** (people who rent GPU compute). The REQUIREMENTS.md Out of Scope section explicitly excludes "Host-side machine operations (list, unlist, maintenance, defjob)" with the reason "Wrong audience -- this provider targets tenants, not GPU hosts."

The STOR-01 and STOR-03 requirement text includes "list/unlist for marketplace" but this conflicts with the explicit Out of Scope ruling. The planner should note this discrepancy and implement CRUD (create, read, delete) without marketplace list/unlist. If the user later wants host-side storage management, it would be a separate scope decision.

### D-04/D-05: Volume offer data sources -- separate or combined

**Recommendation: Separate data sources: `vastai_volume_offers` and `vastai_network_volume_offers`.**

Evidence:
1. Different API endpoints: `POST /volumes/search/` vs `POST /network_volumes/search/`
2. Different response fields: Network volumes include `cluster_id`, `nw_disk_min_bw`, `nw_disk_max_bw`, `nw_disk_avg_bw` which local volumes do not have
3. Same query engine and filter fields (`vol_offers_fields` is shared)
4. Consistent with D-03 decision for separate resources
5. Mirrors the clear endpoint separation in the API

The filter attributes can share the same set (since `vol_offers_fields` is used for both), but network volume offers should expose additional computed attributes for bandwidth metrics.

### Timeout defaults for volume operations

**Recommendation: 5 minutes for create/read/delete, 30 minutes for clone.**

Rationale: Volume creation is likely fast (allocating disk space), similar to template creation. Clone operations involve data copy across machines and could take significant time depending on volume size. 30 minutes gives large volume clones time to complete. The Python SDK's clone_volume doesn't have a timeout hint, but data copy is inherently slower than metadata operations.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Volume offer query construction | Custom query builder | Replicate GPU offers `buildSearchBody` pattern | Already proven, same API query format |
| Read-via-list filtering | Custom list filter | Replicate SSH keys `List` + client-side ID match | Established pattern, same API shape |
| Timeout configuration | Custom timeout handling | `terraform-plugin-framework-timeouts` | Already in use for templates and instances |
| Attribute validation | Manual validation logic | `terraform-plugin-framework-validators` | Already in use, covers all needed patterns |

## Common Pitfalls

### Pitfall 1: List/Unlist Are Host Operations
**What goes wrong:** Implementing marketplace list/unlist as tenant operations.
**Why it happens:** The requirements text mentions "list/unlist for marketplace" and it's easy to assume these are tenant-facing.
**How to avoid:** Read the SDK carefully -- all list/unlist volume commands are marked `[Host]`. This provider targets tenants only.
**Warning signs:** The SDK help text starts with `[Host]`.

### Pitfall 2: DELETE Uses Query Parameter, Not Path Parameter
**What goes wrong:** Calling `DELETE /volumes/{id}/` instead of `DELETE /volumes/?id={id}`.
**Why it happens:** Most other resources use path parameters for DELETE (e.g., `/instances/{id}/`).
**How to avoid:** The SDK uses `apiurl(args, "/volumes/", query_args={"id": args.id})` which puts the ID in query params. Pass `/volumes/?id=123` as the path to `client.Delete()`.
**Warning signs:** 404 errors on volume delete.

### Pitfall 3: No Single-Volume Read Endpoint
**What goes wrong:** Trying to GET a single volume by ID.
**Why it happens:** Instances have `GET /instances/{id}/` but volumes do not.
**How to avoid:** Use `GET /volumes?owner=me&type=local_volume` (or `network_volume`) and filter client-side by ID. Same pattern as SSH keys.
**Warning signs:** 404 or unexpected JSON format on read.

### Pitfall 4: Clone Requires Both Source ID AND Destination Offer ID
**What goes wrong:** Modeling clone as only needing a source volume ID.
**Why it happens:** The D-01 decision says "clone_from_id" which suggests only a source, but the API requires `src_id` (source volume) AND `dst_id` (destination offer).
**How to avoid:** When `clone_from_id` is set, `offer_id` is still required (it becomes `dst_id`). The resource always requires `offer_id`.
**Warning signs:** API errors about missing `dst_id`.

### Pitfall 5: allocated_storage Query Parameter
**What goes wrong:** Omitting `allocated_storage` from search queries.
**Why it happens:** This parameter isn't in the query filter fields -- it's a top-level parameter added separately.
**How to avoid:** The SDK always adds `query["allocated_storage"] = args.storage` (default 1.0 GiB). Include it in search requests. It affects pricing calculations.
**Warning signs:** Unexpected pricing in search results.

### Pitfall 6: Volume Type Filter on Show
**What goes wrong:** Getting both local and network volumes mixed together.
**Why it happens:** The `show__volumes` endpoint supports a `type` parameter: `local_volume`, `network_volume`, `all_volume` (maps to `all`).
**How to avoid:** Always pass the appropriate `type` query parameter when reading volumes. For `vastai_volume`, use `type=local_volume`. For `vastai_network_volume`, use `type=network_volume`.
**Warning signs:** Reading a network volume resource returns a local volume, or vice versa.

### Pitfall 7: Volume Size is Integer for Create, Float for Search
**What goes wrong:** Sending float size on create.
**Why it happens:** Search uses float for `allocated_storage` but create uses `int(args.size)`.
**How to avoid:** Cast size to int for create API calls. The Go client should use `int` for the create request struct's Size field.
**Warning signs:** API validation errors on create.

## Code Examples

### Volume Client Service
```go
// Source: Extracted from vast-sdk/vastai/vast.py create__volume, delete__volume, show__volumes
package client

type VolumeService struct {
    client *VastAIClient
}

type CreateVolumeRequest struct {
    Size    int    `json:"size"`
    OfferID int    `json:"id"`       // Offer ID from search
    Name    string `json:"name,omitempty"`
}

type CloneVolumeRequest struct {
    SourceID           int     `json:"src_id"`
    DestOfferID        int     `json:"dst_id"`
    Size               float64 `json:"size,omitempty"`
    DisableCompression bool    `json:"disable_compression,omitempty"`
}

type Volume struct {
    ID             int       `json:"id"`
    ClusterID      int       `json:"cluster_id"`
    Label          string    `json:"label"`     // Name/label
    DiskSpace      float64   `json:"disk_space"`
    Status         string    `json:"status"`
    DiskName       string    `json:"disk_name"`
    DriverVersion  string    `json:"driver_version"`
    InetUp         float64   `json:"inet_up"`
    InetDown       float64   `json:"inet_down"`
    Reliability    float64   `json:"reliability2"`
    StartDate      float64   `json:"start_date"`
    MachineID      int       `json:"machine_id"`
    Verification   string    `json:"verification"`
    HostID         int       `json:"host_id"`
    Geolocation    string    `json:"geolocation"`
    Instances      []int     `json:"instances"`
}

type volumeListResponse struct {
    Volumes []Volume `json:"volumes"`
}

// Create creates a volume from an offer.
// PUT /volumes/ with {size, id, name?}
func (s *VolumeService) Create(ctx context.Context, req *CreateVolumeRequest) (*CreateVolumeResponse, error) {
    // ...
}

// Clone creates a volume by cloning an existing one.
// POST /volumes/copy/ with {src_id, dst_id, size?, disable_compression?}
func (s *VolumeService) Clone(ctx context.Context, req *CloneVolumeRequest) (*CloneVolumeResponse, error) {
    // ...
}

// List retrieves all volumes owned by the user, filtered by type.
// GET /volumes?owner=me&type={volumeType}
func (s *VolumeService) List(ctx context.Context, volumeType string) ([]Volume, error) {
    path := fmt.Sprintf("/volumes?owner=me&type=%s", volumeType)
    // ...
}

// Delete deletes a volume by ID.
// DELETE /volumes/?id={id}
func (s *VolumeService) Delete(ctx context.Context, id int) error {
    path := fmt.Sprintf("/volumes/?id=%d", id)
    return s.client.Delete(ctx, path, nil)
}
```

### Volume Offer Search Service
```go
// Source: Extracted from vast-sdk/vastai/vast.py search__volumes
type VolumeOfferSearchParams struct {
    DiskSpace      *float64 `json:"-"` // Minimum disk space in GB
    StorageCost    *float64 `json:"-"` // Max $/GB/month
    InetUp         *float64 `json:"-"` // Min upload Mb/s
    InetDown       *float64 `json:"-"` // Min download Mb/s
    Reliability    *float64 `json:"-"` // Min reliability (0-1)
    Geolocation    string   `json:"-"` // Country code
    Verified       *bool    `json:"-"` // Machine verified
    StaticIP       *bool    `json:"-"` // Static IP
    DiskBW         *float64 `json:"-"` // Min disk bandwidth MB/s
    OrderBy        string   `json:"-"` // Sort field
    Limit          int      `json:"-"` // Max results
    AllocatedStorage float64 `json:"-"` // Storage amount for pricing (default 1.0)
    RawQuery       string   `json:"-"` // Bypass structured filters
}

type VolumeOffer struct {
    ID             int     `json:"id"`
    CUDAMaxGood    float64 `json:"cuda_max_good"`
    CPUGhz         float64 `json:"cpu_ghz"`
    DiskBW         float64 `json:"disk_bw"`
    DiskSpace      float64 `json:"disk_space"`
    DiskName       string  `json:"disk_name"`
    StorageCost    float64 `json:"storage_cost"`
    DriverVersion  string  `json:"driver_version"`
    InetUp         float64 `json:"inet_up"`
    InetDown       float64 `json:"inet_down"`
    Reliability    float64 `json:"reliability"`
    Duration       float64 `json:"duration"`
    MachineID      int     `json:"machine_id"`
    Verification   string  `json:"verification"`
    HostID         int     `json:"host_id"`
    Geolocation    string  `json:"geolocation"`
}

// Search searches for volume offers.
// POST /volumes/search/ with structured query
func (s *VolumeService) SearchOffers(ctx context.Context, params *VolumeOfferSearchParams) ([]VolumeOffer, error) {
    body := s.buildSearchBody(params)
    var resp struct {
        Offers []VolumeOffer `json:"offers"`
    }
    if err := s.client.Post(ctx, "/volumes/search/", body, &resp); err != nil {
        return nil, fmt.Errorf("searching volume offers: %w", err)
    }
    return resp.Offers, nil
}
```

### Network Volume Offer (additional fields)
```go
// Source: Extracted from vast-sdk/vastai/vast.py nw_vol_displayable_fields
type NetworkVolumeOffer struct {
    ID           int     `json:"id"`
    DiskSpace    float64 `json:"disk_space"`
    StorageCost  float64 `json:"storage_cost"`
    InetUp       float64 `json:"inet_up"`
    InetDown     float64 `json:"inet_down"`
    Reliability  float64 `json:"reliability"`
    Duration     float64 `json:"duration"`
    Verification string  `json:"verification"`
    HostID       int     `json:"host_id"`
    ClusterID    int     `json:"cluster_id"`
    Geolocation  string  `json:"geolocation"`
    // Network-volume-specific bandwidth metrics
    NWDiskMinBW  float64 `json:"nw_disk_min_bw"`
    NWDiskMaxBW  float64 `json:"nw_disk_max_bw"`
    NWDiskAvgBW  float64 `json:"nw_disk_avg_bw"`
}
```

### Volume Resource Schema (key attributes)
```go
// Source: Derived from SDK create__volume args + show__volumes fields
schema.Schema{
    Attributes: map[string]schema.Attribute{
        "id": schema.Int64Attribute{
            Description: "Volume contract ID.",
            Computed:    true,
            PlanModifiers: []planmodifier.Int64{
                int64planmodifier.UseStateForUnknown(),
            },
        },
        "offer_id": schema.Int64Attribute{
            Description: "ID of the volume offer to create from (from vastai_volume_offers).",
            Required:    true,
            PlanModifiers: []planmodifier.Int64{
                int64planmodifier.RequiresReplace(),
            },
        },
        "size": schema.Int64Attribute{
            Description: "Volume size in GB.",
            Required:    true,
            PlanModifiers: []planmodifier.Int64{
                int64planmodifier.RequiresReplace(),
            },
            Validators: []validator.Int64{
                int64validator.AtLeast(1),
            },
        },
        "name": schema.StringAttribute{
            Description: "Optional name for the volume.",
            Optional:    true,
        },
        "clone_from_id": schema.Int64Attribute{
            Description: "Source volume ID to clone from. When set, creates via clone (POST /volumes/copy/). ForceNew.",
            Optional:    true,
            PlanModifiers: []planmodifier.Int64{
                int64planmodifier.RequiresReplace(),
            },
        },
        "disable_compression": schema.BoolAttribute{
            Description: "Disable compression during clone. Only used with clone_from_id.",
            Optional:    true,
        },
        // Computed fields from show__volumes
        "status":       schema.StringAttribute{Computed: true, Description: "Volume status."},
        "disk_space":   schema.Float64Attribute{Computed: true, Description: "Actual disk space in GB."},
        "machine_id":   schema.Int64Attribute{Computed: true, Description: "Machine hosting this volume."},
        "geolocation":  schema.StringAttribute{Computed: true, Description: "Geographic location."},
        "inet_up":      schema.Float64Attribute{Computed: true, Description: "Upload speed in Mb/s."},
        "inet_down":    schema.Float64Attribute{Computed: true, Description: "Download speed in Mb/s."},
        "reliability":  schema.Float64Attribute{Computed: true, Description: "Host reliability score."},
        // ... more computed fields
    },
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single volume type | Separate local + network volumes | Current API | Must implement as separate resources |
| GPU-only offers | Volume + Network Volume offer search | Current API | Need two new data sources |

## Open Questions

1. **Network volume deletion endpoint**
   - What we know: The SDK has `delete__volume` for local volumes using `DELETE /volumes/?id=X`. There is no `delete__network_volume` function in the SDK.
   - What's unclear: Whether network volumes are deleted via the same endpoint or a different one.
   - Recommendation: Try `DELETE /volumes/?id=X` for network volumes as well (the show endpoint already uses `/volumes?type=network_volume` for both types). If this fails during acceptance testing, investigate `/network_volumes/?id=X` as an alternative. **LOW confidence** on this specific point.

2. **Create response format**
   - What we know: The SDK prints `r.json()` after create but doesn't parse specific fields.
   - What's unclear: Exact JSON structure returned by PUT /volumes/ and PUT /network_volumes/.
   - Recommendation: Assume response includes at least `{id, success}` similar to instance creation. After create, immediately call List to get full volume details. Validate during acceptance testing.

3. **Volume update capabilities**
   - What we know: The SDK has no `update__volume` or `modify__volume` command.
   - What's unclear: Whether any volume attributes can be updated in-place.
   - Recommendation: Treat volumes as immutable (all mutable attributes trigger ForceNew). The only updateable field might be `name` -- test during acceptance testing. If name is updateable, add a separate Update path for it.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | terraform-plugin-testing v1.15.0 + Go testing |
| Config file | None needed -- uses Go test conventions |
| Quick run command | `go test ./internal/services/volume/... ./internal/services/networkvolume/... ./internal/client/... -v -count=1` |
| Full suite command | `go test ./... -v -count=1` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| STOR-01 | Volume CRUD (create from offer, read, delete) | unit + acceptance | `go test ./internal/services/volume/... -v -run TestVolumeResource` | Wave 0 |
| STOR-02 | Volume clone via clone_from_id | unit + acceptance | `go test ./internal/services/volume/... -v -run TestVolumeResource.*clone` | Wave 0 |
| STOR-03 | Network volume CRUD | unit + acceptance | `go test ./internal/services/networkvolume/... -v -run TestNetworkVolumeResource` | Wave 0 |
| DATA-05 | Volume offers search with filters | unit + acceptance | `go test ./internal/services/volume/... -v -run TestVolumeOffers` | Wave 0 |
| DATA-06 | Network volume offers search | unit + acceptance | `go test ./internal/services/networkvolume/... -v -run TestNetworkVolumeOffers` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/services/volume/... ./internal/services/networkvolume/... ./internal/client/... -v -count=1`
- **Per wave merge:** `go test ./... -v -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/services/volume/resource_volume_test.go` -- covers STOR-01, STOR-02
- [ ] `internal/services/volume/resource_volume_acc_test.go` -- covers STOR-01, STOR-02
- [ ] `internal/services/volume/data_source_volume_offers_test.go` -- covers DATA-05
- [ ] `internal/services/volume/data_source_volume_offers_acc_test.go` -- covers DATA-05
- [ ] `internal/services/networkvolume/resource_network_volume_test.go` -- covers STOR-03
- [ ] `internal/services/networkvolume/resource_network_volume_acc_test.go` -- covers STOR-03
- [ ] `internal/services/networkvolume/data_source_network_volume_offers_test.go` -- covers DATA-06
- [ ] `internal/services/networkvolume/data_source_network_volume_offers_acc_test.go` -- covers DATA-06

## Sources

### Primary (HIGH confidence)
- `vast-sdk/vastai/vast.py` -- Full API contracts for all volume operations (lines 1540, 2773, 2807, 3068, 4587, 4684, 5854, 7373, 7410, 8094, 8117)
- `internal/services/offer/data_source_gpu_offers.go` -- GPU offer search pattern to replicate
- `internal/services/template/resource_template.go` -- Resource CRUD pattern to replicate
- `internal/client/offers.go` -- Client service and search body construction pattern
- `internal/client/instances.go` -- Service struct pattern with typed methods
- `internal/client/client.go` -- Client infrastructure (Delete, Post, Put, Get methods)

### Secondary (MEDIUM confidence)
- Volume field names from `vol_displayable_fields`, `nw_vol_displayable_fields`, `volume_fields` in vast.py (lines 741-850) -- display fields may not cover all API response fields

### Tertiary (LOW confidence)
- Network volume deletion endpoint -- assumed same as local volume DELETE based on shared show endpoint, needs validation
- Create response format -- SDK doesn't parse response fields, needs live API testing
- Volume update capabilities -- no SDK evidence of updates, assumed immutable

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all patterns established
- Architecture: HIGH -- clear API contracts from SDK, established patterns from Phase 2
- Pitfalls: HIGH -- derived directly from SDK code analysis
- Volume offers: HIGH -- well-defined search API with documented fields
- Network volume delete: LOW -- no SDK function found, extrapolating from shared endpoint
- Create response: MEDIUM -- SDK uses r.json() but doesn't parse specific fields

**Research date:** 2026-03-27
**Valid until:** 2026-04-27 (stable -- Vast.ai API rarely changes)
