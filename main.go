package main

import (
	"context"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/cmd/app"
)

const (
	NumSpotRegistry = "registry.terraform.io/numspot/numspot"
	ProviderVersion = "1"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have Terraform installed, you can remove the formatting command, but it's suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs --ignore-deprecated

// goreleaser can pass other information to the main package, such as the specific commit
// https://goreleaser.com/cookbooks/using-main.version/

func main() {
	app.ProvideApp(context.Background(), ProviderVersion, NumSpotRegistry).Start()
}
