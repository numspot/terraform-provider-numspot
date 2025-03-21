package vm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/vm/datasource_vm"
	"terraform-provider-numspot/internal/utils"
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
	resp.Schema = datasource_vm.VmDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vmsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_vm.VmModel
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
	state.Items = objectItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeVmParams(ctx context.Context, tf datasource_vm.VmModel, diags *diag.Diagnostics) api.ReadVmsParams {
	return api.ReadVmsParams{
		TagKeys:                              utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                            utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                                 utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:                                  utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Architectures:                        utils.ConvertTfListToArrayOfString(ctx, tf.Architectures, diags),
		BlockDeviceMappingDeleteOnVmDeletion: tf.BlockDeviceMappingDeleteOnVmDeletion.ValueBoolPointer(),
		BlockDeviceMappingDeviceNames:        utils.ConvertTfListToArrayOfString(ctx, tf.BlockDeviceMappingDeviceNames, diags),
		BlockDeviceMappingStates:             utils.ConvertTfListToArrayOfString(ctx, tf.BlockDeviceMappingStates, diags),
		BlockDeviceMappingVolumeIds:          utils.ConvertTfListToArrayOfString(ctx, tf.BlockDeviceMappingVolumeIds, diags),
		ClientTokens:                         utils.ConvertTfListToArrayOfString(ctx, tf.ClientTokens, diags),
		ImageIds:                             utils.ConvertTfListToArrayOfString(ctx, tf.ImageIds, diags),
		IsSourceDestChecked:                  tf.IsSourceDestChecked.ValueBoolPointer(),
		KeypairNames:                         utils.ConvertTfListToArrayOfString(ctx, tf.KeypairNames, diags),
		LaunchNumbers:                        utils.ConvertTfListToArrayOfInt(ctx, tf.LaunchNumbers, diags),
		NicDescriptions:                      utils.ConvertTfListToArrayOfString(ctx, tf.NicDescriptions, diags),
		NicIsSourceDestChecked:               tf.NicIsSourceDestChecked.ValueBoolPointer(),
		NicLinkNicDeleteOnVmDeletion:         tf.NicLinkNicDeleteOnVmDeletion.ValueBoolPointer(),
		NicLinkNicDeviceNumbers:              utils.ConvertTfListToArrayOfInt(ctx, tf.NicLinkNicDeviceNumbers, diags),
		NicLinkNicLinkNicIds:                 utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkNicLinkNicIds, diags),
		NicLinkNicStates:                     utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkNicStates, diags),
		NicLinkPublicIpPublicIps:             utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkPublicIpPublicIpIds, diags),
		NicMacAddresses:                      utils.ConvertTfListToArrayOfString(ctx, tf.NicMacAddresses, diags),
		NicNicIds:                            utils.ConvertTfListToArrayOfString(ctx, tf.NicNicIds, diags),
		NicPrivateIpsLinkPublicIpIds:         utils.ConvertTfListToArrayOfString(ctx, tf.NicPrivateIpsLinkPublicIpIds, diags),
		NicPrivateIpsPrimaryIp:               tf.NicPrivateIpsPrimaryIp.ValueBoolPointer(),
		NicPrivateIpsPrivateIps:              utils.ConvertTfListToArrayOfString(ctx, tf.NicPrivateIpsPrivateIps, diags),
		NicSecurityGroupIds:                  utils.ConvertTfListToArrayOfString(ctx, tf.NicSecurityGroupIds, diags),
		NicSecurityGroupNames:                utils.ConvertTfListToArrayOfString(ctx, tf.NicSecurityGroupNames, diags),
		NicStates:                            utils.ConvertTfListToArrayOfString(ctx, tf.NicStates, diags),
		NicSubnetIds:                         utils.ConvertTfListToArrayOfString(ctx, tf.NicSubnetIds, diags),
		Platforms:                            utils.ConvertTfListToArrayOfString(ctx, tf.Platforms, diags),
		PrivateIps:                           utils.ConvertTfListToArrayOfString(ctx, tf.PrivateIps, diags),
		ProductCodes:                         utils.ConvertTfListToArrayOfString(ctx, tf.ProductCodes, diags),
		PublicIps:                            utils.ConvertTfListToArrayOfString(ctx, tf.PublicIps, diags),
		ReservationIds:                       utils.ConvertTfListToArrayOfString(ctx, tf.ReservationIds, diags),
		RootDeviceNames:                      utils.ConvertTfListToArrayOfString(ctx, tf.RootDeviceNames, diags),
		RootDeviceTypes:                      utils.ConvertTfListToArrayOfString(ctx, tf.RootDeviceTypes, diags),
		SecurityGroupIds:                     utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:                   utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupNames, diags),
		StateReasonMessages:                  utils.ConvertTfListToArrayOfString(ctx, tf.StateReasonMessages, diags),
		SubnetIds:                            utils.ConvertTfListToArrayOfString(ctx, tf.SubnetIds, diags),
		Tenancies:                            utils.ConvertTfListToArrayOfString(ctx, tf.Tenancies, diags),
		VmStateNames:                         utils.ConvertTfListToArrayOfString(ctx, tf.VmStateNames, diags),
		Types:                                utils.ConvertTfListToArrayOfString(ctx, tf.Types, diags),
		VpcIds:                               utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		NicVpcIds:                            utils.ConvertTfListToArrayOfString(ctx, tf.NicVpcIds, diags),
		AvailabilityZoneNames:                utils.ConvertTfListToArrayOfString(ctx, tf.AvailabilityZoneNames, diags),
	}
}

func serializeVms(ctx context.Context, vms *[]api.Vm, diags *diag.Diagnostics) datasource_vm.VmModel {
	var vmsList types.List
	var serializeDiags diag.Diagnostics

	tagsList := types.List{}
	blockDeviceMappingsList := types.List{}
	nicsList := types.ListNull(datasource_vm.NicsValue{}.Type(ctx))
	securityGroupsList := types.List{}
	productCodesList := types.ListNull(types.String{}.Type(ctx))
	creationDateTf := types.String{}
	placement := basetypes.ObjectValue{}

	if len(*vms) != 0 {
		ll := len(*vms)
		itemsValue := make([]datasource_vm.ItemsValue, ll)
		for i := 0; ll > i; i++ {
			if (*vms)[i].Tags != nil {
				tagsList, serializeDiags = mappingVmTags(ctx, vms, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*vms)[i].BlockDeviceMappings != nil {
				blockDeviceMappingsList, serializeDiags = mappingBlockDeviceMappings(ctx, vms, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*vms)[i].Nics != nil {
				nicsList, serializeDiags = mappingNics(ctx, vms, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*vms)[i].SecurityGroups != nil {
				securityGroupsList, serializeDiags = mappingSecurityGroups(ctx, (*vms)[i].SecurityGroups, diags)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*vms)[i].CreationDate != nil {
				creationDate := (*vms)[i].CreationDate.String()
				creationDateTf = types.StringPointerValue(&creationDate)
			}

			if (*vms)[i].ProductCodes != nil {
				productCodesList, serializeDiags = types.ListValueFrom(ctx, types.StringType, (*vms)[i].ProductCodes)
				diags.Append(serializeDiags...)
			}

			if (*vms)[i].Placement != nil {
				placementValue, serializePlacementDiags := mappingPlacement(ctx, (*vms)[i].Placement, diags)
				if serializePlacementDiags.HasError() {
					diags.Append(serializePlacementDiags...)
				}

				placement, serializeDiags = placementValue.ToObjectValue(ctx)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			itemsValue[i], serializeDiags = datasource_vm.NewItemsValue(datasource_vm.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"architecture":                types.StringValue(utils.ConvertStringPtrToString((*vms)[i].Architecture)),
				"block_device_mappings":       blockDeviceMappingsList,
				"bsu_optimized":               types.BoolPointerValue((*vms)[i].BsuOptimized),
				"client_token":                types.StringValue(utils.ConvertStringPtrToString((*vms)[i].ClientToken)),
				"creation_date":               creationDateTf,
				"deletion_protection":         types.BoolPointerValue((*vms)[i].DeletionProtection),
				"hypervisor":                  types.StringValue(utils.ConvertStringPtrToString((*vms)[i].Hypervisor)),
				"id":                          types.StringValue(utils.ConvertStringPtrToString((*vms)[i].Id)),
				"image_id":                    types.StringValue(utils.ConvertStringPtrToString((*vms)[i].ImageId)),
				"initiated_shutdown_behavior": types.StringValue(utils.ConvertStringPtrToString((*vms)[i].InitiatedShutdownBehavior)),
				"is_source_dest_checked":      types.BoolPointerValue((*vms)[i].IsSourceDestChecked),
				"keypair_name":                types.StringValue(utils.ConvertStringPtrToString((*vms)[i].KeypairName)),
				"launch_number":               types.Int64Value(utils.ConvertIntPtrToInt64((*vms)[i].LaunchNumber)),
				"nested_virtualization":       types.BoolPointerValue((*vms)[i].NestedVirtualization),
				"nics":                        nicsList,
				"os_family":                   types.StringValue(utils.ConvertStringPtrToString((*vms)[i].OsFamily)),
				"performance":                 types.StringValue(utils.ConvertStringPtrToString((*vms)[i].Performance)),
				"placement":                   placement,
				"private_dns_name":            types.StringValue(utils.ConvertStringPtrToString((*vms)[i].PrivateDnsName)),
				"private_ip":                  types.StringValue(utils.ConvertStringPtrToString((*vms)[i].PrivateIp)),
				"product_codes":               productCodesList,
				"public_dns_name":             types.StringValue(utils.ConvertStringPtrToString((*vms)[i].PublicDnsName)),
				"public_ip":                   types.StringValue(utils.ConvertStringPtrToString((*vms)[i].PublicIp)),
				"reservation_id":              types.StringValue(utils.ConvertStringPtrToString((*vms)[i].ReservationId)),
				"root_device_name":            types.StringValue(utils.ConvertStringPtrToString((*vms)[i].RootDeviceName)),
				"root_device_type":            types.StringValue(utils.ConvertStringPtrToString((*vms)[i].RootDeviceType)),
				"security_groups":             securityGroupsList,
				"state":                       types.StringValue(utils.ConvertStringPtrToString((*vms)[i].State)),
				"state_reason":                types.StringValue(utils.ConvertStringPtrToString((*vms)[i].StateReason)),
				"subnet_id":                   types.StringValue(utils.ConvertStringPtrToString((*vms)[i].SubnetId)),
				"tags":                        tagsList,
				"type":                        types.StringValue(utils.ConvertStringPtrToString((*vms)[i].Type)),
				"user_data":                   types.StringValue(utils.ConvertStringPtrToString((*vms)[i].UserData)),
				"vpc_id":                      types.StringValue(utils.ConvertStringPtrToString((*vms)[i].VpcId)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		vmsList, serializeDiags = types.ListValueFrom(ctx, new(datasource_vm.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		vmsList = types.ListNull(new(datasource_vm.ItemsValue).Type(ctx))
	}

	return datasource_vm.VmModel{
		Items: vmsList,
	}
}

func mappingVmTags(ctx context.Context, vms *[]api.Vm, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	lt := len(*(*vms)[i].Tags)
	elementValue := make([]datasource_vm.TagsValue, lt)
	for y, tag := range *(*vms)[i].Tags {
		elementValue[y], *diags = datasource_vm.NewTagsValue(datasource_vm.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"key":   types.StringValue(tag.Key),
			"value": types.StringValue(tag.Value),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_vm.TagsValue).Type(ctx), elementValue)
}

func mappingNics(ctx context.Context, vms *[]api.Vm, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	var mappingDiags diag.Diagnostics
	var linkPublicIp basetypes.ObjectValue
	var nicLink basetypes.ObjectValue

	ln := len(*(*vms)[i].Nics)
	elementValue := make([]datasource_vm.NicsValue, ln)
	securityGroupsList := types.ListNull(new(datasource_vm.NicSecurityGroupsValue).Type(ctx))
	privateIpsList := types.ListNull(datasource_vm.PrivateIpsValue{}.Type(ctx))

	for y, nic := range *(*vms)[i].Nics {
		if nic.LinkNic != nil {
			nicLinkValue, mappingNicLinkDiags := mappingNicLink(ctx, nic.LinkNic, diags)
			if mappingNicLinkDiags.HasError() {
				diags.Append(mappingNicLinkDiags...)
			}

			nicLink, mappingDiags = nicLinkValue.ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		} else {
			nicLink, mappingDiags = datasource_vm.NewLinkNicValueNull().ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		}

		if nic.LinkPublicIp != nil {
			linkPublicIpValue, mappingLinkPublicIpDiags := mappingLinkPublicIp(ctx, nic.LinkPublicIp, diags)
			if mappingLinkPublicIpDiags.HasError() {
				diags.Append(mappingLinkPublicIpDiags...)
			}

			linkPublicIp, mappingDiags = linkPublicIpValue.ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		} else {
			linkPublicIp, mappingDiags = datasource_vm.NewNicLinkPublicIpValueNull().ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		}

		if nic.SecurityGroups != nil {
			securityGroupsList, mappingDiags = mappingNicSecurityGroups(ctx, nic.SecurityGroups, diags)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		}

		if nic.PrivateIps != nil {
			privateIpsList, mappingDiags = mappingPrivateIps(ctx, nic.PrivateIps, diags)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		}

		elementValue[y], *diags = datasource_vm.NewNicsValue(datasource_vm.NicsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"description":            types.StringPointerValue(nic.Description),
			"is_source_dest_checked": types.BoolPointerValue(nic.IsSourceDestChecked),
			"link_nic":               nicLink,
			"mac_address":            types.StringPointerValue(nic.MacAddress),
			"nic_id":                 types.StringPointerValue(nic.NicId),
			"nic_link_public_ip":     linkPublicIp,
			"nic_security_groups":    securityGroupsList,
			"private_dns_name":       types.StringPointerValue(nic.PrivateDnsName),
			"private_ips":            privateIpsList,
			"state":                  types.StringPointerValue(nic.State),
			"subnet_id":              types.StringPointerValue(nic.SubnetId),
			"vpc_id":                 types.StringPointerValue(nic.VpcId),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_vm.NicsValue).Type(ctx), elementValue)
}

func mappingNicLink(ctx context.Context, nicLight *api.LinkNicLight, diags *diag.Diagnostics) (datasource_vm.LinkNicValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_vm.NewLinkNicValue(datasource_vm.LinkNicValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(nicLight.DeleteOnVmDeletion),
		"device_number":         types.Int64Value(utils.ConvertIntPtrToInt64(nicLight.DeviceNumber)),
		"link_nic_id":           types.StringPointerValue(nicLight.LinkNicId),
		"state":                 types.StringPointerValue(nicLight.State),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingLinkPublicIp(ctx context.Context, publicIpLightForVm *api.LinkPublicIpLightForVm, diags *diag.Diagnostics) (datasource_vm.NicLinkPublicIpValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_vm.NewNicLinkPublicIpValue(datasource_vm.NicLinkPublicIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"public_dns_name": types.StringPointerValue(publicIpLightForVm.PublicDnsName),
		"public_ip":       types.StringPointerValue(publicIpLightForVm.PublicIp),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingSecurityGroups(ctx context.Context, securityGroups *[]api.SecurityGroupLight, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	ls := len(*securityGroups)
	elementValue := make([]datasource_vm.SecurityGroupsValue, ls)

	for y, securityGroup := range *securityGroups {
		elementValue[y], *diags = datasource_vm.NewSecurityGroupsValue(datasource_vm.SecurityGroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(securityGroup.SecurityGroupId),
			"security_group_name": types.StringPointerValue(securityGroup.SecurityGroupName),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_vm.SecurityGroupsValue).Type(ctx), elementValue)
}

func mappingNicSecurityGroups(ctx context.Context, securityGroups *[]api.SecurityGroupLight, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	ls := len(*securityGroups)
	elementValue := make([]datasource_vm.NicSecurityGroupsValue, ls)

	for y, securityGroup := range *securityGroups {
		elementValue[y], *diags = datasource_vm.NewNicSecurityGroupsValue(datasource_vm.NicSecurityGroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(securityGroup.SecurityGroupId),
			"security_group_name": types.StringPointerValue(securityGroup.SecurityGroupName),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_vm.NicSecurityGroupsValue).Type(ctx), elementValue)
}

func mappingPrivateIps(ctx context.Context, privateIpLight *[]api.PrivateIpLightForVm, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	var mappingDiags diag.Diagnostics
	var linkPublic basetypes.ObjectValue

	ls := len(*privateIpLight)
	elementValue := make([]datasource_vm.PrivateIpsValue, ls)

	for y, privateIp := range *privateIpLight {

		if privateIp.LinkPublicIp != nil {
			linkPublicValue, serializeLinkPublicDiags := mappingLinkPublicIp(ctx, privateIp.LinkPublicIp, diags)
			if serializeLinkPublicDiags.HasError() {
				diags.Append(serializeLinkPublicDiags...)
			}
			linkPublic, mappingDiags = linkPublicValue.ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		} else {
			linkPublic, mappingDiags = datasource_vm.NewNicLinkPublicIpValueNull().ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		}

		elementValue[y], *diags = datasource_vm.NewPrivateIpsValue(datasource_vm.PrivateIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"is_primary":                types.BoolPointerValue(privateIp.IsPrimary),
			"private_dns_name":          types.StringPointerValue(privateIp.PrivateDnsName),
			"private_ip":                types.StringPointerValue(privateIp.PrivateIp),
			"private_ip_link_public_ip": linkPublic,
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_vm.PrivateIpsValue).Type(ctx), elementValue)
}

func mappingBlockDeviceMappings(ctx context.Context, vms *[]api.Vm, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	var mappingDiags diag.Diagnostics

	lb := len(*(*vms)[i].BlockDeviceMappings)
	elementValue := make([]datasource_vm.BlockDeviceMappingsValue, lb)
	bsu := basetypes.ObjectValue{}

	for y, blockDeviceMapping := range *(*vms)[i].BlockDeviceMappings {
		if blockDeviceMapping.Bsu != nil {
			bsuValue, serializeBsuDiags := mappingBsu(ctx, blockDeviceMapping, diags)
			if serializeBsuDiags.HasError() {
				diags.Append(serializeBsuDiags...)
			}
			bsu, mappingDiags = bsuValue.ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		}

		elementValue[y], *diags = datasource_vm.NewBlockDeviceMappingsValue(datasource_vm.BlockDeviceMappingsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"bsu":         bsu,
			"device_name": types.StringPointerValue(blockDeviceMapping.DeviceName),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_vm.BlockDeviceMappingsValue).Type(ctx), elementValue)
}

func mappingBsu(ctx context.Context, blockDeviceMappingCreated api.BlockDeviceMappingCreated, diags *diag.Diagnostics) (datasource_vm.BsuValue, diag.Diagnostics) {
	linkDateTf := types.String{}

	if blockDeviceMappingCreated.Bsu.LinkDate != nil {
		linkDate := blockDeviceMappingCreated.Bsu.LinkDate.String()
		linkDateTf = types.StringPointerValue(&linkDate)
	}

	elementValue, mappingDiags := datasource_vm.NewBsuValue(datasource_vm.BsuValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(blockDeviceMappingCreated.Bsu.DeleteOnVmDeletion),
		"link_date":             linkDateTf,
		"state":                 types.StringPointerValue(blockDeviceMappingCreated.Bsu.State),
		"volume_id":             types.StringPointerValue(blockDeviceMappingCreated.Bsu.VolumeId),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingPlacement(ctx context.Context, placement *api.Placement, diags *diag.Diagnostics) (datasource_vm.PlacementValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_vm.NewPlacementValue(datasource_vm.PlacementValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringPointerValue(placement.AvailabilityZoneName),
		"tenancy":                types.StringPointerValue(placement.Tenancy),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}
