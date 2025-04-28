package services

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"terraform-provider-numspot/internal/client"
)

func ConfigureProviderDatasource(request datasource.ConfigureRequest, response *datasource.ConfigureResponse) *client.NumSpotSDK {
	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Datasource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return nil
	}

	return provider
}

func ConfigureProviderResource(request resource.ConfigureRequest, response *resource.ConfigureResponse) *client.NumSpotSDK {
	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return nil
	}

	return provider
}
