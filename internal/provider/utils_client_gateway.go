package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_client_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func ClientGatewayFromTfToHttp(tf resource_client_gateway.ClientGatewayModel) *api.ClientGatewaySchema {
	return &api.ClientGatewaySchema{
		BgpAsn:         utils.FromTfInt64ToIntPtr(tf.BgpAsn),
		ConnectionType: tf.ConnectionType.ValueStringPointer(),
		Id:             tf.Id.ValueStringPointer(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		State:          tf.State.ValueStringPointer(),
	}
}

func ClientGatewayFromHttpToTf(http *api.ClientGatewaySchema) resource_client_gateway.ClientGatewayModel {
	tmp := resource_client_gateway.ClientGatewayModel{
		BgpAsn:         utils.FromIntPtrToTfInt64(http.BgpAsn),
		ConnectionType: types.StringPointerValue(http.ConnectionType),
		Id:             types.StringPointerValue(http.Id),
		PublicIp:       types.StringPointerValue(http.PublicIp),
		State:          types.StringPointerValue(http.State),
	}
	return tmp
}

func ClientGatewayFromTfToCreateRequest(tf resource_client_gateway.ClientGatewayModel) api.CreateClientGatewayJSONRequestBody {
	return api.CreateClientGatewayJSONRequestBody{
		BgpAsn:         utils.FromTfInt64ToInt(tf.BgpAsn),
		ConnectionType: tf.ConnectionType.ValueString(),
		PublicIp:       tf.PublicIp.ValueString(),
	}
}
