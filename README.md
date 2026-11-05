<div align="center">

# Numspot Terraform Provider

Terraform provider to provision and manage [Numspot Cloud](https://numspot.com/) resources.

[![CI](https://github.com/numspot/terraform-provider-numspot/actions/workflows/ci.yml/badge.svg)](https://github.com/numspot/terraform-provider-numspot/actions/workflows/ci.yml)
[![Security](https://github.com/numspot/terraform-provider-numspot/actions/workflows/security.yml/badge.svg)](https://github.com/numspot/terraform-provider-numspot/actions/workflows/security.yml)
[![Release](https://img.shields.io/github/v/release/numspot/terraform-provider-numspot?sort=semver&label=release)](https://github.com/numspot/terraform-provider-numspot/releases)
[![Terraform Registry](https://img.shields.io/badge/Terraform%20Registry-numspot%2Fnumspot-844FBA?logo=terraform&logoColor=white)](https://registry.terraform.io/providers/numspot/numspot/latest/docs)
[![Go Reference](https://pkg.go.dev/badge/github.com/numspot/terraform-provider-numspot.svg)](https://pkg.go.dev/github.com/numspot/terraform-provider-numspot)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL--2.0-blue.svg)](LICENSE)

</div>

## About

The Numspot provider allows Terraform to manage [Numspot Cloud](https://numspot.com/) resources:
compute, networking, storage, security and more — declared as code, versioned with your
infrastructure, and applied through the standard Terraform workflow.

Full provider documentation, including every resource and data source reference, lives in the
[Terraform Registry](https://registry.terraform.io/providers/numspot/numspot/latest/docs).

## Quick start

1. Get your [Numspot service-account credentials](https://console.eu-west-2.numspot.com/fr/iam/service-accounts)
   and your [Space ID](https://console.eu-west-2.numspot.com/fr/iam/spaces).
2. Declare the provider and a first resource:

```hcl
terraform {
  required_providers {
    numspot = {
      source = "numspot/numspot"
    }
  }
}

provider "numspot" {}

resource "numspot_vpc" "vpc" {
  ip_range = "10.0.0.0/16"

  tags = [{
    key   = "name"
    value = "terraform-demo"
  }]
}
```

3. Authenticate through environment variables:

```bash
$ export NUMSPOT_HOST="https://api.eu-west-2.numspot.com"
$ export NUMSPOT_HOST_OS="https://objectstorage.eu-west-2.numspot.com"
$ export NUMSPOT_CLIENT_ID="my-svc-account-id"
$ export NUMSPOT_CLIENT_SECRET="my-svc-account-secret"
$ export NUMSPOT_SPACE_ID="my-space-id"
$ terraform init
$ terraform plan
```

Credentials can also be set in the provider block (static credentials) — see the
[authentication documentation](https://registry.terraform.io/providers/numspot/numspot/latest/docs)
for all supported methods.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.27

Go-based tooling is managed automatically by this repository:

- `terrafmt` and `tfplugindocs` are pinned via `go tool` in `go.mod`.
- `golangci-lint` is installed locally into `bin/` by `make setup` (or on demand by
  `make fmt`/`make lint`) using the [official install script](https://golangci-lint.run/welcome/install/).

## Building the provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

To generate or update documentation, run `go generate`.

## Testing

Acceptance tests run in **VCR replay mode**: HTTP interactions are recorded in committed
cassettes, so the suite runs fully offline.

```sh
make testacc            # replay the recorded cassettes (offline)
make record-cassettes   # record new cassettes (VCR_MODE=record, requires credentials)
```

## Documentation

- [Terraform Registry docs](https://registry.terraform.io/providers/numspot/numspot/latest/docs) —
  provider, resource and data-source reference.
- [`docs/`](docs/) — the in-repo copy of the registry documentation.
- [`docs/internal/`](docs/internal/) — contributor documentation: the descriptor-driven code
  generation pipeline, commands, and the resource registry.
- [`VERSION.md`](VERSION.md) — currently pinned OpenAPI Specification versions.
- [`AGENTS.md`](AGENTS.md) — contributor- and agent-friendly overview of commands and conventions.

## Contributing

Contributions are welcome! Please open a pull request — Conventional Commits are enforced on
every commit and PR title. See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow,
and our [Code of Conduct](.github/CODE_OF_CONDUCT.md). Security issues follow the process in
[SECURITY.md](SECURITY.md).

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE).
