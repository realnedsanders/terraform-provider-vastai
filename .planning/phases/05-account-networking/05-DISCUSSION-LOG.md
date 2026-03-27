# Phase 5: Account & Networking - Discussion Log

> **Audit trail only.**

**Date:** 2026-03-27
**Phase:** 05-account-networking
**Areas discussed:** Team RBAC model, Cluster/overlay networking, Data sources scope

---

## Team RBAC Model

### Invite Flow
| Option | Description | Selected |
|--------|-------------|----------|
| Create = invite | Single resource, create sends invite | |
| Separate invite + member | Two resources | |
| You decide | Based on API behavior | ✓ |

### Permissions
| Option | Description | Selected |
|--------|-------------|----------|
| String set | permissions = ["create_instance", ...] | ✓ |
| Structured block | Nested booleans per category | |
| You decide | Based on API | |

**User's choice:** String set for permissions

---

## Cluster/Overlay Networking

### Cluster Membership
| Option | Description | Selected |
|--------|-------------|----------|
| Separate resource | vastai_cluster_member | |
| Inline attribute | machine_ids on cluster | |
| You decide | Based on API + Terraform patterns | ✓ |

### Overlay Joining
| Option | Description | Selected |
|--------|-------------|----------|
| Separate resource | vastai_overlay_member | |
| Inline attribute | instance_ids on overlay | |
| You decide | Consistent with cluster decision | ✓ |

---

## Data Sources Scope

### User Profile, Invoices, Audit Logs
All three: You decide based on API capabilities.

---

## Claude's Discretion
- Team member invite modeling
- Cluster/overlay membership patterns
- Data source depth and filtering
- Timeout defaults

## Deferred Ideas
None
