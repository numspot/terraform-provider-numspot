package app

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	numSpot "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

type terraformServe func(ctx context.Context, providerFunc func() provider.Provider, opts providerserver.ServeOpts) error

type App struct {
	terraformServe
	version                string
	numSpotProviderAddress string
	dev                    bool
	debug                  bool
}

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

func (a *App) init() (func() provider.Provider, providerserver.ServeOpts) {
	numSpotProvider := numSpot.ProvideNumSpotProvider()
	serverOptions := providerserver.ServeOpts{
		Address: a.numSpotProviderAddress,
		Debug:   a.debug,
	}

	return numSpotProvider, serverOptions
}

func (a *App) Start() {
	numSpotProvider, serverOptions := a.init()
	if err := a.terraformServe(context.Background(), numSpotProvider, serverOptions); err != nil {
		slog.Error(fmt.Sprintf("unable to start provider: %v", err.Error()))
		return
	}
}
