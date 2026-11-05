package app

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	numspot "github.com/numspot/terraform-provider-numspot/internal/provider"
)

type terraformServe func(ctx context.Context, providerFunc func() provider.Provider, opts providerserver.ServeOpts) error

// App handles app.
type App struct {
	terraformServe
	version                string
	numspotProviderAddress string
	dev                    bool
	debug                  bool
}

// ProvideApp handles provide app.
func ProvideApp(_ context.Context, version, numspotProviderAddress string) *App {
	var debug bool
	var development bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.BoolVar(&development, "dev", false, "set to true to run the provider in development mode")
	flag.Parse()

	return &App{
		terraformServe:         providerserver.Serve,
		version:                version,
		numspotProviderAddress: numspotProviderAddress,
		dev:                    development,
		debug:                  debug,
	}
}

// Start handles start.
func (a *App) Start() {
	numspotProvider, serverOptions := a.init()
	if err := a.terraformServe(context.Background(), numspotProvider, serverOptions); err != nil {
		slog.Error(fmt.Sprintf("unable to start provider: %v", err.Error()))
		return
	}
}

func (a *App) init() (func() provider.Provider, providerserver.ServeOpts) {
	numspotProvider := numspot.ProvideNumspotProvider()
	serverOptions := providerserver.ServeOpts{
		Address: a.numspotProviderAddress,
		Debug:   a.debug,
	}

	return numspotProvider, serverOptions
}
