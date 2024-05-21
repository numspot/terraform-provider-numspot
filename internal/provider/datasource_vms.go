package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VmsDataSourceModel struct {
	Vms                                   []datasource_vm.VmModel `tfsdk:"vms"`
	Architectures                         types.List              `tfsdk:"architectures"`
	BlockDeviceMappingsDeleteOnVmDeletion types.Bool              `tfsdk:"block_device_mappings_delete_on_vm_deletion"`
	BlockDeviceMappingsDeviceNames        types.List              `tfsdk:"block_device_mappings_device_names"`
	BlockDeviceMappingsLinkDates          types.List              `tfsdk:"block_device_mappings_link_dates"`
	BlockDeviceMappingsStates             types.List              `tfsdk:"block_device_mappings_states"`
	BlockDeviceMappingsVolumeIds          types.List              `tfsdk:"block_device_mappings_volume_ids"`
	ClientTokens                          types.List              `tfsdk:"client_tokens"`
	CreationDates                         types.List              `tfsdk:"creation_dates"`
	ImageIds                              types.List              `tfsdk:"image_ids"`
	IsSourceDestChecked                   types.Bool              `tfsdk:"is_source_dest_checked"`
	KeypairNames                          types.List              `tfsdk:"keypair_names"`
	LaunchNumbers                         types.List              `tfsdk:"launch_numbers"`
	NicAccountIds                         types.List              `tfsdk:"nic_account_ids"`
	NicDescriptions                       types.List              `tfsdk:"nic_descriptions"`
	NicIsSourceDestChecked                types.Bool              `tfsdk:"nic_is_source_dest_checked"`
	NicLinkNicDeleteOnVmDeletion          types.Bool              `tfsdk:"nic_link_nic_delete_on_vm_deletion"`
	NicLinkNicDeviceNumbers               types.List              `tfsdk:"nic_link_nic_device_numbers"`
	NicLinkNicLinkNicIds                  types.List              `tfsdk:"nic_link_nic_link_nic_ids"`
	NicLinkNicStates                      types.List              `tfsdk:"nic_link_nic_states"`
	NicLinkPublicIpAccountIds             types.List              `tfsdk:"nic_link_public_ip_account_ids"`
	NicLinkPublicIpsPublicIps             types.List              `tfsdk:"nic_link_public_ips_public_ips"`
	NicMacAddresses                       types.List              `tfsdk:"nic_mac_addresses"`
	NicNicIds                             types.List              `tfsdk:"nic_nic_ids"`
	NicPrivateIpsLinkPublicIpAccountId    types.List              `tfsdk:"nic_private_ips_link_public_ip_account_id"`
	NicPrivateIpsLinkPublicIps            types.List              `tfsdk:"nic_private_ips_link_public_ips"`
	NicPrivateIpsIsPrimary                types.Bool              `tfsdk:"nic_private_ips_is_primary"`
	NicPrivateIpsPrivateIps               types.List              `tfsdk:"nic_private_ips_private_ips"`
	NicSecurityGroupIds                   types.List              `tfsdk:"nic_security_group_ids"`
	NicSecurityGroupNames                 types.List              `tfsdk:"nic_security_group_names"`
	NicStates                             types.List              `tfsdk:"nic_states"`
	NicSubnetIds                          types.List              `tfsdk:"nic_subnet_ids"`
	OsFamilies                            types.List              `tfsdk:"os_families"`
	PrivateIps                            types.List              `tfsdk:"private_ips"`
	ProductCodes                          types.List              `tfsdk:"product_codes"`
	PublicIps                             types.List              `tfsdk:"public_ips"`
	ReservationIds                        types.List              `tfsdk:"reservation_ids"`
	RootDeviceNames                       types.List              `tfsdk:"root_device_names"`
	RootDeviceTypes                       types.List              `tfsdk:"root_device_types"`
	SecurityGroupIds                      types.List              `tfsdk:"security_group_ids"`
	SecurityGroupNames                    types.List              `tfsdk:"security_group_names"`
	StateReasonMessages                   types.List              `tfsdk:"state_reason_messages"`
	SubnetIds                             types.List              `tfsdk:"subnet_ids"`
	TagKeys                               types.List              `tfsdk:"tag_keys"`
	TagValues                             types.List              `tfsdk:"tag_values"`
	Tags                                  types.List              `tfsdk:"tags"`
	Tenancies                             types.List              `tfsdk:"tenancies"`
	VmStateNames                          types.List              `tfsdk:"vm_state_names"`
	VmTypes                               types.List              `tfsdk:"vm_types"`
	VpcIds                                types.List              `tfsdk:"vpc_ids"`
	NicVpcIds                             types.List              `tfsdk:"nic_vpc_ids"`
	AvailabilityZoneNames                 types.List              `tfsdk:"availability_zone_names"`
	IDs                                   types.List              `tfsdk:"ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vmsDataSource{}
)

func (d *vmsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVmsDataSource() datasource.DataSource {
	return &vmsDataSource{}
}

type vmsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *vmsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vms"
}

// Schema defines the schema for the data source.
func (d *vmsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vm.VmDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vmsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VmsDataSourceModel
	request.Config.Get(ctx, &plan)

	params := VmsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadVmsResponse, error) {
		return d.provider.ApiClient.ReadVmsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty VM list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := VmsFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting VM HTTP object to Terraform object", diags.Errors()[0].Detail())
			return
		}
		state.Vms = append(state.Vms, utils.GetPtrValue(tf))
	}
	state.Architectures = plan.Architectures
	state.BlockDeviceMappingsDeleteOnVmDeletion = plan.BlockDeviceMappingsDeleteOnVmDeletion
	state.BlockDeviceMappingsDeviceNames = plan.BlockDeviceMappingsDeviceNames
	state.BlockDeviceMappingsLinkDates = plan.BlockDeviceMappingsLinkDates
	state.BlockDeviceMappingsStates = plan.BlockDeviceMappingsStates
	state.BlockDeviceMappingsVolumeIds = plan.BlockDeviceMappingsVolumeIds
	state.ClientTokens = plan.ClientTokens
	state.CreationDates = plan.CreationDates
	state.ImageIds = plan.ImageIds
	state.IsSourceDestChecked = plan.IsSourceDestChecked
	state.KeypairNames = plan.KeypairNames
	state.LaunchNumbers = plan.LaunchNumbers
	state.NicAccountIds = plan.NicAccountIds
	state.NicDescriptions = plan.NicDescriptions
	state.NicIsSourceDestChecked = plan.NicIsSourceDestChecked
	state.NicLinkNicDeleteOnVmDeletion = plan.NicLinkNicDeleteOnVmDeletion
	state.NicLinkNicDeviceNumbers = plan.NicLinkNicDeviceNumbers
	state.NicLinkNicLinkNicIds = plan.NicLinkNicLinkNicIds
	state.NicLinkNicStates = plan.NicLinkNicStates
	state.NicLinkPublicIpAccountIds = plan.NicLinkPublicIpAccountIds
	state.NicLinkPublicIpsPublicIps = plan.NicLinkPublicIpsPublicIps
	state.NicMacAddresses = plan.NicMacAddresses
	state.NicNicIds = plan.NicNicIds
	state.NicPrivateIpsLinkPublicIpAccountId = plan.NicPrivateIpsLinkPublicIpAccountId
	state.NicPrivateIpsLinkPublicIps = plan.NicPrivateIpsLinkPublicIps
	state.NicPrivateIpsIsPrimary = plan.NicPrivateIpsIsPrimary
	state.NicPrivateIpsPrivateIps = plan.NicPrivateIpsPrivateIps
	state.NicSecurityGroupIds = plan.NicSecurityGroupIds
	state.NicSecurityGroupNames = plan.NicSecurityGroupNames
	state.NicStates = plan.NicStates
	state.NicSubnetIds = plan.NicSubnetIds
	state.OsFamilies = plan.OsFamilies
	state.PrivateIps = plan.PrivateIps
	state.ProductCodes = plan.ProductCodes
	state.PublicIps = plan.PublicIps
	state.ReservationIds = plan.ReservationIds
	state.RootDeviceNames = plan.RootDeviceNames
	state.RootDeviceTypes = plan.RootDeviceTypes
	state.SecurityGroupIds = plan.SecurityGroupIds
	state.SecurityGroupNames = plan.SecurityGroupNames
	state.StateReasonMessages = plan.StateReasonMessages
	state.SubnetIds = plan.SubnetIds
	state.TagKeys = plan.TagKeys
	state.TagValues = plan.TagValues
	state.Tags = plan.Tags
	state.Tenancies = plan.Tenancies
	state.VmStateNames = plan.VmStateNames
	state.VmTypes = plan.VmTypes
	state.VpcIds = plan.VpcIds
	state.NicVpcIds = plan.NicVpcIds
	state.AvailabilityZoneNames = plan.AvailabilityZoneNames
	state.IDs = plan.IDs

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
