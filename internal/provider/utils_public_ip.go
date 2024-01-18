package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_public_ip"
)

func PublicIpFromTfToHttp(tf resource_public_ip.PublicIpModel) *api.PublicIpSchema {
	return &api.PublicIpSchema{
		Id:           nil,
		NicAccountId: nil,
		NicId:        nil,
		PrivateIp:    nil,
		PublicIp:     nil,
		Tags:         nil,
		VmId:         nil,
	}
}

func PublicIpFromHttpToTf(http *api.PublicIpSchema) resource_public_ip.PublicIpModel {
	return resource_public_ip.PublicIpModel{
		Id:           types.StringPointerValue(http.Id),
		NicAccountId: types.StringPointerValue(http.NicAccountId),
		NicId:        types.StringPointerValue(http.NicId),
		PrivateIp:    types.StringPointerValue(http.PrivateIp),
		PublicIp:     types.StringPointerValue(http.PublicIp),
		VmId:         types.StringPointerValue(http.VmId),
	}
}

func PublicIpFromTfToCreateRequest(_ resource_public_ip.PublicIpModel) api.CreatePublicIpJSONRequestBody {
	return api.CreatePublicIpJSONRequestBody{}
}
