# Phase 2: Core Compute - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-25
**Phase:** 02-core-compute
**Areas discussed:** Offer search DX, Instance lifecycle, Schema conventions, Acceptance tests

---

## Offer Search DX

### Filter Approach
| Option | Description | Selected |
|--------|-------------|----------|
| Structured only | Typed attributes with full validation | |
| Structured + raw escape | Typed attributes + raw_query for power users | |
| You decide | Claude picks | ✓ |

**User's choice:** You decide

### Return Shape
| Option | Description | Selected |
|--------|-------------|----------|
| List + most_affordable | All matches as list + cheapest convenience attribute | ✓ |
| Single best match | Only the cheapest matching offer | |
| You decide | Claude picks | |

**User's choice:** List + most_affordable

### Sort Order
| Option | Description | Selected |
|--------|-------------|----------|
| Price ascending | Cheapest first | |
| Score descending | Vast.ai's built-in score | |
| Configurable | User picks via order_by attribute | ✓ |
| You decide | Claude picks | |

**User's choice:** Configurable

### Result Limit
| Option | Description | Selected |
|--------|-------------|----------|
| Default 10, configurable | Top 10 by default, user can change | ✓ |
| Return all matches | No limit | |
| You decide | Claude picks | |

**User's choice:** Default 10, configurable

### Staleness Handling
| Option | Description | Selected |
|--------|-------------|----------|
| Re-query on apply | Standard Terraform behavior, live queries | ✓ |
| Cache with TTL | Cache results for N minutes | |
| You decide | Claude picks | |

**User's choice:** Re-query on apply

### Offer Expiry
| Option | Description | Selected |
|--------|-------------|----------|
| Error with guidance | Clear error suggesting re-run | |
| Auto-retry next offer | Fall back to next offer in list | |
| You decide | Claude picks best fit for Terraform model | ✓ |

**User's choice:** You decide

### Offer Data Richness
| Option | Description | Selected |
|--------|-------------|----------|
| Full details | All available data including benchmarks | ✓ |
| Essential only | Just what's needed for instance creation | |
| You decide | Claude picks | |

**User's choice:** Full details

---

## Instance Lifecycle

### Start/Stop Model
| Option | Description | Selected |
|--------|-------------|----------|
| Status attribute | Single attribute on instance resource | |
| Separate resource | vastai_instance_state resource | |
| You decide | Claude picks based on cloud provider patterns | ✓ |

**User's choice:** You decide

### Preemption Handling
| Option | Description | Selected |
|--------|-------------|----------|
| Remove from state | Silent removal, next apply recreates | |
| Error with message | Error explaining preemption | |
| You decide | Based on conventions, gated to actual preemption | ✓ |

**User's choice:** You decide, but ONLY for actual preemption — not for successful exits on spot instances
**Notes:** User explicitly wants to distinguish preemption (outbid/evicted) from normal termination. Silent removal should ONLY apply when the instance was actually preempted.

### Create Timeout
| Option | Description | Selected |
|--------|-------------|----------|
| 10 minutes | Standard | |
| 30 minutes | Conservative | |
| You decide | Claude picks based on GPU provisioning times | ✓ |

### Bid Price Update
| Option | Description | Selected |
|--------|-------------|----------|
| In-place update | change_bid API | |
| Replace | New instance on price change | |
| You decide | Based on API support | ✓ |

### SSH Attachment
| Option | Description | Selected |
|--------|-------------|----------|
| Inline attribute | ssh_key_ids list on instance | |
| Separate resource | vastai_instance_ssh_key | |
| Both | Inline + separate resource | |
| You decide | Based on Terraform patterns | ✓ |

### Immutable Attributes
| Option | Description | Selected |
|--------|-------------|----------|
| Offer ID immutable | Only offer_id requires replace | |
| Offer + image immutable | Both require replace | |
| You decide | Based on API behavior | ✓ |

---

## Schema Conventions

### Attribute Naming
| Option | Description | Selected |
|--------|-------------|----------|
| Always snake_case | Terraform convention, mapping in models | ✓ |
| Match API names | Use whatever API returns | |
| You decide | Claude picks | |

**User's choice:** Always snake_case

### Computed Handling
| Option | Description | Selected |
|--------|-------------|----------|
| Optional+Computed | User can set, server default if not | ✓ |
| Computed only | Read-only for server-set attributes | |
| You decide | Per-attribute decision | |

**User's choice:** Optional+Computed

### Validators
| Option | Description | Selected |
|--------|-------------|----------|
| Comprehensive | Validate GPU names, regions, etc. at plan time | ✓ |
| Basic only | Just type constraints | |
| You decide | Balance of plan-time vs API-time | |

**User's choice:** Comprehensive

### Model Types
| Option | Description | Selected |
|--------|-------------|----------|
| Per-resource models | Each resource gets own struct | ✓ |
| Shared common types | Extract common patterns | |
| You decide | Based on actual duplication | |

**User's choice:** Per-resource models

### Model Layer Organization
| Option | Description | Selected |
|--------|-------------|----------|
| Models in service dir | Both TF and API models colocated | |
| API models in client | API types in client, TF models in service dirs | ✓ |
| You decide | Based on standalone client decision | |

**User's choice:** API models in client

### Lists vs Sets
| Option | Description | Selected |
|--------|-------------|----------|
| Set where order doesn't matter | SSH keys→Set, env vars→List | ✓ |
| Always List | Everything is a list | |
| You decide | Per-attribute decision | |

**User's choice:** Set where order doesn't matter

### Documentation Links
| Option | Description | Selected |
|--------|-------------|----------|
| Yes, link to docs | Include Vast.ai API URLs in descriptions | ✓ |
| Self-contained | Complete without external links | |
| You decide | Based on conventions | |

**User's choice:** Yes, link to docs

---

## Acceptance Tests

### Cost Control
| Option | Description | Selected |
|--------|-------------|----------|
| Cheapest offer filter | Tests search for cheapest available | |
| Mock API for most tests | httptest mocks + small smoke suite | |
| Both | Mocks for logic + cheapest offers for integration | ✓ |
| You decide | Balance coverage vs cost | |

**User's choice:** Both

### Test Parallelism
| Option | Description | Selected |
|--------|-------------|----------|
| Sequential | One at a time, avoids rate limits | |
| Parallel | resource.ParallelTest() where safe | |
| You decide | Based on API rate limit behavior | ✓ |

**User's choice:** You decide

### Test Scope
| Option | Description | Selected |
|--------|-------------|----------|
| All resources | Full CRUD + import tests for everything | ✓ |
| Critical paths first | Full tests for instance + offers only | |
| You decide | Based on effort vs value | |

**User's choice:** All resources

---

## Claude's Discretion

- Offer filter approach (structured + raw escape balance)
- Offer expiry handling
- Start/stop modeling
- Create timeout
- Bid price update behavior
- SSH attachment model
- Immutable attribute classification
- Test parallelism

## Deferred Ideas

None
