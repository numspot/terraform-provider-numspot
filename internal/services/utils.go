package services

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/numspot/terraform-provider-numspot/internal/client"
)

// ConfigureProviderDatasource handles configure provider datasource.
func ConfigureProviderDatasource(request datasource.ConfigureRequest, response *datasource.ConfigureResponse) *client.NumspotSDK {
	provider, ok := request.ProviderData.(*client.NumspotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Datasource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return nil
	}

	return provider
}

// ConfigureProviderResource is the Terraform resource implementation.
func ConfigureProviderResource(request resource.ConfigureRequest, response *resource.ConfigureResponse) *client.NumspotSDK {
	provider, ok := request.ProviderData.(*client.NumspotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return nil
	}

	return provider
}
