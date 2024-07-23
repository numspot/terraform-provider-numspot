# NumSpot Terraform Provider

The Numspot Provider allows Terraform to manage [Numspot Cloud](https://numspot.com/) resources.

- [Provider documentation](https://registry.terraform.io/providers/numspot/numspot/latest/docs)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= go 1.22.0

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```
To generate or update documentation, run `go generate`.
## Using the provider
Check the documentation in the [Terraform registry](https://registry.terraform.io/providers/numspot/numspot/latest/docs).

## Using locally built provider
In order to use the locally build provider follow this steps.

First build the provider:
```sh
$ go install
```

Add development override to the .terraformrc CLI config file, check the [docs](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) for more details:
```sh

$ cd ~
$ cat > .terraformrc <<EOF
provider_installation {

  dev_overrides {
      "numspot.cloud/dev/numspot" = "/home/$USER/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
EOF
```

You could either pass provider parameters as environment variables like following or set them in the provider block in the terraform script file:
```sh
$ export NUMSPOT_HOST="..."
$ export NUMSPOT_CLIENT_ID="..."
$ export NUMSPOT_CLIENT_SECRET="..."
$ export NUMSPOT_SPACE_ID="..."
```
Now it's time to use the provider
```sh
# Provider configuration
$ mkdir ~/test-terraform-provider-numspot
$ cd ~/test-terraform-provider-numspot

$ cat > main.tf <<EOF
terraform {
  required_providers {
    numspot = {
      source = "numspot.cloud/dev/numspot"
      version = "dev"
    }
  }
}

# If env variables not set, specify provider parameters here
provider "numspot" {
  numspot_host  = ""
  client_id     = ""
  client_secret = ""
  space_id      = ""
}
EOF

# Init project
$ terraform init

# Apply your resources & datasources
$ terraform apply
```

## Testing
Acceptance tests creates real infrastructure objects, beware that you will be charged for this :moneybag:

Before you run acc tests you have to set necessary env variables like mentioned in the previous section and run:
```shell
$ make testacc
```
In order to run a single test, use:
```shell
$ make testacc TESTARGS="-run TestAccKeypairDatasource"
```
