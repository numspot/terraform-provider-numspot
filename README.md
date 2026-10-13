# NumSpot Terraform Provider

The Numspot Provider allows Terraform to manage [Numspot Cloud](https://numspot.com/) resources.

- [Provider documentation](https://registry.terraform.io/providers/numspot/numspot/latest/docs)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= go 1.27

Go-based tools (`terrafmt`, `tfplugindocs`, `golangci-lint`) are managed automatically by this repository:

- `terrafmt` and `tfplugindocs` are pinned via `go tool` in `go.mod`.
- `golangci-lint` is installed locally into `bin/` by `make setup` (or on demand by `make fmt`/`make lint`) using the [official install script](https://golangci-lint.run/welcome/install/).

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

To generate or update documentation, run `go generate`.

## Using the provider

Check the documentation in the [Terraform registry](https://registry.terraform.io/providers/numspot/numspot/latest/docs).

## Contributing / internals

See [AGENTS.md](AGENTS.md) for an agent- and contributor-friendly overview (commands, conventions, the `helpers.go` pattern), and [docs/internal/](docs/internal/) for the full developer docs.

## Proxy Configuration

**Linux/macOS**

```sh
export HTTPS_PROXY=http://192.168.1.24:3128
```

**Windows**

```cmd
set HTTPS_PROXY=http://192.168.1.24:3128
```

## OAS Versions

Current pinned OpenAPI Specification versions are tracked in [VERSION.md](VERSION.md).
