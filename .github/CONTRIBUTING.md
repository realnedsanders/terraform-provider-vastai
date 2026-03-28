# Contributing to terraform-provider-vastai

Thank you for your interest in contributing! This document covers the development setup, testing process, and pull request guidelines.

## Development Setup

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.25
- [GNU Make](https://www.gnu.org/software/make/)
- [golangci-lint](https://golangci-lint.run/usage/install/)
- A [Vast.ai](https://vast.ai) account and API key (for acceptance tests)

### Clone and Build

```bash
git clone https://github.com/realnedsanders/terraform-provider-vastai.git
cd terraform-provider-vastai
make build
```

### Project Structure

```
internal/
  client/          # Go API client for Vast.ai REST API
  provider/        # Terraform provider configuration
  services/        # Per-resource Terraform implementations
    instance/      # vastai_instance resource + data sources
    template/      # vastai_template resource
    offer/         # vastai_gpu_offers data source
    ...            # One directory per resource type
  acctest/         # Acceptance test helpers
  sweep/           # Test sweeper client
examples/          # Example Terraform configurations
templates/         # tfplugindocs documentation templates
docs/              # Generated documentation (do not edit directly)
```

## Running Tests

### Unit Tests

```bash
make test
```

Unit tests use `httptest` mock servers and do not require API credentials.

### Acceptance Tests

Acceptance tests create real resources on Vast.ai and may incur costs.

```bash
export VASTAI_API_KEY="your-api-key"
make testacc
```

### Linting

```bash
make lint
```

### Test Sweepers

Clean up leaked test resources:

```bash
export VASTAI_API_KEY="your-api-key"
make sweep
```

## Making Changes

### Adding a New Resource

1. Create a client service in `internal/client/<resource>.go`
2. Create a service directory at `internal/services/<resource>/`
3. Add `models.go`, `resource_<name>.go`, and `resource_<name>_test.go`
4. Register the resource in `internal/provider/provider.go`
5. Add example at `examples/resources/vastai_<name>/resource.tf`
6. Add template at `templates/resources/<name>.md.tmpl`
7. Run `make generate` to update docs

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `tflog` for logging (never `fmt.Println` or `log.Printf`)
- Mark sensitive attributes with `Sensitive: true`
- Use `snake_case` for all Terraform attribute names
- Add validators for constrained fields
- Include meaningful descriptions on all schema attributes

## Pull Request Process

1. Fork the repository and create a feature branch
2. Make your changes with atomic commits
3. Ensure `make test` and `make lint` pass
4. Add or update tests for your changes
5. Run `make generate` to update documentation
6. Submit a pull request with a clear description

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features or resources
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions or changes
- `chore:` Maintenance tasks

## License

By contributing, you agree that your contributions will be licensed under the [MPL-2.0 License](../LICENSE).
