---
phase: 01-foundation
plan: 02
subsystem: api
tags: [go, http-client, retryablehttp, bearer-auth, exponential-backoff, tflog]

# Dependency graph
requires:
  - phase: 01-foundation-01
    provides: "Compilable Terraform provider binary with VastaiProvider struct and schema"
provides:
  - "VastAIClient with Bearer auth, exponential backoff retry, structured APIError"
  - "HTTP methods Get, Post, Put, Delete wired through retryablehttp"
  - "Provider Configure injects VastAIClient as ResourceData and DataSourceData"
  - "20 unit tests covering auth, retry, backoff, error handling"
affects: [01-03, all-subsequent-phases]

# Tech tracking
tech-stack:
  added: [hashicorp/go-retryablehttp v0.7.8]
  patterns: [bearer-auth-only, exponential-backoff-1.5x, structured-api-errors, retry-after-header, tflog-debug-trace-levels]

key-files:
  created: [internal/client/client.go, internal/client/auth.go, internal/client/errors.go, internal/client/client_test.go]
  modified: [internal/provider/provider.go, go.mod, go.sum]

key-decisions:
  - "Bearer auth only (never query params) per D-09 and FOUND-05 -- prevents credential leaks in logs"
  - "150ms base, 1.5x multiplier, 5 max retries matching Python SDK battle-tested config per D-07"
  - "Retry 429 and 5xx (not 501 Not Implemented) -- aligns with Vast.ai API behavior"
  - "Retry-After header respected on 429 for server-directed backoff"
  - "tflog.Debug for request/response, tflog.Trace for response body per D-08"
  - "Overflow guard in backoff prevents negative duration from math.Pow overflow"

patterns-established:
  - "Bearer auth: Authorization header set in newRequest, never in URL"
  - "Structured errors: APIError{StatusCode, Method, Path, Message} on all 4xx/5xx"
  - "Error message extraction: tries {error} then {msg} JSON fields"
  - "Client injection: resp.ResourceData = c / resp.DataSourceData = c pattern"
  - "Retry policy: context check -> connection error -> 429 -> 5xx (not 501) -> no retry"

requirements-completed: [FOUND-03, FOUND-04, FOUND-05, FOUND-06]

# Metrics
duration: 6min
completed: 2026-03-25
---

# Phase 1 Plan 2: API Client Summary

**REST API client with Bearer auth, 150ms/1.5x exponential backoff, structured APIError, and provider wiring via go-retryablehttp**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-25T19:21:04Z
- **Completed:** 2026-03-25T19:27:22Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- VastAIClient authenticates via Bearer header exclusively, never exposing API key in URLs
- Exponential backoff retry (150ms base, 1.5x multiplier, max 5 retries) with Retry-After header support on 429 responses
- Structured APIError with status code, method, path, and message on all non-2xx responses
- Provider Configure creates client and injects as ResourceData/DataSourceData for all future resources and data sources
- 20 unit tests pass covering auth headers, retry policy, backoff math, error handling, and httptest integration

## Task Commits

Each task was committed atomically:

1. **Task 1: Create API client with auth, retry, backoff, errors, and unit tests** - `fb3f96a` (test: RED), `000e89c` (feat: GREEN)
2. **Task 2: Wire API client into provider Configure method** - `87087c6` (feat)

_Note: Task 1 used TDD with separate RED/GREEN commits._

## Files Created/Modified
- `internal/client/client.go` - VastAIClient struct, NewVastAIClient constructor, retry policy, backoff, do/Get/Post/Put/Delete methods
- `internal/client/auth.go` - newRequest with Bearer auth, User-Agent, Content-Type, Accept headers
- `internal/client/errors.go` - APIError struct with Error() method
- `internal/client/client_test.go` - 20 unit tests using net/http/httptest
- `internal/provider/provider.go` - Updated Configure to create and inject VastAIClient
- `go.mod` - Added go-retryablehttp v0.7.8 dependency
- `go.sum` - Updated dependency checksums

## Decisions Made
- Bearer auth only (never query params) per D-09 and FOUND-05 to prevent credential leaks in logs
- 150ms base, 1.5x multiplier matching Python SDK battle-tested values per D-07
- Retry 429 and 5xx (not 501) -- 501 Not Implemented is not transient
- Overflow guard added in backoff function to prevent negative duration from math.Pow overflow at high attempt counts
- Error message extraction tries {"error": "..."} then {"msg": "..."} JSON patterns matching Vast.ai API response formats

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed floating-point overflow in vastaiBackoff at high attempt numbers**
- **Found during:** Task 1 (GREEN phase)
- **Issue:** math.Pow(1.5, 100) overflows float64 when multiplied by min duration, producing negative time.Duration
- **Fix:** Added math.IsInf/math.IsNaN guard and float64(max) comparison before duration cast
- **Files modified:** internal/client/client.go
- **Verification:** TestBackoff_MaxCap passes with attempt=100
- **Committed in:** 000e89c (Task 1 GREEN commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Essential correctness fix for edge case in backoff math. No scope creep.

## Issues Encountered
- retryablehttp.Request wraps body for replayability; test needed to use req.Request.GetBody() instead of direct io.ReadAll(req.Body) -- adjusted test approach during GREEN phase

## Known Stubs

None -- all functionality fully implemented.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- API client fully functional with auth, retry, and error handling
- Provider Configure injects client for all resources and data sources to consume
- Ready for Plan 01-03 (CI/CD and release pipeline) and Phase 2+ resource implementation

## Self-Check: PASSED

All 6 created/modified files verified present. All 3 task commits (fb3f96a, 000e89c, 87087c6) verified in git log. 20/20 tests pass. Build clean.

---
*Phase: 01-foundation*
*Completed: 2026-03-25*
