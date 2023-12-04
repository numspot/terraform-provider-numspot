// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api_client "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api_client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/iam"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/service/dhcp_options_set"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/service/internet_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/service/key_pair"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/service/route_table"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/service/security_group"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/service/subnet"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/service/virtual_private_cloud"
)

// Ensure NumspotProvider satisfies various provider interfaces.
var _ provider.Provider = &NumspotProvider{}

// NumspotProvider defines the provider implementation.
type NumspotProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version     string
	development bool
}

// NumspotProviderModel describes the provider data model.
type NumspotProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (p *NumspotProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "numspot"
	resp.Version = p.version
}

func (p *NumspotProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Client ID to authenticate user.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Client secret to authenticate user.",
				Optional:            true,
			},
		},
	}
}

func buildBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func AddSecurityCredentialsToRequestHeaders(ctx context.Context, req *http.Request) error {
	clientId, ok := ctx.Value("client_id").(string)
	if !ok {
		return errors.New("Can't find client_id")
	}

	clientSecret, ok := ctx.Value("client_secret").(string)
	if !ok {
		return errors.New("Can't find client_secret")
	}

	req.Header.Add("Authorization", buildBasicAuth(clientId, clientSecret))
	return nil
}

func (p *NumspotProvider) apiClientWithAuth(ctx context.Context, diag diag.Diagnostics, data NumspotProviderModel) *api_client.ClientWithResponses {
	var endpoint string
	if data.Endpoint.IsNull() {
		endpoint = "http://localhost:8080/v0/"
	} else {
		endpoint = data.Endpoint.ValueString()
	}

	err, accessToken := p.authenticateUser(ctx, data)
	if err != nil {
		diag.AddError("Failed to authenticate", err.Error())
		return nil
	}

	bearerProvider, err := securityprovider.NewSecurityProviderBearerToken(*accessToken)
	if err != nil {
		diag.AddError("Failed to create bearer provider token", err.Error())
		return nil
	}

	apiClient, err := api_client.NewClientWithResponses(endpoint, api_client.WithRequestEditorFn(bearerProvider.Intercept),
		api_client.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}),
	)
	if err != nil {
		diag.AddError("Failed to create NumSpot api client", err.Error())
		return nil
	}

	return apiClient
}

func faker(_ context.Context, req *http.Request) error {
	req.Header.Add("Authorization", "Bearer token_200")
	return nil
}

func (p *NumspotProvider) apiClientWithFakeAuth(diag diag.Diagnostics, data NumspotProviderModel) *api_client.ClientWithResponses {
	var endpoint string
	if data.Endpoint.IsNull() {
		endpoint = "http://localhost:8080/v0/"
	} else {
		endpoint = data.Endpoint.ValueString()
	}

	apiClient, err := api_client.NewClientWithResponses(endpoint, api_client.WithRequestEditorFn(faker))
	if err != nil {
		diag.AddError("Failed to create NumSpot api client", err.Error())
		return nil
	}

	return apiClient
}

func (p *NumspotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data NumspotProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiClient *api_client.ClientWithResponses
	if p.development {
		apiClient = p.apiClientWithFakeAuth(resp.Diagnostics, data)
	} else {
		apiClient = p.apiClientWithAuth(ctx, resp.Diagnostics, data)
	}

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to create NumSpot api client", "")
	}

	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *NumspotProvider) authenticateUser(ctx context.Context, data NumspotProviderModel) (error, *string) {
	ctx = context.WithValue(ctx, "client_id", data.ClientId.ValueString())
	ctx = context.WithValue(ctx, "client_secret", data.ClientSecret.ValueString())

	iamEndpoint := "https://authentication-manager.integration.numspot.dev"
	tmp := func(c *iam.Client) error {
		c.Client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		return nil
	}

	iamClient, err := iam.NewClientWithResponses(iamEndpoint, tmp)
	if err != nil {
		return err, nil
	}

	body := iam.Oauth2TokenExchangeFormdataRequestBody{
		GrantType: "client_credentials",
	}

	response, err := iamClient.Oauth2TokenExchangeWithFormdataBodyWithResponse(ctx, body, AddSecurityCredentialsToRequestHeaders)
	if err != nil {
		return err, nil
	}

	accessToken := response.JSON200.AccessToken
	return err, accessToken
}

func (p *NumspotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		key_pair.NewKeyPairResource,
		virtual_private_cloud.NewVirtualPrivateCloudResource,
		security_group.NewSecurityGroupResource,
		dhcp_options_set.NewDhcpOptionsSetResource,
		subnet.NewSubnetResource,
		route_table.NewRouteTableResource,
		internet_gateway.NewInternetGatewayResource,
	}
}

func (p *NumspotProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func New(version string, development bool) func() provider.Provider {
	return func() provider.Provider {
		return &NumspotProvider{
			version:     version,
			development: development,
		}
	}
}
