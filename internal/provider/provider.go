// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-scaffolding-framework/internal/conns"
	"github.com/hashicorp/terraform-provider-scaffolding-framework/internal/service/dhcp_options_set"
	"github.com/hashicorp/terraform-provider-scaffolding-framework/internal/service/key_pair"
	"github.com/hashicorp/terraform-provider-scaffolding-framework/internal/service/security_group"
	"github.com/hashicorp/terraform-provider-scaffolding-framework/internal/service/virtual_private_cloud"
)

// Ensure NumspotProvider satisfies various provider interfaces.
var _ provider.Provider = &NumspotProvider{}

// NumspotProvider defines the provider implementation.
type NumspotProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// NumspotProviderModel describes the provider data model.
type NumspotProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
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
		},
	}
}

func Faker(ctx context.Context, req *http.Request) error {
	req.Header.Add("Authorization", "Bearer token_200")
	return nil
}

func (p *NumspotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data NumspotProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	var endpoint string
	if data.Endpoint.IsNull() {
		endpoint = os.Getenv("NUMSPOT_ENDPOINT")
		if endpoint == "" {
			endpoint = "http://localhost:8080/v0/"
		}
	} else {
		endpoint = data.Endpoint.ValueString()
	}

	/*	accessKey := os.Getenv("NUMSPOT_ACCESS_KEY")
		if !data.AccessKey.IsNull() {
			accessKey = data.AccessKey.ValueString()
		}

		secretKey := os.Getenv("NUMSPOT_SECRET_KEY")
		if !data.SecretKey.IsNull() {
			secretKey = data.SecretKey.ValueString()
		}*/

	// Example client configuration for data sources and resources
	client, err := conns.NewClientWithResponses(endpoint, conns.WithRequestEditorFn(Faker))

	if err != nil {
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *NumspotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		key_pair.NewKeyPairResource,
		virtual_private_cloud.NewVirtualPrivateCloudResource,
		security_group.NewSecurityGroupResource,
		dhcp_options_set.NewDhcpOptionsSetResource,
	}
}

func (p *NumspotProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NumspotProvider{
			version: version,
		}
	}
}
