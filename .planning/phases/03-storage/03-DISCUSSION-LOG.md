# Phase 3: Storage - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-27
**Phase:** 03-storage
**Areas discussed:** Volume lifecycle, Offer search parity

---

## Volume Lifecycle

### Clone Modeling
| Option | Description | Selected |
|--------|-------------|----------|
| clone_from_id attribute | Optional attribute, ForceNew on change | ✓ |
| Separate clone resource | vastai_volume_clone as different type | |
| You decide | Claude picks | |

**User's choice:** clone_from_id attribute

### Marketplace List/Unlist
| Option | Description | Selected |
|--------|-------------|----------|
| Boolean attribute | is_listed on volume resource | |
| Separate resource | vastai_volume_listing | |
| You decide | Claude picks based on API behavior | ✓ |

**User's choice:** You decide

### Resource Split
| Option | Description | Selected |
|--------|-------------|----------|
| Separate resources | vastai_volume and vastai_network_volume | ✓ |
| Single resource | vastai_volume with type attribute | |
| You decide | Claude picks | |

**User's choice:** Separate resources

---

## Offer Search Parity

### Mirror Pattern
| Option | Description | Selected |
|--------|-------------|----------|
| Mirror GPU pattern | Same structure, consistent DX | |
| Simpler version | Fewer filters for storage | |
| You decide | Based on API support | ✓ |

### Data Source Split
| Option | Description | Selected |
|--------|-------------|----------|
| Separate | vastai_volume_offers and vastai_network_volume_offers | |
| Combined | vastai_storage_offers with type filter | |
| You decide | Based on API differences | ✓ |

---

## Claude's Discretion

- Marketplace list/unlist modeling
- Offer search pattern details
- Offer data source structure
- Timeout defaults

## Deferred Ideas

None
