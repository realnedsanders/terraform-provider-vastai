---
status: resolved
trigger: "GitHub issue #2: preserve Vast.ai create-instance API error details and classify 404/410 as an unavailable offer"
created: 2026-09-20T02:06:08+00:00
updated: 2026-09-20T02:18:46+00:00
---

## Current Focus
<!-- OVERWRITE on each update - reflects NOW -->

bug_class: Bohrbug (deterministic data-shape/API-contract and error-handling defect)
reasoning_checkpoint:
  hypothesis: "Two code defects cause the misleading create diagnostic: InstanceResource.Create classifies every HTTP 400 as offer expiry, while extractErrorMessage selects the machine `error` field before the human `msg`; HTTP 410 is omitted from the true unavailable set."
  confirming_evidence:
    - "internal/services/instance/resource_instance.go:456-470 maps APIError status 404 OR 400 to the fixed Offer No Longer Available diagnostic, discarding apiErr.Message; the issue's HTTP 400 necessarily takes this branch."
    - "A no-source-change httptest probe through InstanceService.Create reproduced APIError.Message=`invalid_args` for the issue body containing error=invalid_args and msg=invalid env arguments..., matching client.go:229-233's error-before-msg order."
    - "The current OpenAPI distinguishes 400 Bad Request (lines 181-202) from unavailable 404 (228-246) and 410 (247-265), so the implemented 400/404 set is not the API contract's unavailable set."
  falsification_test: "Serve the reported dual-field body with HTTP 400 through create: the hypothesis would be false if current code produced a normal Error Creating Instance diagnostic containing the nonempty `msg`, or if 410 already produced Offer No Longer Available. Static branch tracing and the client probe show the opposite."
  fix_rationale: "Restrict offer-unavailable classification to status 404/410 and make generic extraction choose the first nonempty human field (`msg`, `message`, `detail`) before the machine `error`; then 400 reaches the normal diagnostic with useful text while documented unavailable statuses retain guidance."
  blind_spots: "No real Vast.ai request was sent and no Terraform resource-level diagnostic harness was added in this investigation-only pass. The issue report supplies the real TRACE response/output; httptest directly verified the client half."
  candidate_causes:
    - "code: wrong status predicate in InstanceResource.Create plus wrong extraction precedence/coverage in extractErrorMessage"
    - "config/specification: phase-2 research/plan incorrectly prescribed treating both 404 and 400 as expiry, while the current OpenAPI documents 400 as general Bad Request and 410 as unavailable"
    - "data: Vast.ai error bodies legitimately contain both machine `error` and human `msg`, and may instead contain `message`, `detail`, or plain text"
  and_gate: "yes: excluding 400 from the expiry branch alone would expose only `invalid_args`; preferring `msg` alone would still be masked by the resource's fixed 400 diagnostic. Both changes are needed for the reported 400 to show its human reason. Replacing omitted 400 with documented 410 is additionally needed for correct unavailable classification."
tdd_checkpoint:
  test_file: "internal/client/client_test.go, internal/client/instances_test.go, internal/services/instance/resource_instance_test.go"
  test_name: "TestExtractErrorMessage, TestInstanceService_Create_BadRequestUsesHumanMessage, TestInstanceResourceCreate_APIErrorDiagnostics"
  failure_summary: "RED confirmed: msg/message/detail precedence failed; APIError lacked Code; 400 was classified as unavailable; 410 was classified as generic create failure. The documented 404 case already followed unavailable guidance."
  command: "go test -count=1 ./internal/client ./internal/services/instance"
  exit_code: 1
next_action: "Resolved: report changed files and red/green verification without committing or pushing."

## Symptoms
<!-- Written during gathering, then IMMUTABLE -->

expected: PUT /asks/{id} HTTP 400 reports the human validation `msg`; HTTP 404 and 410 report Offer No Longer Available guidance; extraction prefers nonempty `msg`, then `message`, `detail`, `error`, then raw body; machine error code may remain separately available.
actual: Current create error handling does not reliably expose human `msg` and does not map status 404/410 to offer-unavailable guidance.
errors: Vast.ai create-instance error responses can include both a machine error code and a human-readable `msg`; plain/non-JSON bodies also occur.
reproduction: Exercise the Go client CreateInstance path against HTTP 400, 404, and 410 test responses, plus the generic API error extractor against each documented field and a plain body.
started: Existing behavior on branch fix/issue-2-create-errors; issue #2 reports it.

## Eliminated
<!-- APPEND only - prevents re-investigating -->

- hypothesis: "The HTTP/retry layer changes the 400 into an unavailable error."
  evidence: "vastaiRetryPolicy does not retry ordinary 4xx; VastAIClient.do returns an APIError with the original status/body-derived message, and InstanceService.Create only wraps it with %w. The misleading classification occurs later in InstanceResource.Create."
  timestamp: 2026-09-20T02:09:02+00:00
- hypothesis: "Changing only the create-resource 400 predicate will fully surface the human validation text."
  evidence: "The probe returned Message=`invalid_args` for a body that also had msg=`invalid env arguments, total length > 32KB` because extractErrorMessage checks `error` first."
  timestamp: 2026-09-20T02:09:02+00:00
- hypothesis: "Changing only generic error-message precedence will fix the issue."
  evidence: "InstanceResource.Create discards APIError.Message for every status 400 and emits a fixed Offer No Longer Available diagnostic, so a corrected message would remain hidden."
  timestamp: 2026-09-20T02:09:02+00:00
- hypothesis: "The API contract defines HTTP 400 as offer unavailable."
  evidence: "create_instance.yaml labels 400 as Bad Request with invalid_args/invalid_price/no_ssh_key_for_vm, while it separately labels 404 and 410 as unavailable."
  timestamp: 2026-09-20T02:09:02+00:00

## Evidence
<!-- APPEND only - facts discovered -->

- timestamp: 2026-09-20T02:09:02+00:00
  checked: "GitHub issue #2 via gh issue view"
  found: "A real PUT /asks/28705908 returned HTTP 400 with {success:false,error:invalid_args,msg:invalid env arguments, total length > 32KB}, but Terraform showed Offer No Longer Available."
  implication: "The offer diagnostic replaced a genuine validation reason; this is not merely hypothetical extractor behavior."
- timestamp: 2026-09-20T02:09:02+00:00
  checked: "internal/services/instance/resource_instance.go:456-470"
  found: "Create uses errors.As to get *client.APIError, then treats StatusCode == 404 || StatusCode == 400 as offer expiry and returns a fixed diagnostic without apiErr.Message."
  implication: "Every create 400, regardless of its body or machine code, is deterministically misclassified and its details are discarded."
- timestamp: 2026-09-20T02:09:02+00:00
  checked: "internal/client/client.go:177-185,221-236 and internal/client/errors.go:5-15"
  found: "Both normal and raw HTTP paths call extractErrorMessage; it returns any string `error` before checking `msg`, accepts empty strings, does not recognize `message` or `detail`, and APIError has no separate machine-code field."
  implication: "Even after 400 stops being classified as expiry, the reported dual-field body becomes `invalid_args`, not its human `msg`; some documented JSON errors degrade to raw JSON."
- timestamp: 2026-09-20T02:09:02+00:00
  checked: "No-source-change Go httptest probe through client.InstanceService.Create"
  found: "For the issue body, Status=400 Message=`invalid_args`. message-only returned the entire JSON object, detail-only returned the entire JSON object, empty error plus nonempty msg returned an empty message, and a plain body was preserved verbatim."
  implication: "The extractor defects and raw-body fallback are directly reproduced through the production client call path."
- timestamp: 2026-09-20T02:09:02+00:00
  checked: "/home/user/code/github.com/vast-ai/docs/api-reference/openapi/yaml/create_instance.yaml:181-275"
  found: "400 is Bad Request and contains machine `error` plus human `msg`; 404 is Offer not found/not available; 410 is Offer no longer available; 429 uses `detail`."
  implication: "Correct status classification is 404/410, and useful generic extraction must cover `msg` and `detail` while retaining raw text fallback."
- timestamp: 2026-09-20T02:09:02+00:00
  checked: "/home/user/code/github.com/vast-ai/docs/api-reference/rate-limits-and-errors.mdx:9-20,35-45"
  found: "The common shape has both error and msg, some endpoints use msg or message, and rate-limit responses may be plain text."
  implication: "Human-first field precedence and non-JSON fallback are cross-endpoint requirements, not a create-only special case."
- timestamp: 2026-09-20T02:09:02+00:00
  checked: "git history and phase planning"
  found: "The faulty resource branch originated in commit 3340250c; .planning/phases/02-core-compute/02-RESEARCH.md:576-579 and 02-04-PLAN.md prescribed 404/400 as expiry. extractErrorMessage's error-before-msg behavior originated in client commit 000e89c8 and remains unchanged."
  implication: "The runtime behavior follows an earlier inaccurate planning assumption and lacks later contract-aligned regression coverage."
- timestamp: 2026-09-20T02:09:02+00:00
  checked: "current tests and baseline"
  found: "go test -count=1 ./internal/client/... ./internal/services/instance/... passes. TestGet_APIError covers only an error-only JSON body; InstanceService Create covers only success; resource_instance_test.go has no create-error diagnostic tests."
  implication: "Existing tests do not cover conflicting error/msg fields, alternate fields/plain bodies, or 400/404/410 classification, so they pass despite the bug."

- timestamp: 2026-09-20T02:18:46+00:00
  checked: "Focused red/green tests plus full Go validation"
  found: "The new focused tests failed before production edits and passed after the surgical fix. go test -count=1 ./... and go vet ./... both passed."
  implication: "The regression is covered at extraction, wrapped client error, and Terraform resource diagnostic boundaries without retry changes."

## Resolution
<!-- OVERWRITE as understanding evolves -->

root_cause: "An AND-gated pair of error-handling defects: internal/services/instance/resource_instance.go wrongly equated every create HTTP 400 with an expired offer and omitted documented 410; internal/client/client.go prioritized machine `error` over human `msg`, accepted blank values, and missed documented `message`/`detail` fields."
fix: "The client now selects the first nonblank msg, message, detail, or error value and falls back to the raw body, while preserving the machine `error` in APIError.Code. Instance creation now classifies only HTTP 404 and 410 as unavailable; HTTP 400 uses the normal diagnostic and surfaces its human message."
verification: "RED: go test -count=1 ./internal/client ./internal/services/instance failed on message precedence, missing APIError.Code, 400 classification, and 410 guidance. GREEN: the same focused command passed after the fix. Full go test -count=1 ./... passed. go vet ./... passed. gofmt and git diff --check passed."
files_changed:
  - internal/client/client.go
  - internal/client/client_test.go
  - internal/client/errors.go
  - internal/client/instances_test.go
  - internal/services/instance/resource_instance.go
  - internal/services/instance/resource_instance_test.go
  - .planning/debug/resolved/issue-2-create-errors.md
