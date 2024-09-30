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

### HTTP Call Mocking
We use the [go-vcr](https://github.com/dnaeon/go-vcr) library to mock HTTP calls. This allows us to record real HTTP interactions and replay them during subsequent test runs, making tests faster and more reliable by avoiding real API calls every time.

The recorded HTTP interactions, called "cassettes," are stored in the `test/testdata` directory. If you modify or add new resources or data sources in the provider, you'll need to update the corresponding cassette by rerunning the tests with the `VCR_MODE` environment variable set to `"record"`. This re-records the interactions for the updated resource or data source.

### Running Tests Using Cassettes
If you want to run tests or the full test suite without making actual API calls and using the pre-recorded cassettes, set the `VCR_MODE` environment variable to "replay". This mode will replay the previously recorded interactions stored in the cassettes.

To run the full test suite or a specific test using the generated cassettes:
```shell
$ VCR_MODE=replay make testacc
```

Or to run a specific test:
```shell
$ VCR_MODE=replay make testacc TESTARGS="-run TestAccKeypairDatasource"
```

This is useful when you want to test without incurring additional charges or waiting for real API responses.

### Running Tests Concurrently
If you want to speed up the test execution by running tests concurrently, you can set the `PARALLEL_TEST` environment variable to "true". This will enable parallel test execution across multiple resources or data sources.

To run tests in parallel:
```shell
$ PARALLEL_TEST=true make testacc
```
Or to run a specific test in parallel:
```shell
$ PARALLEL_TEST=true make testacc TESTARGS="-run TestAccKeypairDatasource"
```

### Updating Cassettes
If you update a resource or data source, follow these steps to regenerate the relevant cassette:

1. Set the `VCR_MODE` environment variable to "record" to create new cassettes.
2. Run the test for the specific resource or data source that was modified.

Example:
```shell
$ VCR_MODE=record make testacc TESTARGS="-run TestAccKeypairDatasource"
```
This will re-record the HTTP interactions for `TestAccKeypairDatasource` and save the updated cassette to the `test/testdata` directory.

