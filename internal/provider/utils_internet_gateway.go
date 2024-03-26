package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_gateway"
)

func InternetServiceFromTfToHttp(tf resource_internet_gateway.InternetGatewayModel) *iaas.InternetGateway {
	return &iaas.InternetGateway{
		Id:    tf.Id.ValueStringPointer(),
		VpcId: tf.VpcIp.ValueStringPointer(),
		State: tf.State.ValueStringPointer(),
	}
}

func InternetServiceFromHttpToTf(http *iaas.InternetGateway) resource_internet_gateway.InternetGatewayModel {
	return resource_internet_gateway.InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		VpcIp: types.StringPointerValue(http.VpcId),
		State: types.StringPointerValue(http.State),
	}
}
