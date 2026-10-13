package services

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-numspot/internal/client"
)

// BaseDataSource is the Terraform data source implementation.
type BaseDataSource struct {
	Provider *client.NumSpotSDK
}

// Configure configures the provider data.
func (b *BaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	b.Provider = ConfigureProviderDatasource(req, resp)
}
