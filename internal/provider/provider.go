package provider

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var _ provider.Provider = (*numspotProvider)(nil)

type Key string

func New(version string, development bool) func() provider.Provider {
	return func() provider.Provider {
		return &numspotProvider{
			version:     version,
			development: development,
		}
	}
}

type numspotProvider struct {
	version     string
	development bool
}

type NumspotProviderModel struct {
	NumSpotHost  types.String `tfsdk:"numspot_host"`
	IAASHost     types.String `tfsdk:"iaas_host"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	SpaceId      types.String `tfsdk:"space_id"`
}

func (p *numspotProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"numspot_host": schema.StringAttribute{
				MarkdownDescription: "Numspot API Host",
				Optional:            true,
			},
			"iaas_host": schema.StringAttribute{
				MarkdownDescription: "Numspot API IAAS Host",
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
			"space_id": schema.StringAttribute{
				MarkdownDescription: "Space ID.",
				Optional:            true,
			},
		},
	}
}

func (p *numspotProvider) authenticateUser(ctx context.Context, numSpotClient *numspot.ClientWithResponses, data *NumspotProviderModel) (*string, error) {
	clientUuid, diags := utils.ParseUUID(data.ClientId.ValueString())
	if diags.HasError() {
		return nil, fmt.Errorf("Error while parsing %s as UUID", data.ClientId.ValueString())
	}
	body := numspot.TokenReq{
		GrantType:    "client_credentials",
		ClientId:     &clientUuid,
		ClientSecret: data.ClientSecret.ValueStringPointer(),
	}

	basicAuth := buildBasicAuth(data.ClientId.ValueString(), data.ClientSecret.ValueString())
	response, err := numSpotClient.TokenWithFormdataBodyWithResponse(ctx, &numspot.TokenParams{
		Authorization: &basicAuth,
	}, body)
	if err != nil {
		return nil, err
	}

	return &response.JSON200.AccessToken, nil
}

func buildBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func (p *numspotProvider) apiClientWithAuth(accessToken string, diag *diag.Diagnostics, data *NumspotProviderModel) *iaas.ClientWithResponses {
	bearerProvider, err := securityprovider.NewSecurityProviderBearerToken(accessToken)
	if err != nil {
		diag.AddError("Failed to create bearer provider token", err.Error())
		return nil
	}

	numspotClient, err := iaas.NewClientWithResponses(data.IAASHost.ValueString(), iaas.WithRequestEditorFn(bearerProvider.Intercept),
		iaas.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}),
	)
	if err != nil {
		diag.AddError("Failed to create NumSpot api provider", err.Error())
		return nil
	}

	return numspotClient
}

type Provider struct {
	SpaceID       iaas.SpaceId
	NumSpotClient *numspot.ClientWithResponses
	IaasClient    *iaas.ClientWithResponses
}

func (p *numspotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config NumspotProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.IAASHost.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("iaas_host"),
			"Unknown Numspot API IAAS host",
			"The provider cannot create the Numspot API provider as there is an unknown configuration value for the Numspot IAAS host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NUMSPOT_IAAS_HOST environment variable.",
		)
	}

	if config.ClientId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Numspot Client id",
			"The provider cannot create the Numspot API provider as there is an unknown configuration value for the Numspot IAM API provider ID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NUMSPOT_CLIENT_ID environment variable.",
		)
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown Numspot Client secret",
			"The provider cannot create the Numspot API provider as there is an unknown configuration value for the Numspot IAM API provider secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NUMSPOT_CLIENT_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	numSpotHost := os.Getenv("NUMSPOT_HOST")
	iaasHost := os.Getenv("NUMSPOT_IAAS_HOST")
	clientID := os.Getenv("NUMSPOT_CLIENT_ID")
	clientSecret := os.Getenv("NUMSPOT_CLIENT_SECRET")
	spaceId := os.Getenv("NUMSPOT_SPACE_ID")

	if !config.NumSpotHost.IsNull() {
		numSpotHost = config.NumSpotHost.ValueString()
	}

	if !config.IAASHost.IsNull() {
		iaasHost = config.IAASHost.ValueString()
	}

	if !config.ClientId.IsNull() {
		clientID = config.ClientId.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	if !config.SpaceId.IsNull() {
		spaceId = config.SpaceId.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if iaasHost == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("iaas_host"),
			"Missing Numspot API Host",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot IAAS host. "+
				"Set the host value in the configuration or use the NUMSPOT_IAAS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if numSpotHost == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("numspot_host"),
			"Missing Numspot API Host",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the NumSpot host. "+
				"Set the host value in the configuration or use the NUMSPOT_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Numspot Client ID",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot Client ID. "+
				"Set the client_id value in the configuration or use the NUMSPOT_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Numspot Client Secret",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot Client Secret. "+
				"Set the client_secret value in the configuration or use the NUMSPOT_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if spaceId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("space_id"),
			"Missing Numspot Space ID",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot Space ID. "+
				"Set the space_id value in the configuration or use the NUMSPOT_SPACE_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	config.NumSpotHost = types.StringValue(numSpotHost)
	config.IAASHost = types.StringValue(iaasHost)
	config.ClientId = types.StringValue(clientID)
	config.ClientSecret = types.StringValue(clientSecret)

	// Create a new Numspot provider using the configuration values
	var (
		iaasClient    *iaas.ClientWithResponses
		numSpotClient *numspot.ClientWithResponses
	)

	httpClient := func(c *numspot.Client) error {
		c.Client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		return nil
	}

	requestEditor := numspot.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add("User-Agent", "TERRAFORM-NUMSPOT")
		return nil
	})

	numSpotClient, err := numspot.NewClientWithResponses(config.NumSpotHost.ValueString(), httpClient, requestEditor)
	if err != nil {
		resp.Diagnostics.AddError("Auth error", "failed to get access token")
		return
	}

	accessToken, err := p.authenticateUser(ctx, numSpotClient, &config)
	if err != nil {
		resp.Diagnostics.AddError("Auth error", "failed to get access token")
		return
	}

	if accessToken == nil {
		resp.Diagnostics.AddError("Failed to retrieve access token", "returned access token is nil")
		return
	}

	if *accessToken == "" {
		resp.Diagnostics.AddError("Failed to retrieve access token", "returned access token is nil")
		return
	}

	bearerProvider, err := securityprovider.NewSecurityProviderBearerToken(*accessToken)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create bearer provider token", err.Error())
		return
	}

	numSpotClient, err = numspot.NewClientWithResponses(config.NumSpotHost.ValueString(), httpClient, requestEditor, numspot.WithRequestEditorFn(bearerProvider.Intercept))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create bearer provider token", err.Error())
		return
	}

	iaasClient = p.apiClientWithAuth(*accessToken, &resp.Diagnostics, &config)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make the Numspot provider available during DataSource and Resource
	// type Configure methods.
	spaceUuid, err := uuid.Parse(spaceId)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("space_id"),
			"Missing Numspot Client ID",
			"Failed to parse Space ID, please provide a valid Space ID.",
		)
		return
	}

	providerData := Provider{
		SpaceID:       spaceUuid,
		NumSpotClient: numSpotClient,
		IaasClient:    iaasClient,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *numspotProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "numspot"
	resp.Version = p.version
}

func (p *numspotProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLoadBalancersDataSource,
		NewDHCPOptionsDataSource,
		NewVolumesDataSource,
		NewVPCsDataSource,
		NewSubnetsDataSource,
		NewPublicIpsDataSource,
		NewVirtualGatewaysDataSource,
		NewNicsDataSource,
		NewNatGatewaysDataSource,
		NewVpcPeeringsDataSource,
		NewInternetGatewaysDataSource,
		NewSnapshotsDataSource,
		NewKeypairsDataSource,
		NewClientGatewaysDataSource,
		NewSecurityGroupsDataSource,
		NewRouteTablesDataSource,
		NewVpnConnectionsDataSource,
		NewSpaceDataSource,
		NewVmsDataSource,
		NewProductTypesDataSource,
		NewFlexibleGpusDataSource,
		NewServiceAccountsDataSource,
	}
}

func (p *numspotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClientGatewayResource,
		NewDirectLinkResource,
		NewDirectLinkInterfaceResource,
		NewFlexibleGpuResource,
		NewImageResource,
		NewInternetGatewayResource,
		NewListenerRuleResource,
		NewLoadBalancerResource,
		NewNatGatewayResource,
		NewNetResource,
		NewNetAccessPointResource,
		NewNicResource,
		NewPublicIpResource,
		NewRouteTableResource,
		NewSecurityGroupResource,
		NewSnapshotResource,
		NewSubnetResource,
		NewVolumeResource,
		NewVpnConnectionResource,
		NewVmResource,
		NewKeyPairResource,
		NewDhcpOptionsResource,
		NewVirtualGatewayResource,
		NewVpcPeeringResource,
		NewSpaceResource,
		NewServiceAccountResource,
		NewAclsResource,
	}
}
