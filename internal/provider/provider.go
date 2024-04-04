package provider

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
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
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iam"
)

var _ provider.Provider = (*numspotProvider)(nil)

type Key string

var (
	clientIdKey     = Key("client_id")
	clientSecretKey = Key("client_secret")
)

var (
	errClientIdNotFound     = errors.New("can't find client_id")
	errClientSecretNotFound = errors.New("can't find client_secret")
)

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
	Host         types.String `tfsdk:"host"`
	IAMHost      types.String `tfsdk:"iam_host"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	SpaceId      types.String `tfsdk:"space_id"`
}

func (p *numspotProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Numspot API host",
				Optional:            true,
			},
			"iam_host": schema.StringAttribute{
				MarkdownDescription: "Numspot IAM host",
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

func (p *numspotProvider) authenticateUser(ctx context.Context, data *NumspotProviderModel) (error, *string) {
	ctx = context.WithValue(ctx, clientIdKey, data.ClientId.ValueString())
	ctx = context.WithValue(ctx, clientSecretKey, data.ClientSecret.ValueString())

	iamEndpoint := data.IAMHost.ValueString()
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

	if response.JSON200 != nil {
		return nil, response.JSON200.AccessToken
	}

	return err, nil
}

func buildBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func AddSecurityCredentialsToRequestHeaders(ctx context.Context, req *http.Request) error {
	clientId, ok := ctx.Value(clientIdKey).(string)
	if !ok {
		return errClientIdNotFound
	}

	clientSecret, ok := ctx.Value(clientSecretKey).(string)
	if !ok {
		return errClientSecretNotFound
	}

	req.Header.Add("Authorization", buildBasicAuth(clientId, clientSecret))
	return nil
}

func (p *numspotProvider) apiClientWithAuth(ctx context.Context, diag *diag.Diagnostics, data *NumspotProviderModel) *iaas.ClientWithResponses {
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

	numspotClient, err := iaas.NewClientWithResponses(data.Host.ValueString(), iaas.WithRequestEditorFn(bearerProvider.Intercept),
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

func (p *numspotProvider) apiClientWithFakeAuth(data *NumspotProviderModel, diag diag.Diagnostics) *iaas.ClientWithResponses {
	numspotClient, err := iaas.NewClientWithResponses(data.Host.ValueString(), iaas.WithRequestEditorFn(faker),
		iaas.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}))
	if err != nil {
		diag.AddError("Failed to create NumSpot api provider", err.Error())
		return nil
	}

	return numspotClient
}

func faker(_ context.Context, req *http.Request) error {
	delete(req.Header, "Authorization")

	req.Header.Add("Authorization", "Bearer token_200")
	return nil
}

type Provider struct {
	SpaceID   iaas.SpaceId
	ApiClient *iaas.ClientWithResponses
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

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Numspot API Host",
			"The provider cannot create the Numspot API provider as there is an unknown configuration value for the Numspot API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NUMSPOT_API_HOST environment variable.",
		)
	}

	if config.IAMHost.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_host"),
			"Unknown IAM API Host",
			"The provider cannot create the IAM API provider as there is an unknown configuration value for the IAM API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NUMSPOT_IAM_HOST environment variable.",
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

	host := os.Getenv("NUMSPOT_API_HOST")
	iamHost := os.Getenv("NUMSPOT_IAM_HOST")
	clientID := os.Getenv("NUMSPOT_CLIENT_ID")
	clientSecret := os.Getenv("NUMSPOT_CLIENT_SECRET")
	spaceId := os.Getenv("NUMSPOT_SPACE_ID")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.IAMHost.IsNull() {
		iamHost = config.IAMHost.ValueString()
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

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Numspot API Host",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot API host. "+
				"Set the host value in the configuration or use the NUMSPOT_API_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if iamHost == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_host"),
			"Missing Numspot IAM API Host",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot IAM API host. "+
				"Set the iam_host value in the configuration or use the NUMSPOT_IAM_HOST environment variable. "+
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

	config.IAMHost = types.StringValue(iamHost)
	config.Host = types.StringValue(host)
	config.ClientId = types.StringValue(clientID)
	config.ClientSecret = types.StringValue(clientSecret)

	// Create a new Numspot provider using the configuration values
	var client *iaas.ClientWithResponses
	if p.development {
		client = p.apiClientWithFakeAuth(&config, resp.Diagnostics)
	} else {
		client = p.apiClientWithAuth(ctx, &resp.Diagnostics, &config)
	}

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to create NumSpot api provider", "")
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
		ApiClient: client,
		SpaceID:   spaceUuid,
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
	}
}
