---
status: awaiting_human_verify
trigger: "Implement GitHub issue #3: correct destroy completion and nullable show-instance absence semantics"
created: 2026-09-20T02:05:08+00:00
updated: 2026-09-20T02:15:23+00:00
---

## Current Focus
<!-- OVERWRITE on each update - reflects NOW -->

hypothesis: Confirmed fix passes all automated verification signals.
test: Parent/user review and, when credentials are available, a live Terraform destroy against Vast.ai.
expecting: DELETE 200/404 completes without any GET poll; a later Read sees 200-null as absence; populated blank/null status remains present.
next_action: Await human review or live-environment confirmation before archiving the GSD debug session.
reasoning_checkpoint:
  hypothesis: "The pointerless instanceGetWrapper loses the JSON null/presence distinction, while InstanceResource.Delete adds an invalid asynchronous poll after a completed synchronous DELETE response."
  confirming_evidence:
    - 'A direct Go probe decodes both {instances:null} and a missing instances field into Instance{ID:0, ActualStatus:""} without error.'
    - "destroy_instance.yaml defines HTTP 200 as 'Instance destroyed successfully'; the official CLI returns the DELETE JSON directly and performs no follow-up status poll."
    - "The official CLI show_instance explicitly maps response instances:null to None, while the provider returns &wrapper.Instances."
  falsification_test: "The hypothesis would be false if focused tests showed null already produces a not-found error or resource Delete makes only the DELETE request."
  fix_rationale: "Preserving the nullable JSON shape and returning a domain absence sentinel fixes the bad value at the decode boundary; removing the post-200 poll aligns lifecycle behavior with the endpoint contract."
  blind_spots: "No live Vast.ai acceptance run is available; malformed payload policy and Terraform framework state-removal behavior must be covered locally."
  candidate_causes:
    - "code: non-pointer response wrapper and post-DELETE WaitForStatus call"
    - "data: API represents absent instance as HTTP 200 with instances:null and allows nullable actual_status"
    - "config/environment: five-minute delete timeout only exposes the poll mismatch; it does not cause it"
  and_gate: "yes: the visible destroy delay requires both a synchronous deletion contract and the provider's extra status poll; the ID-zero bug independently requires nullable data plus lossy decoding"

## Symptoms
<!-- Written during gathering, then IMMUTABLE -->

expected: DELETE 200 and 404 complete successfully without polling; GET 200 with `{"instances":null}` yields a truthful typed absent condition shared by resource Read and data source; populated instances with blank/null actual_status remain present.
actual: Destroy performs a five-minute destroyed-status poll after DELETE; nullable show responses decode toward an ID-zero instance instead of absence.
errors: Terraform destroy can wait five minutes or fail after the API has completed deletion; missing instances can be represented in state as instance ID 0.
reproduction: Exercise DELETE 200/404, and decode GET `/instances/{id}` responses with null versus populated `instances` values.
started: Present in the current implementation; exact introducing commit not yet investigated.

## Eliminated
<!-- APPEND only - prevents re-investigating -->

## Evidence
<!-- APPEND only - facts discovered -->

- timestamp: 2026-09-20T02:08:19+00:00
  checked: Baseline focused tests
  found: `go test ./internal/client ./internal/services/instance` passed before regression tests.
  implication: Existing coverage does not exercise nullable show responses or resource Delete request count.
- timestamp: 2026-09-20T02:08:19+00:00
  checked: OpenAPI destroy/show schemas and official Vast CLI implementation
  found: DELETE 200 is documented as completed success; CLI destroy returns the DELETE response without polling; CLI show returns None when `instances` is null; `actual_status` is nullable.
  implication: Empty/null status is valid instance data and cannot prove absence; the nullable wrapper itself is the absence discriminator.
- timestamp: 2026-09-20T02:08:19+00:00
  checked: Go JSON decoding and provider call paths
  found: A direct Go probe produced ID 0 for both `{"instances":null}` and a missing wrapper field; resource Delete calls WaitForStatus after successful Destroy; resource Read only recognizes HTTP 404 and the data source has no absence branch.
  implication: Absence must be preserved at InstanceService.Get and shared truthfully without fabricating an HTTP 404.

- timestamp: 2026-09-20T02:11:17+00:00
  checked: Focused RED regression run
  found: Client tests failed to compile because ErrInstanceNotFound does not exist; data-source null produced no diagnostic; resource Read wrote ID 0; resource Delete made 2 HTTP requests instead of 1.
  implication: The tests independently reproduce each required behavior before production changes.

- timestamp: 2026-09-20T02:13:36+00:00
  checked: Focused GREEN regression run
  found: `go test ./internal/client ./internal/services/instance` passed after implementing RawMessage decoding, ErrInstanceNotFound/IsInstanceNotFound, idempotent Destroy, consumer handling, and removal of the resource delete poll.
  implication: The exact RED reproductions now pass while existing start/create polling tests remain green.

- timestamp: 2026-09-20T02:15:23+00:00
  checked: Final verification guardrail
  found: Focused tests passed uncached; `go test ./...`, `go vet ./...`, and `git diff --check` all passed; source search found no resource destroyed-status poll or empty-status absence heuristic.
  implication: Automated verification accepts the fix; only live Vast.ai acceptance testing remains unavailable without account credentials.

## Resolution
<!-- OVERWRITE as understanding evolves -->

root_cause: InstanceResource.Delete treated synchronous DELETE 200 as async and polled; InstanceService.Get decoded nullable `instances` into a zero-value struct, losing absence.
fix: Preserve the `instances` envelope with json.RawMessage; return ErrInstanceNotFound for explicit null; share IsInstanceNotFound across resource Read, data source Read, Destroy, WaitForStatus, and acceptance checks; remove the post-DELETE poll.
verification: Focused RED reproduced compile/ID-zero/extra-GET/data-source failures; focused GREEN passed uncached; full tests, vet, formatting, and diff checks passed. Live acceptance not run because it requires a real Vast.ai account/API key and billable GPU lifecycle.
files_changed:
  - internal/client/instances.go
  - internal/client/instances_test.go
  - internal/services/instance/resource_instance.go
  - internal/services/instance/resource_instance_test.go
  - internal/services/instance/data_source_instance.go
  - internal/services/instance/data_source_instance_test.go
  - internal/services/instance/resource_instance_acc_test.go
