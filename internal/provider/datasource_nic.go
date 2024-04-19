package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_nic"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type NicsDataSourceModel struct {
	Nics                           []datasource_nic.NicModel `tfsdk:"nics"`
	Descriptions                   types.List                `tfsdk:"descriptions"`
	IsSourceDestChecked            types.Bool                `tfsdk:"is_source_dest_checked"`
	LinkNicDeleteOnVMDeletion      types.Bool                `tfsdk:"link_nic_delete_on_vm_deletion"`
	LinkNicDeviceNumbers           types.List                `tfsdk:"link_nic_device_numbers"`
	LinkNicIds                     types.List                `tfsdk:"link_nic_link_nic_ids"`
	LinkNicStates                  types.List                `tfsdk:"link_nic_states"`
	LinkNicVMIds                   types.List                `tfsdk:"link_nic_vm_ids"`
	LinkPublicIpLinkPublicIpIds    types.List                `tfsdk:"link_public_ip_ids"`
	LinkPublicIpPublicIpIds        types.List                `tfsdk:"link_public_ip_public_ip_ids"`
	LinkPublicIpPublicIps          types.List                `tfsdk:"link_public_ip_public_ips"`
	MacAddresses                   types.List                `tfsdk:"mac_addresses"`
	PrivateDnsNames                types.List                `tfsdk:"private_dns_names"`
	PrivateIpIsPrimary             types.Bool                `tfsdk:"private_ips_is_primary"`
	PrivateIpLinkPublicIpPublicIps types.List                `tfsdk:"private_ips_link_public_ip_public_ips"`
	PrivateIpPrivateIps            types.List                `tfsdk:"private_ips_private_ips"`
	SecurityGroupIds               types.List                `tfsdk:"security_group_ids"`
	SecurityGroupNames             types.List                `tfsdk:"security_group_names"`
	States                         types.List                `tfsdk:"states"`
	SubnetIds                      types.List                `tfsdk:"subnet_ids"`
	VpcIds                         types.List                `tfsdk:"vpc_ids"`
	IDs                            types.List                `tfsdk:"ids"`
	AvailabilityZoneNames          types.List                `tfsdk:"tags"`
	Tags                           types.List                `tfsdk:"availability_zone_names"`
	TagKeys                        types.List                `tfsdk:"tag_keys"`
	TagValues                      types.List                `tfsdk:"tag_values"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &nicsDataSource{}
)

func (d *nicsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewNicsDataSource() datasource.DataSource {
	return &nicsDataSource{}
}

// coffeesDataSource is the data source implementation.
type nicsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *nicsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nics"
}

// Schema defines the schema for the data source.
func (d *nicsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_nic.NicDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *nicsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan NicsDataSourceModel
	request.Config.Get(ctx, &plan)

	params := NicsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadNicsResponse, error) {
		return d.provider.ApiClient.ReadNicsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Nic list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := NicsFromHttpToTfDatasource(ctx, &item)
		if diags.HasError() {
			response.Diagnostics.AddError("Error while converting Nic HTTP object to Terraform object", diags.Errors()[0].Detail())
			return
		}
		state.Nics = append(state.Nics, *tf)
	}

	state.Descriptions = plan.Descriptions
	state.IsSourceDestChecked = plan.IsSourceDestChecked
	state.LinkNicDeleteOnVMDeletion = plan.LinkNicDeleteOnVMDeletion
	state.LinkNicDeviceNumbers = plan.LinkNicDeviceNumbers
	state.LinkNicIds = plan.LinkNicIds
	state.LinkNicStates = plan.LinkNicStates
	state.LinkNicVMIds = plan.LinkNicVMIds
	state.LinkPublicIpLinkPublicIpIds = plan.LinkPublicIpLinkPublicIpIds
	state.LinkPublicIpPublicIpIds = plan.LinkPublicIpPublicIpIds
	state.LinkPublicIpPublicIps = plan.LinkPublicIpPublicIps
	state.MacAddresses = plan.MacAddresses
	state.PrivateDnsNames = plan.PrivateDnsNames
	state.PrivateIpIsPrimary = plan.PrivateIpIsPrimary
	state.PrivateIpLinkPublicIpPublicIps = plan.PrivateIpLinkPublicIpPublicIps
	state.PrivateIpPrivateIps = plan.PrivateIpPrivateIps
	state.SecurityGroupIds = plan.SecurityGroupIds
	state.SecurityGroupNames = plan.SecurityGroupNames
	state.States = plan.States
	state.SubnetIds = plan.SubnetIds
	state.VpcIds = plan.VpcIds
	state.IDs = plan.IDs
	state.AvailabilityZoneNames = plan.AvailabilityZoneNames
	state.Tags = plan.Tags
	state.TagKeys = plan.TagKeys
	state.TagValues = plan.TagValues

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
