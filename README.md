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
To generate or update documentation, run `go generate`.

## Using the provider
Check the documentation in the [Terraform registry](https://registry.terraform.io/providers/numspot/numspot/latest/docs).

## Proxy Configuration

**Linux/macOS**
```sh
export HTTPS_PROXY=http://192.168.1.24:3128
```

**Windows**
```cmd
set HTTPS_PROXY=http://192.168.1.24:3128
```