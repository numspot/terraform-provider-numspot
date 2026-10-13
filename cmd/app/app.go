package app

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	numSpot "terraform-provider-numspot/internal/provider"
)

type terraformServe func(ctx context.Context, providerFunc func() provider.Provider, opts providerserver.ServeOpts) error

// App handles app.
type App struct {
	terraformServe
	version                string
	numSpotProviderAddress string
	dev                    bool
	debug                  bool
}

// ProvideApp handles provide app.
func ProvideApp(_ context.Context, version, numSpotProviderAddress string) *App {
	var debug bool
	var development bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.BoolVar(&development, "dev", false, "set to true to run the provider in development mode")
	flag.Parse()

	return &App{
		terraformServe:         providerserver.Serve,
		version:                version,
		numSpotProviderAddress: numSpotProviderAddress,
		dev:                    development,
		debug:                  debug,
	}
}

// Start handles start.
func (a *App) Start() {
	numSpotProvider, serverOptions := a.init()
	if err := a.terraformServe(context.Background(), numSpotProvider, serverOptions); err != nil {
		slog.Error(fmt.Sprintf("unable to start provider: %v", err.Error()))
		return
	}
}

func (a *App) init() (func() provider.Provider, providerserver.ServeOpts) {
	numSpotProvider := numSpot.ProvideNumSpotProvider()
	serverOptions := providerserver.ServeOpts{
		Address: a.numSpotProviderAddress,
		Debug:   a.debug,
	}

	return numSpotProvider, serverOptions
}
