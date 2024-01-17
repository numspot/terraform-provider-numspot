package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_security_group"
)

func SecurityGroupFromTfToHttp(tf resource_security_group.SecurityGroupModel) *api.SecurityGroupSchema {
	return &api.SecurityGroupSchema{}
}

func SecurityGroupFromHttpToTf(http *api.SecurityGroupSchema) resource_security_group.SecurityGroupModel {
	return resource_security_group.SecurityGroupModel{}
}

func SecurityGroupFromTfToCreateRequest(tf resource_security_group.SecurityGroupModel) api.CreateSecurityGroupJSONRequestBody {
	return api.CreateSecurityGroupJSONRequestBody{}
}
