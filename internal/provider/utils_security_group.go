package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_security_group"
)

func SecurityGroupFromTfToHttp(tf resource_security_group.SecurityGroupModel) *api.SecurityGroupSchema {
	return &api.SecurityGroupSchema{
		Id:            tf.Id.ValueStringPointer(),
		AccountId:     tf.AccountId.ValueStringPointer(),
		Description:   tf.Description.ValueStringPointer(),
		Name:          tf.Name.ValueStringPointer(),
		NetId:         tf.NetId.ValueStringPointer(),
		InboundRules:  nil,
		OutboundRules: nil,
	}
}

func SecurityGroupFromHttpToTf(http *api.SecurityGroupSchema) resource_security_group.SecurityGroupModel {
	return resource_security_group.SecurityGroupModel{
		AccountId:         types.StringPointerValue(http.AccountId),
		Description:       types.StringPointerValue(http.Description),
		Id:                types.StringPointerValue(http.Id),
		InboundRules:      types.List{},
		Name:              types.StringPointerValue(http.Name),
		NetId:             types.StringPointerValue(http.NetId),
		OutboundRules:     types.List{},
		SecurityGroupName: types.StringPointerValue(http.Name),
	}
}

func SecurityGroupFromTfToCreateRequest(tf resource_security_group.SecurityGroupModel) api.CreateSecurityGroupJSONRequestBody {
	return api.CreateSecurityGroupJSONRequestBody{
		Description:       tf.Description.ValueString(),
		NetId:             tf.NetId.ValueStringPointer(),
		SecurityGroupName: tf.Name.ValueString(),
	}
}
