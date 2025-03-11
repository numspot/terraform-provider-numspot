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

## Using locally built provider
In order to use the locally build provider follow this steps.

First this is building the provider:
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

## Step by step implementation of a new resource
Here is a step by step tutorial that explains how to add a new resource/datasource in this repository. From a change in Numspot's OAS to the deployment of a new version of the provider.

Note : Multiple steps here are optional

1) Generate a new SDK (if needed) : **On repository `numspot-sdk-go`**
    - Import the new OAS in `api/numspot_public_openapi.yaml`
    - Build the SDK with the new OAS (see `README` on this repo) 
    - `Commit` and `Tag` with a new version the SDK on Gitlab 
2) Scaffold resource model (if needed) : **On repository `terraform-provider-generator`** 
    - Edit configuration file `generator.numspot.yaml` to add your new resource (and datasource if wanted)
      - Add a section with your resource name under the `resources` section (same for datasource if wanted). Get inspiration from other resources section if needed.
    - If SDK version changed, edit the `TAG_VERSION` variable in `Makefile`  
    - Scaffold the new files (see `README` on this repo) 
    - If needed fix the scaffolded file (the scaffolding tool generates duplicate functions when a same objectType is used multiple time in the schema)
    - Don't forget to push the changes you made on `generator.numspot.yaml` and `Makefile` to gitlab
3) Implement the new resource 
    - If SDK version changed, update `go.mod` file to use the latest version of `numspot-sdk-go` 
      - Note : An SDK change might have impact on other resources (depending on the changes on the OAS). You might have errors to fix at this step. 
    - Create a folder `internal/services/[MY RESOURCE]`
    - In this folder, copy/paste the Scaffolded file for your resource (and datasource if needed)
    - Create files `internal/services/[MY RESOURCE]/resource_[my resource].go` (and datasource if needed) and `internal/core/[my resource].go`
    - Implement the CRUD operations for this resource (Get inspiration from other resources if needed)
        - Files in `services` folder must be as simple as possible, all functional logic must be in `core` folder.
        - File in `core` must be **completely independant** from Terraform framework. In the future this part might be extracted from the provider and used by some other services.
4) Tests 
    - Create a file `internal/test/resource_[my resource]_test.go` (same for datasource if needed)
    - Implement acceptance tests in this file (same for datasource if needed). Get inspiration from other tests if needed
    - Once your test is working, create a `cassette` file that will be used to replay tests with a mock of numspot APIs (used by our CI/CD)
        - To do change the start of your test file with `acct := acctest.NewAccTest(t, true, "record")` (for more information about VCR read `Testing` section above)
        - Tests must be idempotents (no attributes with random values for example) for VCR to work
5) Documentation 
    - Add un example of Terraform plan to deploy your resource in `examples/resource/resource_[my resource]/resource.tf` (and datasource if needed)
    - Execute `go generate` to generate the doc (if you created your resource model without using Scaffolding, you will need to write description for each attributes manually so that it appears in the doc)
6) Commit 
    - Before pushing your new resource, you must :
        - Be sure that acceptance tests works in replay mode (if you changed a resource used in other tests, you might need to re-generate cassettes files for these other tests)
        - Execute `make fmt` to ensure the code is properly formatted 
        - Execute `make lint` to ensure the code is properly linted 
        - Execute `go generate` to ensure the doc is properly generated
    - You can then push your changes on a branch
    - After merging the branche on `main`, you can tag a new version of the provider (on gitlab). Tagging will trigger a pipeline which will push changes on Github public repository and make it available publicly on Hashicorp registry
