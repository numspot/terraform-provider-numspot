package provider

/*
 DIRECT LINKS are not handled for now

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link_interface"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func DirectLinkInterfaceFromTfToHttp(tf *resource_direct_link_interface.DirectLinkInterfaceModel) *numspot.DirectLinkInterfaces {
	return &numspot.DirectLinkInterfaces{
		BgpAsn:                  utils.FromTfInt64ToIntPtr(tf.BgpAsn),
		BgpKey:                  tf.BgpKey.ValueStringPointer(),
		ClientPrivateIp:         tf.ClientPrivateIp.ValueStringPointer(),
		DirectLinkId:            tf.DirectLinkId.ValueStringPointer(),
		Id:                      tf.DirectLinkInterfaceId.ValueStringPointer(),
		DirectLinkInterfaceName: tf.DirectLinkInterfaceName.ValueStringPointer(),
		InterfaceType:           tf.InterfaceType.ValueStringPointer(),
		Location:                tf.Location.ValueStringPointer(),
		Mtu:                     utils.FromTfInt64ToIntPtr(tf.Mtu),
		NumspotPrivateIp:        tf.OutscalePrivateIp.ValueStringPointer(),
		State:                   tf.State.ValueStringPointer(),
		VirtualGatewayId:        tf.VirtualGatewayId.ValueStringPointer(),
		Vlan:                    utils.FromTfInt64ToIntPtr(tf.Vlan),
	}
}

func DirectLinkInterfaceFromHttpToTf(http *numspot.DirectLinkInterfaces) resource_direct_link_interface.DirectLinkInterfaceModel {
	return resource_direct_link_interface.DirectLinkInterfaceModel{
		BgpAsn:                  utils.FromIntPtrToTfInt64(http.BgpAsn),
		BgpKey:                  types.StringPointerValue(http.BgpKey),
		ClientPrivateIp:         types.StringPointerValue(http.ClientPrivateIp),
		DirectLinkId:            types.StringPointerValue(http.DirectLinkId),
		DirectLinkInterface:     resource_direct_link_interface.NewDirectLinkInterfaceValueUnknown(),
		DirectLinkInterfaceId:   types.StringPointerValue(http.Id),
		DirectLinkInterfaceName: types.StringPointerValue(http.DirectLinkInterfaceName),
		Id:                      types.StringPointerValue(http.DirectLinkId),
		InterfaceType:           types.StringPointerValue(http.InterfaceType),
		Location:                types.StringPointerValue(http.Location),
		OutscalePrivateIp:       types.StringPointerValue(http.NumspotPrivateIp),
		State:                   types.StringPointerValue(http.State),
		VirtualGatewayId:        types.StringPointerValue(http.VirtualGatewayId),
		Vlan:                    utils.FromIntPtrToTfInt64(http.Vlan),
	}
}

func DirectLinkInterfaceFromTfToCreateRequest(tf *resource_direct_link_interface.DirectLinkInterfaceModel) numspot.CreateDirectLinkInterfaceJSONRequestBody {
	return numspot.CreateDirectLinkInterfaceJSONRequestBody{
		DirectLinkId:        tf.DirectLinkId.ValueString(),
		DirectLinkInterface: numspot.DirectLinkInterface{},
	}
}
*/
