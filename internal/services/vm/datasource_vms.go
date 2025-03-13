package vm

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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vmsDataSource{}
)

func (d *vmsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVmsDataSource() datasource.DataSource {
	return &vmsDataSource{}
}

type vmsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *vmsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vms"
}

// Schema defines the schema for the data source.
func (d *vmsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = VmDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vmsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VmsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeVmParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numspotVm, err := core.ReadVMsWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read vms", err.Error())
		return
	}

	objectItems := serializeVms(ctx, numspotVm, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeVmParams(ctx context.Context, tf VmsDataSourceModel, diags *diag.Diagnostics) numspot.ReadVmsParams {
	return numspot.ReadVmsParams{
		TagKeys:                              utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:                            utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:                                 utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:                                  utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		Architectures:                        utils.TfStringListToStringPtrList(ctx, tf.Architectures, diags),
		BlockDeviceMappingDeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.BlockDeviceMappingDeleteOnVmDeletion),
		BlockDeviceMappingDeviceNames:        utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingDeviceNames, diags),
		BlockDeviceMappingStates:             utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingsDataSourcetates, diags),
		BlockDeviceMappingVolumeIds:          utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingVolumeIds, diags),
		ClientTokens:                         utils.TfStringListToStringPtrList(ctx, tf.ClientTokens, diags),
		ImageIds:                             utils.TfStringListToStringPtrList(ctx, tf.ImageIds, diags),
		IsSourceDestChecked:                  utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
		KeypairNames:                         utils.TfStringListToStringPtrList(ctx, tf.KeypairNames, diags),
		LaunchNumbers:                        utils.TFInt64ListToIntListPointer(ctx, tf.LaunchNumbers, diags),
		NicDescriptions:                      utils.TfStringListToStringPtrList(ctx, tf.NicDescriptions, diags),
		NicIsSourceDestChecked:               utils.FromTfBoolToBoolPtr(tf.NicIsSourceDestChecked),
		NicLinkNicDeleteOnVmDeletion:         utils.FromTfBoolToBoolPtr(tf.NicLinkNicDeleteOnVmDeletion),
		NicLinkNicDeviceNumbers:              utils.TFInt64ListToIntListPointer(ctx, tf.NicLinkNicDeviceNumbers, diags),
		NicLinkNicLinkNicIds:                 utils.TfStringListToStringPtrList(ctx, tf.NicLinkNicLinkNicIds, diags),
		NicLinkNicStates:                     utils.TfStringListToStringPtrList(ctx, tf.NicLinkNicStates, diags),
		NicLinkPublicIpPublicIps:             utils.TfStringListToStringPtrList(ctx, tf.NicLinkPublicIpPublicIpIds, diags),
		NicMacAddresses:                      utils.TfStringListToStringPtrList(ctx, tf.NicMacAddresses, diags),
		NicNicIds:                            utils.TfStringListToStringPtrList(ctx, tf.NicNicIds, diags),
		NicPrivateIpsLinkPublicIpIds:         utils.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsLinkPublicIpIds, diags),
		NicPrivateIpsPrimaryIp:               utils.FromTfBoolToBoolPtr(tf.NicPrivateIpsPrimaryIp),
		NicPrivateIpsPrivateIps:              utils.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsPrivateIps, diags),
		NicSecurityGroupIds:                  utils.TfStringListToStringPtrList(ctx, tf.NicSecurityGroupIds, diags),
		NicSecurityGroupNames:                utils.TfStringListToStringPtrList(ctx, tf.NicSecurityGroupNames, diags),
		NicStates:                            utils.TfStringListToStringPtrList(ctx, tf.NicStates, diags),
		NicSubnetIds:                         utils.TfStringListToStringPtrList(ctx, tf.NicSubnetIds, diags),
		Platforms:                            utils.TfStringListToStringPtrList(ctx, tf.Platforms, diags),
		PrivateIps:                           utils.TfStringListToStringPtrList(ctx, tf.PrivateIps, diags),
		ProductCodes:                         utils.TfStringListToStringPtrList(ctx, tf.ProductCodes, diags),
		PublicIps:                            utils.TfStringListToStringPtrList(ctx, tf.PublicIps, diags),
		ReservationIds:                       utils.TfStringListToStringPtrList(ctx, tf.ReservationIds, diags),
		RootDeviceNames:                      utils.TfStringListToStringPtrList(ctx, tf.RootDeviceNames, diags),
		RootDeviceTypes:                      utils.TfStringListToStringPtrList(ctx, tf.RootDeviceTypes, diags),
		SecurityGroupIds:                     utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:                   utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames, diags),
		StateReasonMessages:                  utils.TfStringListToStringPtrList(ctx, tf.StateReasonMessages, diags),
		SubnetIds:                            utils.TfStringListToStringPtrList(ctx, tf.SubnetIds, diags),
		Tenancies:                            utils.TfStringListToStringPtrList(ctx, tf.Tenancies, diags),
		VmStateNames:                         utils.TfStringListToStringPtrList(ctx, tf.VmStateNames, diags),
		Types:                                utils.TfStringListToStringPtrList(ctx, tf.Types, diags),
		VpcIds:                               utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		NicVpcIds:                            utils.TfStringListToStringPtrList(ctx, tf.NicVpcIds, diags),
		AvailabilityZoneNames:                utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames, diags),
	}
}

func serializeVms(ctx context.Context, vms *[]numspot.Vm, diags *diag.Diagnostics) []VmModelItemDataSource {
	return utils.FromHttpGenericListToTfList(ctx, vms, func(ctx context.Context, vm *numspot.Vm, diags *diag.Diagnostics) *VmModelItemDataSource {
		var (
			blockDeviceMappings = types.ListNull(BlockDeviceMappingsValue{}.Type(ctx))
			nics                = types.ListNull(NicsValue{}.Type(ctx))
			securityGroups      = types.ListNull(SecurityGroupsValue{}.Type(ctx))
			productCodes        types.List
			placement           PlacementValue
			tagsList            types.List
			launchNumberTf      types.Int64
			creationDateTf      types.String
		)

		if vm.BlockDeviceMappings != nil {
			blockDeviceMappings = utils.GenericListToTfListValue(
				ctx,
				fromBlockDeviceMappingsToBlockDeviceMappingsList,
				*vm.BlockDeviceMappings,
				diags,
			)
		}

		if vm.Nics != nil {
			nics = utils.GenericListToTfListValue(
				ctx,
				fromNicsToNicsList,
				*vm.Nics,
				diags,
			)
		}

		if vm.SecurityGroups != nil {
			securityGroups = utils.GenericListToTfListValue(
				ctx,
				fromSecurityGroupToTFSecurityGroupList,
				*vm.SecurityGroups,
				diags,
			)
		}

		if vm.Placement != nil {
			var diagnostics diag.Diagnostics
			placement, diagnostics = NewPlacementValue(
				PlacementValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"availability_zone_name": types.StringPointerValue(vm.Placement.AvailabilityZoneName),
					"tenancy":                types.StringPointerValue(vm.Placement.Tenancy),
				},
			)
			diags.Append(diagnostics...)
		}

		if vm.ProductCodes != nil {
			productCodes = utils.StringListToTfListValue(ctx, *vm.ProductCodes, diags)
		}

		if vm.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *vm.Tags, diags)
			if diags.HasError() {
				return nil
			}
		}

		if vm.LaunchNumber != nil {
			launchNumber := int64(*vm.LaunchNumber)
			launchNumberTf = types.Int64PointerValue(&launchNumber)
		}

		if vm.CreationDate != nil {
			creationDate := vm.CreationDate.String()
			creationDateTf = types.StringPointerValue(&creationDate)
		}

		return &VmModelItemDataSource{
			Id:                            types.StringPointerValue(vm.Id),
			State:                         types.StringPointerValue(vm.State),
			BsuOptimized:                  types.BoolPointerValue(vm.BsuOptimized),
			Performance:                   types.StringPointerValue(vm.Performance),
			Tags:                          tagsList,
			Architecture:                  types.StringPointerValue(vm.Architecture),
			BlockDeviceMappingsDataSource: blockDeviceMappings,
			ClientToken:                   types.StringPointerValue(vm.ClientToken),
			CreationDate:                  creationDateTf,
			DeletionProtection:            types.BoolPointerValue(vm.DeletionProtection),
			Hypervisor:                    types.StringPointerValue(vm.Hypervisor),
			ImageId:                       types.StringPointerValue(vm.ImageId),
			InitiatedShutdownBehavior:     types.StringPointerValue(vm.InitiatedShutdownBehavior),
			IsSourceDestChecked:           types.BoolPointerValue(vm.IsSourceDestChecked),
			KeypairName:                   types.StringPointerValue(vm.KeypairName),
			LaunchNumber:                  launchNumberTf,
			NestedVirtualization:          types.BoolPointerValue(vm.NestedVirtualization),
			Nics:                          nics,
			OsFamily:                      types.StringPointerValue(vm.OsFamily),
			Placement:                     placement,
			PrivateDnsName:                types.StringPointerValue(vm.PrivateDnsName),
			PrivateIp:                     types.StringPointerValue(vm.PrivateIp),
			ProductCodes:                  productCodes,
			PublicDnsName:                 types.StringPointerValue(vm.PublicDnsName),
			PublicIp:                      types.StringPointerValue(vm.PublicIp),
			ReservationId:                 types.StringPointerValue(vm.ReservationId),
			RootDeviceName:                types.StringPointerValue(vm.RootDeviceName),
			RootDeviceType:                types.StringPointerValue(vm.RootDeviceType),
			SecurityGroups:                securityGroups,
			StateReason:                   types.StringPointerValue(vm.StateReason),
			SubnetId:                      types.StringPointerValue(vm.SubnetId),
			Type:                          types.StringPointerValue(vm.Type),
			UserData:                      types.StringPointerValue(vm.UserData),
			VpcId:                         types.StringPointerValue(vm.VpcId),
		}
	}, diags)
}

func fromBlockDeviceMappingsToBlockDeviceMappingsList(ctx context.Context, http numspot.BlockDeviceMappingCreated, diags *diag.Diagnostics) BlockDeviceMappingsDataSourceValue {
	bsu := fromBsuToTFBsu(ctx, http.Bsu, diags)
	if diags.HasError() {
		return BlockDeviceMappingsDataSourceValue{}
	}
	bsuObjectValue, diagnostics := bsu.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return BlockDeviceMappingsDataSourceValue{}
	}

	value, diagnostics := NewBlockDeviceMappingsDataSourceValue(
		BlockDeviceMappingsDataSourceValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":         bsuObjectValue,
			"device_name": types.StringPointerValue(http.DeviceName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func fromBsuToTFBsu(ctx context.Context, http *numspot.BsuCreated, diags *diag.Diagnostics) BsuDataSourceValue {
	if http == nil {
		return BsuDataSourceValue{}
	}

	var linkDateTf types.String

	if http.LinkDate != nil {
		linkDate := http.LinkDate.String()
		linkDateTf = types.StringPointerValue(&linkDate)
	}

	value, diagnostics := NewBsuDataSourceValue(
		BsuDataSourceValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"link_date":             linkDateTf,
			"state":                 types.StringPointerValue(http.State),
			"volume_id":             types.StringPointerValue(http.VolumeId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func fromSecurityGroupToTFSecurityGroupList(ctx context.Context, http numspot.SecurityGroupLight, diags *diag.Diagnostics) SecurityGroupsValue {
	value, diagnostics := NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(http.SecurityGroupId),
			"security_group_name": types.StringPointerValue(http.SecurityGroupName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func fromNicsToNicsList(ctx context.Context, http numspot.NicLight, diags *diag.Diagnostics) NicsValue {
	linkNic := fromLinkNicToTFLinkNic(ctx, http.LinkNic, diags)
	linkNICObject, diagnostics := linkNic.ToObjectValue(ctx)
	diags.Append(diagnostics...)

	linkPublicIP := linkPublicIpForVmFromHTTPDatasource(ctx, http.LinkPublicIp, diags)
	linkPublicIPObject, diagnostics := linkPublicIP.ToObjectValue(ctx)
	diags.Append(diagnostics...)

	privateIps := utils.GenericListToTfListValue(ctx, privateIpsFromApi, utils.GetPtrValue(http.PrivateIps), diags)
	securityGroups := utils.GenericListToTfListValue(ctx, securityGroupsForVmFromHTTP, utils.GetPtrValue(http.SecurityGroups), diags)

	value, diagnostics := NewNicsValue(
		NicsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"description":            types.StringPointerValue(http.Description),
			"is_source_dest_checked": types.BoolPointerValue(http.IsSourceDestChecked),
			"link_nic":               linkNICObject,
			"link_public_ip":         linkPublicIPObject,
			"mac_address":            types.StringPointerValue(http.MacAddress),
			"nic_id":                 types.StringPointerValue(http.NicId),
			"private_dns_name":       types.StringPointerValue(http.PrivateDnsName),
			"private_ips":            privateIps,
			"security_groups":        securityGroups,
			"state":                  types.StringPointerValue(http.State),
			"subnet_id":              types.StringPointerValue(http.SubnetId),
			"vpc_id":                 types.StringPointerValue(http.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func fromLinkNicToTFLinkNic(ctx context.Context, http *numspot.LinkNicLight, diags *diag.Diagnostics) LinkNicValue {
	if http == nil {
		return LinkNicValue{}
	}
	var deviceNumberTf types.Int64

	if http.DeviceNumber != nil {
		deviceNumber := int64(*http.DeviceNumber)
		deviceNumberTf = types.Int64PointerValue(&deviceNumber)
	} else {
		deviceNumberTf = types.Int64Value(0)
	}

	value, diagnostics := NewLinkNicValue(
		LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_number":         deviceNumberTf,
			"link_nic_id":           types.StringPointerValue(http.LinkNicId),
			"state":                 types.StringPointerValue(http.State),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func linkPublicIpForVmFromHTTPDatasource(ctx context.Context, http *numspot.LinkPublicIpLightForVm, diags *diag.Diagnostics) LinkPublicIpValue {
	if http == nil {
		return LinkPublicIpValue{}
	}

	value, diagnostics := NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(http.PublicDnsName),
			"public_ip":       types.StringPointerValue(http.PublicIp),
		})
	diags.Append(diagnostics...)
	return value
}

func securityGroupsForVmFromHTTP(ctx context.Context, elt numspot.SecurityGroupLight, diags *diag.Diagnostics) SecurityGroupsValue {
	value, diagnostics := NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		})
	diags.Append(diagnostics...)
	return value
}
