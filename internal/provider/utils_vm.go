package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func VmFromTfToHttp(tf resource_vm.VmModel) *api.VmSchema {
	return &api.VmSchema{}
}

func VmFromHttpToTf(http *api.VmSchema) resource_vm.VmModel {
	launchNumber := utils.FromIntPtrToTfInt64(http.LaunchNumber)
	vmsCount := utils.FromIntToTfInt64(-1)

	privateIps, _ := types.ListValueFrom(context.Background(), types.StringType, []string{*http.PrivateIp})
	// securityGroupIds, _ := types.ListValueFrom(context.Background(), types.StringType, []string{})
	productCodes, _ := types.ListValueFrom(context.Background(), types.StringType, http.ProductCodes)

	return resource_vm.VmModel{
		//
		Architecture:        types.StringPointerValue(http.Architecture),
		BlockDeviceMappings: types.ListNull(resource_vm.BlockDeviceMappingsValue{}.Type(context.Background())),
		BootOnCreation:      types.BoolValue(true),
		BsuOptimized:        types.BoolPointerValue(http.BsuOptimized),
		ClientToken:         types.StringPointerValue(http.ClientToken),
		CreationDate:        types.StringValue(""),
		//
		DeletionProtection:        types.BoolPointerValue(http.DeletionProtection),
		Hypervisor:                types.StringPointerValue(http.Hypervisor),
		Id:                        types.StringPointerValue(http.Id),
		ImageId:                   types.StringPointerValue(http.ImageId),
		InitiatedShutdownBehavior: types.StringPointerValue(http.InitiatedShutdownBehavior),
		IsSourceDestChecked:       types.BoolPointerValue(http.IsSourceDestChecked),
		KeypairName:               types.StringPointerValue(http.KeypairName),
		//
		LaunchNumber:         launchNumber,
		NestedVirtualization: types.BoolPointerValue(http.NestedVirtualization),
		NetId:                types.StringPointerValue(http.NetId),
		Nics:                 types.ListNull(resource_vm.NicsValue{}.Type(context.Background())),
		OsFamily:             types.StringPointerValue(http.OsFamily),
		Performance:          types.StringPointerValue(http.Performance),
		Placement:            resource_vm.PlacementValue{},
		PrivateDnsName:       types.StringPointerValue(http.PrivateDnsName),
		PrivateIp:            types.StringPointerValue(http.PrivateIp),
		//
		PrivateIps:                  privateIps,
		ProductCodes:                productCodes,
		PublicDnsName:               types.StringPointerValue(http.PublicDnsName),
		PublicIp:                    types.StringPointerValue(http.PublicIp),
		ReservationId:               types.StringPointerValue(http.ReservationId),
		RootDeviceName:              types.StringPointerValue(http.RootDeviceName),
		RootDeviceType:              types.StringPointerValue(http.RootDeviceType),
		SecurityGroupIds:            types.ListNull(types.StringType),
		SecurityGroups:              types.ListNull(types.StringType),
		State:                       types.StringPointerValue(http.State),
		StateReason:                 types.StringPointerValue(http.StateReason),
		SubnetId:                    types.StringPointerValue(http.SubnetId),
		Type:                        types.StringPointerValue(http.Type),
		UserData:                    types.StringPointerValue(http.UserData),
		VmInitiatedShutdownBehavior: types.StringPointerValue(http.InitiatedShutdownBehavior),
		VmType:                      types.StringPointerValue(http.Type),
		VmsCount:                    vmsCount,
	}
}

func VmFromTfToCreateRequest(tf resource_vm.VmModel) api.CreateVmsJSONRequestBody {
	return api.CreateVmsJSONRequestBody{
		BootOnCreation:              nil,
		BsuOptimized:                nil,
		ClientToken:                 nil,
		DeletionProtection:          nil,
		ImageId:                     tf.ImageId.ValueString(),
		KeypairName:                 tf.KeypairName.ValueStringPointer(),
		NestedVirtualization:        nil,
		Nics:                        nil,
		Performance:                 nil,
		Placement:                   nil,
		PrivateIps:                  nil,
		SecurityGroupIds:            nil,
		SecurityGroups:              nil,
		SubnetId:                    tf.SubnetId.ValueStringPointer(),
		UserData:                    nil,
		VmInitiatedShutdownBehavior: nil,
		VmType:                      tf.VmType.ValueStringPointer(),
	}
}
