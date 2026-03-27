# Phase 4: Serverless - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.

**Date:** 2026-03-27
**Phase:** 04-serverless
**Areas discussed:** Endpoint/worker model, Autoscaling params

---

## Endpoint/Worker Model

### Resource Separation
| Option | Description | Selected |
|--------|-------------|----------|
| All separate | vastai_endpoint, vastai_worker_group, vastai_autogroup | |
| Endpoint + workers | Combined autogroup as attribute on worker group | |
| You decide | Claude picks based on API structure | ✓ |

### Cascade Deletion
| Option | Description | Selected |
|--------|-------------|----------|
| Explicit deletion | Workers deleted separately before endpoint | ✓ |
| Cascade via API | Endpoint delete cascades to workers | |
| You decide | Based on API behavior | |

**User's choice:** Explicit deletion — Terraform manages dependencies

### Status Data Source Depth
| Option | Description | Selected |
|--------|-------------|----------|
| Full status | Metadata + worker health/load | |
| Metadata only | Just endpoint info | |
| You decide | Based on API response | ✓ |

---

## Autoscaling Params

### Required vs Optional
| Option | Description | Selected |
|--------|-------------|----------|
| All optional with defaults | Sensible defaults for everything | |
| Some required | min_load and max_workers required, rest optional | ✓ |
| You decide | Based on API requirements | |

### Validation Ranges
| Option | Description | Selected |
|--------|-------------|----------|
| Strict ranges | Validate all params at plan time | ✓ |
| Basic only | Just type validation | |
| You decide | Based on what makes sense | |

---

## Claude's Discretion
- Resource model separation
- Status data source depth
- Default values for optional params
- Timeout defaults

## Deferred Ideas
None
