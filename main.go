package main

import (
	"context"

	"github.com/numspot/terraform-provider-numspot/cmd/app"
)

const (
	NumspotRegistry = "registry.terraform.io/numspot/numspot"
)

// ProviderVersion is stamped at build time by goreleaser
// (-X main.ProviderVersion={{ .Version }}). It stays a variable so the
// linker can override it; dev builds fall back to this default.
var ProviderVersion = "1"

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have Terraform installed, you can remove the formatting command, but it's suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go tool tfplugindocs --ignore-deprecated

// goreleaser can pass other information to the main package, such as the specific commit
// https://goreleaser.com/cookbooks/using-main.version/

func main() {
	app.ProvideApp(context.Background(), ProviderVersion, NumspotRegistry).Start()
}
