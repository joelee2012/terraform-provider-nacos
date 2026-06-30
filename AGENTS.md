# Agent Guidelines for Terraform Provider for Nacos

## Essential Commands

- `make build` - Build the project (`go build -v ./...`)
- `make install` - Build and install the provider binary (`go install -v ./...`)
- `make lint` - Run linters (`golangci-lint run`)
- `make test` - Run unit tests (`go test -v -cover -timeout=120s -parallel=10 ./...`)
- `make testacc` - Run acceptance tests (requires running Nacos server and env vars)
- `make generate` - Generate documentation and copyright headers (requires Terraform CLI installed)
- `make fmt` - Format Go code (`gofmt -s -w -e .`)
- `make` (default) - Runs `fmt lint install generate`

## Code Organization

- `internal/provider/` - Main provider implementation
  - Data sources: `data_source_*.go`
  - Resources: `resource_*.go`
  - Provider config: `provider.go`
- `docs/` - Generated Terraform documentation (do not hand-edit; run `make generate`)
- `examples/` - HCL usage examples (formatted by `make generate`)
- `verify/` - Manual Terraform configurations for local provider testing
- `tools/` - Tool dependencies and code generation scripts (`tools.go`)

## Testing

### Unit Tests
- Standard Go tests in `*_test.go` files
- Run with: `make test` (120s timeout, 10 parallel)

### Acceptance Tests
- Require a running Nacos server and `TF_ACC=1`
- Required environment variables:
  ```
  export NACOS_HOST=http://127.0.0.1:8848/nacos
  export NACOS_USERNAME=nacos
  export NACOS_PASSWORD=nacos
  ```
- Run with: `make testacc` (120m timeout, produces `coverage.txt` and `cover.html`)
- To start a local Nacos server:
  ```bash
  docker compose up -d
  ```
  The `docker-compose.yaml` uses the `NACOS_VERSION` env variable (defaults to `latest` if unset).
- The CI workflow waits for Nacos readiness by polling `NACOS_STATE` and initializes the admin password via `NACOS_AUTH_URL` for v2.4+ and v3.

### Nacos Version Quirks
- **Nacos v2.x**: Use port `8848` and path `/nacos` (e.g., `http://127.0.0.1:8848/nacos`)
- **Nacos v3.x**: Use port `8080` and no `/nacos` path (e.g., `http://127.0.0.1:8080`)
- CI tests against a matrix of Terraform versions (1.0.*, 1.5.*, 1.6.*, 1.14.*) and Nacos versions (v2.1.2, v2.3.2, v2.4.3, v2.5.1, v3.1.0)

## Code Generation

- Always run `make generate` after modifying provider schema or resource/data source definitions
- `make generate` runs from the `tools/` directory and executes `go generate ./...`
- Requires **Terraform CLI** to be installed (it runs `terraform fmt -recursive ../examples/`)
- Generates provider docs via `terraform-plugin-docs` into `docs/`
- Also runs `copywrite headers` for copyright headers

## Linting & Style

- Linter config: `.golangci.yml` (golangci-lint v2 format)
- Excluded paths: `examples/`, `third_party/`, `builtin/`
- Follow existing Terraform Plugin Framework patterns in `internal/provider/`

## Gotchas

- The provider reads `NACOS_HOST`, `NACOS_USERNAME`, `NACOS_PASSWORD`, and `NACOS_API_VERSION` from environment variables if not set in Terraform configuration
- `go.mod` specifies `go 1.25.0`
- `tools/go.mod` specifies `go 1.24.0`
- Release builds use GoReleaser (`.goreleaser.yml`) and require `GPG_FINGERPRINT` for signing
