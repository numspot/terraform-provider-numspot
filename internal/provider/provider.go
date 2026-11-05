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

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/services/bucket"
	"github.com/numspot/terraform-provider-numspot/internal/services/clientgateway"
	"github.com/numspot/terraform-provider-numspot/internal/services/computebridge"
	"github.com/numspot/terraform-provider-numspot/internal/services/dhcpoptions"
	"github.com/numspot/terraform-provider-numspot/internal/services/flexiblegpu"
	"github.com/numspot/terraform-provider-numspot/internal/services/hybridbridge"
	"github.com/numspot/terraform-provider-numspot/internal/services/image"
	"github.com/numspot/terraform-provider-numspot/internal/services/internetgateway"
	"github.com/numspot/terraform-provider-numspot/internal/services/keypair"
	"github.com/numspot/terraform-provider-numspot/internal/services/kubernetescluster"
	"github.com/numspot/terraform-provider-numspot/internal/services/kubernetesnodepool"
	"github.com/numspot/terraform-provider-numspot/internal/services/loadbalancer"
	"github.com/numspot/terraform-provider-numspot/internal/services/natgateway"
	"github.com/numspot/terraform-provider-numspot/internal/services/nic"
	"github.com/numspot/terraform-provider-numspot/internal/services/postgrescluster"
	"github.com/numspot/terraform-provider-numspot/internal/services/publicip"
	"github.com/numspot/terraform-provider-numspot/internal/services/routetable"
	"github.com/numspot/terraform-provider-numspot/internal/services/securitygroup"
	"github.com/numspot/terraform-provider-numspot/internal/services/securitygrouprule"
	"github.com/numspot/terraform-provider-numspot/internal/services/servercertificate"
	"github.com/numspot/terraform-provider-numspot/internal/services/snapshot"
	"github.com/numspot/terraform-provider-numspot/internal/services/subnet"
	"github.com/numspot/terraform-provider-numspot/internal/services/virtualgateway"
	"github.com/numspot/terraform-provider-numspot/internal/services/vm"
	"github.com/numspot/terraform-provider-numspot/internal/services/volume"
	"github.com/numspot/terraform-provider-numspot/internal/services/vpc"
	"github.com/numspot/terraform-provider-numspot/internal/services/vpnconnection"
)

// NumspotProviderModel handles numspot provider model.
type NumspotProviderModel struct {
	NumspotHost    types.String `tfsdk:"numspot_host"`
	ClientId       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	SpaceId        types.String `tfsdk:"space_id"`
	OrganisationId types.String `tfsdk:"organisation_id"`
	NumspotHostOs  types.String `tfsdk:"numspot_host_os"`
}

var _ provider.Provider = (*numspotProvider)(nil)

type numspotProvider struct {
	version    string
	httpClient *http.Client
}

// ProvideNumspotProvider returns a factory that creates a Numspot provider instance.
func ProvideNumspotProvider() func() provider.Provider {
	return func() provider.Provider {
		return &numspotProvider{}
	}
}

// ProvideNumspotProviderWithHTTPClient returns a factory that creates a Numspot provider instance using the provided HTTP client.
func ProvideNumspotProviderWithHTTPClient(httpClient *http.Client) func() provider.Provider {
	return func() provider.Provider {
		return &numspotProvider{httpClient: httpClient}
	}
}

// Schema defines the schema.
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
			"organisation_id": schema.StringAttribute{
				MarkdownDescription: "Organisation ID. Required for IAM organisation-scoped resources.",
				Optional:            true,
			},
			"numspot_host_os": schema.StringAttribute{
				MarkdownDescription: "Numspot API Host of Object Storage.",
				Optional:            true,
			},
		},
	}
}

// providerValues holds the resolved string values for provider configuration.
type providerValues struct {
	numspotHost    string
	clientID       string
	clientSecret   string
	spaceID        string
	organisationID string
	numspotHostOs  string
}

// Configure configures the provider data.
func (p *numspotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config NumspotProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateNoUnknownProviderValues(resp, &config)
	if resp.Diagnostics.HasError() {
		return
	}

	values := resolveProviderValues(&config)
	validateRequiredProviderValues(resp, &values)
	if resp.Diagnostics.HasError() {
		return
	}

	numspotSDK, err := client.NewNumspotSDK(ctx, buildClientOptions(p, &values)...)
	if err != nil {
		resp.Diagnostics.AddError("Error initializing Numspot SDK", err.Error())
		return
	}

	resp.DataSourceData = numspotSDK
	resp.ResourceData = numspotSDK
}

func validateNoUnknownProviderValues(resp *provider.ConfigureResponse, config *NumspotProviderModel) {
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

	if config.NumspotHost.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("numspot_host"),
			"Unknown Numspot API Host",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot host. "+
				"Set the host value in the configuration or use the NUMSPOT_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if config.NumspotHostOs.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("numspot_host_os"),
			"Missing Numspot API Host Object Storage",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot object storage host. "+
				"Set the host value in the configuration or use the NUMSPOT_HOST_OS environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if config.SpaceId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("space_id"),
			"Missing Numspot Space ID",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot Space ID. "+
				"Set the space_id value in the configuration or use the NUMSPOT_SPACE_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if config.OrganisationId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("organisation_id"),
			"Unknown Numspot Organisation ID",
			"The provider cannot create the Numspot API provider as there is an unknown configuration value for the Numspot Organisation ID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NUMSPOT_ORGANISATION_ID environment variable.",
		)
	}
}

func resolveProviderValues(config *NumspotProviderModel) providerValues {
	values := providerValues{
		numspotHost:    os.Getenv("NUMSPOT_HOST"),
		clientID:       os.Getenv("NUMSPOT_CLIENT_ID"),
		clientSecret:   os.Getenv("NUMSPOT_CLIENT_SECRET"),
		spaceID:        os.Getenv("NUMSPOT_SPACE_ID"),
		organisationID: os.Getenv("NUMSPOT_ORGANISATION_ID"),
		numspotHostOs:  os.Getenv("NUMSPOT_HOST_OS"),
	}

	if !config.NumspotHost.IsNull() {
		values.numspotHost = config.NumspotHost.ValueString()
	}

	if !config.ClientId.IsNull() {
		values.clientID = config.ClientId.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		values.clientSecret = config.ClientSecret.ValueString()
	}

	if !config.SpaceId.IsNull() {
		values.spaceID = config.SpaceId.ValueString()
	}

	if !config.OrganisationId.IsNull() {
		values.organisationID = config.OrganisationId.ValueString()
	}

	if !config.NumspotHostOs.IsNull() {
		values.numspotHostOs = config.NumspotHostOs.ValueString()
	}

	return values
}

func validateRequiredProviderValues(resp *provider.ConfigureResponse, values *providerValues) {
	if values.numspotHost == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("numspot_host"),
			"Missing Numspot API Host",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot host. "+
				"Set the host value in the configuration or use the NUMSPOT_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if values.clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Numspot Client ID",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot Client ID. "+
				"Set the client_id value in the configuration or use the NUMSPOT_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if values.clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Numspot Client Secret",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot Client Secret. "+
				"Set the client_secret value in the configuration or use the NUMSPOT_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if values.spaceID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("space_id"),
			"Missing Numspot Space ID",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot Space ID. "+
				"Set the space_id value in the configuration or use the NUMSPOT_SPACE_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if values.numspotHostOs == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("numspot_host_os"),
			"Missing Numspot API Host Object Storage",
			"The provider cannot create the Numspot API provider as there is a missing or empty value for the Numspot object storage host. "+
				"Set the host value in the configuration or use the NUMSPOT_HOST_OS environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
}

func buildClientOptions(p *numspotProvider, values *providerValues) []client.Option {
	options := []client.Option{
		client.WithHost(values.numspotHost),
		client.WithSpaceID(values.spaceID),
		client.WithOrganisationID(values.organisationID),
		client.WithClientID(values.clientID),
		client.WithClientSecret(values.clientSecret),
		client.WithHostOs(values.numspotHostOs),
	}

	if p.httpClient != nil {
		options = append(options, client.WithHTTPClient(p.httpClient))
	}

	return options
}

// Metadata sets the type name.
func (p *numspotProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "numspot"
	resp.Version = p.version
}

// DataSources returns the available data sources.
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
		vm.NewVMsDataSource,
		flexiblegpu.NewFlexibleGpusDataSource,
		bucket.NewBucketsDataSource,
		clientgateway.NewClientGatewaysDataSource,
		virtualgateway.NewVirtualGatewaysDataSource,
		vpnconnection.NewVpnConnectionsDataSource,
		computebridge.NewComputeBridgesDataSource,
		hybridbridge.NewHybridBridgesDataSource,
		kubernetesnodepool.NewKubernetesNodepoolsDataSource,
		servercertificate.NewServerCertificatesDataSource,
		postgrescluster.NewPostgresClustersDataSource,
		kubernetescluster.NewKubernetesClustersDataSource,
	}
}

// Resources returns the available resources.
func (p *numspotProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		flexiblegpu.NewFlexibleGpuResource,
		image.NewImageResource,
		internetgateway.NewInternetGatewayResource,
		loadbalancer.NewLoadBalancerResource,
		natgateway.NewNatGatewayResource,
		vpc.NewVPCResource,
		nic.NewNicResource,
		publicip.NewPublicIpResource,
		routetable.NewRouteTableResource,
		securitygroup.NewSecurityGroupResource,
		securitygrouprule.NewSecurityGroupRuleResource,
		snapshot.NewSnapshotResource,
		subnet.NewSubnetResource,
		volume.NewVolumeResource,
		vm.NewVMResource,
		keypair.NewKeypairResource,
		dhcpoptions.NewDhcpOptionsResource,
		servercertificate.NewServerCertificateResource,
		bucket.NewBucketResource,
		clientgateway.NewClientGatewayResource,
		virtualgateway.NewVirtualGatewayResource,
		vpnconnection.NewVpnConnectionResource,
		computebridge.NewComputeBridgeResource,
		hybridbridge.NewHybridBridgeResource,
		kubernetescluster.NewKubernetesClusterResource,
		kubernetesnodepool.NewKubernetesNodepoolResource,
		postgrescluster.NewPostgresClusterResource,
	}
}
