package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

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

func invokeLinkPublicIP(ctx context.Context, client *api.ClientWithResponses, data resource_public_ip.PublicIpModel) error {
	var payload = api.LinkPublicIpJSONRequestBody{}
	if !data.VmId.IsNull() {
		payload = api.LinkPublicIpJSONRequestBody{VmId: data.VmId.ValueStringPointer()}
	} else {
		payload = api.LinkPublicIpJSONRequestBody{
			NicId:     data.NicId.ValueStringPointer(),
			PrivateIp: data.PrivateIp.ValueStringPointer(),
		}
	}
	if !data.PublicIp.IsNull() {
		payload.PublicIp = data.PublicIp.ValueStringPointer()
	}
	res, err := client.LinkPublicIpWithResponse(ctx, data.Id.ValueString(), payload)
	if err != nil {
		return err
	}
	expectedStatusCode := 200
	if res.StatusCode() != expectedStatusCode {
		return utils.HandleError(res.Body)
	}

	return nil
}
