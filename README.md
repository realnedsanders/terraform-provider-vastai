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
