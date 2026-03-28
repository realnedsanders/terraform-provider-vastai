# Terraform Provider for Vast.ai

A Terraform/OpenTofu provider for managing [Vast.ai](https://vast.ai) GPU compute infrastructure.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (to build the provider plugin)

## Usage

```terraform
terraform {
  required_providers {
    vastai = {
      source  = "realnedsanders/vastai"
      version = "~> 0.1"
    }
  }
}

provider "vastai" {
  # API key can be set via VASTAI_API_KEY environment variable (recommended)
  # api_key = "your-api-key-here"
}
```

## Authentication

The provider requires a Vast.ai API key. Set it via environment variable:

```bash
export VASTAI_API_KEY="your-api-key"
terraform plan
```

Or configure it directly in the provider block (not recommended for version control):

```terraform
provider "vastai" {
  api_key = "your-api-key-here"
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `vastai_instance` | GPU compute instance with full lifecycle management |
| `vastai_template` | Reusable container configuration template |
| `vastai_ssh_key` | SSH public key for instance access |
| `vastai_volume` | Persistent local storage volume |
| `vastai_network_volume` | Network-attached storage volume |
| `vastai_endpoint` | Serverless inference endpoint with autoscaling |
| `vastai_worker_group` | Worker group bound to a serverless endpoint |
| `vastai_api_key` | API key with permission scoping |
| `vastai_environment_variable` | Account-level environment variable |
| `vastai_team` | Team for multi-user collaboration |
| `vastai_team_role` | Role with granular permissions for a team |
| `vastai_team_member` | Team member (invite on create, remove on destroy) |
| `vastai_subaccount` | Sub-account for organizational hierarchy |
| `vastai_cluster` | Physical GPU cluster |
| `vastai_cluster_member` | Machine membership in a cluster |
| `vastai_overlay` | Overlay network on top of a cluster |
| `vastai_overlay_member` | Instance membership in an overlay network |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `vastai_gpu_offers` | Search GPU marketplace offers with filters |
| `vastai_instance` | Look up a single instance by ID |
| `vastai_instances` | List instances with optional filtering |
| `vastai_templates` | Search templates by query |
| `vastai_ssh_keys` | List all SSH keys |
| `vastai_volume_offers` | Search volume offers with filters |
| `vastai_network_volume_offers` | Search network volume offers |
| `vastai_endpoints` | List serverless endpoints |
| `vastai_user` | Current account profile |
| `vastai_invoices` | Billing invoice history |
| `vastai_audit_logs` | Account activity audit trail |

## Building The Provider

```bash
git clone https://github.com/realnedsanders/terraform-provider-vastai.git
cd terraform-provider-vastai
make build
```

## Development

### Prerequisites

- Go 1.25+
- GNU Make

### Building

```bash
make build     # Build the provider binary
make lint      # Run golangci-lint
make test      # Run unit tests
make testacc   # Run acceptance tests (requires VASTAI_API_KEY)
```

### Running Tests

Unit tests:

```bash
make test
```

Acceptance tests require a Vast.ai account and API key:

```bash
export VASTAI_API_KEY="your-api-key"
make testacc
```

### Installing Locally

To install the provider locally for testing:

```bash
make install
```

## Documentation

Full documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/realnedsanders/vastai/latest/docs).

## License

This project is licensed under the MPL-2.0 License — see the [LICENSE](LICENSE) file for details.
