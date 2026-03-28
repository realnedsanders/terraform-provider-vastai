# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-28

### Added

#### Resources
- `vastai_instance` - GPU compute instance with full lifecycle (create, start/stop, update, destroy, import)
- `vastai_template` - Reusable container configuration template
- `vastai_ssh_key` - SSH public key management with instance attachment
- `vastai_volume` - Persistent local storage volume with clone support
- `vastai_network_volume` - Network-attached storage volume
- `vastai_endpoint` - Serverless inference endpoint with autoscaling parameters
- `vastai_worker_group` - Worker group bound to serverless endpoints
- `vastai_api_key` - API key with permission scoping
- `vastai_environment_variable` - Account-level environment variable
- `vastai_team` - Team for multi-user collaboration
- `vastai_team_role` - Role with granular JSON permissions
- `vastai_team_member` - Team member management (invite/remove)
- `vastai_subaccount` - Sub-account for organizational hierarchy
- `vastai_cluster` - Physical GPU cluster
- `vastai_cluster_member` - Machine membership in a cluster
- `vastai_overlay` - Overlay network on a cluster
- `vastai_overlay_member` - Instance membership in an overlay

#### Data Sources
- `vastai_gpu_offers` - GPU marketplace search with structured filters and most_affordable
- `vastai_instance` - Single instance lookup by ID
- `vastai_instances` - Instance list with optional filtering
- `vastai_templates` - Template search by query
- `vastai_ssh_keys` - List all SSH keys
- `vastai_volume_offers` - Volume offer search with filters
- `vastai_network_volume_offers` - Network volume offer search
- `vastai_endpoints` - List serverless endpoints
- `vastai_user` - Current account profile
- `vastai_invoices` - Billing invoice history with date filtering
- `vastai_audit_logs` - Account activity audit trail

#### Provider Features
- Bearer token authentication (API key never in URL)
- Configurable API endpoint URL
- Exponential backoff retry on 429/5xx errors
- Structured error diagnostics
- Full import support for all resources
- Comprehensive attribute validators
- Plan modifiers (UseStateForUnknown, RequiresReplace)
- Configurable timeouts per resource
- Resource sweepers for CI cleanup

[Unreleased]: https://github.com/realnedsanders/terraform-provider-vastai/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/realnedsanders/terraform-provider-vastai/releases/tag/v0.1.0
