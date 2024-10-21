package nic

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type NicsDataSourceModel struct {
	Items                           []NicModelDatasource `tfsdk:"items"`
	AvailabilityZoneNames           types.List           `tfsdk:"availability_zone_names"`
	Descriptions                    types.List           `tfsdk:"descriptions"`
	Ids                             types.List           `tfsdk:"ids"`
	IsSourceDestCheck               types.Bool           `tfsdk:"is_source_dest_check"`
	LinkNicDeleteOnVmDeletion       types.Bool           `tfsdk:"link_nic_delete_on_vm_deletion"`
	LinkNicDeviceNumbers            types.List           `tfsdk:"link_nic_device_numbers"`
	LinkNicLinkNicIds               types.List           `tfsdk:"link_nic_link_nic_ids"`
	LinkNicStates                   types.List           `tfsdk:"link_nic_states"`
	LinkNicVmIds                    types.List           `tfsdk:"link_nic_vm_ids"`
	LinkPublicIpLinkPublicIpIds     types.List           `tfsdk:"link_public_ip_link_public_ip_ids"`
	LinkPublicIpPublicIpIds         types.List           `tfsdk:"link_public_ip_public_ip_ids"`
	LinkPublicIpPublicIps           types.List           `tfsdk:"link_public_ip_public_ips"`
	MacAddresses                    types.List           `tfsdk:"mac_addresses"`
	PrivateDnsNames                 types.List           `tfsdk:"private_dns_names"`
	PrivateIpsLinkPublicIpPublicIps types.List           `tfsdk:"private_ips_link_public_ip_public_ips"`
	PrivateIpsPrimaryIp             types.Bool           `tfsdk:"private_ips_primary_ip"`
	PrivateIpsPrivateIps            types.List           `tfsdk:"private_ips_private_ips"`
	SecurityGroupIds                types.List           `tfsdk:"security_group_ids"`
	SecurityGroupNames              types.List           `tfsdk:"security_group_names"`
	States                          types.List           `tfsdk:"states"`
	SubnetIds                       types.List           `tfsdk:"subnet_ids"`
	TagKeys                         types.List           `tfsdk:"tag_keys"`
	TagValues                       types.List           `tfsdk:"tag_values"`
	Tags                            types.List           `tfsdk:"tags"`
	VpcIds                          types.List           `tfsdk:"vpc_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &nicsDataSource{}
)

func (d *nicsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

func NewNicsDataSource() datasource.DataSource {
	return &nicsDataSource{}
}

type nicsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *nicsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nics"
}

// Schema defines the schema for the data source.
func (d *nicsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = NicDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *nicsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var plan, state NicsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeReadParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	nics, err := core.ReadNicsWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read nic with params", err.Error())
		return
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, nics, serializeNicDatasource, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}
	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeReadParams(ctx context.Context, tf NicsDataSourceModel, diags *diag.Diagnostics) numspot.ReadNicsParams {
	return numspot.ReadNicsParams{
		Descriptions:                    utils.TfStringListToStringPtrList(ctx, tf.Descriptions, diags),
		IsSourceDestCheck:               utils.FromTfBoolToBoolPtr(tf.IsSourceDestCheck),
		LinkNicDeleteOnVmDeletion:       utils.FromTfBoolToBoolPtr(tf.LinkNicDeleteOnVmDeletion),
		LinkNicDeviceNumbers:            utils.TFInt64ListToIntListPointer(ctx, tf.LinkNicDeviceNumbers, diags),
		LinkNicLinkNicIds:               utils.TfStringListToStringPtrList(ctx, tf.LinkNicLinkNicIds, diags),
		LinkNicStates:                   utils.TfStringListToStringPtrList(ctx, tf.LinkNicStates, diags),
		LinkNicVmIds:                    utils.TfStringListToStringPtrList(ctx, tf.LinkNicVmIds, diags),
		MacAddresses:                    utils.TfStringListToStringPtrList(ctx, tf.MacAddresses, diags),
		LinkPublicIpLinkPublicIpIds:     utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpLinkPublicIpIds, diags),
		LinkPublicIpPublicIpIds:         utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpPublicIpIds, diags),
		LinkPublicIpPublicIps:           utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpPublicIps, diags),
		PrivateDnsNames:                 utils.TfStringListToStringPtrList(ctx, tf.PrivateDnsNames, diags),
		PrivateIpsPrimaryIp:             utils.FromTfBoolToBoolPtr(tf.PrivateIpsPrimaryIp),
		PrivateIpsLinkPublicIpPublicIps: utils.TfStringListToStringPtrList(ctx, tf.PrivateIpsLinkPublicIpPublicIps, diags),
		PrivateIpsPrivateIps:            utils.TfStringListToStringPtrList(ctx, tf.PrivateIpsPrivateIps, diags),
		SecurityGroupIds:                utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:              utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames, diags),
		States:                          utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		SubnetIds:                       utils.TfStringListToStringPtrList(ctx, tf.SubnetIds, diags),
		VpcIds:                          utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		Ids:                             utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		AvailabilityZoneNames:           utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames, diags),
		Tags:                            utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		TagKeys:                         utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:                       utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
	}
}

func serializeNicDatasource(ctx context.Context, http *numspot.Nic, diags *diag.Diagnostics) *NicModelDatasource {
	if http == nil {
		return nil
	}

	var (
		tagsList     types.List
		linkNic      LinkNicValue
		linkPublicIp LinkPublicIpValue
	)

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	if http.LinkNic != nil {
		var diagnostics diag.Diagnostics
		linkNic, diagnostics = NewLinkNicValue(LinkNicValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"delete_on_vm_deletion": types.BoolPointerValue(http.LinkNic.DeleteOnVmDeletion),
				"device_number":         utils.FromIntPtrToTfInt64(http.LinkNic.DeviceNumber),
				"id":                    types.StringPointerValue(http.LinkNic.Id),
				"state":                 types.StringPointerValue(http.LinkNic.State),
				"vm_id":                 types.StringPointerValue(http.LinkNic.VmId),
			})
		diags.Append(diagnostics...)
		if diags.HasError() {
			return nil
		}
	}

	if http.LinkPublicIp != nil {
		var diagnostics diag.Diagnostics
		linkPublicIp, diagnostics = NewLinkPublicIpValue(LinkPublicIpValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"id":              types.StringPointerValue(http.LinkPublicIp.Id),
				"public_dns_name": types.StringPointerValue(http.LinkPublicIp.PublicDnsName),
				"public_ip":       types.StringPointerValue(http.LinkPublicIp.PublicIp),
				"public_ip_id":    types.StringPointerValue(http.LinkPublicIp.PublicIpId),
			})
		diags.Append(diagnostics...)
		if diags.HasError() {
			return nil
		}
	}

	privateIps := utils.GenericListToTfListValue(ctx, serializeNumspotPrivateIps, utils.GetPtrValue(http.PrivateIps), diags)
	if diags.HasError() {
		return nil
	}

	securityGroups := utils.GenericListToTfListValue(ctx, serializeNumspotSecurityGroups, utils.GetPtrValue(http.SecurityGroups), diags)
	if diags.HasError() {
		return nil
	}

	return &NicModelDatasource{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Description:          types.StringPointerValue(http.Description),
		Id:                   types.StringPointerValue(http.Id),
		IsSourceDestChecked:  types.BoolPointerValue(http.IsSourceDestChecked),
		LinkNic:              linkNic,
		LinkPublicIp:         linkPublicIp,
		MacAddress:           types.StringPointerValue(http.MacAddress),
		PrivateDnsName:       types.StringPointerValue(http.PrivateDnsName),
		PrivateIps:           privateIps,
		SecurityGroups:       securityGroups,
		State:                types.StringPointerValue(http.State),
		SubnetId:             types.StringPointerValue(http.SubnetId),
		VpcId:                types.StringPointerValue(http.VpcId),
		Tags:                 tagsList,
	}
}
