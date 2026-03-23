# NumSpot Terraform Provider

The Numspot Provider allows Terraform to manage [Numspot Cloud](https://numspot.com/) resources.

- [Provider documentation](https://registry.terraform.io/providers/numspot/numspot/latest/docs)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= go 1.22.0

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```
To generate or updqate documentation, run `go generate`.

## Using the provider
Check the documentation in the [Terraform registry](https://registry.terraform.io/providers/numspot/numspot/latest/docs).

## Acceptance Tests

Acceptance tests use the VCR system (go-vcr) to record and replay API calls.

### Prerequisites

Set the following environment variables:

```sh
export NUMSPOT_CLIENT_ID=<your_client_id>
export NUMSPOT_CLIENT_SECRET=<your_client_secret>
export NUMSPOT_SPACE_ID=<your_space_id>
export NUMSPOT_REGION=<your_region>  # e.g.: eu-west-2
export NUMSPOT_HOST_OS=<your_os_host> # e.g.: https://objectstorage.preprod.eu-west-2.numspot.com
export NUMSPOT_HOST=<your_iaas_host> # e.g.: https://api.preprod.eu-west-2.numspot.com
```

### Running Tests

```sh
# Run all tests with a specific VCR mode
VCR_MODE=replay make testacc

# Run a specific test
VCR_MODE=replay make testacc TESTARGS="-run TestAccVpcResource"

# Run in parallel
PARALLEL_TEST=true VCR_MODE=replay make testacc
```

### VCR Modes

| Mode | Description |
|------|-------------|
| `replay` | Use recorded cassettes (no API calls) |
| `record` | Make API calls and record cassettes |
| `passthrough` | Real API calls without recording |

### Cassettes

Cassettes are stored in `internal/test/testdata/*.cassette.yaml`.

## Proxy Configuration

**Linux/macOS**
```sh
export HTTPS_PROXY=http://192.168.1.24:3128
```

**Windows**
```cmd
set HTTPS_PROXY=http://192.168.1.24:3128
```