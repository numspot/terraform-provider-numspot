package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/services/bucket"
	"terraform-provider-numspot/internal/services/clientgateway"
	"terraform-provider-numspot/internal/services/computebridge"
	"terraform-provider-numspot/internal/services/dhcpoptions"
	"terraform-provider-numspot/internal/services/flexiblegpu"
	"terraform-provider-numspot/internal/services/hybridbridge"
	"terraform-provider-numspot/internal/services/image"
	"terraform-provider-numspot/internal/services/internetgateway"
	"terraform-provider-numspot/internal/services/keypair"
	"terraform-provider-numspot/internal/services/loadbalancer"
	"terraform-provider-numspot/internal/services/managedservicebridge"
	"terraform-provider-numspot/internal/services/natgateway"
	"terraform-provider-numspot/internal/services/nic"
	"terraform-provider-numspot/internal/services/publicip"
	"terraform-provider-numspot/internal/services/routetable"
	"terraform-provider-numspot/internal/services/securitygroup"
	"terraform-provider-numspot/internal/services/servercertificate"
	"terraform-provider-numspot/internal/services/snapshot"
	"terraform-provider-numspot/internal/services/subnet"
	"terraform-provider-numspot/internal/services/virtualgateway"
	"terraform-provider-numspot/internal/services/vm"
	"terraform-provider-numspot/internal/services/volume"
	"terraform-provider-numspot/internal/services/vpc"
	"terraform-provider-numspot/internal/services/vpnconnection"
)

type NumspotProviderModel struct {
	NumSpotHost   types.String `tfsdk:"numspot_host"`
	ClientId      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	SpaceId       types.String `tfsdk:"space_id"`
	NumSpotHostOs types.String `tfsdk:"numspot_host_os"`
}

var _ provider.Provider = (*numspotProvider)(nil)

type numspotProvider struct {
	version    string
	httpClient *http.Client
}

func ProvideNumSpotProvider() func() provider.Provider {
	return func() provider.Provider {
		return &numspotProvider{}
	}
}

func ProvideNumSpotProviderWithHTTPClient(client *http.Client) func() provider.Provider {
	return func() provider.Provider {
		return &numspotProvider{httpClient: client}
	}
}

func (p *numspotProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"numspot_host": schema.StringAttribute{
				MarkdownDescription: "Numspot API Host",
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
			"numspot_host_os": schema.StringAttribute{
				MarkdownDescription: "Numspot API Host of Object Storage.",
				Optional:            true,
			},
		},
	}
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
	clientID := os.Getenv("NUMSPOT_CLIENT_ID")
	clientSecret := os.Getenv("NUMSPOT_CLIENT_SECRET")
	spaceId := os.Getenv("NUMSPOT_SPACE_ID")
	numSpotHostOs := os.Getenv("NUMSPOT_HOST_OS")

	if !config.NumSpotHost.IsNull() {
		numSpotHost = config.NumSpotHost.ValueString()
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

	if !config.NumSpotHostOs.IsNull() {
		numSpotHostOs = config.NumSpotHostOs.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
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

	if numSpotHostOs == "" {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("numspot_host_os"),
			"Missing Numspot API Host Object Storage",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the NumSpot object storage host. "+
				"Set the host value in the configuration or use the NUMSPOT_HOST_OS environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	config.NumSpotHost = types.StringValue(numSpotHost)
	config.ClientId = types.StringValue(clientID)
	config.ClientSecret = types.StringValue(clientSecret)
	config.NumSpotHostOs = types.StringValue(numSpotHostOs)

	options := []client.Option{
		client.WithHost(config.NumSpotHost.ValueString()),
		client.WithSpaceID(spaceId),
		client.WithClientID(config.ClientId.ValueString()),
		client.WithClientSecret(config.ClientSecret.ValueString()),
		client.WithHostOs(config.NumSpotHostOs.ValueString()),
	}

	if p.httpClient != nil {
		options = append(options, client.WithHTTPClient(p.httpClient))
	}

	numSpotSDK, err := client.NewNumSpotSDK(ctx, options...)
	if err != nil {
		resp.Diagnostics.AddError("Error initializing Numspot SDK", err.Error())
		return
	}

	resp.DataSourceData = numSpotSDK
	resp.ResourceData = numSpotSDK
}

func (p *numspotProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "numspot"
	resp.Version = p.version
}

func (p *numspotProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		loadbalancer.NewLoadBalancersDataSource,
		dhcpoptions.NewDHCPOptionsDataSource,
		volume.NewVolumesDataSource,
		vpc.NewVPCsDataSource,
		subnet.NewSubnetsDataSource,
		publicip.NewPublicIpsDataSource,
		nic.NewNicsDataSource,
		natgateway.NewNatGatewaysDataSource,
		internetgateway.NewInternetGatewaysDataSource,
		snapshot.NewSnapshotsDataSource,
		keypair.NewKeypairsDataSource,
		securitygroup.NewSecurityGroupsDataSource,
		routetable.NewRouteTablesDataSource,
		vm.NewVmsDataSource,
		flexiblegpu.NewFlexibleGpusDataSource,
		bucket.NewBucketsDataSource,
		servercertificate.NewServerCertificateDataSource,
		clientgateway.NewClientGatewaysDataSource,
		virtualgateway.NewVirtualGatewaysDataSource,
		vpnconnection.NewVpnConnectionsDataSource,
		computebridge.NewComputeBridgeDataSource,
		hybridbridge.NewHybridBridgeDataSource,
		managedservicebridge.NewManagedServiceBridgeDataSource,
	}
}

func (p *numspotProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		flexiblegpu.NewFlexibleGpuResource,
		image.NewImageResource,
		internetgateway.NewInternetGatewayResource,
		loadbalancer.NewLoadBalancerResource,
		natgateway.NewNatGatewayResource,
		vpc.NewNetResource,
		nic.NewNicResource,
		publicip.NewPublicIpResource,
		routetable.NewRouteTableResource,
		securitygroup.NewSecurityGroupResource,
		snapshot.NewSnapshotResource,
		subnet.NewSubnetResource,
		volume.NewVolumeResource,
		vm.NewVmResource,
		keypair.NewKeyPairResource,
		dhcpoptions.NewDhcpOptionsResource,
		servercertificate.NewServerCertificateResource,
		bucket.NewBucketResource,
		clientgateway.NewClientGatewayResource,
		virtualgateway.NewVirtualGatewayResource,
		vpnconnection.NewVpnConnectionResource,
		computebridge.NewComputeBridgeResource,
		hybridbridge.NewHybridBridgeResource,
		managedservicebridge.NewManagedServiceBridgeResource,
	}
}
