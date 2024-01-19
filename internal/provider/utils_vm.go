package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
)

func VmFromTfToHttp(tf resource_vm.VmModel) *api.VmSchema {
	return &api.VmSchema{}
}

func VmFromHttpToTf(http *api.VmSchema) resource_vm.VmModel {
	return resource_vm.VmModel{}
}

func VmFromTfToCreateRequest(tf resource_vm.VmModel) api.CreateVmsJSONRequestBody {
	return api.CreateVmsJSONRequestBody{}
}
